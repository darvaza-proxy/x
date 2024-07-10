package reconnect

import (
	"bufio"
	"io"
	"net"
	"time"

	"darvaza.org/core"
	// "darvaza.org/x/fs"
)

var (
	_ io.Reader = (*Client)(nil)
	_ io.Writer = (*Client)(nil)
	// _ fs.Flusher = (*Client)(nil)
	_ io.Closer = (*Client)(nil)
)

// dial attempts to stablish a connection to the server.
func (c *Client) dial(network, addr string) (net.Conn, error) {
	conn, err := c.dialer.DialContext(c.ctx, network, addr)
	switch {
	case conn != nil:
		return conn, nil
	case err == nil:
		err = &net.OpError{
			Op:  "dial",
			Net: network,
			Err: core.Wrap(ErrAbnormalConnect, addr),
		}

		fallthrough
	default:
		return nil, err
	}
}

// reconnect waits before dialing
func (c *Client) reconnect(network, addr string) (net.Conn, error) {
	if fn := c.getWaitReconnect(); fn != nil {
		if err := fn(c.ctx); err != nil {
			return nil, err
		}
	}

	return c.dial(network, addr)
}

// setConn prepares the Client to use the new net.Conn
// and returns the previous, if any.
func (c *Client) setConn(conn net.Conn) net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.unsafeSetConn(conn)
}

func (c *Client) unsafeSetConn(conn net.Conn) (prev net.Conn) {
	prev, c.conn = c.conn, conn
	if conn == nil {
		c.in, c.out = nil, nil
	} else {
		c.in = bufio.NewReader(conn)
		c.out = bufio.NewWriter(conn)
	}
	return prev
}

// ResetDeadline sets the connection's read and write deadlines using
// the default values.
func (c *Client) ResetDeadline() error {
	return c.SetDeadline(c.readTimeout, c.writeTimeout)
}

// ResetReadDeadline resets the connection's read deadline using
// the default duration.
func (c *Client) ResetReadDeadline() error {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	t := TimeoutToAbsoluteTime(now, c.readTimeout)
	return c.conn.SetReadDeadline(t)
}

// ResetWriteDeadline resets the connection's write deadline using
// the default duration.
func (c *Client) ResetWriteDeadline() error {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	t := TimeoutToAbsoluteTime(now, c.writeTimeout)
	return c.conn.SetWriteDeadline(t)
}

// SetDeadline sets the connections's read and write deadlines.
// if write is zero but read is positive, write is set using the same
// value as read.
// zero or negative can be used to disable the deadline.
func (c *Client) SetDeadline(read, write time.Duration) error {
	if read > 0 && write == 0 {
		write = read
	}

	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	t := TimeoutToAbsoluteTime(now, read)
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}

	t = TimeoutToAbsoluteTime(now, write)
	return c.conn.SetWriteDeadline(t)
}

// Read implements a buffered io.Reader
func (c *Client) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.in != nil {
		return c.in.Read(p)
	}

	return 0, ErrNotConnected
}

// Write implements a buffered io.Writer
// warrantied to buffer all the given data or fail.
func (c *Client) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	w := c.out
	if w == nil {
		return 0, ErrNotConnected
	}

	total := 0
	for len(p) > 0 {
		n, err := w.Write(p)

		switch {
		case err != nil:
			return total, err
		default:
			total += n
			p = p[n:]
		}
	}

	return total, nil
}

// Flush blocks until all the buffered output
// has been written, or an error occurs.
func (c *Client) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.out != nil {
		return c.out.Flush()
	}

	return ErrNotConnected
}

// Close terminates the current connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
