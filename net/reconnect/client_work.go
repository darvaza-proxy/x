package reconnect

import (
	"context"

	"darvaza.org/core"
)

// Wait blocks until the [Client] workers have finished,
// and returns the cancellation reason.
func (c *Client) Wait() error {
	c.wg.Wait()
	return c.Err()
}

// Err returns the cancellation reason.
// it will return nil if the cause was initiated
// by the user.
func (c *Client) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return filterNonError(c.err)
}

// Done returns a channel that watches the [Client] workers,
// and provides the cancellation reason.
func (c *Client) Done() <-chan error {
	var barrier chan error

	go func() {
		defer close(barrier)
		barrier <- c.Wait()
	}()

	return barrier
}

// Shutdown initiates a shutdown and wait until the workers
// are done, or the given context times out.
func (c *Client) Shutdown(ctx context.Context) error {
	_ = c.terminate(nil)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.Done():
		return c.Err()
	}
}

func (c *Client) terminate(cause error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.err; err != nil {
		// already cancelled
		return filterNonError(err)
	}

	if cause == nil {
		cause = context.Canceled
	}

	c.err = cause
	c.cancel(cause)
	return filterNonError(cause)
}

// Go spawns a goroutine within the [Client]'s context.
func (c *Client) Go(_ string, fn func(context.Context) error) {
	if fn == nil {
		core.Panic("Client.Go called without a handler")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		var catcher core.Catcher

		defer c.wg.Done()

		err := catcher.Do(func() error {
			return fn(c.ctx)
		})

		if IsFatal(err) {
			_ = c.terminate(err)
		}
	}()
}
