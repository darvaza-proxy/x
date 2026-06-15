package reconnect

import (
	"context"
	"net"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Wait blocks until the [Client] workers have finished,
// and returns the cancellation reason, nil if the shutdown
// was user-initiated.
func (c *Client) Wait() error {
	return filterNonError(c.wg.Wait())
}

// Err returns the cancellation reason.
// It will return nil if the cause was initiated
// by the user.
func (c *Client) Err() error {
	return filterNonError(c.wg.Err())
}

// Done returns a channel that is closed once the [Client]
// workers have finished. Use [Client.Err] to learn the
// cancellation reason.
func (c *Client) Done() <-chan struct{} {
	return c.wg.Done()
}

// Shutdown initiates a shutdown and waits until the workers
// are done, or the given context times out.
func (c *Client) Shutdown(ctx context.Context) error {
	c.terminate(nil)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.Done():
		return c.Err()
	}
}

// terminate cancels the [Client] with the given cause, defaulting
// to [context.Canceled]. It is idempotent: only the first call
// records a cause and logs the termination.
func (c *Client) terminate(cause error) {
	if cause == nil {
		cause = context.Canceled
	}

	if c.wg.Cancel(cause) {
		// first cancellation
		if l, ok := c.WithDebug(nil); ok {
			l = l.WithField(slog.ErrorFieldName, cause)
			l.Print("client terminated")
		}
	}
}

// Go spawns a goroutine within the [Client]'s context. Submissions
// after shutdown are no-ops: the worker is dropped rather than run
// with an already-cancelled context.
func (c *Client) Go(funcs ...WorkerFunc) {
	for _, fn := range funcs {
		if fn != nil {
			c.spawnOne(fn, nil)
		}
	}
}

// GoCatch spawns a goroutine within the [Client]'s context,
// optionally allowing filtering the error to stop cascading.
// Submissions after shutdown are no-ops, as with [Client.Go].
func (c *Client) GoCatch(run WorkerFunc, catch CatcherFunc) {
	if run != nil {
		c.spawnOne(run, catch)
	}
}

func (c *Client) spawnOne(run WorkerFunc, catch CatcherFunc) {
	err := c.wg.Go(func(ctx context.Context) {
		var catcher core.Catcher

		err := catcher.Do(func() error {
			return run(ctx)
		})

		if err != nil && catch != nil {
			err = catch(ctx, err)
		}

		_ = c.handlePossiblyFatalError(nil, err)
	})
	if err != nil {
		// the group is already shutting down; the worker is dropped
		// rather than run with a cancelled context. Surface it at
		// debug so a late submission isn't lost silently.
		if l, ok := c.WithDebug(nil); ok {
			l = l.WithField(slog.ErrorFieldName, err)
			l.Print("worker submitted after shutdown; dropped")
		}
	}
}

// run implements the main loop.
func (c *Client) run(conn net.Conn) {
	var abort bool

	defer func() {
		// panic
		if err := core.AsRecovered(recover()); err != nil {
			// report and terminate
			_ = c.doOnError(nil, err, "reconnect.Client")
			c.terminate(err)
		}
	}()

	for !abort {
		if conn != nil {
			// run
			e1 := c.runSession(conn)
			e2 := c.doOnDisconnect()

			abort = c.runError(conn, e1, e2)
		}

		if !abort {
			// reconnect
			conn, abort = c.tryReconnect()
		}
	}
}

func (c *Client) runError(conn net.Conn, e1, e2 error) bool {
	e1 = c.handlePossiblyFatalError(conn, e1)
	e2 = c.handlePossiblyFatalError(conn, e2)
	return e1 != nil || e2 != nil
}

func (c *Client) runSession(conn net.Conn) error {
	defer unsafeClose(conn)

	// initialize
	c.setConn(conn)

	if fn := c.getOnSession(); fn != nil {
		var catch core.Catcher

		// hand over, surfacing recovered panics as errors
		return catch.Do(func() error {
			return fn(c.ctx)
		})
	}

	// no handler, try logging it before closing the connection.
	if l, ok := c.WithInfo(conn.LocalAddr()); ok {
		l.Print("connected")
	}

	return nil
}

func (c *Client) tryReconnect() (net.Conn, bool) {
	if fn := c.getWaitReconnect(); fn != nil {
		if err := fn(c.ctx); err != nil {
			// Per the [Waiter] contract any error means "stop
			// reconnecting". Unlike a dial failure it is terminal
			// regardless of [IsFatal], which classifies connection
			// errors, not the waiter's stop decision.
			return nil, c.handleWaiterError(err)
		}
	}

	network, address := c.getRemote()
	conn, err := c.dial(network, address)
	if conn != nil {
		// ready
		return conn, false
	}

	return nil, c.handleReconnectError(err)
}

func (c *Client) doOnDisconnect() error {
	conn := c.setConn(nil)

	if fn := c.getOnDisconnect(); fn != nil {
		return fn(c.ctx, conn)
	}

	// no handler, try logging it before closing the connection.
	if l, ok := c.WithInfo(conn.LocalAddr()); ok {
		l.Print("disconnected")
	}

	return nil
}

func (c *Client) doOnError(conn net.Conn, err error, note string, args ...any) error {
	var addr net.Addr

	if fn := c.getOnError(); fn != nil {
		return fn(c.ctx, conn, core.Wrapf(err, note, args...))
	}

	if conn != nil {
		addr = conn.LocalAddr()
	}

	// no handler, try log the error here and pass it through.
	if l, ok := c.WithError(addr, err); ok {
		l.Printf(note, args...)
	}

	return err
}

// handlePossiblyFatalError handles an error and returns nil if it wasn't fatal.
// Fatal errors should terminate the worker immediately.
// The returned error is unfiltered.
func (c *Client) handlePossiblyFatalError(conn net.Conn, err error) error {
	if err != nil {
		err = c.doOnError(conn, err, "")
		if IsFatal(err) {
			c.terminate(err)
			return err // unfiltered
		}
	}

	return nil
}

func (c *Client) handleConnectError(conn net.Conn, err error) error {
	if err == nil {
		return nil
	}

	// remember context terminations before the non-fatal
	// handling swallows the error.
	cancellation := checkCancellation(err)

	if err = c.handlePossiblyFatalError(conn, err); err != nil {
		return err
	}

	switch {
	case cancellation != nil:
		return cancellation
	case c.ctx.Err() != nil:
		// the [Client]'s context is done; stop instead of
		// spinning on doomed reconnection attempts.
		return c.ctx.Err()
	default:
		return nil
	}
}

// checkCancellation extracts a context termination from the
// error chain, if any.
func checkCancellation(err error) error {
	switch {
	case core.IsError(err, context.Canceled):
		return context.Canceled
	case core.IsError(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	default:
		return nil
	}
}

func (c *Client) handleReconnectError(err error) bool {
	err = c.handleConnectError(nil, err)
	return err != nil
}

// handleWaiterError stops the [Client] after a [Waiter] declined to
// continue. The waiter is the reconnection-policy authority, so any
// error terminates unconditionally; [Config.OnError] still observes
// it and may rewrite the recorded cause.
func (c *Client) handleWaiterError(err error) bool {
	err = c.doOnError(nil, err, "reconnect waiter")
	c.terminate(err)
	return true
}
