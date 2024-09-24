// Package flock provides a wrapper around syscall.Flock
package flock

import (
	syscall "darvaza.org/x/fs/fssyscall"
)

// LockEx locks a file by name
func LockEx(filename string) (syscall.Handle, error) {
	zero := syscall.ZeroHandle

	h, err := syscall.Open(filename)
	if err != nil {
		return zero, err
	}

	if err = syscall.LockEx(h); err != nil {
		_ = h.Close()
		return zero, err
	}

	return h, nil
}
