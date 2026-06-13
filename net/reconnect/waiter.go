package reconnect

import (
	"context"
	"time"
)

const (
	// DefaultWaitReconnect specifies how long [NewConstantWaiter]
	// waits between reconnection attempts by default.
	DefaultWaitReconnect = 5 * time.Second
)

// A Waiter is a function that blocks between reconnection
// attempts. It returns nil when the [Client] is good to try
// again, or an error to stop reconnecting.
type Waiter func(context.Context) error

// NewConstantWaiter blocks for a given amount of time, or until
// the context is cancelled.
// If the given duration is zero, [DefaultWaitReconnect] is used.
// If negative, reconnecting is disabled, failing with
// [ErrDoNotReconnect].
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

// NewImmediateErrorWaiter returns a [Waiter] that doesn't wait.
// It returns the context's error if the context has already
// terminated, or the given error otherwise. A nil error allows
// an immediate reconnection attempt.
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

// NewDoNotReconnectWaiter returns a [Waiter] that stops
// reconnection attempts, failing with the given error, or
// [ErrDoNotReconnect] when nil. The context's error takes
// precedence if the context has already terminated.
func NewDoNotReconnectWaiter(err error) func(context.Context) error {
	if err == nil {
		err = ErrDoNotReconnect
	}

	return NewImmediateErrorWaiter(err)
}
