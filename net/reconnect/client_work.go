package reconnect

import (
	"context"
	"net"

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
func (c *Client) Go(name string, fn func(context.Context) error) {
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

		if c.handlePossiblyFatalError(nil, err, name) != nil {
			return
		}
	}()
}

// run implements the main loop
func (c *Client) run() {
	var conn net.Conn
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
		conn, abort = c.doConnect()

		if conn != nil {
			e1 := c.runSession(conn)
			e2 := c.doOnDisconnect()

			abort = c.runError(conn, e1, e2)
		}
	}
}

func (c *Client) runError(conn net.Conn, e1, e2 error) bool {
	e1 = c.handlePossiblyFatalError(conn, e1, "")
	e2 = c.handlePossiblyFatalError(conn, e2, "")
	return e1 != nil || e2 != nil
}

func (c *Client) runSession(conn net.Conn) error {
	defer unsafeClose(conn)

	if fn := c.getOnSession(); fn != nil {
		var catch core.Catcher

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

func (c *Client) doConnect() (net.Conn, bool) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		var err error

		network, address := c.getRemote()
		conn, err = c.reconnect(network, address)
		if abort := c.runError(nil, err, nil); abort {
			// abort
			return nil, true
		}

		if conn != nil {
			c.setConn(conn)
		}
	}

	// ready
	return conn, false
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
		return fn(c.ctx, conn, core.Wrap(err, note, args...))
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
// fatal errors should terminate the worker immediately.
// the returned error is unfiltered.
func (c *Client) handlePossiblyFatalError(conn net.Conn, err error, note string) error {
	if err != nil {
		err = c.doOnError(conn, err, note)
		if IsFatal(err) {
			_ = c.terminate(err)
			return err // unfiltered
		}
	}

	return nil
}
