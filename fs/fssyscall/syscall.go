// Package fssyscall abstracts syscall file descriptors
// across platform.Handles
package fssyscall

import (
	"os"
	"syscall"
)

// Open performs [syscall.Open] returning a [Handle]
func Open(filename string) (Handle, error) {
	h, err := syscall.Open(filename, os.O_RDONLY, 0)
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
