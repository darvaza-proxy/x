package reconnect

import (
	"net"
	"testing"

	"darvaza.org/core"
)

// TestUnsafeClose lives in the white-box test package because
// unsafeClose is unexported and has no public equivalent to drive its
// nil-safe, variadic behaviour through.
func TestUnsafeClose(t *testing.T) {
	t.Run("nil is a no-op", runTestUnsafeCloseNil)
	t.Run("closes the connections", runTestUnsafeCloseConn)
}

func runTestUnsafeCloseNil(t *testing.T) {
	t.Helper()

	// a nil closer must be a no-op, not a panic: Connect closes the
	// conn on a failed worker spawn even after a non-fatal dial miss
	// left it nil.
	core.AssertNoPanic(t, func() { unsafeClose(nil) }, "unsafeClose(nil)")
}

func runTestUnsafeCloseConn(t *testing.T) {
	t.Helper()

	c1, c2 := net.Pipe()

	// a nil closer mixed into the variadic list is skipped; the real
	// connections still close.
	unsafeClose(c1, nil, c2)

	// both ends are closed: a write now fails on either.
	_, err := c1.Write([]byte("x"))
	core.AssertError(t, err, "write after close (c1)")
	_, err = c2.Write([]byte("x"))
	core.AssertError(t, err, "write after close (c2)")
}
