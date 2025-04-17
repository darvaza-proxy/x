package spinlock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/x/sync/mutex"

	"github.com/stretchr/testify/assert"
)

func TestSpinLock_Basic(t *testing.T) {
	var sl SpinLock

	// At creation, spinlock should be unlocked
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))

	// Lock acquisition
	sl.Lock()
	assert.Equal(t, uint32(1), atomic.LoadUint32((*uint32)(&sl)))

	// Lock release
	sl.Unlock()
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))
}

func TestSpinLock_TryLock(t *testing.T) {
	var sl SpinLock

	// TryLock should succeed when spinlock is free
	assert.True(t, sl.TryLock())
	assert.Equal(t, uint32(1), atomic.LoadUint32((*uint32)(&sl)))

	// TryLock should fail when spinlock is already held
	assert.False(t, sl.TryLock())

	// After releasing, TryLock should succeed again
	sl.Unlock()
	assert.True(t, sl.TryLock())
	sl.Unlock()
}

func TestSpinLock_Concurrent(t *testing.T) {
	var sl SpinLock
	var counter int32

	// Test with multiple concurrent goroutines
	const numGoroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
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

	// Verify counter matches expected value
	assert.Equal(t, int32(numGoroutines*iterations), counter)
}

//revive:disable-next-line:cognitive-complexity
func TestSpinLock_TryLockConcurrent(t *testing.T) {
	var sl SpinLock
	var counter int32
	var attempts int32

	// Track goroutines currently in critical section and maximum seen
	var currentGoroutines int32
	var maxGoroutines int32

	const numGoroutines = 100
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()

			// Increment count of goroutines, and update maximum if needed
			goroutinesNow := atomic.AddInt32(&currentGoroutines, 1)
			atomicUpdateMax(&maxGoroutines, goroutinesNow)
			defer atomic.AddInt32(&currentGoroutines, -1)

			for range iterations {
				atomic.AddInt32(&attempts, 1)
				if sl.TryLock() {
					counter++
					sl.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	t.Logf("Max goroutines in critical section: %d", maxGoroutines)
	t.Logf("Counter: %d, Attempts: %d", counter, attempts)

	// If maxGoroutines never exceeded 1, then counter should equal attempts
	if maxGoroutines <= 1 {
		assert.Equal(t, attempts, counter, "Only one goroutine happened at the time, all attempts should succeed")
	} else {
		// Due to contention, successful locks should be fewer than attempts
		assert.Less(t, counter, attempts, "With contention, successful locks should be fewer than attempts")
		// Some attempts should succeed
		assert.Greater(t, counter, int32(0), "Some lock attempts should succeed")
	}
}

// atomicUpdateMax atomically updates *ptr with val if val is greater than the current value
func atomicUpdateMax(ptr *int32, val int32) {
	for {
		currentMax := atomic.LoadInt32(ptr)
		if val <= currentMax || atomic.CompareAndSwapInt32(ptr, currentMax, val) {
			break
		}
	}
}

func TestSpinLock_NilReceiver(t *testing.T) {
	var sl *SpinLock

	// All operations on nil receiver should panic
	assert.Panics(t, func() {
		sl.Lock()
	})

	assert.Panics(t, func() {
		sl.TryLock()
	})

	assert.Panics(t, func() {
		sl.Unlock()
	})
}

func TestSpinLock_UnlockUnlockedSpinLock(t *testing.T) {
	var sl SpinLock

	// Unlocking an already unlocked spinlock should panic
	assert.Panics(t, func() {
		sl.Unlock()
	})
}

func TestSpinLock_LockDefer(t *testing.T) {
	var sl SpinLock
	var executed bool

	func() {
		sl.Lock()
		defer sl.Unlock()

		executed = true
	}()

	// Spinlock should be released after deferred Unlock
	assert.True(t, executed)
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))

	// Lock should be acquirable again
	assert.True(t, sl.TryLock())
	sl.Unlock()
}

func TestSpinLock_Interfaces(t *testing.T) {
	// Verify SpinLock implements sync.Locker interface
	var sl SpinLock

	var locker sync.Locker = &sl

	// Use through the sync.Locker interface
	locker.Lock()
	assert.Equal(t, uint32(1), atomic.LoadUint32((*uint32)(&sl)))
	locker.Unlock()
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))
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

// runBenchmarkTryLock benchmarks TryLock success
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

// ----- Benchmark implementations -----

// Basic lock benchmarks
func BenchmarkLock_SpinLock(b *testing.B) {
	var sl SpinLock
	runBenchmarkBasicLock(b, &sl)
}

func BenchmarkLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkBasicLock(b, &mu)
}

// Deferred unlock benchmarks
func BenchmarkLockWithDefer_SpinLock(b *testing.B) {
	var sl SpinLock
	runBenchmarkLockWithDefer(b, &sl)
}

func BenchmarkLockWithDefer_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkLockWithDefer(b, &mu)
}

// Contention benchmarks
func BenchmarkContention_SpinLock(b *testing.B) {
	var sl SpinLock
	runBenchmarkContention(b, &sl)
}

func BenchmarkContention_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkContention(b, &mu)
}

// TryLock benchmark with retry
func BenchmarkRetryLock_SpinLock(b *testing.B) {
	var sl SpinLock
	runBenchmarkRetryLock(b, &sl)
}

func BenchmarkRetryLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkRetryLock(b, &mu)
}

// TryLock benchmark with target counter
func BenchmarkTryLock_SpinLock(b *testing.B) {
	var sl SpinLock
	runBenchmarkTryLock(b, &sl)
}

func BenchmarkTryLock_StdMutex(b *testing.B) {
	var mu sync.Mutex
	runBenchmarkTryLock(b, &mu)
}
