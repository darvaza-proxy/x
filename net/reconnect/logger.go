package reconnect

import (
	"net"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

const (
	// LogFieldAddress is the field name used to store the address
	// when logging.
	LogFieldAddress = "addr"

	// LogFieldError is the field name used to store the error
	// when logging.
	LogFieldError = slog.ErrorFieldName
)

// WithDebug gets a logger at Debug level optionally annotated
// by an IP address. If the Debug log-level is disabled, it
// will return `nil, false`
func (c *Client) WithDebug(addr net.Addr) (slog.Logger, bool) {
	if l, ok := c.logger.Debug().WithEnabled(); ok {
		l = logWithAddress(l, addr)
		return l, true
	}
	return nil, false
}

// WithInfo gets a logger at Info level optionally annotated
// by an IP address. If the Info log-level is disabled, it
// will return `nil, false`
func (c *Client) WithInfo(addr net.Addr) (slog.Logger, bool) {
	if l, ok := c.logger.Info().WithEnabled(); ok {
		l = logWithAddress(l, addr)
		return l, true
	}
	return nil, false
}

// WithError gets a logger at Error level optionally annotated
// by an IP address. If the Error log-level is disabled, it
// will return `nil, false`
func (c *Client) WithError(addr net.Addr, err error) (slog.Logger, bool) {
	if l, ok := c.logger.Error().WithEnabled(); ok {
		l = logWithAddress(l, addr)
		l = logWithError(l, err)
		return l, true
	}
	return nil, false
}

// SayRemote makes a log entry optionally including the remote's address from
// the [net.Conn]
func (c *Client) SayRemote(conn net.Conn, note string, args ...any) {
	var ra net.Addr
	if conn != nil {
		ra = conn.RemoteAddr()
	}

	if l, ok := c.WithInfo(ra); ok {
		if len(args) > 0 {
			l.Printf(note, args...)
		} else {
			l.Print(note)
		}
	}
}

// SayRemoteError makes an error log entry optionally include the
// remote's address from the [net.Conn].
func (c *Client) SayRemoteError(conn net.Conn, err error, note string, args ...any) {
	var ra net.Addr
	if conn != nil {
		ra = conn.RemoteAddr()
	}

	if l, ok := c.WithError(ra, err); ok {
		if len(args) > 0 {
			l.Printf(note, args...)
		} else {
			l.Print(note)
		}
	}
}

func logWithAddress(l slog.Logger, addr net.Addr) slog.Logger {
	if l != nil && !core.IsZero(addr) {
		if s := addr.String(); s != "" {
			l = l.WithField(LogFieldAddress, s)
		}
	}
	return l
}

func logWithError(l slog.Logger, err error) slog.Logger {
	if l != nil && !IsNonError(err) {
		l = l.WithField(LogFieldError, err)
	}
	return l
}

func newDefaultLogger() slog.Logger {
	return discard.New()
}
