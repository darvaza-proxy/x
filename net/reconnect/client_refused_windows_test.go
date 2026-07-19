package reconnect_test

// cspell:words WSAECONNREFUSED

import "golang.org/x/sys/windows"

// errConnRefused is the errno a dial to a closed port reports on
// Windows: the Winsock WSAECONNREFUSED. Go's std syscall.ECONNREFUSED is
// an invented placeholder there, not the value the socket layer returns.
var errConnRefused error = windows.WSAECONNREFUSED
