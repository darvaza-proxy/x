package spinlock_test

import (
	"runtime"
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

// TestSpinLock_TryLockDoesNotBlock is the deterministic counterpart to
// TestSpinLock_LockBlocks: against a held spinlock, TryLock reports
// failure from a second goroutine and returns within the budget instead
// of spinning until release. Both halves hold on any schedule, so this
// pins the contract the contention test can only sample.
func TestSpinLock_TryLockDoesNotBlock(t *testing.T) {
	var sl spinlock.SpinLock

	core.AssertMustTrue(t, sl.TryLock(), "initial TryLock")

	var acquired atomic.Bool
	done := make(chan struct{})
	go func() {
		acquired.Store(sl.TryLock())
		close(done)
	}()

	synctesting.AssertMustClosed(t, done, spinlockTestTimeout,
		"TryLock returns while the spinlock is held")
	core.AssertFalse(t, acquired.Load(),
		"TryLock while held by another goroutine")

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
// test: the total attempts, an atomic tally of successful acquisitions,
// a lock-protected counter bumped inside the critical section, a live
// occupancy gauge, and a count of exclusion breaches.
//
// occupancy is raised on entry to the critical section and lowered on
// exit; a correct lock never lets it exceed one, so any entry that
// observes a higher value records a violation. The gauge never
// false-fails, though it only catches admissions whose gauge windows
// overlap, which the yield in run widens deliberately.
//
// The counter/successes equality is the belt-and-braces check: a
// double-admit races the plain counter — the race detector flags the
// race, and a lost increment breaks the equality.
type tryLockStats struct {
	counter int

	sl         spinlock.SpinLock
	attempts   atomic.Int32
	successes  atomic.Int32
	occupancy  atomic.Int32
	violations atomic.Int32
}

// run performs iterations TryLock attempts. Each success raises the
// occupancy gauge inside the critical section — a value above one is a
// double-admit and records a violation — then bumps the success tally and
// the lock-protected counter before releasing.
//
// The yield before the gauge is lowered hands the processor to another
// worker while this one is admitted. Without it a schedule offering no
// parallelism runs each worker's attempts back to back, none of them
// ever meets a held lock, and every oracle here passes vacuously.
func (s *tryLockStats) run(iterations int) {
	for range iterations {
		s.attempts.Add(1)
		if s.sl.TryLock() {
			if s.occupancy.Add(1) != 1 {
				s.violations.Add(1)
			}
			s.successes.Add(1)
			s.counter++
			runtime.Gosched()
			s.occupancy.Add(-1)
			s.sl.Unlock()
		}
	}
}

// TestSpinLock_TryLockConcurrent verifies TryLock never double-admits
// under contention. The occupancy gauge asserts mutual exclusion
// directly: no entry ever saw a second worker already inside, and run's
// yield keeps that window open even where the schedule offers no
// parallelism. The counter/successes equality backs that up — the race
// detector flags a double-admit racing the plain counter, and a lost
// increment breaks the equality. The non-blocking half of the contract
// belongs to TestSpinLock_TryLockDoesNotBlock, which pins it on every
// schedule rather than sampling it here.
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

	attempts := stats.attempts.Load()
	successes := stats.successes.Load()
	violations := stats.violations.Load()
	counter := int32(stats.counter)
	t.Logf("attempts: %d, successes: %d", attempts, successes)

	core.AssertEqual(t, int32(goroutines*iterations), attempts,
		"every attempt counted")
	core.AssertTrue(t, successes > 0, "some attempts succeed")
	core.AssertEqual(t, int32(0), violations,
		"mutual exclusion: no double-admit observed")
	core.AssertEqual(t, successes, counter,
		"mutual exclusion: counter matches successes")
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
