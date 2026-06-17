package reconnect_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
)

// errCollector records every error reported via Config.OnError.
type errCollector struct {
	mu   sync.Mutex
	errs []error
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

// TestClientShutdownUnblocksSession verifies Shutdown closes the live
// connection so an OnSession parked on a blocking Read unwinds. Without
// it, cancelling the context leaves the read parked, the run loop never
// returns, and Shutdown blocks until its own deadline instead of
// completing cleanly.
func TestClientShutdownUnblocksSession(t *testing.T) {
	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	// accept and hold the peer open without ever writing, so the
	// client's Read blocks until the connection is closed.
	peerDone := make(chan struct{})
	defer close(peerDone)
	go func() {
		conn, err := lsn.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		<-peerDone
	}()

	var c *reconnect.Client
	var once sync.Once
	reading := make(chan struct{})

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		// stop after the session ends rather than redialling.
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),

		OnSession: func(context.Context) error {
			once.Do(func() { close(reading) })
			var buf [1]byte
			_, err := c.Read(buf[:])
			return err
		},
	}

	c, err = reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	// the session is established and about to block on Read.
	<-reading

	// Shutdown's own deadline bounds the wait and, on success, only
	// returns once the workers are done — so it both proves the parked
	// Read was released and fails cleanly (DeadlineExceeded) on a
	// regression rather than hanging the suite on an unbounded Wait.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
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

// TestClientWaiterErrorStops verifies a Waiter error stops the
// client instead of busy-looping. Per the Waiter contract any error
// means "stop reconnecting", even a non-fatal one that IsFatal would
// otherwise let the connection retry. Regression test for the 100%
// CPU spin where a custom non-fatal waiter error skipped the dial
// yet never terminated the client.
func TestClientWaiterErrorStops(t *testing.T) {
	errStopWaiting := errors.New("waiter stop sentinel")

	var waiterCalls atomic.Int32
	waiter := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			waiterCalls.Add(1)
			return errStopWaiting
		}
	}

	// unreachable: bind a port then release it, so the first dial
	// fails non-fatally (ECONNREFUSED) and the run loop reaches the
	// waiter.
	lsn, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	addr := lsn.Addr().String()
	core.AssertMustNoError(t, lsn.Close(), "close listener")

	collector := new(errCollector)

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        addr,
		WaitReconnect: waiter,
		OnError:       collector.OnError,
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	var stopped bool
	select {
	case <-c.Done():
		stopped = true
	case <-time.After(2 * time.Second):
	}

	core.AssertMustTrue(t, stopped, "stopped after waiter error")

	// the waiter is consulted exactly once and its error ends the
	// client; a regression would spin it without bound.
	core.AssertEqual(t, 1, int(waiterCalls.Load()), "waiter calls")
	core.AssertErrorIs(t, c.Wait(), errStopWaiting, "Wait")

	var sawWaiterErr bool
	for _, e := range collector.Errors() {
		if errors.Is(e, errStopWaiting) {
			sawWaiterErr = true
		}
	}
	core.AssertTrue(t, sawWaiterErr, "OnError observed the waiter error")
}
