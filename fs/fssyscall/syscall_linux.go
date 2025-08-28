//go:build linux

package fssyscall

import (
	"syscall"

	"darvaza.org/core"
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

// TryLockEx attempts to create an advisory exclusive lock on
// the file associated to the given [Handle] without blocking.
// Returns syscall.EBUSY if the lock cannot be acquired immediately.
func TryLockEx(h Handle) error {
	err := syscall.Flock(h.Sys(), syscall.LOCK_EX|syscall.LOCK_NB)
	if core.IsError(err, syscall.EAGAIN, syscall.EWOULDBLOCK) {
		return syscall.EBUSY
	}
	return err
}
