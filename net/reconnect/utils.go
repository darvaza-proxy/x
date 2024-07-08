package reconnect

import (
	"io"
	"net"
	"time"
)

func newDialer(keepalive, timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		KeepAlive: keepalive,
		Timeout:   timeout,
	}
}

func unsafeClose(f io.Closer) {
	_ = f.Close()
}

// TimeoutToAbsoluteTime adds the given [time.Duration] to a
// base [time.Time].
// if the duration is negative, a zero [time.Time] will
// be returned.
// if the base is zero, the current time will be used.
func TimeoutToAbsoluteTime(base time.Time, d time.Duration) time.Time {
	if d > 0 {
		if base.IsZero() {
			base = time.Now()
		}

		return base.Add(d)
	}

	return time.Time{} // isZero()
}
