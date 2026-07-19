package reconnect

// cspell:words WSAECONNABORTED WSAECONNREFUSED WSAECONNRESET

import "golang.org/x/sys/windows"

// expectableConnErrorsOS holds the Windows Sockets connection errnos. A
// dial or an established connection reports the WSAE* errnos, not the
// POSIX ECONN* values expectableConnErrors matches: Go's std syscall
// package defines those as invented placeholders on Windows, distinct
// from the numbers the socket layer actually returns. Listing the real
// Winsock codes keeps the recoverable-connection contract holding on
// Windows too.
var expectableConnErrorsOS = []error{
	windows.WSAECONNABORTED,
	windows.WSAECONNREFUSED,
	windows.WSAECONNRESET,
}
