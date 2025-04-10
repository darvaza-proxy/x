package semaphore

//revive:disable:cyclomatic
//revive:disable:cognitive-complexity

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// doSomethingBriefly simulates a short operation by sleeping for 10ms.
func doSomethingBriefly() {
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
		doSomethingBriefly()
		s.Unlock()
	})

	t.Run("sequential locks", func(_ *testing.T) {
		s := &Semaphore{}
		for i := 0; i < 5; i++ {
			s.Lock()
			doSomethingBriefly()
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
		doSomethingBriefly()
		s.RUnlock()
	})

	t.Run("multiple readers", func(_ *testing.T) {
		s := &Semaphore{}

		// Acquire three read locks
		s.RLock()
		s.RLock()
		s.RLock()

		doSomethingBriefly()

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
			doSomethingBriefly()
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

func TestSemaphore_Concurrency(t *testing.T) {
	t.Run("multiple writers", func(t *testing.T) {
		s := &Semaphore{}
		count := 0
		const numGoroutines = 10
		const iterations = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
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
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					s.RLock()
					// Just read the counter, don't modify
					_ = counter
					time.Sleep(1 * time.Millisecond)
					s.RUnlock()
				}
			}()
		}

		// Start writers
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
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

func TestSemaphore_Acquire_Release(t *testing.T) {
	t.Run("acquire and release", func(t *testing.T) {
		s := &Semaphore{}
		ctx := context.Background()

		err := s.Acquire(ctx)
		require.NoError(t, err)

		err = s.Release()
		require.NoError(t, err)
	})

	t.Run("release without acquire should panic", func(t *testing.T) {
		s := &Semaphore{}
		assert.Panics(t, func() {
			_ = s.Release()
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
		doSomethingBriefly()
		locker.Unlock()
	})
}

func TestSemaphore_EdgeCases(t *testing.T) {
	t.Run("writer after readers", func(_ *testing.T) {
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

		// Writer should still be blocked
		select {
		case <-writerDone:
			t.Fatal("Writer acquired lock while readers held it")
		default:
			// Expected behaviour
		}

		// Release all read locks
		s.RUnlock()
		s.RUnlock()
		s.RUnlock()

		// Now writer should be able to proceed
		select {
		case <-writerDone:
			// Expected behaviour
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Writer failed to acquire lock after all readers released")
		}
	})

	t.Run("readers after writer", func(t *testing.T) {
		s := &Semaphore{}
		s.Lock()

		done := make(chan struct{})
		ready := make(chan struct{})

		// Readers should block until writer is done
		for i := 0; i < 3; i++ {
			go func() {
				ready <- struct{}{}
				s.RLock()
				defer s.RUnlock()
				done <- struct{}{}
			}()
		}

		// Wait for all readers to be ready
		for i := 0; i < 3; i++ {
			<-ready
		}

		time.Sleep(10 * time.Millisecond)

		// No readers should have acquired the lock
		select {
		case <-done:
			t.Fatal("Reader acquired lock while writer held it")
		default:
			// Expected behaviour
		}

		// Release the write lock
		s.Unlock()

		// Now all readers should proceed
		timeout := time.After(100 * time.Millisecond)
		for i := 0; i < 3; i++ {
			select {
			case <-done:
				// Expected behaviour
			case <-timeout:
				t.Fatal("Readers failed to acquire lock after writer released")
			}
		}
	})
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

		for i := 0; i < numWorkers; i++ {
			go func(id int) {
				defer wg.Done()

				// Each goroutine has a different pattern of operations
				for j := 0; j < iterations; j++ {
					switch (id + j) % 7 {
					case 0: // Lock + Unlock
						s.Lock()
						counterMutex.Lock()
						counter++
						counterMutex.Unlock()
						s.Unlock()
					case 1: // TryLock
						if s.TryLock() {
							counterMutex.Lock()
							counter++
							counterMutex.Unlock()
							s.Unlock()
						}
					case 2: // RLock + RUnlock
						s.RLock()
						// Just read the counter
						counterMutex.Lock()
						_ = counter
						counterMutex.Unlock()
						s.RUnlock()
					case 3: // TryRLock
						if s.TryRLock() {
							counterMutex.Lock()
							_ = counter
							counterMutex.Unlock()
							s.RUnlock()
						}
					case 4: // LockContext with short timeout
						ctx, cancel := context.WithTimeout(
							context.Background(),
							1*time.Millisecond,
						)
						if s.LockContext(ctx) == nil {
							counterMutex.Lock()
							counter++
							counterMutex.Unlock()
							s.Unlock()
						}
						cancel()
					case 5: // RLockContext with short timeout
						ctx, cancel := context.WithTimeout(
							context.Background(),
							1*time.Millisecond,
						)
						if s.RLockContext(ctx) == nil {
							counterMutex.Lock()
							_ = counter
							counterMutex.Unlock()
							s.RUnlock()
						}
						cancel()
					case 6: // Acquire and Release
						ctx := context.Background()
						if s.Acquire(ctx) == nil {
							counterMutex.Lock()
							counter++
							counterMutex.Unlock()
							_ = s.Release()
						}
					}
				}
			}(i)
		}

		wg.Wait()

		// We don't know exactly what the counter will be due to TryLock/TryRLock
		// and timeouts, but it should be > 0
		assert.Greater(t, counter, 0)
	})
}

func TestSemaphore_BoundaryConditions(t *testing.T) {
	t.Run("many consecutive read locks", func(t *testing.T) {
		s := &Semaphore{}

		// Acquire a lot of read locks
		const numLocks = 1000
		for i := 0; i < numLocks; i++ {
			s.RLock()
		}

		// Try to get a write lock - should block
		writerDone := make(chan struct{})
		go func() {
			s.Lock()
			doSomethingBriefly()
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
		for i := 0; i < numLocks; i++ {
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

// TestSemaphore_WriterStarvationPrevention verifies that writers aren't starved
// when there's a continuous stream of reader lock requests.
func TestSemaphore_WriterStarvationPrevention(t *testing.T) {
	s := &Semaphore{}
	const numReaders = 10

	// Channel to coordinate test steps
	writerAcquired := make(chan struct{})

	// Start a writer that will try to acquire the lock
	go func() {
		// Signal we want to write
		writerReady := make(chan struct{})
		go func() {
			close(writerReady)
			s.Lock()
			close(writerAcquired)

			// Hold the lock for a moment
			time.Sleep(10 * time.Millisecond)
			s.Unlock()
		}()

		// Wait until writer is ready to attempt acquisition
		<-writerReady

		// Give time for writer to register its intent
		time.Sleep(20 * time.Millisecond)
	}()

	// Start a continuous stream of readers
	readersStarted := &sync.WaitGroup{}
	readersStarted.Add(numReaders)

	for range numReaders {
		go func() {
			readersStarted.Done() // Signal this reader is ready

			// Start acquiring read locks in a loop
			for {
				select {
				case <-writerAcquired:
					// Writer got the lock, we can stop
					return
				default:
					// Try to get a read lock
					if s.TryRLock() {
						// Hold briefly then release
						time.Sleep(5 * time.Millisecond)
						s.RUnlock()
					}
					// Small pause to prevent CPU spinning
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}

	// Wait for all readers to start
	readersStarted.Wait()

	// Writer should eventually acquire the lock despite continuous readers
	select {
	case <-writerAcquired:
		// Success - writer wasn't starved
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Writer appears to be starved by readers")
	}
}

// TestSemaphore_ReaderPreferenceWithoutWriters verifies that readers can
// acquire the lock freely when no writers are waiting.
func TestSemaphore_ReaderPreferenceWithoutWriters(t *testing.T) {
	s := &Semaphore{}
	const numReaders = 5

	// First, acquire a read lock
	s.RLock()

	// Now try to acquire more read locks in parallel
	var wg sync.WaitGroup
	wg.Add(numReaders)

	success := make([]bool, numReaders)

	for i := 0; i < numReaders; i++ {
		go func(id int) {
			defer wg.Done()
			// With no writers waiting, TryRLock should succeed immediately
			success[id] = s.TryRLock()
			if success[id] {
				s.RUnlock()
			}
		}(i)
	}

	wg.Wait()
	s.RUnlock() // Release the initial read lock

	// All readers should have succeeded
	for i := 0; i < numReaders; i++ {
		assert.True(t, success[i], "Reader %d failed to acquire lock when no writers were waiting", i)
	}
}
