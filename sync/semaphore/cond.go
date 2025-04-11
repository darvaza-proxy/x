package semaphore

// This file includes a condition variable implementation (Cond) that
// allows goroutines to coordinate and wait for specific conditions to be met
// on a shared atomic counter.

import (
	"context"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/spinlock"
)

// CondFunc is a function type used to define a condition for waiting on a Cond
// value. It takes an int32 value and returns a boolean indicating whether the
// condition is met. When true, waiters will be unblocked.
type CondFunc func(int32) bool

// Cond is a synchronisation primitive that allows atomic operations and waiting
// on an int32 value until a condition is met. It combines features of a
// condition variable and an atomic counter, allowing goroutines to coordinate
// based on a numeric value and user-defined conditions.
//
// All methods are safe for concurrent use from multiple goroutines.
type Cond struct {
	// atomic value
	value int32

	// sl protects lazy initialisation
	sl spinlock.SpinLock
	// mu protects map access
	mu sync.Mutex

	// waiters map of channels to signal waiting goroutines
	waiters map[chan struct{}]struct{}
}

// NewCond creates a new Cond with the specified initial value.
// Returns a ready-to-use Cond instance with initialised internal structures.
func NewCond(initialValue int) *Cond {
	return &Cond{
		value:   int32(initialValue),
		waiters: make(map[chan struct{}]struct{}),
	}
}

// lazyInit initialises the Cond's internal structures if not yet initialised.
// Returns an error if the receiver is nil, otherwise ensures waiters map exists.
func (c *Cond) lazyInit() error {
	if c == nil {
		return errors.ErrNilReceiver
	}

	c.sl.Lock()
	if c.waiters == nil {
		c.waiters = make(map[chan struct{}]struct{})
	}
	c.sl.Unlock()

	return nil
}

// Add atomically adds the given integer to the Cond value and returns the new
// value. It notifies all waiters through a broadcast only when the counter
// becomes exactly zero after the addition. If n is 0, it simply returns the
// current value without broadcasting.
// Panics if the receiver is nil.
func (c *Cond) Add(n int) int {
	err := c.lazyInit()
	switch {
	case err != nil:
		core.Panic(core.NewPanicError(1, err))
		return 0 // unreachable
	case n == 0:
		return c.load()
	default:
		return c.doAdd(n)
	}
}

func (c *Cond) doAdd(n int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	v := c.add(n)
	if v == 0 {
		c.unsafeBroadcast()
	}
	return v
}

// add performs the atomic addition without broadcasting.
// Returns the new value after the addition operation.
//
//revive:disable-next-line:confusing-naming
func (c *Cond) add(n int) int {
	v := atomic.AddInt32(&c.value, int32(n))
	return int(v)
}

// Inc atomically increments the Cond value by 1 and returns the new value.
// It notifies all waiters only if the resulting value becomes zero.
// Panics if the receiver is nil.
func (c *Cond) Inc() int {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doAdd(1)
}

// Dec atomically decrements the Cond value by 1 and returns the new value.
// It notifies all waiters only if the resulting value becomes zero.
// Panics if the receiver is nil.
func (c *Cond) Dec() int {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doAdd(-1)
}

// Value atomically returns the current value of the Cond.
// This operation does not affect waiters.
// Panics if the receiver is nil.
func (c *Cond) Value() int {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.load()
}

// load returns the current atomic value without broadcasting.
func (c *Cond) load() int {
	v := atomic.LoadInt32(&c.value)
	return int(v)
}

// WaitZero blocks the calling goroutine until the Cond value becomes zero.
// Panics if the receiver is nil.
func (c *Cond) WaitZero() {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.doWait(nil, nil)
}

// WaitFn blocks the calling goroutine until the provided condition function
// returns true. If no condition function is provided (nil), it waits until
// the Cond value becomes zero. Panics if the receiver is nil.
func (c *Cond) WaitFn(until CondFunc) {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.doWait(nil, until)
}

// WaitFnAbort blocks until the condition function returns true or the abort
// channel is closed. Returns context.Canceled if aborted, or nil if the
// condition is met. If until is nil, waits for the value to become zero.
func (c *Cond) WaitFnAbort(abort <-chan struct{}, until CondFunc) error {
	err := c.lazyInit()
	switch {
	case err != nil:
		// invalid
		return err
	case c.doWait(abort, until):
		// cancelled
		return context.Canceled
	default:
		// success
		return nil
	}
}

// WaitFnContext blocks until the condition function returns true or the context
// is canceled. Returns the context's error if canceled, or nil if the condition
// is met. If until is nil, waits for the value to become zero.
// Returns utils.ErrNilContext if the provided context is nil.
func (c *Cond) WaitFnContext(ctx context.Context, until CondFunc) error {
	err := c.lazyInit()
	switch {
	case err != nil:
		// invalid
		return err
	case ctx == nil:
		// required
		return errors.ErrNilContext
	case c.doWait(ctx.Done(), until):
		// cancelled
		return ctx.Err()
	default:
		// success
		return nil
	}
}

// doWait implements the waiting logic for all Wait methods.
// It waits until either the condition is met or the abort channel is closed.
// Returns true if aborted, false if the condition is met.
func (c *Cond) doWait(abort <-chan struct{}, until CondFunc) bool {
	switch {
	case cancelled(abort):
		// cancelled already
		return true
	case c.doMatch(until):
		// ready
		return false
	}

	// Create a channel to wait on
	ch := c.newWaiter()
	defer c.removeWaiter(ch)

	for {
		select {
		case <-ch:
			// signaled
			switch {
			case cancelled(abort):
				// but cancelled
				return true
			case c.doMatch(until):
				// condition met
				return false
			default:
				// back to sleep
				continue
			}
		case <-abort:
			// cancelled.
			return true
		}
	}
}

// newWaiter registers a new waiter channel and returns it.
// The channel is used to notify the waiter about state changes.
func (c *Cond) newWaiter() chan struct{} {
	ch := make(chan struct{}, 1)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.waiters[ch] = struct{}{}
	return ch
}

// removeWaiter deregisters and closes a waiter channel.
func (c *Cond) removeWaiter(ch chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.waiters, ch)
	close(ch)
}

// Match tests the value against the given condition function and returns
// the result. If no condition function is provided, it tests if the Cond
// value is zero. Panics if the receiver is nil.
func (c *Cond) Match(cond CondFunc) bool {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doMatch(cond)
}

// doMatch implements the condition check logic.
// If until is nil, it checks if the value is zero.
func (c *Cond) doMatch(until CondFunc) bool {
	v := c.load()
	if until == nil {
		return v == 0
	}
	return until(c.value)
}

// IsZero checks if the Cond value is zero.
// Panics if the receiver is nil.
func (c *Cond) IsZero() bool {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return c.doMatch(nil)
}

// Signal notifies a single waiter about a state change.
// The waiter is selected at random from the pool of waiting goroutines.
// Returns true if a waiter was notified, false otherwise.
// Panics if the receiver is nil.
func (c *Cond) Signal() bool {
	if err := c.lazyInit(); err != nil {
		core.Panic(err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.unsafeSignal()
}

// unsafeSignal sends a signal to one waiter without locking.
// It assumes the mutex is already held by the caller.
func (c *Cond) unsafeSignal() bool {
	// iteration over maps is guaranteed to be random.
	for ch := range c.waiters {
		select {
		case ch <- struct{}{}:
			// signal sent
			return true
		default:
			// skip, already ready to run
			continue
		}
	}

	return false
}

// Broadcast notifies all waiters about a state change.
// Returns true if at least one waiter was notified, false otherwise.
// Panics if the receiver is nil.
func (c *Cond) Broadcast() bool {
	if err := c.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.unsafeBroadcast()
}

// unsafeBroadcast sends a signal to all waiters without locking.
// It assumes the mutex is already held by the caller.
func (c *Cond) unsafeBroadcast() bool {
	var signaled bool
	for ch := range c.waiters {
		select {
		case ch <- struct{}{}:
			// Signal sent
			signaled = true
		default:
			// skip, already ready to run
			continue
		}
	}
	return signaled
}
