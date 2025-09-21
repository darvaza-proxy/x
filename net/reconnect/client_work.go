package reconnect

import (
	"context"
	"net"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Wait blocks until the [Client] workers have finished,
// and returns the cancellation reason.
func (c *Client) Wait() error {
	c.wg.Wait()
	return c.Err()
}

// Err returns the cancellation reason.
// It will return nil if the cause was initiated
// by the user.
func (c *Client) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	return filterNonError(err)
}

// Done returns a channel that watches the [Client] workers,
// and provides the cancellation reason.
func (c *Client) Done() <-chan struct{} {
	barrier := make(chan struct{})

	go func() {
		defer close(barrier)
		c.wg.Wait()
	}()

	return barrier
}

// Shutdown initiates a shutdown and waits until the workers
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

	if l, ok := c.WithDebug(nil); ok {
		l = l.WithField(slog.ErrorFieldName, cause)
		l.Print("client terminated")
	}

	c.err = cause
	c.cancel(cause)
	return filterNonError(cause)
}

// Go spawns a goroutine within the [Client]'s context.
func (c *Client) Go(funcs ...WorkerFunc) {
	for _, fn := range funcs {
		if fn != nil {
			c.spawnOne(fn, nil)
		}
	}
}

// GoCatch spawns a goroutine within the [Client]'s context,
// optionally allowing filtering the error to stop cascading.
func (c *Client) GoCatch(run WorkerFunc, catch CatcherFunc) {
	if run != nil {
		c.spawnOne(run, catch)
	}
}

func (c *Client) spawnOne(run WorkerFunc, catch CatcherFunc) {
	c.wg.Add(1)

	go func() {
		var catcher core.Catcher

		defer c.wg.Done()

		err := catcher.Do(func() error {
			return run(c.ctx)
		})

		if err != nil && catch != nil {
			err = catch(c.ctx, err)
		}

		_ = c.handlePossiblyFatalError(nil, err)
	}()
}

// run implements the main loop.
func (c *Client) run(conn net.Conn) {
	var abort bool

	defer func() {
		// panic
		if err := core.AsRecovered(recover()); err != nil {
			// report and terminate
			_ = c.doOnError(nil, err, "reconnect.Client")
			_ = c.terminate(err)
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

		// hand over
		return catch.Try(func() error {
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
	network, address := c.getRemote()
	conn, err := c.reconnect(network, address)
	if conn != nil {
		// ready
		return conn, false
	}

	abort := c.handleReconnectError(err)
	return nil, abort
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
			_ = c.terminate(err)
			return err // unfiltered
		}
	}

	return nil
}

func (c *Client) handleConnectError(conn net.Conn, err error) error {
	var cancelled bool

	switch err {
	case nil:
		return nil
	case context.Canceled:
		cancelled = true
	}

	err = c.handlePossiblyFatalError(conn, err)
	switch {
	case err != nil:
		return err
	case cancelled:
		return context.Canceled
	default:
		return nil
	}
}

func (c *Client) handleReconnectError(err error) bool {
	err = c.handleConnectError(nil, err)
	return err != nil
}
