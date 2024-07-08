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

// IsFatal tells if the error means the connection
// should be closed and not retried.
func IsFatal(err error) bool {
	if err != nil {
		is, _ := core.IsErrorFn2(checkIsFatal, err)
		return is
	}
	return false
}

func checkIsFatal(err error) (is, certainly bool) {
	switch err {
	case ErrDoNotReconnect:
		// do-not-reconnect
		return true, true
	default:
		// normal errors are never fatal
		if is, _ := checkIsExpectable(err); is {
			return false, true
		}

		// reconnect if temporary
		return core.CheckIsTemporary(err)
	}
}

func checkIsExpectable(err error) (is, certainly bool) {
	switch err {
	case fs.ErrClosed,
		syscall.ECONNABORTED,
		syscall.ECONNREFUSED,
		syscall.ECONNRESET:
		// reconnect
		return true, true
	default:
		// unknown
		return false, false
	}
}
