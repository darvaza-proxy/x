package reconnect

import (
	"context"
	"time"
)

const (
	// DefaultWaitReconnect specifies how long we will wait for
	// to reconnect by default
	DefaultWaitReconnect = 5 * time.Second
)

// A Waiter is a function that blocks and returns an
// error when cancelled or nil when we are good to continue.
type Waiter func(context.Context) error

// NewConstantWaiter blocks for a given amount of time, or until
// the context is cancelled.
// If the given duration is negative, the [Waiter] won't wait, but
// it will still check for context terminations.
// If zero, the [Waiter] will wait the default amount.
func NewConstantWaiter(d time.Duration) func(context.Context) error {
	if d < 0 {
		return NewImmediateErrorWaiter(ErrDoNotReconnect)
	}

	if d == 0 {
		d = DefaultWaitReconnect
	}

	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
			return nil
		}
	}
}

// NewImmediateErrorWaiter returns a Waiter that will return the
// context cancellation cause or the specified error, if any.
// There is no actual waiting.
func NewImmediateErrorWaiter(err error) func(context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}
}

// NewDoNotReconnectWaiter returns a Waiter that will return the
// context cancellation cause, the specified error, or ErrDoNotReconnect.
func NewDoNotReconnectWaiter(err error) func(context.Context) error {
	if err == nil {
		err = ErrDoNotReconnect
	}

	return NewImmediateErrorWaiter(err)
}
