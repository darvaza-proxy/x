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
