package reconnect

import (
	"net"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var (
	_ fs.Reader = (*Client)(nil)
	_ fs.Writer = (*Client)(nil)
	_ fs.Closer = (*Client)(nil)
)

// dial attempts to stablish a connection to the server.
func (c *Client) dial(network, addr string) (net.Conn, error) {
	conn, err := c.dialer.DialContext(c.ctx, network, addr)
	switch {
	case err != nil:
		return nil, err
	case conn == nil:
		err = &net.OpError{
			Op:  "dial",
			Net: network,
			Err: core.Wrap(ErrAbnormalConnect, addr),
		}
		return nil, err
	}

	if fn := c.getOnConnect(); fn != nil {
		if err := fn(c.ctx, conn); err != nil {
			defer unsafeClose(conn)

			return nil, err
		}
	}

	return conn, nil
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
	return prev
}

// getConn returns the current connection, or an ErrNotConnected error
func (c *Client) getConn() (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn, nil
	}

	return nil, ErrNotConnected
}

// ResetDeadline sets the connection's read and write deadlines using
// the default values.
func (c *Client) ResetDeadline() error {
	return c.SetDeadline(c.readTimeout, c.writeTimeout)
}

// ResetReadDeadline resets the connection's read deadline using
// the default duration.
func (c *Client) ResetReadDeadline() error {
	return c.SetReadDeadline(c.readTimeout)
}

// SetReadDeadline sets the connections' read deadline to
// the specified duration. Use zero or negative to disable it.
func (c *Client) SetReadDeadline(d time.Duration) error {
	now := time.Now()
	conn, err := c.getConn()
	if err != nil {
		return err
	}

	t := TimeoutToAbsoluteTime(now, d)
	return conn.SetReadDeadline(t)
}

// ResetWriteDeadline resets the connection's write deadline using
// the default duration.
func (c *Client) ResetWriteDeadline() error {
	return c.SetWriteDeadline(c.writeTimeout)
}

// SetWriteDeadline sets the connections' write deadline to
// the specified duration. Use zero or negative to disable it.
func (c *Client) SetWriteDeadline(d time.Duration) error {
	now := time.Now()
	conn, err := c.getConn()
	if err != nil {
		return err
	}

	t := TimeoutToAbsoluteTime(now, d)
	return conn.SetWriteDeadline(t)
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
	conn, err := c.getConn()
	if err != nil {
		return err
	}

	t := TimeoutToAbsoluteTime(now, read)
	if err := conn.SetReadDeadline(t); err != nil {
		return err
	}

	t = TimeoutToAbsoluteTime(now, write)
	return conn.SetWriteDeadline(t)
}

// Read reads from the TCP connection, if connected.
func (c *Client) Read(p []byte) (int, error) {
	conn, err := c.getConn()
	if err != nil {
		return 0, err
	}

	return conn.Read(p)
}

// Write writes to the TCP connection, if connected.
func (c *Client) Write(p []byte) (int, error) {
	conn, err := c.getConn()
	if err != nil {
		return 0, err
	}

	return conn.Write(p)
}

// Close terminates the current connection
func (c *Client) Close() error {
	if conn, _ := c.getConn(); conn != nil {
		return conn.Close()
	}

	return nil
}

// RemoteAddr returns the remote address if connected.
func (c *Client) RemoteAddr() net.Addr {
	if conn, _ := c.getConn(); conn != nil {
		return conn.RemoteAddr()
	}
	return nil
}

// LocalAddr returns the local address if connected.
func (c *Client) LocalAddr() net.Addr {
	if conn, _ := c.getConn(); conn != nil {
		return conn.LocalAddr()
	}
	return nil
}
