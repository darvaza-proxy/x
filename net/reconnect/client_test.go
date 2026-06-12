package reconnect_test

import (
	"context"
	"net"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
	"darvaza.org/x/sync/errors"
)

// newOnError returns a Config.OnError hook that records every error it
// receives into errs, so a test can assert on them once the client stops.
// errs is the concurrency-safe accumulator; the hook may fire from several
// client goroutines at once.
func newOnError(
	errs *errors.CompoundError,
) func(context.Context, net.Conn, error) error {
	return func(_ context.Context, _ net.Conn, err error) error {
		_ = errs.AppendError(err)
		return err
	}
}

// TestClientOnSessionPanic verifies a panic inside OnSession is
// surfaced through OnError as a recovered error instead of being
// silently swallowed as a clean session end.
func TestClientOnSessionPanic(t *testing.T) {
	errSessionPanic := errors.New("session panic sentinel")

	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	defer func() {
		_ = lsn.Close()
	}()

	errs := new(errors.CompoundError)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		// stop after the first session instead of retrying
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),

		OnSession: func(context.Context) error {
			panic(errSessionPanic)
		},
		OnError: newOnError(errs),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNotNil(t, c, "client")

	core.AssertMustNoError(t, c.Connect(), "Connect")

	// the panic is not fatal, so the do-not-reconnect waiter stops
	// the client and Wait reports a user-initiated shutdown.
	core.AssertNoError(t, c.Wait(), "Wait")

	// the sentinel only reaches OnError if the panic was recovered and
	// surfaced there; a swallowed panic leaves errs empty. errors.Is unwraps
	// the CompoundError to inspect every error it collected.
	core.AssertErrorIs(t, errs.AsError(), errSessionPanic,
		"OnSession panic reported via OnError")
}

// runTestClientConnectExpiredContext verifies Connect refuses to
// start when the context deadline has already expired.
func runTestClientConnectExpiredContext(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	defer func() {
		_ = lsn.Close()
	}()

	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(-time.Second))
	defer cancel()

	cfg := &reconnect.Config{
		Context: ctx,
		Remote:  lsn.Addr().String(),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	err = c.Connect()
	core.AssertError(t, err, "Connect")
	core.AssertErrorIs(t, err, context.DeadlineExceeded, "Connect")
}

// runTestClientDeadlineStopsReconnecting verifies the client stops
// instead of spinning on reconnection attempts once the context
// deadline expires.
func runTestClientDeadlineStopsReconnecting(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	defer func() {
		_ = lsn.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(),
		150*time.Millisecond)
	defer cancel()

	cfg := &reconnect.Config{
		Context: ctx,
		Remote:  lsn.Addr().String(),

		OnSession: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	var stopped bool
	select {
	case <-c.Done():
		stopped = true
	case <-time.After(5 * time.Second):
	}

	core.AssertMustTrue(t, stopped, "stopped after context deadline")
}

func TestClientContextDeadline(t *testing.T) {
	t.Run("expired at Connect", runTestClientConnectExpiredContext)
	t.Run("expires during session", runTestClientDeadlineStopsReconnecting)
}
