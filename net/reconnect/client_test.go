package reconnect_test

import (
	"context"
	"net"
	"testing"

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
