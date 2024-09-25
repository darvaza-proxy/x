// Package fssyscall abstracts syscall file descriptors
// across platform.Handles
package fssyscall

import (
	"io/fs"
	"syscall"
)

// Open performs [syscall.Open] returning a [Handle]
func Open(filename string, mode int, perm fs.FileMode) (Handle, error) {
	h, err := syscall.Open(filename, mode, uint32(perm))
	if err != nil {
		return ZeroHandle, err
	}
	return Handle(h), nil
}

// Close attempts to close the file associated to the [Handle]
func (h Handle) Close() error {
	return syscall.Close(h.Sys())
}

// IsZero indicates the [Handle] doesn't refer to a possibly valid
// descriptor
func (h Handle) IsZero() bool {
	return h <= ZeroHandle
}
