package cond

// This file includes a condition variable implementation (Count) that
// allows goroutines to coordinate and wait for specific conditions to be met
// on a shared atomic counter.

import (
	"context"
	"runtime"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// Count is a synchronisation primitive that allows atomic operations and waiting
// on an int32 value until a condition is met. It combines features of a
// condition variable and an atomic counter, allowing goroutines to coordinate
// based on a numeric value and user-defined conditions.
//
// All methods are safe for concurrent use from multiple goroutines.
//
// It's based on Roberto Clapis' series on [advanced concurrency patterns].
//
// [advanced concurrency patterns]: https://blogtitle.github.io/go-advanced-concurrency-patterns-part-3-channels/
type Count struct {
	v int32
	b Barrier

	m func(int32) bool
}

// NewCount creates a new Count with an initial value and optional broadcast
// conditions. It initialises the Count with the given value and optional
// matching functions. If initialisation fails, it panics with the encountered
// error, though no errors are anticipated.
// If no matching functions are provided, every change of value broadcasts.
// The returned Count is ready for use in concurrent synchronisation scenarios
// and it's self closing.
func NewCount(initialValue int, broadcast ...func(int32) bool) *Count {
	c := new(Count)
	if err := c.doInit(initialValue, broadcast); err != nil {
		core.Panic(err)
	}

	runtime.SetFinalizer(c, func(c *Count) {
		_ = c.b.Close()
	})

	return c
}

// IsNil reports whether the Count is nil or its underlying Barrier is nil or
// uninitialised. It provides a safe way to check the initialisation status of
// a Count instance.
func (c *Count) IsNil() bool {
	if c == nil {
		return true
	}
	return c.b.IsNil()
}

// IsClosed reports whether the Count is closed and no longer usable for
// synchronisation. It returns true if the Count is nil or its underlying
// Barrier is closed.
func (c *Count) IsClosed() bool {
	if c == nil {
		return true
	}
	return c.b.IsClosed()
}

// check validates the Count instance, ensuring it is not nil and has been
// properly initialised. It returns an error if the Count is nil or its
// underlying Barrier is not initialised.
func (c *Count) check() error {
	switch {
	case c == nil:
		return errors.ErrNilReceiver
	case c.b.IsNil():
		return errors.ErrNotInitialised
	default:
		return nil
	}
}

// Init initialises the Count with a given initial value and optional broadcast
// conditions. It sets up the Count for use in concurrent synchronisation
// scenarios. Returns an error if the receiver is nil or already initialised.
// Close needs to be called at the end to clear the internal channel.
func (c *Count) Init(initialValue int, broadcast ...func(int32) bool) error {
	if c == nil {
		return errors.ErrNilReceiver
	}

	return c.doInit(initialValue, broadcast)
}

func (c *Count) doInit(initialValue int, broadcast []func(int32) bool) error {
	// barrier
	if err := c.b.Init(); err != nil {
		return err
	}

	c.v = int32(initialValue)
	c.m = makeAnyMatch(broadcast)
	return nil
}

// Reset sets the Count to a new value, validating the Count instance before resetting.
// It returns an error if the Count is not properly initialised or is nil.
func (c *Count) Reset(value int) error {
	if err := c.check(); err != nil {
		return err
	}

	return c.doReset(value)
}

func (c *Count) doReset(value int) error {
	b, ok := <-c.b.Acquire()
	if !ok {
		return errors.ErrClosed
	}

	// set new value
	atomic.StoreInt32(&c.v, int32(value))
	// set new barrier
	c.b.Release(make(Token))

	// and signal all waiters
	close(b)
	return nil
}

// Close releases the resources associated with the Count.
// It returns an error if the receiver is nil.
func (c *Count) Close() error {
	if c == nil {
		return errors.ErrNilReceiver
	}

	return c.b.Close()
}

// Add atomically adds the given integer to the Cond value and returns the new
// value. It notifies all waiters about the state change through a broadcast
// unless custom conditions were specified during initialisation.
// If n is 0, it simply returns the current value without broadcasting.
func (c *Count) Add(n int) int {
	if n == 0 {
		return int(atomic.LoadInt32(&c.v))
	}

	return c.doAdd(n)
}

// doAdd atomically adds n to the value and broadcasts to all waiters
// unless custom conditions were specified during initialisation.
// Returns the new value after the addition operation.
func (c *Count) doAdd(n int) int {
	v := atomic.AddInt32(&c.v, int32(n))
	if c.m(v) {
		// wake all waiters
		c.b.Broadcast()
	}
	return int(v)
}

// Inc atomically increments the Cond value by 1 and returns the new value.
// It notifies all waiters unless custom conditions were specified during
// initialisation.
func (c *Count) Inc() int {
	return c.doAdd(1)
}

// Dec atomically decrements the Cond value by 1 and returns the new value.
// It notifies all waiters about the state change unless custom conditions
// were specified during initialisation.
func (c *Count) Dec() int {
	return c.doAdd(-1)
}

// Value atomically returns the current value of the Cond.
// This operation does not affect waiters.
func (c *Count) Value() int {
	return int(atomic.LoadInt32(&c.v))
}

// WaitFnAbort blocks until the condition function returns true or the abort
// channel is closed. Returns context.Canceled if aborted, or nil if the
// condition is met. If until is nil, waits for the value to become zero.
func (c *Count) WaitFnAbort(abort <-chan struct{}, until func(int32) bool) error {
	err := c.check()
	switch {
	case err != nil:
		return err
	case c.doWaitFn(abort, until).IsCancelled():
		return context.Canceled
	default:
		return nil
	}
}

func (c *Count) doWaitFn(abort <-chan struct{}, until func(int32) bool) waitResult {
	switch {
	case isCancelled(abort):
		return waitCancelled
	case c.doMatch(until):
		return waitSuccess
	default:
		return c.doWaitFnLoop(abort, until)
	}
}

func (c *Count) doWaitFnLoop(abort <-chan struct{}, until func(int32) bool) waitResult {
	var b Token
	var res waitResult

	for res.IsContinue() {
		if b == nil {
			b, res = c.doWaitFnPass1(abort, until)
		} else {
			res = c.doWaitFnPass2(abort, until, b)
			b = nil
		}
	}

	return res
}

func (c *Count) doWaitFnPass1(abort <-chan struct{}, until func(int32) bool) (Token, waitResult) {
	// acquire fresh barrier
	select {
	case b, ok := <-c.b.Acquire():
		var res waitResult

		switch {
		case !ok, isCancelled(abort):
			res = waitCancelled
		case c.doMatch(until):
			res = waitSuccess
		default:
			res = waitContinue
		}
		c.b.Release(b)

		if res.IsContinue() {
			return b, waitContinue
		}

		// done
		return nil, res
	case <-abort:
		return nil, waitCancelled
	}
}

func (c *Count) doWaitFnPass2(abort <-chan struct{}, until func(int32) bool, b Token) waitResult {
	// wait to be signaled
	select {
	case <-b.Signaled():
		// and check without delay
		if c.doMatch(until) {
			return waitSuccess
		}
		// try again later.
		return waitContinue
	case <-abort:
		return waitCancelled
	}
}

// WaitFnContext blocks until the condition function returns true or the context
// is cancelled. Returns the context's error if cancelled, or nil if the
// condition is met. If until is nil, waits for the value to become zero.
// Returns utils.ErrNilContext if the provided context is nil.
func (c *Count) WaitFnContext(ctx context.Context, until func(int32) bool) error {
	err := c.check()
	switch {
	case err != nil:
		return err
	case ctx == nil:
		return errors.ErrNilContext
	case c.doWaitFn(ctx.Done(), until).IsCancelled():
		return ctx.Err()
	default:
		return nil
	}
}

// WaitFn blocks the calling goroutine until the provided condition function
// returns true. If no condition function is provided (nil), it waits until
// the Count value becomes zero. Panics if the receiver is nil or
// uninitialised.
func (c *Count) WaitFn(until func(int32) bool) {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.doWaitFn(nil, until)
}

// Wait blocks the calling goroutine until the Cond value becomes zero.
// Panics if the receiver is nil or uninitialised.
func (c *Count) Wait() {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.doWaitFn(nil, nil)
}

// Match tests the value against the given condition function and returns
// the result. If no condition function is provided, it tests if the Cond
// value is zero. Panics if the receiver is nil or uninitialised.
func (c *Count) Match(fn func(int32) bool) bool {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doMatch(fn)
}

// doMatch implements the condition check logic.
// If until is nil, it checks if the value is zero.
func (c *Count) doMatch(fn func(int32) bool) bool {
	v := atomic.LoadInt32(&c.v)
	if fn == nil {
		return v == 0
	}
	return fn(v)
}

// IsZero checks if the Count value is zero.
// Panics if the receiver is nil or uninitialised.
func (c *Count) IsZero() bool {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doMatch(nil)
}

// Signal notifies a single waiter about a state change.
// The waiter is selected at random from the pool of waiting goroutines.
// Panics if the receiver is nil or uninitialised.
func (c *Count) Signal() bool {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.b.Signal()
}

// Broadcast notifies all waiters about a state change.
// Panics if the receiver is nil or uninitialised.
func (c *Count) Broadcast() {
	if err := c.check(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.b.Broadcast()
}

// waitResult represents the outcome of a wait operation.
type waitResult int

const (
	waitContinue  waitResult = iota
	waitSuccess              // Counter matched the condition
	waitCancelled            // Wait was cancelled
)

// IsCancelled returns true if the wait operation was cancelled.
func (r waitResult) IsCancelled() bool {
	return r == waitCancelled
}

// IsContinue returns true if the wait result indicates a
// we should try again.
func (r waitResult) IsContinue() bool {
	return r == waitContinue
}
