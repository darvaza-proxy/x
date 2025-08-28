//go:build windows

package fssyscall

import (
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	// Windows file locking constants
	reserved = 0
	allBytes = ^uint32(0)
)

// ZeroHandle represents a closed Handle
const ZeroHandle Handle = 0

// Handle represents a OS specific file descriptor
type Handle syscall.Handle

// Sys returns the underlying type for syscall operations
func (h Handle) Sys() syscall.Handle { return syscall.Handle(h) }

// LockEx attempts to create an advisory exclusive lock on
// the file associated to the given [Handle].
func LockEx(h Handle) error {
	ol := &windows.Overlapped{}
	err := windows.LockFileEx(
		windows.Handle(h.Sys()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		reserved,
		allBytes,
		allBytes,
		ol,
	)
	return err
}

// UnlockEx releases an advisory lock on the file associated with the given Handle.
func UnlockEx(h Handle) error {
	ol := &windows.Overlapped{}
	err := windows.UnlockFileEx(
		windows.Handle(h.Sys()),
		reserved,
		allBytes,
		allBytes,
		ol,
	)
	return err
}

// TryLockEx attempts to create an advisory exclusive lock on
// the file associated to the given [Handle] without blocking.
// Returns syscall.EBUSY if the lock cannot be acquired immediately.
func TryLockEx(h Handle) error {
	ol := &windows.Overlapped{}
	err := windows.LockFileEx(
		windows.Handle(h.Sys()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		reserved,
		allBytes,
		allBytes,
		ol,
	)
	if err != nil {
		// Convert Windows lock violation to EBUSY for consistency with Linux
		if err == windows.ERROR_LOCK_VIOLATION {
			return syscall.EBUSY
		}
	}
	return err
}
