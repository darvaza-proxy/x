package spinlock

import (
	"context"
	"runtime"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// CondFunc is a function type used to define a condition for waiting on a Cond value.
// It takes an int32 value and returns a boolean indicating whether the condition is met.
type CondFunc func(int32) bool

// Cond is a synchronization primitive that allows atomic operations and waiting
// on an int32 value until a condition is met.
// It implements a spinlock pattern with atomic operations rather than
// traditional blocking mechanisms.
type Cond int32

// NewCond creates a new Cond with the specified initial value.
// It returns a pointer to the created Cond.
func NewCond(initialValue int) *Cond {
	v := int32(initialValue)
	return (*Cond)(&v)
}

// Add atomically adds the given integer to the Cond value and returns the new value.
// It panics if the receiver is nil.
func (c *Cond) Add(n int) int32 {
	v, err := c.checkAndAdd(n)
	if err != nil {
		core.Panic(err)
	}
	return v
}

func (c *Cond) checkAndAdd(n int) (int32, error) {
	ptr, err := c.check()
	switch {
	case err != nil:
		return 0, err
	case n == 0:
		return atomic.LoadInt32(ptr), nil
	default:
		return atomic.AddInt32(ptr, int32(n)), nil
	}
}

// Inc atomically increments the Cond value by 1 and returns the new value.
// It panics if the receiver is nil.
func (c *Cond) Inc() int32 {
	v, err := c.checkAndAdd(1)
	if err != nil {
		core.Panic(err)
	}
	return v
}

// Dec atomically decrements the Cond value by 1 and returns the new value.
// It panics if the receiver is nil.
func (c *Cond) Dec() int32 {
	v, err := c.checkAndAdd(-1)
	if err != nil {
		core.Panic(err)
	}
	return v
}

// Value atomically returns the current value of the Cond.
// It panics if the receiver is nil.
func (c *Cond) Value() int32 {
	ptr, err := c.check()
	if err != nil {
		core.Panic(err)
	}

	return atomic.LoadInt32(ptr)
}

// WaitZero waits until the Cond value becomes zero.
// It blocks the current goroutine until the condition is met.
// It panics if the receiver is nil.
// This method uses a spin-waiting pattern instead of blocking.
func (c *Cond) WaitZero() {
	ptr, err := c.check()
	if err != nil {
		core.Panic(err)
	}

	c.doWait(nil, ptr, nil)
}

// WaitFn waits until the provided condition function returns true.
// It blocks the current goroutine until the condition is met.
// If no condition function is provided, it waits until the Cond value becomes zero.
// It panics if the receiver is nil.
// This method uses a spin-waiting pattern instead of blocking.
func (c *Cond) WaitFn(until CondFunc) {
	ptr, err := c.check()
	if err != nil {
		core.Panic(err)
	}
	c.doWait(nil, ptr, until)
}

// WaitFnAbort waits until the provided condition function returns true or the abort channel is closed.
// It returns an error if the Cond is invalid, returns context.Canceled if aborted, or nil if the condition is met.
// This provides a cancellable wait mechanism using channels.
func (c *Cond) WaitFnAbort(abort <-chan struct{}, until CondFunc) error {
	ptr, err := c.check()
	switch {
	case err != nil:
		// invalid
		return err
	case c.doWait(abort, ptr, until):
		// aborted
		return context.Canceled
	default:
		// success
		return nil
	}
}

// WaitFnContext waits until the provided condition function returns true or the context is canceled.
// It returns an error if the Cond is invalid, returns the context's error if canceled, or nil if the condition is met.
// This provides a context-aware wait mechanism for better integration with context-based cancellation.
func (c *Cond) WaitFnContext(ctx context.Context, until CondFunc) error {
	ptr, err := c.check()
	switch {
	case err != nil:
		// invalid
		return err
	case ctx == nil:
		// required
		return errors.ErrNilContext
	case c.doWait(ctx.Done(), ptr, until):
		// aborted
		return ctx.Err()
	default:
		// success
		return nil
	}
}

// Is tests the value against the given function and passes through the result.
// If no condition function is provided, it tests if the Cond value is zero.
// It panics if the receiver is nil.
// Unlike Wait methods, this is a non-blocking check.
func (c *Cond) Is(cond CondFunc) bool {
	ptr, err := c.check()
	if err != nil {
		core.Panic(err)
	}

	return c.match(ptr, cond)
}

// IsZero checks if the Cond value is zero.
// It returns true if the value is zero, false otherwise.
// It panics if the receiver is nil.
// This is a convenience method equivalent to Is(func(v int32) bool { return v == 0 }).
func (c *Cond) IsZero() bool {
	ptr, err := c.check()
	if err != nil {
		core.Panic(err)
	}

	return c.match(ptr, nil)
}

// match checks if the current value satisfies the provided condition.
// If cond is nil, it checks if the value is zero.
// This is a helper function used internally by Is and IsZero.
func (*Cond) match(ptr *int32, cond CondFunc) bool {
	v := atomic.LoadInt32(ptr)
	if cond == nil {
		return v == 0
	}
	return cond(v)
}

// doWait implements the waiting logic for all Wait methods.
// It repeatedly checks the condition, yielding the processor between checks.
// It returns true if the wait was aborted via the abort channel, false otherwise.
// This uses a spin-wait pattern with runtime.Gosched() rather than blocking.
func (c *Cond) doWait(abort <-chan struct{}, ptr *int32, until CondFunc) bool {
	if cancelled(abort) {
		return true
	}

	for !c.match(ptr, until) {
		if cancelled(abort) {
			return true
		}
		runtime.Gosched()
	}

	return false
}

// check validates that the receiver is not nil and returns a usable pointer.
// Returns a pointer to the underlying int32 and an error if validation fails.
func (c *Cond) check() (*int32, error) {
	switch {
	case c == nil:
		return nil, errors.ErrNilReceiver
	default:
		return (*int32)(c), nil
	}
}

// cancelled checks if the abort channel is closed without blocking.
// It returns true if the channel is closed, false otherwise.
// A nil channel will be ignored and false returned.
// This is a helper function used in the doWait method.
func cancelled(abort <-chan struct{}) bool {
	select {
	case <-abort:
		return true
	default:
		return false
	}
}
