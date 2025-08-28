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

// LockEx is a no-op function that simulates an advisory exclusive lock on the file
// associated to the given [Handle].
//
// Note: This function does not provide actual locking functionality on Windows.
// It exists for API compatibility with other platforms.
//
// TODO: implement
func LockEx(Handle) error { return nil }

// UnlockEx is a no-op function that releases an advisory lock on the file associated
// with the given Handle.
//
// Note: This function does not provide actual locking functionality on Windows.
// It exists for API compatibility with other platforms.
//
// TODO: implement
func UnlockEx(Handle) error { return nil }

// TryLockEx is a no-op function that attempts to create an advisory exclusive lock on
// the file associated to the given [Handle] without blocking.
//
// Note: This function does not provide actual locking functionality on Windows.
// It exists for API compatibility with other platforms.
//
// TODO: implement
func TryLockEx(Handle) error { return nil }
