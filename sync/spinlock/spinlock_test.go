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

	// Should start unlocked
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))

	// Lock
	sl.Lock()
	assert.Equal(t, uint32(1), atomic.LoadUint32((*uint32)(&sl)))

	// Unlock
	sl.Unlock()
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))
}

func TestSpinLock_TryLock(t *testing.T) {
	var sl SpinLock

	// When unlocked, TryLock should succeed
	assert.True(t, sl.TryLock())
	assert.Equal(t, uint32(1), atomic.LoadUint32((*uint32)(&sl)))

	// When locked, TryLock should fail
	assert.False(t, sl.TryLock())

	// After unlock, TryLock should succeed again
	sl.Unlock()
	assert.True(t, sl.TryLock())
	sl.Unlock()
}

func TestSpinLock_Concurrent(t *testing.T) {
	var sl SpinLock
	var counter int32

	// Test with multiple goroutines
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

	// Ensure all operations were performed correctly
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
					// Simulate some work
					time.Sleep(10 * time.Microsecond)
					sl.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// The counter should be less than attempts due to TryLock failures
	assert.Less(t, counter, attempts)
	// But we should have had some successful acquisitions
	assert.Greater(t, counter, int32(0))
}

func TestSpinLock_NilReceiver(t *testing.T) {
	var sl *SpinLock

	// Lock on nil receiver should panic
	assert.Panics(t, func() {
		sl.Lock()
	})

	// TryLock on nil receiver should panic
	assert.Panics(t, func() {
		sl.TryLock()
	})

	// Unlock on nil receiver should panic
	assert.Panics(t, func() {
		sl.Unlock()
	})
}

func TestSpinLock_UnlockUnlockedSpinLock(t *testing.T) {
	var sl SpinLock

	// Unlocking an unlocked spinlock should panic
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

	// The spinlock should be unlocked after the deferred Unlock
	assert.True(t, executed)
	assert.Equal(t, uint32(0), atomic.LoadUint32((*uint32)(&sl)))

	// We should be able to lock again
	assert.True(t, sl.TryLock())
	sl.Unlock()
}

func TestSpinLock_Interfaces(t *testing.T) {
	// Test that SpinLock properly implements required interfaces
	var sl SpinLock

	var locker sync.Locker = &sl

	// Use the locker through the interface
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
