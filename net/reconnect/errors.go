package reconnect

import (
	"errors"
	"io/fs"
	"syscall"

	"darvaza.org/core"
)

var (
	// ErrAbnormalConnect indicates the dialer didn't return error
	// nor connection.
	ErrAbnormalConnect = core.QuietWrap(syscall.ECONNABORTED, "abnormal response")

	// ErrDoNotReconnect indicates the Waiter
	// instructed us to not reconnect
	ErrDoNotReconnect = errors.New("don't reconnect")

	// ErrNotConnected indicates the [Client] isn't currently connected.
	ErrNotConnected = core.QuietWrap(fs.ErrClosed, "connection closed")
)
