//go:build linux

package fssyscall

// ZeroHandle represents a closed [Handle]
const ZeroHandle Handle = -1

// Handle represents a OS specific file descriptor
type Handle int

// Sys returns the underlying type for syscall operations
func (h Handle) Sys() int { return int(h) }
