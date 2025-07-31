package semaphore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/x/sync/mutex"

	"github.com/stretchr/testify/assert"
)

// doSomething simulates a short operation by sleeping for 10ms.
func doSomething() {
	time.Sleep(10 * time.Millisecond)
}

func TestSemaphore_Init(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var s *Semaphore
		err := s.lazyInit()
		assert.Error(t, err)
	})

	t.Run("initialise channels", func(t *testing.T) {
		s := &Semaphore{}
		err := s.lazyInit()
		assert.NoError(t, err)
		assert.NotNil(t, s.global)
		assert.NotNil(t, s.readers)
	})

	t.Run("idempotent", func(t *testing.T) {
		s := &Semaphore{}
		err := s.lazyInit()
		assert.NoError(t, err)

		global := s.global
		readers := s.readers

		err = s.lazyInit()
		assert.NoError(t, err)
		assert.True(t, global == s.global, "global channel reference changed")
		assert.True(t, readers == s.readers, "readers channel reference changed")
	})
}

func TestSemaphore_Lock_Unlock(t *testing.T) {
	t.Run("basic lock unlock", func(_ *testing.T) {
		s := &Semaphore{}
		s.Lock()
		doSomething()
		s.Unlock()
	})

	t.Run("sequential locks", func(_ *testing.T) {
		s := &Semaphore{}
		for range 5 {
			s.Lock()
			doSomething()
			s.Unlock()
		}
	})
}

func TestSemaphore_TryLock(t *testing.T) {
	t.Run("acquire free lock", func(t *testing.T) {
		s := &Semaphore{}
		acquired := s.TryLock()
		assert.True(t, acquired)
		s.Unlock()
	})

	t.Run("fail to acquire held lock", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock()

		acquired := s.TryLock()
		assert.False(t, acquired)

		s.Unlock()
	})
}

func TestSemaphore_RLock_RUnlock(t *testing.T) {
	t.Run("basic rlock runlock", func(_ *testing.T) {
		s := &Semaphore{}
		s.RLock()
		doSomething()
		s.RUnlock()
	})

	t.Run("multiple readers", func(_ *testing.T) {
		s := &Semaphore{}

		// Acquire three read locks
		s.RLock()
		s.RLock()
		s.RLock()

		doSomething()

		// Release all three
		s.RUnlock()
		s.RUnlock()
		s.RUnlock()
	})

	t.Run("readers block writer", func(t *testing.T) {
		s := &Semaphore{}
		s.RLock()

		writerDone := make(chan struct{})
		go func() {
			defer close(writerDone)
			// This should block until the read lock is released
			s.Lock()
			doSomething()
			s.Unlock()
		}()

		// Writer shouldn't be able to acquire lock immediately
		select {
		case <-writerDone:
			t.Fatal("Writer acquired lock while reader held it")
		case <-time.After(50 * time.Millisecond):
			// Expected behaviour
		}

		// Release reader and writer should proceed
		s.RUnlock()

		select {
		case <-writerDone:
			// Expected behaviour
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Writer failed to acquire lock after reader released it")
		}
	})
}

func TestSemaphore_TryRLock(t *testing.T) {
	t.Run("acquire free lock", func(t *testing.T) {
		s := &Semaphore{}
		acquired := s.TryRLock()
		assert.True(t, acquired)
		s.RUnlock()
	})

	t.Run("acquire with existing readers", func(t *testing.T) {
		s := &Semaphore{}
		s.RLock()

		acquired := s.TryRLock()
		assert.True(t, acquired)

		s.RUnlock()
		s.RUnlock()
	})

	t.Run("fail with writer", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock()

		acquired := s.TryRLock()
		assert.False(t, acquired)

		s.Unlock()
	})
}

// TestSemaphore_ContextLocking tests both LockContext and RLockContext
//
//revive:disable-next-line:cognitive-complexity
func TestSemaphore_ContextLocking(t *testing.T) {
	testCases := []lockFunctions{
		{
			name:        "exclusive",
			contextLock: (*Semaphore).LockContext,
			lock:        (*Semaphore).Lock,
			unlock:      (*Semaphore).Unlock,
		},
		{
			name:        "shared",
			contextLock: (*Semaphore).RLockContext,
			lock:        (*Semaphore).RLock,
			unlock:      (*Semaphore).RUnlock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("acquire with context", func(t *testing.T) {
				s := &Semaphore{}
				ctx := context.Background()

				err := tc.contextLock(s, ctx)
				assert.NoError(t, err)
				tc.unlock(s)
			})

			t.Run("nil context", func(t *testing.T) {
				var ctx context.Context
				s := &Semaphore{}
				err := tc.contextLock(s, ctx)
				assert.Error(t, err)
			})

			t.Run("cancelled context", func(t *testing.T) {
				s := &Semaphore{}
				// Hold the lock
				tc.lock(s)

				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				// Should fail due to cancelled context
				err := tc.contextLock(s, ctx)
				assert.Error(t, err)
				assert.Equal(t, context.Canceled, err)

				tc.unlock(s)
			})

			t.Run("timeout context", func(t *testing.T) {
				s := &Semaphore{}
				// Hold the lock
				tc.lock(s)

				ctx, cancel := context.WithTimeout(
					context.Background(),
					50*time.Millisecond,
				)
				defer cancel()

				time.Sleep(100 * time.Millisecond)

				// Should timeout
				err := tc.contextLock(s, ctx)
				assert.Error(t, err)
				assert.Equal(t, context.DeadlineExceeded, err)

				tc.unlock(s)
			})
		})
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSemaphore_Concurrency(t *testing.T) {
	t.Run("multiple writers", func(t *testing.T) {
		s := &Semaphore{}
		count := 0
		const numGoroutines = 10
		const iterations = 100

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
		assert.Equal(t, numGoroutines*iterations, count)
	})

	t.Run("readers and writers", func(t *testing.T) {
		s := &Semaphore{}
		counter := 0
		const numReaders = 10
		const numWriters = 5
		const iterations = 50

		var wg sync.WaitGroup
		wg.Add(numReaders + numWriters)

		// Start readers
		for range numReaders {
			go func() {
				defer wg.Done()
				for range iterations {
					s.RLock()
					// Just read the counter, don't modify
					_ = counter
					time.Sleep(1 * time.Millisecond)
					s.RUnlock()
				}
			}()
		}

		// Start writers
		for range numWriters {
			go func() {
				defer wg.Done()
				for range iterations {
					s.Lock()
					counter++
					time.Sleep(2 * time.Millisecond)
					s.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, numWriters*iterations, counter)
	})
}

func TestSemaphore_ErrorCases(t *testing.T) {
	t.Run("unlock without lock should panic", func(t *testing.T) {
		s := &Semaphore{}
		assert.Panics(t, func() {
			s.Unlock()
		})
	})

	t.Run("runlock without rlock should panic", func(t *testing.T) {
		s := &Semaphore{}
		assert.Panics(t, func() {
			s.RUnlock()
		})
	})

	t.Run("unlock when read-locked should panic", func(t *testing.T) {
		s := &Semaphore{}
		s.RLock()
		assert.Panics(t, func() {
			s.Unlock()
		})
	})
}

// lockFunctions groups related locking and unlocking functions together
// for testing purposes.
type lockFunctions struct {
	name        string
	contextLock func(*Semaphore, context.Context) error
	lock        func(*Semaphore)
	unlock      func(*Semaphore)
}

func TestSemaphore_InterfaceCompliance(t *testing.T) {
	var s Semaphore

	// Should satisfy sync.Locker interface
	var _ sync.Locker = &s

	// Test API compatibility
	t.Run("lock unlock pattern", func(_ *testing.T) {
		var locker sync.Locker = &Semaphore{}
		locker.Lock()
		doSomething()
		locker.Unlock()
	})
}

func TestSemaphore_EdgeCases(t *testing.T) {
	t.Run("writer after readers", func(t *testing.T) {
		testWriterAfterReaders(t)
	})

	t.Run("readers after writer", func(t *testing.T) {
		testReadersAfterWriter(t)
	})
}

// testWriterAfterReaders tests that writers block until all readers release their locks
func testWriterAfterReaders(t *testing.T) {
	s := &Semaphore{}

	// Acquire multiple read locks
	s.RLock()
	s.RLock()
	s.RLock()

	writerReady := make(chan struct{})
	writerDone := make(chan struct{})

	// Writer should block until all readers are done
	go func() {
		close(writerReady)
		s.Lock()
		defer s.Unlock()
		close(writerDone)
	}()

	// Wait for writer to be ready
	<-writerReady
	time.Sleep(10 * time.Millisecond)

	// Verify writer is blocked
	verifyChannelNotClosed(t, writerDone, "Writer acquired lock while readers held it")

	// Release all read locks
	s.RUnlock()
	s.RUnlock()
	s.RUnlock()

	// Verify writer proceeds after readers release
	verifyChannelCloses(t, writerDone, 100*time.Millisecond,
		"Writer failed to acquire lock after all readers released")
}

// testReadersAfterWriter tests that readers block until writer releases its lock
func testReadersAfterWriter(t *testing.T) {
	s := &Semaphore{}
	s.Lock()

	done := make(chan struct{})
	ready := make(chan struct{})

	// Start readers that will block
	const numReaders = 3
	for range numReaders {
		go func() {
			ready <- struct{}{}
			s.RLock()
			defer s.RUnlock()
			done <- struct{}{}
		}()
	}

	// Wait for all readers to be ready
	waitForReadersReady(numReaders, ready)
	time.Sleep(10 * time.Millisecond)

	// Verify readers are blocked
	verifyChannelNotClosed(t, done, "Reader acquired lock while writer held it")

	// Release the write lock
	s.Unlock()

	// Verify all readers proceed
	verifyAllReadersAcquiredLocks(t, numReaders, done)
}

// waitForReadersReady waits for the specified number of readers to signal readiness
func waitForReadersReady(count int, ready chan struct{}) {
	for range count {
		<-ready
	}
}

// verifyChannelNotClosed checks that a channel has not been closed
func verifyChannelNotClosed(t *testing.T, ch chan struct{}, failMessage string) {
	select {
	case <-ch:
		t.Fatal(failMessage)
	default:
		// Expected behaviour
	}
}

// verifyChannelCloses checks that a channel is closed within the timeout period
func verifyChannelCloses(t *testing.T, ch chan struct{}, timeout time.Duration, failMessage string) {
	select {
	case <-ch:
		// Expected behaviour
	case <-time.After(timeout):
		t.Fatal(failMessage)
	}
}

// verifyAllReadersAcquiredLocks verifies that all readers acquired their locks
func verifyAllReadersAcquiredLocks(t *testing.T, numReaders int, done chan struct{}) {
	timeout := time.After(100 * time.Millisecond)
	for range numReaders {
		select {
		case <-done:
			// Expected behaviour
		case <-timeout:
			t.Fatal("Readers failed to acquire lock after writer released")
		}
	}
}

func TestSemaphore_ContextCancellation(t *testing.T) {
	t.Run("cancel during wait for exclusive lock", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock() // Lock is held

		ctx, cancel := context.WithCancel(context.Background())

		// Start a goroutine that will try to acquire the lock
		errCh := make(chan error)
		go func() {
			errCh <- s.LockContext(ctx)
		}()

		// Give it a moment to block
		time.Sleep(10 * time.Millisecond)

		// Cancel the context
		cancel()

		// Should get context.Canceled error
		select {
		case err := <-errCh:
			assert.Equal(t, context.Canceled, err)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Context cancellation did not unblock LockContext")
		}

		s.Unlock()
	})

	t.Run("cancel during wait for read lock", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock() // Lock is held

		ctx, cancel := context.WithCancel(context.Background())

		// Start a goroutine that will try to acquire the lock
		errCh := make(chan error)
		go func() {
			errCh <- s.RLockContext(ctx)
		}()

		// Give it a moment to block
		time.Sleep(10 * time.Millisecond)

		// Cancel the context
		cancel()

		// Should get context.Canceled error
		select {
		case err := <-errCh:
			assert.Equal(t, context.Canceled, err)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Context cancellation did not unblock RLockContext")
		}

		s.Unlock()
	})

	t.Run("deadline exceeded during lock wait", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock() // Lock is held

		ctx, cancel := context.WithTimeout(
			context.Background(),
			50*time.Millisecond,
		)
		defer cancel()

		// Start a goroutine that will try to acquire the lock
		start := time.Now()
		err := s.LockContext(ctx)
		elapsed := time.Since(start)

		// Should get deadline exceeded error
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(50))

		s.Unlock()
	})
}

//revive:disable-next-line:cognitive-complexity
func TestSemaphore_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Run("mixed concurrent operations", func(t *testing.T) {
		s := &Semaphore{}
		const numWorkers = 20
		const iterations = 500

		var wg sync.WaitGroup
		wg.Add(numWorkers)

		counter := 0
		var counterMutex sync.Mutex

		for i := range numWorkers {
			go func(id int) {
				defer wg.Done()
				for j := range iterations {
					// Determine operation type based on worker ID and iteration
					opType := (id + j) % 6
					performOperation(s, opType, &counter, &counterMutex)
				}
			}(i)
		}

		wg.Wait()

		// We don't know exactly what the counter will be due to TryLock/TryRLock
		// and timeouts, but it should be > 0
		assert.Greater(t, counter, 0)
	})
}

// performOperation executes one of the semaphore operations based on the operation type
func performOperation(s *Semaphore, opType int, counter *int, counterMutex *sync.Mutex) {
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

// exclusiveLockOperation performs a Lock+Unlock operation
func exclusiveLockOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	s.Lock()
	counterMutex.Lock()
	(*counter)++
	counterMutex.Unlock()
	s.Unlock()
}

// tryLockOperation attempts a TryLock operation
func tryLockOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	if s.TryLock() {
		counterMutex.Lock()
		(*counter)++
		counterMutex.Unlock()
		s.Unlock()
	}
}

// readLockOperation performs a RLock+RUnlock operation
func readLockOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	s.RLock()
	// Just read the counter
	counterMutex.Lock()
	_ = *counter
	counterMutex.Unlock()
	s.RUnlock()
}

// tryReadLockOperation attempts a TryRLock operation
func tryReadLockOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	if s.TryRLock() {
		counterMutex.Lock()
		_ = *counter
		counterMutex.Unlock()
		s.RUnlock()
	}
}

// lockContextOperation performs a LockContext with short timeout
func lockContextOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		1*time.Millisecond,
	)
	defer cancel()

	if s.LockContext(ctx) == nil {
		counterMutex.Lock()
		(*counter)++
		counterMutex.Unlock()
		s.Unlock()
	}
}

// readLockContextOperation performs a RLockContext with short timeout
func readLockContextOperation(s *Semaphore, counter *int, counterMutex *sync.Mutex) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		1*time.Millisecond,
	)
	defer cancel()

	if s.RLockContext(ctx) == nil {
		counterMutex.Lock()
		_ = *counter
		counterMutex.Unlock()
		s.RUnlock()
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSemaphore_BoundaryConditions(t *testing.T) {
	t.Run("many consecutive read locks", func(t *testing.T) {
		s := &Semaphore{}

		// Acquire a lot of read locks
		const numLocks = 1000
		for range numLocks {
			s.RLock()
		}

		// Try to get a write lock - should block
		writerDone := make(chan struct{})
		go func() {
			s.Lock()
			doSomething()
			s.Unlock()
			close(writerDone)
		}()

		// Writer shouldn't proceed yet
		time.Sleep(10 * time.Millisecond)
		select {
		case <-writerDone:
			t.Fatal("Writer shouldn't have acquired the lock")
		default:
			// Expected
		}

		// Release all read locks
		for range numLocks {
			s.RUnlock()
		}

		// Now writer should be able to proceed
		select {
		case <-writerDone:
			// Expected behaviour
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Writer failed to acquire lock after all readers released")
		}
	})
}

func TestSemaphore_WriterStarvationPrevention(t *testing.T) {
	t.Run("writers not starved by continuous readers", func(t *testing.T) {
		testWritersNotStarvedByContinuousReaders(t)
	})

	t.Run("writers preferred over new readers", func(t *testing.T) {
		testWritersPreferredOverNewReaders(t)
	})
}

// testWritersNotStarvedByContinuousReaders verifies that writers can acquire
// the lock even with continuous reader traffic.
//
//revive:disable-next-line:cognitive-complexity
func testWritersNotStarvedByContinuousReaders(t *testing.T) {
	s := &Semaphore{}
	const numReaders = 100
	const readerCycles = 5

	// Track when writers acquire the lock
	writerAcquisitions := make([]time.Time, 0, 3)
	writerDone := make(chan struct{})

	// Start a writer that will try to acquire lock periodically
	go func() {
		defer close(writerDone)

		for range 3 {
			s.Lock()
			writerAcquisitions = append(writerAcquisitions, time.Now())
			// Simulate some work
			time.Sleep(5 * time.Millisecond)
			s.Unlock()

			// Small wait between attempts
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Create continuous reader pressure
	var wg sync.WaitGroup
	readersStopped := make(chan struct{})

	wg.Add(numReaders)
	for i := range numReaders {
		go func(id int) {
			defer wg.Done()

			// Each reader does multiple read lock cycles
			for range readerCycles {
				// Small jitter to prevent perfect synchronisation
				time.Sleep(time.Duration(id%3) * time.Millisecond)

				s.RLock()
				// Simulate read operation
				time.Sleep(1 * time.Millisecond)
				s.RUnlock()
			}
		}(i)
	}

	// Wait for all readers to finish and signal completion
	go func() {
		wg.Wait()
		close(readersStopped)
	}()

	// Wait for writer to finish or timeout
	select {
	case <-writerDone:
		// Expected case - writer completed its work
	case <-time.After(2 * time.Second):
		t.Fatal("Writers appear to be starved by readers")
	}

	// Wait for readers to finish or timeout
	select {
	case <-readersStopped:
		// Expected case - readers completed
	case <-time.After(2 * time.Second):
		t.Fatal("Readers failed to complete")
	}

	// Verify writer was able to acquire the lock multiple times
	assert.Equal(t, 3, len(writerAcquisitions),
		"Writer should have acquired the lock 3 times")

	// Check that writer acquisitions weren't all bunched at the end
	if len(writerAcquisitions) >= 2 {
		totalDuration := writerAcquisitions[len(writerAcquisitions)-1].Sub(
			writerAcquisitions[0])
		// 2 intervals of at least 10ms
		expectedMinDuration := 20 * time.Millisecond
		assert.GreaterOrEqual(t, totalDuration, expectedMinDuration,
			"Writer acquisitions should be spread out, not clustered")
	}
}

// testWritersPreferredOverNewReaders verifies that when writers are waiting,
// new readers are blocked until writers complete.
func testWritersPreferredOverNewReaders(t *testing.T) {
	s := &Semaphore{}

	// First get some existing readers
	for range 5 {
		s.RLock()
	}

	// Signal when a writer is waiting
	writerWaiting := make(chan struct{})
	writerDone := make(chan struct{})

	// Start a writer that will signal when it's waiting
	go func() {
		close(writerWaiting)
		s.Lock()
		defer s.Unlock()
		close(writerDone)
	}()

	// Wait for writer to be waiting
	<-writerWaiting
	time.Sleep(10 * time.Millisecond)

	// Start a reader that should be blocked by waiting writer
	readerBlocked := make(chan struct{})
	readerAcquired := make(chan struct{})
	go func() {
		close(readerBlocked)
		s.RLock()
		defer s.RUnlock()
		close(readerAcquired)
	}()

	// Wait for reader to be blocked
	<-readerBlocked
	time.Sleep(10 * time.Millisecond)

	// Verify new reader doesn't acquire the lock while writer is waiting
	select {
	case <-readerAcquired:
		t.Fatal("New reader shouldn't acquire lock while writers wait")
	default:
		// Expected behaviour
	}

	// Release all initial read locks
	for range 5 {
		s.RUnlock()
	}

	// Verify writer completes before the new reader
	select {
	case <-writerDone:
		// Expected behaviour - writer should get the lock first
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Writer failed to acquire lock after readers released")
	}

	// Now reader should get the lock
	select {
	case <-readerAcquired:
		// Expected behaviour - reader gets lock after writer
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Reader failed to acquire lock after writer completed")
	}
}

// ----- Benchmark functions -----

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

// runBenchmarkRWLock benchmarks read lock operations with occasional writes
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

// runBenchmarkReadOnly benchmarks read-only operations
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

// runBenchmarkTryLock benchmarks TryLock operations
//
//revive:disable:cognitive-complexity
func runBenchmarkTryLock(b *testing.B, mu mutex.Mutex) {
	var lockAttempts atomic.Int32
	var locksCount int32

	// Reset the timer to exclude setup
	b.ResetTimer()
	// Start timing
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

	elapsed := time.Since(startTime)
	if locksCount > 0 {
		// Report attempts per lock
		b.ReportMetric(float64(lockAttempts.Load())/float64(locksCount), "attempts/lock")
		if elapsed > 0 {
			// Report locks per second
			locksPerSec := float64(locksCount) / elapsed.Seconds()
			b.ReportMetric(locksPerSec, "locks/sec")
			// And nanoseconds per lock
			b.ReportMetric(float64(elapsed.Nanoseconds())/float64(lockAttempts.Load()), "ns/lock")
		}
	}
}

// runBenchmarkTryRLock benchmarks TryRLock operations
//
//revive:disable:cognitive-complexity
func runBenchmarkTryRLock(b *testing.B, mu mutex.RWMutex) {
	var lockAttempts atomic.Int32
	var locksCount int32

	// Reset the timer to exclude setup
	b.ResetTimer()
	// Start timing
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

	elapsed := time.Since(startTime)
	if locksCount > 0 {
		// Report attempts per rlock
		b.ReportMetric(float64(lockAttempts.Load())/float64(locksCount), "attempts/rlock")
		if elapsed > 0 {
			// Report rlocks per second
			b.ReportMetric(float64(locksCount)/elapsed.Seconds(), "rlocks/sec")
			// And nanoseconds per rlock
			b.ReportMetric(float64(elapsed.Nanoseconds())/float64(lockAttempts.Load()), "ns/rlock")
		}
	}
}

// runBenchmarkContextLock benchmarks LockContext operations
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

// runBenchmarkContextRLock benchmarks RLockContext operations
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
	s := &Semaphore{}
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
	s := &Semaphore{}
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
	s := &Semaphore{}
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
	s := &Semaphore{}
	runBenchmarkRWLock(b, s)
}

func BenchmarkRWLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkRWLock(b, &mu)
}

// Read-only benchmarks
func BenchmarkReadOnly_Semaphore(b *testing.B) {
	s := &Semaphore{}
	runBenchmarkReadOnly(b, s)
}

func BenchmarkReadOnly_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkReadOnly(b, &mu)
}

// TryLock benchmarks
func BenchmarkTryLock_Semaphore(b *testing.B) {
	s := &Semaphore{}
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
	s := &Semaphore{}
	runBenchmarkTryRLock(b, s)
}

func BenchmarkTryRLock_StdRWMutex(b *testing.B) {
	var mu sync.RWMutex
	runBenchmarkTryRLock(b, &mu)
}

// Context lock benchmarks
func BenchmarkContextLock_Semaphore(b *testing.B) {
	s := &Semaphore{}
	runBenchmarkContextLock(b, s)
}

func BenchmarkContextRLock_Semaphore(b *testing.B) {
	s := &Semaphore{}
	runBenchmarkContextRLock(b, s)
}
