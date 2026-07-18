package reconnect

import (
	"context"
	"io/fs"
	"os"
	"syscall"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

var (
	// ErrAbnormalConnect indicates the dialer didn't return error
	// nor connection.
	ErrAbnormalConnect = core.QuietWrap(syscall.ECONNABORTED, "abnormal response")

	// ErrDoNotReconnect indicates the Waiter
	// instructed us to not reconnect
	ErrDoNotReconnect = errors.New("don't reconnect")

	// ErrNotConnected indicates the [Client] isn't currently connected.
	// It wraps [ErrClosed] so a single errors.Is target covers both a
	// closed client and the not-connected window.
	ErrNotConnected = core.QuietWrap(ErrClosed, "client not connected")

	// ErrRunning indicates the [Client] has already been started.
	ErrRunning = core.QuietWrap(syscall.EBUSY, "client already running")

	// ErrClosed indicates the [Client] or [StreamSession] has
	// already been shut down. It wraps the workgroup's sentinel so
	// the shutdown signal still matches the one the group returns
	// across the lifecycle stack.
	ErrClosed = core.QuietWrap(errors.ErrClosed, "already closed")

	// ErrNameEmpty indicates a name is empty. It wraps [core.ErrInvalid]
	// so a caller matching the invalid-argument family with errors.Is
	// catches it.
	ErrNameEmpty = core.QuietWrap(core.ErrInvalid, "name missing")

	// ErrNameTooLong indicates a name exceeds maximum length. It wraps
	// [core.ErrInvalid] for the same reason.
	ErrNameTooLong = core.QuietWrap(core.ErrInvalid, "name too long")
)

// IsFatal tells if the error means the connection
// should be closed and not retried.
// Only [ErrDoNotReconnect], possibly wrapped, is considered
// fatal; anything else is treated as recoverable.
//
// IsFatal classifies connection errors seen inside the reconnect loop.
// Caller-misuse errors are reported at setup time and never reach that
// decision; they extend [core.ErrInvalid], so match them with errors.Is
// against that sentinel instead.
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

// expectableConnErrors are the connection failures the reconnect loop
// treats as recoverable: the transport or the peer dropped the
// connection in a way a redial can legitimately recover from. The
// per-platform set in expectableConnErrorsOS adds the equivalents that
// exist only on that platform — the WSAE* family on Windows, where a
// dial or an established connection reports those rather than the POSIX
// ECONN* values above.
var expectableConnErrors = append([]error{
	fs.ErrClosed,
	os.ErrDeadlineExceeded,
	syscall.ECONNABORTED,
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
}, expectableConnErrorsOS...)

// checkIsExpectable reports whether err is a recoverable connection
// failure, matching any node in its chain against expectableConnErrors.
var checkIsExpectable = core.NewCheckErrorIsIn2(expectableConnErrors)

// filterNonError checks if the cause of the shutdown is worth
// reporting or it was initiated by the user instead.
func filterNonError(err error) error {
	if IsNonError(err) {
		return nil
	}

	// error
	return err
}

// IsNonError reports whether the error represents a
// user-initiated shutdown instead of an actual failure.
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

// IsNotConnected reports whether err indicates the [Client] had no session
// when a request was attempted, matching [ErrNotConnected] anywhere in the
// chain. A fully shut-down client surfaces [ErrClosed] without
// ErrNotConnected wrapping it; match [ErrClosed] instead to cover both the
// closed and not-connected cases with a single target.
func IsNotConnected(err error) bool {
	return errors.Is(err, ErrNotConnected)
}

// IsClosed reports whether err indicates the [Client] has been shut down,
// matching [ErrClosed] anywhere in the chain. Because [ErrNotConnected]
// wraps [ErrClosed], IsClosed is the broad companion to [IsNotConnected]:
// it is true for both a fully closed client and the transient
// not-connected window, whereas IsNotConnected matches only the latter.
func IsClosed(err error) bool {
	return errors.Is(err, ErrClosed)
}
