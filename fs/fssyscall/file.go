package fssyscall

import (
	"os"

	"darvaza.org/core"
)

// FLockEx attempts to create an advisory exclusive lock on
// the file associated to the given *os.File.
func FLockEx(f *os.File) error {
	if f == nil {
		return core.ErrInvalid
	}
	h := Handle(f.Fd())
	return LockEx(h)
}

// FUnlockEx releases an advisory lock on the file associated
// with the given *os.File.
func FUnlockEx(f *os.File) error {
	if f == nil {
		return core.ErrInvalid
	}
	h := Handle(f.Fd())
	return UnlockEx(h)
}

// FTryLockEx attempts to create an advisory exclusive lock on
// the file associated to the given *os.File without blocking.
// Returns an error if the lock cannot be acquired immediately.
func FTryLockEx(f *os.File) error {
	if f == nil {
		return core.ErrInvalid
	}
	h := Handle(f.Fd())
	return TryLockEx(h)
}
