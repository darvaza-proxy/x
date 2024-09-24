//go:build windows

package fssyscall

import (
	"syscall"
)

// ZeroHandle represents a closed Handle
const ZeroHandle Handle = 0

// Handle represents a OS specific file descriptor
type Handle syscall.Handle

// Sys returns the underlying type for syscall operations
func (h Handle) Sys() syscall.Handle { return syscall.Handle(h) }
