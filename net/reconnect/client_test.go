package reconnect_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
)

// errCollector records every error reported via Config.OnError.
type errCollector struct {
	errs []error
	mu   sync.Mutex
}

func (c *errCollector) OnError(_ context.Context, _ net.Conn, err error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errs = append(c.errs, err)
	return err
}

// Errors returns the recorded errors.
func (c *errCollector) Errors() []error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.errs
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

	collector := new(errCollector)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		// stop after the first session instead of retrying
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),

		OnSession: func(context.Context) error {
			panic(errSessionPanic)
		},
		OnError: collector.OnError,
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNotNil(t, c, "client")

	core.AssertMustNoError(t, c.Connect(), "Connect")

	// the panic is not fatal, so the do-not-reconnect waiter stops
	// the client and Wait reports a user-initiated shutdown.
	core.AssertNoError(t, c.Wait(), "Wait")

	var found *core.PanicError
	for _, err := range collector.Errors() {
		if errors.Is(err, errSessionPanic) {
			core.AssertMustTrue(t, errors.As(err, &found),
				"recovered as PanicError")
		}
	}

	core.AssertNotNil(t, found, "OnSession panic reported via OnError")
}
