package spinlock_test

import (
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/atomic"
	"darvaza.org/x/sync/internal/synctesting"
	"darvaza.org/x/sync/mutex"
	"darvaza.org/x/sync/spinlock"
)

// spinlockTestTimeout caps each foreground synchronisation step. Generous
// enough to absorb scheduler jitter on loaded CI workers, tight enough
// that a hung waiter fails the run instead of stalling.
//
// spinlockOpenGuard bounds negative-case assertions ("still blocked"):
// long enough to catch a spurious early acquisition, short enough to keep
// the test responsive.
const (
	spinlockTestTimeout = time.Second

	spinlockOpenGuard = 20 * time.Millisecond
)

// spinlockPanicTestCase verifies operations that panic on misuse or on a
// nil receiver. setup arranges the receiver state; op selects the method
// exercised; wantPanic pins the panic — an error matches the chain via
// errors.Is, a string by substring.
type spinlockPanicTestCase struct {
	wantPanic any

	setup func() *spinlock.SpinLock
	op    func(*spinlock.SpinLock)

	name string
}

func newSpinlockPanicTestCase(name string, setup func() *spinlock.SpinLock,
	op func(*spinlock.SpinLock), wantPanic any) spinlockPanicTestCase {
	return spinlockPanicTestCase{
		name:      name,
		setup:     setup,
		op:        op,
		wantPanic: wantPanic,
	}
}

func (tc spinlockPanicTestCase) Name() string { return tc.name }

func (tc spinlockPanicTestCase) Test(t *testing.T) {
	t.Helper()
	sl := tc.setup()
	core.AssertPanic(t, func() { tc.op(sl) }, tc.wantPanic, "panic")
}

var _ core.TestCase = spinlockPanicTestCase{}

func newSpinLock() *spinlock.SpinLock { return new(spinlock.SpinLock) }
func nilSpinLock() *spinlock.SpinLock { return nil }

func opLock(sl *spinlock.SpinLock)    { sl.Lock() }
func opTryLock(sl *spinlock.SpinLock) { sl.TryLock() }
func opUnlock(sl *spinlock.SpinLock)  { sl.Unlock() }

func spinlockPanicTestCases() []spinlockPanicTestCase {
	return []spinlockPanicTestCase{
		newSpinlockPanicTestCase("nil receiver Lock", nilSpinLock, opLock,
			core.ErrNilReceiver),
		newSpinlockPanicTestCase("nil receiver TryLock", nilSpinLock, opTryLock,
			core.ErrNilReceiver),
		newSpinlockPanicTestCase("nil receiver Unlock", nilSpinLock, opUnlock,
			core.ErrNilReceiver),
		newSpinlockPanicTestCase("unlock of unlocked", newSpinLock, opUnlock,
			"unlock of unlocked spinlock"),
	}
}

// TestSpinLock_Panics verifies every nil-receiver path and the
// unlock-of-unlocked misuse panic, pinning the wrapped error identity.
func TestSpinLock_Panics(t *testing.T) {
	core.RunTestCases(t, spinlockPanicTestCases())
}

// TestSpinLock_Basic verifies the zero value starts unlocked and a
// Lock/Unlock cycle returns the spinlock to an acquirable state. State is
// observed through TryLock rather than the internal representation.
func TestSpinLock_Basic(t *testing.T) {
	var sl spinlock.SpinLock

	core.AssertMustTrue(t, sl.TryLock(), "TryLock on zero value")
	sl.Unlock()

	sl.Lock()
	core.AssertFalse(t, sl.TryLock(), "TryLock while Lock held")
	sl.Unlock()

	core.AssertMustTrue(t, sl.TryLock(), "TryLock after Unlock")
	sl.Unlock()
}

// TestSpinLock_TryLock verifies a TryLock acquisition excludes a second
// TryLock until released.
func TestSpinLock_TryLock(t *testing.T) {
	var sl spinlock.SpinLock

	core.AssertMustTrue(t, sl.TryLock(), "TryLock when free")
	core.AssertFalse(t, sl.TryLock(), "TryLock while TryLock held")

	sl.Unlock()
	core.AssertMustTrue(t, sl.TryLock(), "TryLock after Unlock")
	sl.Unlock()
}

// TestSpinLock_LockBlocks pins the spin-wait contract deterministically:
// Lock against a held spinlock parks the caller until Unlock, then
// acquires.
func TestSpinLock_LockBlocks(t *testing.T) {
	var sl spinlock.SpinLock

	core.AssertMustTrue(t, sl.TryLock(), "initial TryLock")

	done := make(chan struct{})
	go func() {
		sl.Lock()
		close(done)
	}()

	synctesting.AssertMustOpen(t, done, spinlockOpenGuard,
		"Lock blocks while the spinlock is held")

	sl.Unlock()

	synctesting.AssertMustClosed(t, done, spinlockTestTimeout,
		"Lock acquires once the spinlock is released")

	// release the goroutine's acquisition
	sl.Unlock()
}

// TestSpinLock_Concurrent verifies mutual exclusion under contention: a
// counter incremented only inside the critical section ends at exactly
// goroutines × iterations.
func TestSpinLock_Concurrent(t *testing.T) {
	var sl spinlock.SpinLock
	var counter int

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			for range iterations {
				sl.Lock()
				counter++
				sl.Unlock()
			}
		}()
	}

	wg.Wait()

	core.AssertEqual(t, goroutines*iterations, counter, "counter")
}

// tryLockStats aggregates the shared state of the TryLock contention
// test: attempted and successful acquisitions, the live count of worker
// goroutines, and the highest concurrency level observed.
type tryLockStats struct {
	sl spinlock.SpinLock

	attempts atomic.Int32
	counter  int

	current atomic.Int32
	peak    atomic.Int32
}

// run attempts iterations TryLock acquisitions, incrementing the
// lock-protected counter on each success and tracking peak worker
// concurrency via atomic.UpdateMax.
func (s *tryLockStats) run(iterations int) {
	atomic.UpdateMax(&s.peak, s.current.Add(1))
	defer s.current.Add(-1)

	for range iterations {
		s.attempts.Add(1)
		if s.sl.TryLock() {
			s.counter++
			s.sl.Unlock()
		}
	}
}

// TestSpinLock_TryLockConcurrent verifies TryLock never blocks and never
// double-admits under contention. When the workers genuinely overlapped,
// some attempts must fail; when the scheduler serialised them, every
// attempt must succeed.
func TestSpinLock_TryLockConcurrent(t *testing.T) {
	var stats tryLockStats

	const goroutines = 100
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			stats.run(iterations)
		}()
	}

	wg.Wait()

	attempts, counter := stats.attempts.Load(), int32(stats.counter)
	t.Logf("peak workers: %d, counter: %d, attempts: %d",
		stats.peak.Load(), counter, attempts)

	if stats.peak.Load() <= 1 {
		core.AssertEqual(t, attempts, counter,
			"workers never overlapped: every attempt succeeds")
		return
	}
	core.AssertTrue(t, counter > 0, "some attempts succeed")
	core.AssertTrue(t, counter < attempts,
		"under contention some attempts fail")
}

// TestSpinLock_LockDefer verifies the Lock + deferred Unlock idiom leaves
// the spinlock acquirable.
func TestSpinLock_LockDefer(t *testing.T) {
	var sl spinlock.SpinLock
	var executed bool

	func() {
		sl.Lock()
		defer sl.Unlock()

		executed = true
	}()

	core.AssertTrue(t, executed, "critical section ran")
	core.AssertMustTrue(t, sl.TryLock(), "acquirable after deferred Unlock")
	sl.Unlock()
}

// TestSpinLock_Locker exercises SpinLock through the sync.Locker view;
// the compile-time interface assertions live next to the type.
func TestSpinLock_Locker(t *testing.T) {
	var sl spinlock.SpinLock
	var locker sync.Locker = &sl

	locker.Lock()
	core.AssertFalse(t, sl.TryLock(), "held via sync.Locker")
	locker.Unlock()

	core.AssertMustTrue(t, sl.TryLock(), "released via sync.Locker")
	sl.Unlock()
}

// ----- Shared benchmark functions -----

// runBenchmarkBasicLock benchmarks basic lock/unlock operations
func runBenchmarkBasicLock(b *testing.B, mu mutex.Mutex) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			n = 2*n + 1
			mu.Unlock()
		}
	})
}

// runBenchmarkLockWithDefer benchmarks using defer to unlock
func runBenchmarkLockWithDefer(b *testing.B, mu mutex.Mutex) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			func() {
				mu.Lock()
				defer mu.Unlock()
				n = 2*n + 1
			}()
		}
	})
}

// runBenchmarkContention benchmarks lock under contention with CPU work
func runBenchmarkContention(b *testing.B, mu mutex.Mutex) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			// Simulate work with some CPU-bound operations
			for range 50 {
				n = n*2 + 1
			}
			mu.Unlock()
		}
	})
}

// runBenchmarkRetryLock benchmarks retry-based locking using TryLock
func runBenchmarkRetryLock(b *testing.B, mu mutex.Mutex) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for !mu.TryLock() {
				// Busy-wait until lock is available
				continue
			}
			n = 2*n + 1
			mu.Unlock()
		}
	})
}

// runBenchmarkTryLock benchmarks TryLock operations.
func runBenchmarkTryLock(b *testing.B, mu mutex.Mutex) {
	var lockAttempts atomic.Int32
	var locksCount atomic.Int32

	b.ResetTimer()
	startTime := time.Now()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lockAttempts.Add(1)

			if mu.TryLock() {
				locksCount.Add(1)
				mu.Unlock()
			}
		}
	})

	synctesting.ReportTryMetrics(b, lockAttempts.Load(), locksCount.Load(),
		time.Since(startTime), "lock")
}

// ----- Benchmark implementations -----

// Basic lock benchmarks
func BenchmarkLock_SpinLock(b *testing.B) {
	var sl spinlock.SpinLock
	runBenchmarkBasicLock(b, &sl)
}

func BenchmarkLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkBasicLock(b, &mu)
}

// Deferred unlock benchmarks
func BenchmarkLockWithDefer_SpinLock(b *testing.B) {
	var sl spinlock.SpinLock
	runBenchmarkLockWithDefer(b, &sl)
}

func BenchmarkLockWithDefer_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkLockWithDefer(b, &mu)
}

// Contention benchmarks
func BenchmarkContention_SpinLock(b *testing.B) {
	var sl spinlock.SpinLock
	runBenchmarkContention(b, &sl)
}

func BenchmarkContention_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkContention(b, &mu)
}

// TryLock benchmark with retry
func BenchmarkRetryLock_SpinLock(b *testing.B) {
	var sl spinlock.SpinLock
	runBenchmarkRetryLock(b, &sl)
}

func BenchmarkRetryLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkRetryLock(b, &mu)
}

// TryLock benchmark with target counter
func BenchmarkTryLock_SpinLock(b *testing.B) {
	var sl spinlock.SpinLock
	runBenchmarkTryLock(b, &sl)
}

func BenchmarkTryLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkTryLock(b, &mu)
}
