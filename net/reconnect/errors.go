package reconnect

import (
	"context"
	"errors"
	"io/fs"
	"os"
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
	ErrNotConnected = core.QuietWrap(fs.ErrClosed, "not connected")

	// ErrRunning indicates the [Client] has already being started.
	ErrRunning = core.QuietWrap(syscall.EBUSY, "client already running")

	// ErrNameEmpty indicates a name is empty
	ErrNameEmpty = errors.New("name missing")

	// ErrNameTooLong indicates a name exceeds maximum length
	ErrNameTooLong = errors.New("name too long")
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

		// temporary errors are never fatal
		if is, _ := core.CheckIsTemporary(err); is {
			return false, true
		}

		// unknown
		return false, false
	}
}

func checkIsExpectable(err error) (is, certainly bool) {
	switch err {
	case fs.ErrClosed,
		os.ErrDeadlineExceeded,
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

// filterNonError checks if the cause of the shutdown is worth
// reporting or it was initiated by the user instead.
func filterNonError(err error) error {
	if IsNonError(err) {
		return nil
	}

	// error
	return err
}

// IsNonError checks if the error is an actual error instead of
// a manual shutdown.
func IsNonError(err error) bool {
	if err == nil {
		return true
	}

	is, _ := core.IsErrorFn2(checkIsNonError, err)
	return is
}

func checkIsNonError(err error) (is, certainly bool) {
	switch err {
	case nil, context.Canceled, ErrDoNotReconnect:
		return true, true
	default:
		return false, false
	}
}
