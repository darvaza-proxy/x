package reconnect

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
)

// TestClientRunSessionSkipsCancelledSession reproduces the publish-after-
// cancel window: a cancellation landing between the dial and setConn is
// consumed by the cancel watcher against a not-yet-stored connection, so
// the live connection it stores next is never closed and an OnSession parked
// on a read hangs to the Shutdown deadline. The runSessionBeforeSetConn seam
// cancels the parent context and blocks on c.ctx.Done(), so the cancellation
// has landed before setConn runs — deterministically, without depending on
// the watcher's timing.
//
// The invariant asserted is that runSession does not enter OnSession with an
// already-cancelled context: it fails while runSession runs OnSession
// unconditionally and passes once the post-setConn guard winds the session
// down instead.
func TestClientRunSessionSkipsCancelledSession(t *testing.T) {
	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	// accept every dial and drop it; the session is expected never to
	// start, so the listener only needs its backlog drained.
	go func() {
		for {
			conn, err := lsn.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	var entered atomic.Bool

	// fire in the dial-to-setConn window: cancel the parent, then block
	// until the derived context observes it, so the guard under test sees a
	// cancelled context the moment runSession reaches it.
	runSessionBeforeSetConn = func(ctx context.Context) {
		cancel()
		<-ctx.Done()
	}
	t.Cleanup(func() { runSessionBeforeSetConn = nil })

	cfg := &Config{
		Context: parent,
		Remote:  lsn.Addr().String(),

		// stop after the first session rather than redialling.
		WaitReconnect: NewDoNotReconnectWaiter(nil),

		OnSession: func(context.Context) error {
			entered.Store(true)
			return nil
		},
	}

	c, err := New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	// the cancelled context winds the client down.
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("client did not stop")
	}

	// OnSession must not run once the context is already cancelled.
	core.AssertFalse(t, entered.Load(),
		"OnSession entered with an already-cancelled context")
}
