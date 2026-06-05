package semaphore_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/sync/internal/synctesting"
	"darvaza.org/x/sync/mutex"
	"darvaza.org/x/sync/semaphore"
)

const (
	// semaphoreTestTimeout caps each foreground synchronisation step:
	// generous enough to absorb scheduler jitter on loaded CI workers,
	// tight enough that a hung waiter fails the run instead of stalling.
	semaphoreTestTimeout = time.Second

	// semaphoreOpenGuard bounds negative-case assertions ("did not
	// acquire"): long enough to catch a spurious early acquisition,
	// short enough to keep the test responsive.
	semaphoreOpenGuard = 20 * time.Millisecond

	// workDelay is the brief unit of simulated work held under a lock.
	workDelay = 10 * time.Millisecond
)

// doSomething simulates a short operation held under a lock.
func doSomething() { time.Sleep(workDelay) }

func TestSemaphore_Lock_Unlock(t *testing.T) {
	t.Run("basic lock unlock", runTestLockUnlockBasic)
	t.Run("sequential locks", runTestLockUnlockSequential)
}

func runTestLockUnlockBasic(_ *testing.T) {
	s := &semaphore.Semaphore{}
	s.Lock()
	doSomething()
	s.Unlock()
}

func runTestLockUnlockSequential(_ *testing.T) {
	s := &semaphore.Semaphore{}
	for range 5 {
		s.Lock()
		doSomething()
		s.Unlock()
	}
}

// tryLockTestCase verifies TryLock's non-blocking acquisition against a
// possibly-held exclusive lock.
type tryLockTestCase struct {
	name string

	holdLock bool
	want     bool
}

func newTryLockTestCase(name string, holdLock, want bool) tryLockTestCase {
	return tryLockTestCase{name: name, holdLock: holdLock, want: want}
}

func (tc tryLockTestCase) Name() string { return tc.name }

func (tc tryLockTestCase) Test(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	if tc.holdLock {
		s.Lock()
		defer s.Unlock()
	}

	got := s.TryLock()
	core.AssertEqual(t, tc.want, got, "TryLock")
	if got {
		s.Unlock()
	}
}

var _ core.TestCase = tryLockTestCase{}

func TestSemaphore_TryLock(t *testing.T) {
	core.RunTestCases(t, []tryLockTestCase{
		newTryLockTestCase("acquire free lock", false, true),
		newTryLockTestCase("fail to acquire held lock", true, false),
	})
}

func TestSemaphore_RLock_RUnlock(t *testing.T) {
	t.Run("basic rlock runlock", runTestRLockBasic)
	t.Run("multiple readers", runTestRLockMultiple)
	t.Run("readers block writer", runTestRLockBlocksWriter)
}

func runTestRLockBasic(_ *testing.T) {
	s := &semaphore.Semaphore{}
	s.RLock()
	doSomething()
	s.RUnlock()
}

func runTestRLockMultiple(_ *testing.T) {
	s := &semaphore.Semaphore{}
	s.RLock()
	s.RLock()
	s.RLock()

	doSomething()

	s.RUnlock()
	s.RUnlock()
	s.RUnlock()
}

func runTestRLockBlocksWriter(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.RLock()

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		s.Lock()
		doSomething()
		s.Unlock()
	}()

	synctesting.AssertMustOpen(t, writerDone, semaphoreOpenGuard,
		"writer blocked while reader holds lock")

	s.RUnlock()

	synctesting.AssertMustClosed(t, writerDone, semaphoreTestTimeout,
		"writer proceeds after reader released")
}

// tryRLockTestCase verifies TryRLock's non-blocking acquisition against a
// possibly-held reader or writer lock.
type tryRLockTestCase struct {
	name string

	holdReader bool
	holdWriter bool
	want       bool
}

func newTryRLockTestCase(name string, holdReader, holdWriter,
	want bool) tryRLockTestCase {
	return tryRLockTestCase{
		name:       name,
		holdReader: holdReader,
		holdWriter: holdWriter,
		want:       want,
	}
}

func (tc tryRLockTestCase) Name() string { return tc.name }

func (tc tryRLockTestCase) Test(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	switch {
	case tc.holdWriter:
		s.Lock()
		defer s.Unlock()
	case tc.holdReader:
		s.RLock()
		defer s.RUnlock()
	default:
		// hold nothing; the lock is free
	}

	got := s.TryRLock()
	core.AssertEqual(t, tc.want, got, "TryRLock")
	if got {
		s.RUnlock()
	}
}

var _ core.TestCase = tryRLockTestCase{}

func TestSemaphore_TryRLock(t *testing.T) {
	core.RunTestCases(t, []tryRLockTestCase{
		newTryRLockTestCase("acquire free lock", false, false, true),
		newTryRLockTestCase("acquire with existing readers", true, false, true),
		newTryRLockTestCase("fail with writer", false, true, false),
	})
}

// lockMode bundles a context-aware acquisition entry point with the
// matching blocking lock/unlock pair, so the context scenarios run
// identically against both the exclusive and shared modes.
type lockMode struct {
	name        string
	contextLock func(*semaphore.Semaphore, context.Context) error
	lock        func(*semaphore.Semaphore)
	unlock      func(*semaphore.Semaphore)
}

func lockModes() []lockMode {
	return []lockMode{
		{
			name:        "exclusive",
			contextLock: (*semaphore.Semaphore).LockContext,
			lock:        (*semaphore.Semaphore).Lock,
			unlock:      (*semaphore.Semaphore).Unlock,
		},
		{
			name:        "shared",
			contextLock: (*semaphore.Semaphore).RLockContext,
			lock:        (*semaphore.Semaphore).RLock,
			unlock:      (*semaphore.Semaphore).RUnlock,
		},
	}
}

func TestSemaphore_ContextLocking(t *testing.T) {
	for _, m := range lockModes() {
		t.Run(m.name, m.run)
	}
}

func (m lockMode) run(t *testing.T) {
	t.Helper()
	t.Run("acquire with context", m.testAcquire)
	t.Run("nil context", m.testNilContext)
	t.Run("nil receiver", m.testNilReceiver)
	t.Run("cancelled context", m.testCancelled)
	t.Run("timeout context", m.testTimeout)
}

func (m lockMode) testAcquire(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	core.AssertMustNoError(t, m.contextLock(s, context.Background()), "acquire")
	m.unlock(s)
}

func (m lockMode) testNilContext(t *testing.T) {
	t.Helper()
	var ctx context.Context
	s := &semaphore.Semaphore{}
	core.AssertError(t, m.contextLock(s, ctx), "nil context")
}

func (m lockMode) testNilReceiver(t *testing.T) {
	t.Helper()
	var s *semaphore.Semaphore
	core.AssertError(t, m.contextLock(s, context.Background()), "nil receiver")
}

func (m lockMode) testCancelled(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	m.lock(s)
	defer m.unlock(s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	core.AssertErrorIs(t, m.contextLock(s, ctx), context.Canceled,
		"cancelled context")
}

func (m lockMode) testTimeout(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	m.lock(s)
	defer m.unlock(s)

	// A deadline already in the past forces DeadlineExceeded at entry,
	// deterministically and regardless of mode.
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(-time.Millisecond))
	defer cancel()

	core.AssertErrorIs(t, m.contextLock(s, ctx), context.DeadlineExceeded,
		"timeout context")
}

func TestSemaphore_Concurrency(t *testing.T) {
	t.Run("multiple writers", runTestConcurrentWriters)
	t.Run("readers and writers", runTestConcurrentReadersWriters)
}

func runTestConcurrentWriters(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	count := 0
	const numGoroutines, iterations = 10, 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				s.Lock()
				count++
				s.Unlock()
			}
		}()
	}
	wg.Wait()

	core.AssertEqual(t, numGoroutines*iterations, count, "total increments")
}

func runTestConcurrentReadersWriters(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	counter := 0
	const numReaders, numWriters, iterations = 10, 5, 50

	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)
	for range numReaders {
		go runReaderLoop(&wg, s, &counter, iterations)
	}
	for range numWriters {
		go runWriterLoop(&wg, s, &counter, iterations)
	}
	wg.Wait()

	core.AssertEqual(t, numWriters*iterations, counter, "writer increments")
}

func runReaderLoop(wg *sync.WaitGroup, s *semaphore.Semaphore, counter *int,
	iterations int) {
	defer wg.Done()
	for range iterations {
		s.RLock()
		_ = *counter
		time.Sleep(time.Millisecond)
		s.RUnlock()
	}
}

func runWriterLoop(wg *sync.WaitGroup, s *semaphore.Semaphore, counter *int,
	iterations int) {
	defer wg.Done()
	for range iterations {
		s.Lock()
		*counter++
		time.Sleep(2 * time.Millisecond)
		s.Unlock()
	}
}

// semaphorePanicTestCase verifies operations that panic on misuse or on a
// nil receiver. setup arranges the receiver state; op selects the method
// exercised under the shared assertion path.
type semaphorePanicTestCase struct {
	name  string
	setup func() *semaphore.Semaphore
	op    func(*semaphore.Semaphore)
}

func newSemaphorePanicTestCase(name string, setup func() *semaphore.Semaphore,
	op func(*semaphore.Semaphore)) semaphorePanicTestCase {
	return semaphorePanicTestCase{name: name, setup: setup, op: op}
}

func (tc semaphorePanicTestCase) Name() string { return tc.name }

func (tc semaphorePanicTestCase) Test(t *testing.T) {
	t.Helper()
	s := tc.setup()
	core.AssertPanic(t, func() { tc.op(s) }, nil, "panic")
}

var _ core.TestCase = semaphorePanicTestCase{}

func newSemaphore() *semaphore.Semaphore { return &semaphore.Semaphore{} }
func nilSemaphore() *semaphore.Semaphore { return nil }

func readLockedSemaphore() *semaphore.Semaphore {
	s := &semaphore.Semaphore{}
	s.RLock()
	return s
}

func opLock(s *semaphore.Semaphore)     { s.Lock() }
func opUnlock(s *semaphore.Semaphore)   { s.Unlock() }
func opRLock(s *semaphore.Semaphore)    { s.RLock() }
func opRUnlock(s *semaphore.Semaphore)  { s.RUnlock() }
func opTryLock(s *semaphore.Semaphore)  { s.TryLock() }
func opTryRLock(s *semaphore.Semaphore) { s.TryRLock() }

func semaphorePanicTestCases() []semaphorePanicTestCase {
	return []semaphorePanicTestCase{
		// misuse on an initialised receiver
		newSemaphorePanicTestCase("unlock without lock", newSemaphore, opUnlock),
		newSemaphorePanicTestCase("runlock without rlock", newSemaphore, opRUnlock),
		newSemaphorePanicTestCase("unlock when read-locked", readLockedSemaphore,
			opUnlock),
		// nil receiver panics through the public wrappers
		newSemaphorePanicTestCase("nil receiver Lock", nilSemaphore, opLock),
		newSemaphorePanicTestCase("nil receiver RLock", nilSemaphore, opRLock),
		newSemaphorePanicTestCase("nil receiver Unlock", nilSemaphore, opUnlock),
		newSemaphorePanicTestCase("nil receiver RUnlock", nilSemaphore, opRUnlock),
		newSemaphorePanicTestCase("nil receiver TryLock", nilSemaphore, opTryLock),
		newSemaphorePanicTestCase("nil receiver TryRLock", nilSemaphore, opTryRLock),
	}
}

func TestSemaphore_ErrorCases(t *testing.T) {
	core.RunTestCases(t, semaphorePanicTestCases())
}

func TestSemaphore_InterfaceCompliance(_ *testing.T) {
	var locker sync.Locker = &semaphore.Semaphore{}
	locker.Lock()
	doSomething()
	locker.Unlock()
}

func TestSemaphore_EdgeCases(t *testing.T) {
	t.Run("writer after readers", runTestWriterAfterReaders)
	t.Run("readers after writer", runTestReadersAfterWriter)
}

// runTestWriterAfterReaders verifies a writer blocks until every reader
// releases its lock.
func runTestWriterAfterReaders(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.RLock()
	s.RLock()
	s.RLock()

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		s.Lock()
		defer s.Unlock()
	}()

	synctesting.AssertMustOpen(t, writerDone, semaphoreOpenGuard,
		"writer blocked while readers hold lock")

	s.RUnlock()
	s.RUnlock()
	s.RUnlock()

	synctesting.AssertMustClosed(t, writerDone, semaphoreTestTimeout,
		"writer proceeds after all readers released")
}

// runTestReadersAfterWriter verifies readers block until the writer
// releases its lock, then all acquire.
func runTestReadersAfterWriter(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.Lock()

	const numReaders = 3
	readersDone := make(chan struct{}, numReaders)
	for range numReaders {
		go func() {
			s.RLock()
			defer s.RUnlock()
			readersDone <- struct{}{}
		}()
	}

	synctesting.AssertMustOpen(t, readersDone, semaphoreOpenGuard,
		"readers blocked while writer holds lock")

	s.Unlock()

	synctesting.AssertMustReadersReady(t, readersDone, numReaders,
		semaphoreTestTimeout, "all readers acquire after writer released")
}

func TestSemaphore_ContextCancellation(t *testing.T) {
	t.Run("cancel during wait for exclusive lock", runTestCancelExclusiveWait)
	t.Run("cancel during wait for read lock", runTestCancelReadWait)
	t.Run("deadline exceeded during lock wait", runTestDeadlineExceeded)
}

func runTestCancelExclusiveWait(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.Lock()
	defer s.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	var gotErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		gotErr = s.LockContext(ctx)
	}()

	synctesting.AssertMustOpen(t, done, semaphoreOpenGuard,
		"LockContext blocks while lock held")
	cancel()
	synctesting.AssertMustClosed(t, done, semaphoreTestTimeout,
		"LockContext unblocked by cancel")
	core.AssertErrorIs(t, gotErr, context.Canceled, "LockContext cancelled")
}

func runTestCancelReadWait(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.Lock()
	defer s.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	var gotErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		gotErr = s.RLockContext(ctx)
	}()

	synctesting.AssertMustOpen(t, done, semaphoreOpenGuard,
		"RLockContext blocks while writer holds lock")
	cancel()
	synctesting.AssertMustClosed(t, done, semaphoreTestTimeout,
		"RLockContext unblocked by cancel")
	core.AssertErrorIs(t, gotErr, context.Canceled, "RLockContext cancelled")
}

func runTestDeadlineExceeded(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}
	s.Lock()
	defer s.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), semaphoreOpenGuard)
	defer cancel()

	core.AssertErrorIs(t, s.LockContext(ctx), context.DeadlineExceeded,
		"LockContext deadline exceeded")
}

func TestSemaphore_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	s := &semaphore.Semaphore{}
	const numWorkers, iterations = 20, 500

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	counter := 0
	var counterMutex sync.Mutex

	for i := range numWorkers {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				performOperation(s, (id+j)%6, &counter, &counterMutex)
			}
		}(i)
	}
	wg.Wait()

	// The exact total is unknowable because TryLock/TryRLock and the
	// timeouts may skip increments, but it must have advanced.
	core.AssertTrue(t, counter > 0, "counter advanced")
}

// performOperation executes one of the semaphore operations based on the
// operation type.
func performOperation(s *semaphore.Semaphore, opType int, counter *int,
	counterMutex *sync.Mutex) {
	switch opType {
	case 0:
		exclusiveLockOperation(s, counter, counterMutex)
	case 1:
		tryLockOperation(s, counter, counterMutex)
	case 2:
		readLockOperation(s, counter, counterMutex)
	case 3:
		tryReadLockOperation(s, counter, counterMutex)
	case 4:
		lockContextOperation(s, counter, counterMutex)
	default:
		readLockContextOperation(s, counter, counterMutex)
	}
}

func exclusiveLockOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	s.Lock()
	counterMutex.Lock()
	*counter++
	counterMutex.Unlock()
	s.Unlock()
}

func tryLockOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	if s.TryLock() {
		counterMutex.Lock()
		*counter++
		counterMutex.Unlock()
		s.Unlock()
	}
}

func readLockOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	s.RLock()
	counterMutex.Lock()
	_ = *counter
	counterMutex.Unlock()
	s.RUnlock()
}

func tryReadLockOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	if s.TryRLock() {
		counterMutex.Lock()
		_ = *counter
		counterMutex.Unlock()
		s.RUnlock()
	}
}

func lockContextOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if s.LockContext(ctx) == nil {
		counterMutex.Lock()
		*counter++
		counterMutex.Unlock()
		s.Unlock()
	}
}

func readLockContextOperation(s *semaphore.Semaphore, counter *int,
	counterMutex *sync.Mutex) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if s.RLockContext(ctx) == nil {
		counterMutex.Lock()
		_ = *counter
		counterMutex.Unlock()
		s.RUnlock()
	}
}

func TestSemaphore_BoundaryConditions(t *testing.T) {
	t.Run("many consecutive read locks", runTestManyReadLocks)
}

func runTestManyReadLocks(t *testing.T) {
	t.Helper()
	s := &semaphore.Semaphore{}

	const numLocks = 1000
	for range numLocks {
		s.RLock()
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		s.Lock()
		doSomething()
		s.Unlock()
	}()

	synctesting.AssertMustOpen(t, writerDone, semaphoreOpenGuard,
		"writer blocked while 1000 readers hold lock")

	for range numLocks {
		s.RUnlock()
	}

	synctesting.AssertMustClosed(t, writerDone, semaphoreTestTimeout,
		"writer proceeds after all readers released")
}

// ----- Benchmark functions -----

// runBenchmarkBasicLock benchmarks basic lock/unlock operations.
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

// runBenchmarkLockWithDefer benchmarks using defer to unlock.
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

// runBenchmarkContention benchmarks lock under contention with CPU work.
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

// runBenchmarkRWLock benchmarks read lock operations with occasional writes.
func runBenchmarkRWLock(b *testing.B, mu mutex.RWMutex) {
	n := 1
	const readersPerWriter = 100 // Ratio of reads to writes

	b.RunParallel(func(pb *testing.PB) {
		count := 0
		for pb.Next() {
			if count%readersPerWriter == 0 {
				// Occasionally do a write
				mu.Lock()
				n = 2*n + 1
				mu.Unlock()
			} else {
				// Mostly do reads
				mu.RLock()
				_ = n
				mu.RUnlock()
			}
			count++
		}
	})
}

// runBenchmarkReadOnly benchmarks read-only operations.
func runBenchmarkReadOnly(b *testing.B, mu mutex.RWMutex) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			_ = n
			mu.RUnlock()
		}
	})
}

// reportTryMetrics reports the attempts/lock, locks-per-second and
// nanoseconds-per-lock ratios for a TryLock-style benchmark. unit is the
// per-acquisition noun ("lock" or "rlock"). Guard clauses keep this off
// the benchmark loop's complexity budget.
func reportTryMetrics(b *testing.B, attempts, count int32,
	elapsed time.Duration, unit string) {
	if count <= 0 {
		return
	}
	b.ReportMetric(float64(attempts)/float64(count), "attempts/"+unit)
	if elapsed <= 0 {
		return
	}
	b.ReportMetric(float64(count)/elapsed.Seconds(), unit+"s/sec")
	b.ReportMetric(float64(elapsed.Nanoseconds())/float64(attempts), "ns/"+unit)
}

// runBenchmarkTryLock benchmarks TryLock operations.
func runBenchmarkTryLock(b *testing.B, mu mutex.Mutex) {
	var lockAttempts atomic.Int32
	var locksCount int32

	b.ResetTimer()
	startTime := time.Now()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lockAttempts.Add(1)

			if mu.TryLock() {
				locksCount++
				mu.Unlock()
			}
		}
	})

	reportTryMetrics(b, lockAttempts.Load(), locksCount,
		time.Since(startTime), "lock")
}

// runBenchmarkTryRLock benchmarks TryRLock operations.
func runBenchmarkTryRLock(b *testing.B, mu mutex.RWMutex) {
	var lockAttempts atomic.Int32
	var locksCount int32

	b.ResetTimer()
	startTime := time.Now()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lockAttempts.Add(1)

			if mu.TryRLock() {
				locksCount++
				mu.RUnlock()
			}
		}
	})

	reportTryMetrics(b, lockAttempts.Load(), locksCount,
		time.Since(startTime), "rlock")
}

// runBenchmarkContextLock benchmarks LockContext operations.
func runBenchmarkContextLock(b *testing.B, mu mutex.MutexContext) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			if err := mu.LockContext(ctx); err == nil {
				n = 2*n + 1
				mu.Unlock()
			}
		}
	})
}

// runBenchmarkContextRLock benchmarks RLockContext operations.
func runBenchmarkContextRLock(b *testing.B, mu mutex.RWMutexContext) {
	n := 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			if err := mu.RLockContext(ctx); err == nil {
				_ = n
				mu.RUnlock()
			}
		}
	})
}

// ----- Benchmark implementations -----

// Basic lock/unlock benchmarks
func BenchmarkLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkBasicLock(b, s)
}

func BenchmarkLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkBasicLock(b, &mu)
}

func BenchmarkLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkBasicLock(b, &mu)
}

// Deferred unlock benchmarks
func BenchmarkLockWithDefer_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkLockWithDefer(b, s)
}

func BenchmarkLockWithDefer_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkLockWithDefer(b, &mu)
}

func BenchmarkLockWithDefer_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkLockWithDefer(b, &mu)
}

// Contention benchmarks
func BenchmarkContention_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkContention(b, s)
}

func BenchmarkContention_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkContention(b, &mu)
}

func BenchmarkContention_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkContention(b, &mu)
}

// Reader/Writer lock benchmarks
func BenchmarkRWLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkRWLock(b, s)
}

func BenchmarkRWLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkRWLock(b, &mu)
}

// Read-only benchmarks
func BenchmarkReadOnly_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkReadOnly(b, s)
}

func BenchmarkReadOnly_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkReadOnly(b, &mu)
}

// TryLock benchmarks
func BenchmarkTryLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkTryLock(b, s)
}

func BenchmarkTryLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkTryLock(b, &mu)
}

func BenchmarkTryLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkTryLock(b, &mu)
}

// TryRLock benchmarks
func BenchmarkTryRLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkTryRLock(b, s)
}

func BenchmarkTryRLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkTryRLock(b, &mu)
}

// Context lock benchmarks
func BenchmarkContextLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkContextLock(b, s)
}

func BenchmarkContextRLock_Semaphore(b *testing.B) {
	s := &semaphore.Semaphore{}
	runBenchmarkContextRLock(b, s)
}
