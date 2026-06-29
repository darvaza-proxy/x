package reconnect_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
)

// Compile-time verification that the test case type implements TestCase.
var _ core.TestCase = notConnectedTestCase{}

// notConnectedTestCase asserts how a Client connection accessor behaves
// before any connection exists: the deadline and IO accessors report
// ErrNotConnected, while Close is a no-op.
type notConnectedTestCase struct {
	call    func(*reconnect.Client) error
	wantErr error
	name    string
}

func newNotConnectedTestCase(name string,
	call func(*reconnect.Client) error, wantErr error) notConnectedTestCase {
	return notConnectedTestCase{
		name:    name,
		call:    call,
		wantErr: wantErr,
	}
}

func (tc notConnectedTestCase) Name() string { return tc.name }

func (tc notConnectedTestCase) Test(t *testing.T) {
	t.Helper()

	c, err := reconnect.New(&reconnect.Config{
		Context: context.Background(),
		Remote:  addrUnused,
	})
	core.AssertMustNoError(t, err, "New")

	err = tc.call(c)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, tc.name)
	} else {
		core.AssertNoError(t, err, tc.name)
	}
}

func notConnectedTestCases() []notConnectedTestCase {
	return []notConnectedTestCase{
		newNotConnectedTestCase("SetReadDeadline",
			func(c *reconnect.Client) error { return c.SetReadDeadline(time.Second) },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("SetWriteDeadline",
			func(c *reconnect.Client) error { return c.SetWriteDeadline(time.Second) },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("SetDeadline",
			func(c *reconnect.Client) error { return c.SetDeadline(time.Second, time.Second) },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("ResetReadDeadline",
			func(c *reconnect.Client) error { return c.ResetReadDeadline() },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("ResetWriteDeadline",
			func(c *reconnect.Client) error { return c.ResetWriteDeadline() },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("ResetDeadline",
			func(c *reconnect.Client) error { return c.ResetDeadline() },
			reconnect.ErrNotConnected),
		newNotConnectedTestCase("Read",
			func(c *reconnect.Client) error {
				_, err := c.Read(make([]byte, 1))
				return err
			}, reconnect.ErrNotConnected),
		newNotConnectedTestCase("Write",
			func(c *reconnect.Client) error {
				_, err := c.Write([]byte("x"))
				return err
			}, reconnect.ErrNotConnected),
		newNotConnectedTestCase("Close",
			func(c *reconnect.Client) error { return c.Close() }, nil),
	}
}

// TestClientConnNotConnected exercises every connection accessor on a
// Client that was created but never connected.
func TestClientConnNotConnected(t *testing.T) {
	core.RunTestCases(t, notConnectedTestCases())

	// the address accessors report nil rather than erroring when there
	// is no connection.
	c, err := reconnect.New(&reconnect.Config{
		Context: context.Background(),
		Remote:  addrUnused,
	})
	core.AssertMustNoError(t, err, "New")
	core.AssertNil(t, c.RemoteAddr(), "RemoteAddr")
	core.AssertNil(t, c.LocalAddr(), "LocalAddr")
}

// isTimeout reports whether err is a network timeout, the observable
// effect of a read deadline elapsing on the live connection.
func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// TestClientConnLiveSession drives the connection accessors against a
// live session. OnSession runs after the Client stores the dialled
// connection, so the accessors operate on the real socket: a write is
// echoed back through Read, a short read deadline makes the next Read
// time out, and Close tears the connection down. Assertions here run in
// the Client's goroutine, so they use the non-fatal Assert* helpers.
func TestClientConnLiveSession(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	// the peer echoes whatever the client writes.
	go func() {
		conn, err := lsn.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = io.Copy(conn, conn)
	}()

	var c *reconnect.Client
	done := make(chan struct{})

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        lsn.Addr().String(),
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),

		OnSession: func(context.Context) error {
			defer close(done)

			// the accessors see the live connection.
			core.AssertNotNil(t, c.RemoteAddr(), "RemoteAddr")
			core.AssertNotNil(t, c.LocalAddr(), "LocalAddr")

			// SetDeadline with a zero write timeout substitutes the read
			// value, so both deadlines stay generous and the round trip
			// below succeeds within them.
			core.AssertNoError(t, c.SetDeadline(2*time.Second, 0), "SetDeadline")

			// a write is echoed straight back through Read.
			_, err := c.Write([]byte("ping\n"))
			core.AssertNoError(t, err, "Write")

			buf := make([]byte, len("ping\n"))
			_, err = io.ReadFull(c, buf)
			core.AssertNoError(t, err, "Read")
			core.AssertEqual(t, "ping\n", string(buf), "echo")

			// the remaining deadline setters apply cleanly to the live
			// connection.
			core.AssertNoError(t, c.ResetDeadline(), "ResetDeadline")
			core.AssertNoError(t, c.ResetReadDeadline(), "ResetReadDeadline")
			core.AssertNoError(t, c.ResetWriteDeadline(), "ResetWriteDeadline")
			core.AssertNoError(t, c.SetWriteDeadline(2*time.Second), "SetWriteDeadline")

			// a short read deadline makes the next read time out, proving
			// the deadline reached the live connection.
			core.AssertNoError(t, c.SetReadDeadline(50*time.Millisecond),
				"SetReadDeadline")
			_, err = c.Read(buf)
			core.AssertTrue(t, isTimeout(err), "Read times out")

			// Close tears the live connection down.
			core.AssertNoError(t, c.Close(), "Close")

			// the connection stays stored after Close, so applying a
			// deadline now surfaces the socket's error from the
			// read-deadline arm rather than ErrNotConnected.
			core.AssertError(t, c.SetDeadline(time.Second, 0),
				"SetDeadline after Close")
			return nil
		},
	}

	c, err = reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not run")
	}

	core.AssertNoError(t, c.Wait(), "Wait")
}
