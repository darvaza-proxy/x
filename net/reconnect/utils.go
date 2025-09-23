package reconnect

import (
	"errors"
	"fmt"
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

// ParseRemote determines the network type and address from a remote string.
// It supports:
// - "unix:/path/to/socket" - explicit Unix socket.
// - "/path/to/socket" - Unix socket (absolute path).
// - "path/to/file.sock" - Unix socket (ends with .sock).
// - "host:port" - TCP socket.
func ParseRemote(remote string) (network, address string, err error) {
	if remote == "" {
		return "", "", errors.New("remote address missing")
	}

	// Check for explicit unix: prefix
	if addr, ok := strings.CutPrefix(remote, NetworkUnix+":"); ok {
		return parseRemoteUNIX(addr)
	}

	// Check if it's an absolute path (Unix socket)
	if strings.HasPrefix(remote, "/") {
		return parseRemoteUNIX(remote)
	}

	// Check if it ends with .sock (Unix socket)
	if strings.HasSuffix(remote, ".sock") {
		return parseRemoteUNIX(remote)
	}

	// Fallback to TCP
	return parseRemoteTCP(remote)
}

func parseRemoteUNIX(remote string) (network, address string, err error) {
	// not-empty
	if remote == "" {
		return "", "", errors.New("unix socket path missing")
	}
	// TODO: validate path when it might be valid but not exist?
	return NetworkUnix, remote, nil
}

func parseRemoteTCP(remote string) (network, address string, err error) {
	// TCP - validate host:port
	_, port, err := net.SplitHostPort(remote)
	switch {
	case err != nil:
		return "", "", err
	case port == "":
		return "", "", fmt.Errorf("%q: port missing", remote)
	default:
		return NetworkTCP, remote, nil
	}
}

// ValidateRemote validates a remote address for use with reconnect clients.
// It returns nil if the address is valid for either TCP or Unix socket connection.
// For TCP addresses, it validates host:port format.
// For Unix socket addresses, it accepts the address as-is.
func ValidateRemote(remote string) error {
	_, _, err := ParseRemote(remote)
	return err
}
