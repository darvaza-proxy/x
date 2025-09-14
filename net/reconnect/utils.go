package reconnect

import (
	"io"
	"net"
	"strings"
	"time"
)

const (
	// NetworkTCP represents TCP network type
	NetworkTCP = "tcp"
	// NetworkUnix represents Unix domain socket network type
	NetworkUnix = "unix"
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
// If the duration is negative, a zero [time.Time] will
// be returned.
// If the base is zero, the current time will be used.
func TimeoutToAbsoluteTime(base time.Time, d time.Duration) time.Time {
	if d > 0 {
		if base.IsZero() {
			base = time.Now()
		}

		return base.Add(d)
	}

	return time.Time{} // isZero()
}

// parseRemote determines the network type and address from a remote string.
// It supports:
// - "unix:/path/to/socket" - explicit Unix socket.
// - "/path/to/socket" - Unix socket (absolute path).
// - "path/to/file.sock" - Unix socket (contains .sock).
// - "host:port" - TCP socket.
func parseRemote(remote string) (network, address string) {
	// Check for explicit unix: prefix
	if addr, ok := strings.CutPrefix(remote, NetworkUnix+":"); ok {
		return NetworkUnix, addr
	}

	// Check if it's an absolute path (Unix socket)
	if strings.HasPrefix(remote, "/") {
		return NetworkUnix, remote
	}

	// Check if it contains .sock (Unix socket)
	if strings.Contains(remote, ".sock") {
		return NetworkUnix, remote
	}

	// Default to TCP
	return NetworkTCP, remote
}
