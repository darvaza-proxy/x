//go:build !windows

package bind

import (
	"golang.org/x/sys/unix"
)

func controlSetReuseAddr(fd uintptr) error {
	return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}
