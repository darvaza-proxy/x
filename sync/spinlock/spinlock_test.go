package spinlock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func doSomethingBriefly() {
	time.Sleep(10 * time.Millisecond)
}

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

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				sl.Lock()
				atomic.AddInt32(&counter, 1)
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

	const numGoroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				atomic.AddInt32(&attempts, 1)
				if sl.TryLock() {
					atomic.AddInt32(&counter, 1)
					// Brief work simulation
					time.Sleep(10 * time.Microsecond)
					sl.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Due to contention, successful locks should be fewer than attempts
	assert.Less(t, counter, attempts)
	// Some attempts should succeed
	assert.Greater(t, counter, int32(0))
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

func BenchmarkSpinLock(b *testing.B) {
	var sl SpinLock

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sl.Lock()
			doSomethingBriefly()
			sl.Unlock()
		}
	})
}

func BenchmarkSpinLockWithDefer(b *testing.B) {
	var sl SpinLock

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			func() {
				sl.Lock()
				defer sl.Unlock()
			}()
		}
	})
}

func BenchmarkSpinLockTryLock(b *testing.B) {
	var sl SpinLock

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if sl.TryLock() {
				sl.Unlock()
			}
		}
	})
}
