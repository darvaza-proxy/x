package reconnect

import (
	"errors"
	"fmt"
	"io"
	stdnet "net"
	"strings"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/net"
)

const (
	// NetworkTCP represents TCP network type
	NetworkTCP = "tcp"
	// NetworkUnix represents Unix domain socket network type
	NetworkUnix = "unix"
	// MaxUNIXSocketPathLength is the maximum length for UNIX domain socket paths
	// Limited by sockaddr_un.sun_path (108 bytes including null terminator on Linux)
	MaxUNIXSocketPathLength = 107
)

func newDialer(keepalive, timeout time.Duration) *stdnet.Dialer {
	return &stdnet.Dialer{
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
// - "@abstract-name" - abstract Unix socket.
// - "path/to/file.sock" - Unix socket (ends with .sock).
// - "host:port" - TCP socket.
func ParseRemote(remote string) (network, address string, err error) {
	if remote == "" {
		return "", "", errors.New("remote address missing")
	}

	if addr, skip, err := parseRemoteUNIX(remote); !skip || err != nil {
		return NetworkUnix, addr, err
	}

	// Fallback to TCP
	return parseRemoteTCP(remote)
}

func parseRemoteUNIX(remote string) (address string, skip bool, err error) {
	// Check for explicit unix: prefix
	if addr, ok := strings.CutPrefix(remote, NetworkUnix+":"); ok {
		err = validateMustRemoteUNIX(addr)
		if err != nil {
			// invalid prefixed unix socket
			err = core.Wrapf(err, "%q: invalid prefixed UNIX socket", remote)
		}
		return addr, false, err
	}

	skip, err = validateRemoteUNIX(remote)
	switch {
	case skip:
		// try TCP instead
		return "", true, nil
	case err != nil:
		// invalid unix socket
		return "", false, core.Wrapf(err, "%q: invalid UNIX socket", remote)
	default:
		// good-enough unix socket
		return remote, false, nil
	}
}

func validateMustRemoteUNIX(remote string) error {
	skip, err := validateRemoteUNIX(remote)
	if skip {
		err = validateUNIXName(remote)
	}
	return err
}

func validateRemoteUNIX(remote string) (skip bool, err error) {
	var validatePath bool

	switch {
	case strings.HasPrefix(remote, "@"):
		// abstract socket, must validate
		return false, validateUNIXName(remote[1:])
	case strings.HasPrefix(remote, "/"):
		// absolute path
		validatePath = true
	case strings.HasSuffix(remote, ".sock"):
		// path ends with .sock
		validatePath = true
	}

	if validatePath {
		return false, validateUNIXName(remote)
	}

	// skip others
	return true, nil
}

func validateUNIXName(name string) error {
	switch {
	case name == "":
		return ErrNameEmpty
	case len(name) > MaxUNIXSocketPathLength:
		return ErrNameTooLong
	case strings.ContainsRune(name, 0):
		return core.ErrInvalid
	default:
		return nil
	}
}

func parseRemoteTCP(remote string) (network, address string, err error) {
	// TCP - validate host:port using darvaza.org/x/net
	host, port, err := net.SplitHostPort(remote)
	switch {
	case err != nil:
		return "", "", err
	case host == "", host == "::", host == "0.0.0.0", host == "0":
		return "", "", fmt.Errorf("%q: host missing", remote)
	case port == 0:
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
