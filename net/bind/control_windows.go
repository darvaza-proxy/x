//go:build windows

package bind

import (
	"golang.org/x/sys/windows"
)

func controlSetReuseAddr(fd uintptr) error {
	return windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
}
