// Package cond provides synchronisation primitives for coordinating goroutines.
package cond

import (
	"context"

	"darvaza.org/x/sync/errors"
)

// CountZero is a specialised variant of Count that specifically broadcasts
// when the counter reaches zero. This simplifies coordination in cases
// where completion is signified by a zero value.
type CountZero Count

// isZero returns true when the provided value equals zero.
// Used as the broadcast condition for the CountZero type.
func isZero(n int32) bool { return n == 0 }

// NewCountZero creates a new Count that automatically wakes
// waiters when the count reaches zero.
func NewCountZero(initialValue int) *CountZero {
	c := NewCount(
		initialValue,
		isZero,
	)
	return (*CountZero)(c)
}

// sys returns the underlying Count instance.
// Returns nil if the receiver is nil.
func (c0 *CountZero) sys() *Count {
	if c0 == nil {
		return nil
	}
	return (*Count)(c0)
}

// check validates the CountZero instance and returns the underlying Count.
// Returns an error if the receiver is nil or not properly initialised.
func (c0 *CountZero) check() (*Count, error) {
	c := c0.sys()
	switch {
	case c == nil:
		return nil, errors.ErrNilReceiver
	case c.b.IsNil():
		return nil, errors.ErrNotInitialised
	default:
		return c, nil
	}
}

// Init initialises the CountZero with a given initial value.
// Returns an error if the receiver is nil or already initialised.
func (c0 *CountZero) Init(initialValue int) error {
	if c := c0.sys(); c != nil {
		return c.doInit(
			initialValue,
			[]func(int32) bool{isZero},
		)
	}
	return errors.ErrNilReceiver
}

// Reset sets the CountZero to a new value.
// Returns an error if the instance is not properly initialised or is nil.
func (c0 *CountZero) Reset(value int) error {
	c, err := c0.check()
	if err != nil {
		return err
	}
	return c.doReset(value)
}

// Close releases the resources associated with the CountZero.
// Returns an error if the receiver is nil or not properly initialised.
func (c0 *CountZero) Close() error {
	c, err := c0.check()
	if err != nil {
		return err
	}
	return c.b.Close()
}

// IsNil reports whether the CountZero is nil or uninitialised.
// Provides a safe way to check the initialisation status.
func (c0 *CountZero) IsNil() bool {
	if c := c0.sys(); c != nil {
		return c.b.IsNil()
	}
	return true
}

// IsClosed reports whether the CountZero is closed and no longer usable.
// Returns true if nil or if the underlying barrier is closed.
func (c0 *CountZero) IsClosed() bool {
	if c := c0.sys(); c != nil {
		return c.b.IsClosed()
	}
	return true
}

// Add atomically adds the given value and returns the new value.
// Broadcasts to waiters if the result is zero.
func (c0 *CountZero) Add(n int) int { return c0.sys().Add(n) }

// Inc atomically increments the counter by 1 and returns the new value.
// Broadcasts to waiters if the result is zero.
func (c0 *CountZero) Inc() int { return c0.sys().Inc() }

// Dec atomically decrements the counter by 1 and returns the new value.
// Broadcasts to waiters if the result is zero.
func (c0 *CountZero) Dec() int { return c0.sys().Dec() }

// Value atomically returns the current value without affecting waiters.
func (c0 *CountZero) Value() int { return c0.sys().Value() }

// Wait blocks the calling goroutine until the counter reaches zero.
// Returns an error if the CountZero is nil or not properly initialised.
func (c0 *CountZero) Wait() error {
	return c0.sys().WaitFnAbort(nil, nil)
}

// WaitAbort blocks until the counter reaches zero or the abort channel closes.
// Returns context.Canceled if aborted, or nil if the counter reaches zero.
func (c0 *CountZero) WaitAbort(abort <-chan struct{}) error {
	return c0.sys().WaitFnAbort(abort, nil)
}

// WaitContext blocks until the counter reaches zero or the context is cancelled.
// Returns the context's error if cancelled, or nil if the counter reaches zero.
func (c0 *CountZero) WaitContext(ctx context.Context) error {
	return c0.sys().WaitFnContext(ctx, nil)
}
