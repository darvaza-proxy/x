// Package flock provides a wrapper around syscall.Flock
package flock

import (
	"os"

	syscall "darvaza.org/x/fs/fssyscall"
)

// LockEx locks a file by name
func LockEx(filename string) (syscall.Handle, error) {
	const zero = syscall.ZeroHandle

	h, err := syscall.Open(filename, os.O_RDONLY, 0)
	if err != nil {
		return zero, err
	}

	err = syscall.LockEx(h)
	if err != nil {
		_ = h.Close()
		return zero, err
	}

	return h, nil
}
