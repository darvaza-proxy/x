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

// addrUnused is a placeholder Remote for a [reconnect.Client] that is
// never dialled; only the Config field needs to be populated.
const addrUnused = "127.0.0.1:1"

// addrLoopbackAny asks the OS for any free loopback port, used to bring
// up a real listener the client can dial.
const addrLoopbackAny = "127.0.0.1:0"

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

// acceptAndDrop accepts every connection on lsn and closes it at once,
// until lsn is closed. Sessions that end themselves only need the
// listener backlog kept drained.
func acceptAndDrop(lsn net.Listener) {
	for {
		conn, err := lsn.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}
}

// TestClientOnSessionPanic verifies a panic inside OnSession is
// surfaced through OnError as a recovered error instead of being
// silently swallowed as a clean session end.
func TestClientOnSessionPanic(t *testing.T) {
	errSessionPanic := errors.New("session panic sentinel")

	lsn, err := net.Listen("tcp", addrLoopbackAny)
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
	lsn, err := net.Listen("tcp", addrLoopbackAny)
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

	lsn, err := net.Listen("tcp", addrLoopbackAny)
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

	lsn, err := net.Listen("tcp", addrLoopbackAny)
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

	// bind a port then release it, so the first dial fails
	// non-fatally (ECONNREFUSED) and the run loop reaches the waiter.
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	unreachableAddr := lsn.Addr().String()
	core.AssertMustNoError(t, lsn.Close(), "close listener")

	collector := new(errCollector)

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        unreachableAddr,
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

// TestClientParentCancelCause verifies that cancelling the parent
// context surfaces its cause through Wait and Err. The pre-workgroup
// lifecycle recorded a parent-context cancellation nowhere and
// reported it as nil.
func TestClientParentCancelCause(t *testing.T) {
	errParentCause := errors.New("parent cancel sentinel")

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() {
		_ = lsn.Close()
	}()

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	var once sync.Once
	sessionReady := make(chan struct{})

	cfg := &reconnect.Config{
		Context: ctx,
		Remote:  lsn.Addr().String(),

		OnSession: func(ctx context.Context) error {
			once.Do(func() { close(sessionReady) })
			<-ctx.Done()
			return ctx.Err()
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	// once the session is live, cancel the parent with a custom cause.
	var ready bool
	select {
	case <-sessionReady:
		ready = true
	case <-time.After(2 * time.Second):
	}
	core.AssertMustTrue(t, ready, "session started")
	cancel(errParentCause)

	var stopped bool
	select {
	case <-c.Done():
		stopped = true
	case <-time.After(2 * time.Second):
	}
	core.AssertMustTrue(t, stopped, "stopped after parent cancel")

	core.AssertErrorIs(t, c.Wait(), errParentCause, "Wait")
	core.AssertErrorIs(t, c.Err(), errParentCause, "Err")
}

// TestClientGoAfterShutdownNoop verifies Go and GoCatch are no-ops
// once the client is shut down: the worker is dropped rather than run
// with an already-cancelled context.
func TestClientGoAfterShutdownNoop(t *testing.T) {
	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  addrUnused,
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	// shut down before any work is submitted; with no workers Shutdown
	// returns promptly and reports a clean, user-initiated stop.
	core.AssertNoError(t, c.Shutdown(context.Background()), "Shutdown")

	var ran atomic.Bool
	worker := func(context.Context) error {
		ran.Store(true)
		return nil
	}
	c.Go(worker)
	c.GoCatch(worker, nil)

	// the drop is synchronous, so neither worker can have run.
	core.AssertFalse(t, ran.Load(), "worker after shutdown dropped")
}

// TestClientConnectReturnsClosed verifies Connect reports ErrClosed,
// not the workgroup's internal error, when a shutdown cancels the
// group after the dial has a live connection but before the run loop
// is enrolled. OnConnect provides that window deterministically.
func TestClientConnectReturnsClosed(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() {
		_ = lsn.Close()
	}()

	var c *reconnect.Client

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		OnConnect: func(context.Context, net.Conn) error {
			// cancel the group synchronously, with an already-expired
			// deadline so Shutdown returns at once instead of waiting.
			expired, cancel := context.WithCancel(context.Background())
			cancel()
			_ = c.Shutdown(expired)
			return nil
		},
	}

	c, err = reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	err = c.Connect()
	core.AssertErrorIs(t, err, reconnect.ErrClosed, "Connect")
}

// Compile-time verification that the test case type implements TestCase.
var _ core.TestCase = connectRejectTestCase{}

// connectRejectTestCase drives handleConnectError's cancellation arm:
// OnConnect rejects an established connection with a context
// cancellation, which checkCancellation extracts from the dial error
// so Connect surfaces it rather than treating it as retryable. The
// returned error is the rejection itself, so no separate expectation
// is declared — the invariant is asserted directly.
type connectRejectTestCase struct {
	reject error
	name   string
}

func newConnectRejectTestCase(name string, reject error) connectRejectTestCase {
	return connectRejectTestCase{name: name, reject: reject}
}

func (tc connectRejectTestCase) Name() string { return tc.name }

func (tc connectRejectTestCase) Test(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),
		OnConnect: func(context.Context, net.Conn) error {
			return tc.reject
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	// the cancellation carried by the dial error stops Connect, which
	// surfaces exactly that error.
	core.AssertErrorIs(t, c.Connect(), tc.reject, "Connect")
}

func connectRejectTestCases() []connectRejectTestCase {
	return []connectRejectTestCase{
		newConnectRejectTestCase("cancelled", context.Canceled),
		newConnectRejectTestCase("deadline exceeded", context.DeadlineExceeded),
	}
}

func TestClientConnectReject(t *testing.T) {
	core.RunTestCases(t, connectRejectTestCases())
}

// TestClientConnectStopsWhenContextDone drives handleConnectError's
// context-done arm: OnConnect shuts the client down and then rejects
// the connection with a plain error that carries no cancellation of
// its own. checkCancellation therefore does not fire, so it is the
// already-cancelled client context that stops Connect from looping on
// doomed reconnection attempts.
func TestClientConnectStopsWhenContextDone(t *testing.T) {
	errPlainReject := errors.New("plain reject sentinel")

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	var c *reconnect.Client

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		OnConnect: func(context.Context, net.Conn) error {
			// cancel the client first, with an already-cancelled context
			// so Shutdown returns at once, then reject with a plain
			// error that carries no cancellation itself.
			done, cancel := context.WithCancel(context.Background())
			cancel()
			_ = c.Shutdown(done)
			return errPlainReject
		},
	}

	c, err = reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	// the cancelled context stops Connect; it surfaces the context
	// error rather than the plain rejection or a reconnect spin.
	core.AssertErrorIs(t, c.Connect(), context.Canceled, "Connect")
}
