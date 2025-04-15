// Package spinlock provides a lightweight, efficient spinlock implementation
// for low-contention mutual exclusion scenarios.
//
// Spinlocks are suitable for protecting shared resources that are accessed for
// very short periods. Unlike traditional mutexes that put goroutines to
// sleep during contention, spinlocks actively spin (busy-wait), which can
// be more efficient when the expected wait time is brief.
//
// Usage considerations:
//   - Best for rare lock contention and minimal lock-holding duration
//   - Avoid for long-held locks as they waste CPU cycles during spinning
//   - Ideal for performance-critical paths where context switching is costly
//
// Example usage:
//
//	var lock spinlock.SpinLock
//
//	// In concurrent code:
//	lock.Lock()
//	defer lock.Unlock()
//	// Critical section here
package spinlock

import (
	"runtime"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/mutex"
)

// SpinLock is a lightweight mutual exclusion primitive that uses spinning
// instead of sleeping for efficiency in low-contention scenarios.
// The zero value is an unlocked spinlock.
//
// SpinLock implements both sync.Locker and mutex.Mutex interfaces.
// Unlike sync.Mutex, SpinLock consumes CPU cycles when waiting to
// acquire the lock, making it suitable for protecting resources
// that are held briefly.
type SpinLock uint32

func (sl *SpinLock) ptr() *uint32 {
	if sl == nil {
		return nil
	}
	return (*uint32)(sl)
}

// Lock acquires the spinlock, blocking until the lock is available.
// If the receiver is nil, it will panic with errors.ErrNilReceiver.
//
// When waiting, Lock yields the processor using runtime.Gosched()
// to allow other goroutines to execute, but does not put the
// calling goroutine to sleep.
func (sl *SpinLock) Lock() {
	if err := sl.doLock(); err != nil {
		core.Panic(err)
	}
}

func (sl *SpinLock) doLock() error {
	ptr := sl.ptr()
	if ptr == nil {
		return errors.ErrNilReceiver
	}

	for !atomic.CompareAndSwapUint32(ptr, 0, 1) {
		runtime.Gosched()
	}

	return nil
}

// TryLock attempts to acquire the spinlock without blocking.
// Returns true if the lock was successfully acquired, false otherwise.
// If the receiver is nil, it will panic with core.ErrNilReceiver.
//
// Useful in non-blocking scenarios where the caller can perform
// alternative actions if the lock isn't immediately available.
func (sl *SpinLock) TryLock() bool {
	ok, err := sl.doTryLock()
	if err != nil {
		core.Panic(err)
	}
	return ok
}

func (sl *SpinLock) doTryLock() (bool, error) {
	ptr := sl.ptr()
	switch {
	case ptr == nil:
		return false, core.ErrNilReceiver
	case atomic.CompareAndSwapUint32(ptr, 0, 1):
		return true, nil
	default:
		return false, nil
	}
}

// Unlock releases the spinlock.
// If the receiver is nil, it will panic with core.ErrNilReceiver.
// If the spinlock is not currently locked, it will panic with
// an "unlock of unlocked spinlock" error.
func (sl *SpinLock) Unlock() {
	if err := sl.doUnlock(); err != nil {
		core.Panic(err)
	}
}

func (sl *SpinLock) doUnlock() error {
	ptr := sl.ptr()
	switch {
	case ptr == nil:
		return errors.ErrNilReceiver
	case !atomic.CompareAndSwapUint32(ptr, 1, 0):
		return core.NewPanicError(1, "unlock of unlocked spinlock")
	default:
		return nil
	}
}

var (
	_ sync.Locker = (*SpinLock)(nil)
	_ mutex.Mutex = (*SpinLock)(nil)
)
