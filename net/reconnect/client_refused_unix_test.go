//go:build !windows

package reconnect_test

import "syscall"

// errConnRefused is the errno a dial to a closed port reports on this
// platform: the POSIX ECONNREFUSED on unix-like systems.
var errConnRefused error = syscall.ECONNREFUSED
