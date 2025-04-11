// Package waitgroup provides a lightweight alternative to the standard
// sync.WaitGroup using a condition variable from the semaphore package.
// This implementation offers equivalent functionality with potentially
// fewer resources.
//
// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of goroutines to wait for.
// Then each goroutine runs and calls Done when finished.
// At the same time, Wait can be used to block until all goroutines complete.
package waitgroup

import (
	"fmt"

	"darvaza.org/core"
	"darvaza.org/x/sync/semaphore"
)

// WaitGroup implements a lightweight alternative to the standard
// sync.WaitGroup using a [semaphore.Cond].
//
// A WaitGroup must not be copied after first use. Create a new instance for
// each independent synchronisation task.
type WaitGroup semaphore.Cond

// sys returns the underlying semaphore.Cond or nil if the WaitGroup is nil.
// This is an internal helper method for safer type conversion.
func (wg *WaitGroup) sys() *semaphore.Cond {
	if wg == nil {
		return nil
	}
	return (*semaphore.Cond)(wg)
}

// Count returns the current number of workers in the WaitGroup.
// If the WaitGroup is nil, it returns 0.
func (wg *WaitGroup) Count() int {
	cond := wg.sys()
	if cond == nil {
		return 0
	}
	return cond.Value()
}

// Add increments the WaitGroup counter by the specified number of workers.
// If the WaitGroup is nil, it panics with a nil receiver error.
//
// The counter must be greater than 0. Negative values will cause a panic.
// Zero values are allowed but have no effect.
func (wg *WaitGroup) Add(n int) {
	if err := wg.doAdd(n); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}
}

// doAdd is an internal helper that attempts to add n to the counter
// and returns an error instead of panicking.
func (wg *WaitGroup) doAdd(n int) error {
	cond := wg.sys()
	switch {
	case cond == nil:
		return core.ErrNilReceiver
	case n == 0:
		return nil
	case n <= 0:
		return fmt.Errorf("invalid count: %d", n)
	default:
		cond.Add(n)
		return nil
	}
}

// Done decrements the WaitGroup counter by one. If the counter becomes
// negative, it panics with an error indicating that Done was called without
// corresponding workers. If the WaitGroup is nil, it also panics with a nil
// receiver error.
//
// This method should be called exactly once per worker previously added with
// Add. Typically deferred at the start of a goroutine.
func (wg *WaitGroup) Done() {
	if err := wg.doDec(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}
}

// doDec is an internal helper that attempts to decrement the counter
// and returns an error instead of panicking.
func (wg *WaitGroup) doDec() error {
	if cond := wg.sys(); cond == nil {
		// invalid
		return core.ErrNilReceiver
	} else if n := cond.Dec(); n < 0 {
		// wtf. undo and panic
		cond.Inc()
		return fmt.Errorf("done called without workers")
	}

	return nil
}

// Wait blocks until the WaitGroup counter reaches zero, indicating all
// workers have completed. If the WaitGroup is nil, it panics with a nil
// receiver error.
//
// Once Wait returns, the WaitGroup is in a reset state and can be reused.
func (wg *WaitGroup) Wait() {
	cond := wg.sys()
	if cond == nil {
		core.Panic(core.NewPanicError(1, core.ErrNilReceiver))
	}

	cond.WaitZero()
}
