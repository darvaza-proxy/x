//go:build !windows

package reconnect

// expectableConnErrorsOS holds the platform-specific recoverable
// connection errors. Unix-like systems surface the POSIX ECONN* errnos
// already listed in expectableConnErrors, so there is nothing further
// to add here.
var expectableConnErrorsOS []error
