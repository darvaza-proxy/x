package reconnect

import (
	"time"
)

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
