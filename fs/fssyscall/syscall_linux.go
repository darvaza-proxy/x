//go:build linux

package fssyscall

import (
	"syscall"
)

// ZeroHandle represents a closed [Handle]
const ZeroHandle Handle = -1

// Handle represents a OS specific file descriptor
type Handle int

// Sys returns the underlying type for syscall operations
func (h Handle) Sys() int { return int(h) }

// LockEx attempts to create an advisory exclusive lock on
// the file associated to the given [Handle].
func LockEx(h Handle) error {
	return syscall.Flock(h.Sys(), syscall.LOCK_EX)
}

// UnlockEx releases an advisory lock on the file associated with the given Handle.
func UnlockEx(h Handle) error {
	return syscall.Flock(h.Sys(), syscall.LOCK_UN)
}
