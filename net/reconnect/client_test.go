package reconnect_test

import (
	"context"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"

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

// addrUnused is a placeholder Remote for a [reconnect.Client] that is
// never dialled; only the Config field needs to be populated.
const addrUnused = "127.0.0.1:1"

// addrLoopbackAny asks the OS for any free loopback port, used to bring
// up a real listener the client can dial.
const addrLoopbackAny = "127.0.0.1:0"

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

// TestClientShutdownUnblocksSession verifies Shutdown closes the live
// connection so an OnSession parked on a blocking Read unwinds. Without
// it, cancelling the context leaves the read parked, the run loop never
// returns, and Shutdown blocks until its own deadline instead of
// completing cleanly.
func TestClientShutdownUnblocksSession(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	// the peer never writes, so the client's Read blocks until the
	// connection is closed.
	peerDone := make(chan struct{})
	defer close(peerDone)
	go holdPeer(lsn, peerDone)

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

	errs := new(errors.CompoundError)

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        unreachableAddr,
		WaitReconnect: waiter,
		OnError:       newOnError(errs),
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

	// the waiter error must also reach OnError; errors.Is unwraps the
	// CompoundError to inspect every error the hook collected.
	core.AssertErrorIs(t, errs.AsError(), errStopWaiting,
		"OnError observed the waiter error")
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
// with an already-cancelled context, and the drop is recorded at debug
// level so a late submission is not lost silently.
func TestClientGoAfterShutdownNoop(t *testing.T) {
	logger := mock.NewLogger()

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  addrUnused,
		Logger:  logger,
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	// shut down before any work is submitted; with no workers Shutdown
	// returns promptly and reports a clean, user-initiated stop.
	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")

	var ran atomic.Bool
	worker := func(context.Context) error {
		ran.Store(true)
		return nil
	}
	logger.Clear()
	c.Go(worker)
	c.GoCatch(worker, nil)

	// the drop is synchronous, so neither worker can have run.
	core.AssertFalse(t, ran.Load(), "worker after shutdown dropped")

	// one debug record per dropped submission, carrying the
	// workgroup's refusal as its error field.
	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 2, len(msgs), "dropped records")
	for i, msg := range msgs {
		core.AssertEqual(t, slog.Debug, msg.Level, "record %d level", i)
		field := core.AssertMustTypeIs[error](t, msg.Fields[slog.ErrorFieldName],
			"record %d error field", i)
		core.AssertErrorIs(t, field, errors.ErrClosed, "record %d cause", i)
	}
}

// TestClientGo covers the worker paths of Go and GoCatch on a live
// client: the outcome reaches the catch first and then the error
// hooks, and only a fatal result stops the client.
func TestClientGo(t *testing.T) {
	t.Run("clean worker", runTestClientGoClean)
	t.Run("absorbed error", runTestClientGoAbsorbed)
	t.Run("worker panics", runTestClientGoWorkerPanic)
	t.Run("catch panics", runTestClientGoCatchPanic)
	t.Run("catch rewrites to fatal", runTestClientGoCatchFatal)
}

// newIdleClient returns a client that is never connected, so its
// workgroup is live for Go and GoCatch without a remote to dial.
func newIdleClient(t *testing.T, errs *errors.CompoundError) *reconnect.Client {
	t.Helper()

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  addrUnused,
		OnError: newOnError(errs),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	return c
}

// runTestClientGoClean runs a worker that succeeds and leaves the
// client running.
func runTestClientGoClean(t *testing.T) {
	t.Helper()

	errs := new(errors.CompoundError)
	c := newIdleClient(t, errs)

	ran := make(chan struct{})
	c.Go(func(context.Context) error {
		close(ran)
		return nil
	})
	assertClosedWithin(t, ran, "worker ran")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
	core.AssertTrue(t, errs.OK(), "errors")
}

// runTestClientGoAbsorbed runs a worker that fails with a plain error
// the catch passes through. The error reaches OnError and is otherwise
// absorbed: Shutdown reports a clean stop rather than the worker's
// error, which a terminated client would have recorded as its cause.
func runTestClientGoAbsorbed(t *testing.T) {
	t.Helper()

	errWorker := errors.New("worker failed")
	errs := new(errors.CompoundError)
	c := newIdleClient(t, errs)

	seen := make(chan struct{})
	c.GoCatch(func(context.Context) error {
		return errWorker
	}, func(_ context.Context, err error) error {
		defer close(seen)
		return err
	})
	assertClosedWithin(t, seen, "catch ran")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
	core.AssertErrorIs(t, errs.AsError(), errWorker, "OnError")
}

// runTestClientGoWorkerPanic runs a worker that panics. The recovered
// panic takes the same route as a returned error: it reaches the catch
// and then OnError, and is absorbed as non-fatal.
func runTestClientGoWorkerPanic(t *testing.T) {
	t.Helper()

	errPanic := errors.New("worker panic sentinel")
	errs := new(errors.CompoundError)
	c := newIdleClient(t, errs)

	var caught error
	seen := make(chan struct{})
	c.GoCatch(func(context.Context) error {
		panic(errPanic)
	}, func(_ context.Context, err error) error {
		defer close(seen)
		caught = err
		return err
	})
	assertClosedWithin(t, seen, "catch ran")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
	core.AssertErrorIs(t, caught, errPanic, "caught")
	core.AssertErrorIs(t, errs.AsError(), errPanic, "OnError")
}

// runTestClientGoCatchPanic runs a catch that panics on a failing
// worker. The recovered panic replaces the worker's error and takes the
// worker's own route: it reaches OnError and is absorbed as non-fatal
// instead of the workgroup cancelling the client on it.
func runTestClientGoCatchPanic(t *testing.T) {
	t.Helper()

	errWorker := errors.New("worker failed")
	errPanic := errors.New("catch panic sentinel")
	errs := new(errors.CompoundError)
	c := newIdleClient(t, errs)

	seen := make(chan struct{})
	c.GoCatch(func(context.Context) error {
		return errWorker
	}, func(context.Context, error) error {
		close(seen)
		panic(errPanic)
	})
	assertClosedWithin(t, seen, "catch ran")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
	core.AssertErrorIs(t, errs.AsError(), errPanic, "OnError")
	core.AssertNotErrorIs(t, errs.AsError(), errWorker, "OnError")
}

// runTestClientGoCatchFatal runs a worker whose plain failure the
// catch rewrites into ErrDoNotReconnect. The rewritten error is what
// the client acts on: it stops, as a fatal error demands, and reports
// the stop as user-initiated.
func runTestClientGoCatchFatal(t *testing.T) {
	t.Helper()

	errWorker := errors.New("worker failed")
	errs := new(errors.CompoundError)
	c := newIdleClient(t, errs)

	var caught error
	c.GoCatch(func(context.Context) error {
		return errWorker
	}, func(_ context.Context, err error) error {
		caught = err
		return reconnect.ErrDoNotReconnect
	})
	assertClosedWithin(t, c.Done(), "client stopped")

	core.AssertErrorIs(t, caught, errWorker, "caught")
	core.AssertNoError(t, c.Wait(), "Wait")
	core.AssertErrorIs(t, errs.AsError(), reconnect.ErrDoNotReconnect,
		"OnError")
	core.AssertNotErrorIs(t, errs.AsError(), errWorker, "OnError")
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

// newReconnectWaiter returns a Waiter that permits the first `permit`
// reconnect attempts and then returns stop, alongside a counter of how
// many times it was consulted. A cancelled context always wins.
func newReconnectWaiter(permit int32, stop error) (
	func(context.Context) error, *atomic.Int32) {
	calls := new(atomic.Int32)
	fn := func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if calls.Add(1) <= permit {
			return nil
		}
		return stop
	}
	return fn, calls
}

// clientStopTimeout bounds how long a stopped client may take to close
// its Done channel before the wait is treated as a hang.
const clientStopTimeout = 2 * time.Second

// assertStopped fails unless the client's Done channel closes before
// ctx expires, turning a stuck client into a clean failure rather than
// a hung suite.
func assertStopped(ctx context.Context, t *testing.T, c *reconnect.Client) {
	t.Helper()

	select {
	case <-c.Done():
	case <-ctx.Done():
		t.Fatal("client did not stop in time")
	}
}

// countMatchingErrors reports how many errors in errs match target.
func countMatchingErrors(errs []error, target error) int {
	var n int
	for _, e := range errs {
		if errors.Is(e, target) {
			n++
		}
	}
	return n
}

// TestClientReconnectDialFails drives the reconnect-dial failure path:
// the waiter permits one reconnect attempt, so tryReconnect proceeds to
// the dial, the dial fails, and the failure is routed through
// handleReconnectError before the next waiter call stops the client.
// This is the complement of TestClientWaiterErrorStops, where the
// waiter errors on its first call and the dial is never reached.
func TestClientReconnectDialFails(t *testing.T) {
	errStopWaiting := errors.New("waiter stop sentinel")

	// the waiter permits exactly one reconnect attempt, then stops the
	// client. The permitted attempt dials the unreachable address, so
	// tryReconnect routes that dial failure through handleReconnectError
	// before the second waiter call ends the loop.
	waiter, waiterCalls := newReconnectWaiter(1, errStopWaiting)

	// bind a port then release it, so every dial fails non-fatally
	// (connection refused): the synchronous Connect dial and the
	// reconnect dial alike.
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	unreachableAddr := lsn.Addr().String()
	core.AssertMustNoError(t, lsn.Close(), "close listener")

	errs := new(errors.CompoundError)

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        unreachableAddr,
		WaitReconnect: waiter,
		OnError:       newOnError(errs),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	assertStopped(ctx, t, c)

	// two waiter calls: one permitting the failed reconnect dial, one
	// stopping the loop. A single call would mean the dial — and so
	// handleReconnectError — was never reached.
	core.AssertEqual(t, 2, int(waiterCalls.Load()), "waiter calls")
	core.AssertErrorIs(t, c.Wait(), errStopWaiting, "Wait")

	// both dials are refused and reported: the synchronous Connect dial
	// and the reconnect dial routed through handleReconnectError.
	core.AssertEqual(t, 2,
		countMatchingErrors(errs.Errors(), errConnRefused),
		"refused dials reported")
}

// TestClientReconnectSucceeds drives the reconnect-success path: the
// waiter permits one reconnect, the redial succeeds against the live
// listener, and a second session runs before the waiter stops the
// client. This exercises tryReconnect's ready return and the run
// loop's re-entry with a freshly dialled connection.
func TestClientReconnectSucceeds(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	// the peer accepts every dial and drops it; OnSession ends each
	// session itself, so the listener only needs its backlog drained.
	go acceptAndDrop(lsn)

	errStopWaiting := errors.New("waiter stop sentinel")
	waiter, waiterCalls := newReconnectWaiter(1, errStopWaiting)

	var sessions atomic.Int32

	cfg := &reconnect.Config{
		Context:       context.Background(),
		Remote:        lsn.Addr().String(),
		WaitReconnect: waiter,
		OnSession: func(context.Context) error {
			sessions.Add(1)
			return nil
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	assertStopped(ctx, t, c)

	// the permitted waiter call let tryReconnect redial successfully,
	// so a second session ran before the waiter stopped the client.
	core.AssertEqual(t, 2, int(sessions.Load()), "sessions")
	core.AssertEqual(t, 2, int(waiterCalls.Load()), "waiter calls")
	core.AssertErrorIs(t, c.Wait(), errStopWaiting, "Wait")
}

// TestClientConnectFatalRejection drives handleConnectError's fatal
// arm: OnConnect rejects an established connection with the only fatal
// sentinel, ErrDoNotReconnect. dial surfaces it as the connect error,
// so Connect terminates the client and returns it unfiltered instead
// of retrying.
func TestClientConnectFatalRejection(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	errs := new(errors.CompoundError)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		OnConnect: func(context.Context, net.Conn) error {
			return reconnect.ErrDoNotReconnect
		},
		OnError: newOnError(errs),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")

	// the fatal rejection is returned by Connect itself, not retried.
	core.AssertErrorIs(t, c.Connect(), reconnect.ErrDoNotReconnect, "Connect")

	// the client is terminated, and the fatal error was observed by
	// OnError on its way through handlePossiblyFatalError.
	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	assertStopped(ctx, t, c)
	core.AssertEqual(t, 1,
		countMatchingErrors(errs.Errors(), reconnect.ErrDoNotReconnect),
		"fatal rejection reported")
}

// TestClientDisconnectPanic drives run's recovery arm: a panic from
// OnDisconnect escapes doOnDisconnect, which unlike OnSession has no
// catcher, so it unwinds to run's deferred recover. The recover
// surfaces it through OnError and terminates the client instead of
// leaking the panic out of the worker goroutine.
func TestClientDisconnectPanic(t *testing.T) {
	errDisconnectPanic := errors.New("disconnect panic sentinel")

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	errs := new(errors.CompoundError)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		// the session ends at once; the panic comes from the disconnect
		// hook, which runs outside the session catcher.
		OnSession: func(context.Context) error {
			return nil
		},
		OnDisconnect: func(context.Context, net.Conn) error {
			panic(errDisconnectPanic)
		},
		OnError: newOnError(errs),
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	assertStopped(ctx, t, c)

	// the recovered panic is reported through OnError as a PanicError
	// and recorded as the termination cause. The type match matters:
	// it proves the error travelled through run's deferred recover.
	var found *core.PanicError
	core.AssertMustTrue(t, errors.As(errs.AsError(), &found),
		"OnDisconnect panic reported via OnError")
	core.AssertErrorIs(t, found, errDisconnectPanic, "panic payload")
	core.AssertErrorIs(t, c.Wait(), errDisconnectPanic, "Wait")
}

// holdPeer accepts one connection on lsn and holds it open without
// ever writing, until done is closed. A session against it stays live
// instead of ending at once and driving a reconnect.
func holdPeer(lsn net.Listener, done <-chan struct{}) {
	conn, err := lsn.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	<-done
}

// assertClosedWithin fails unless ch closes before the client-stop
// timeout, turning an event that never happens into a clean failure
// naming it rather than a hung suite.
func assertClosedWithin(t *testing.T, ch <-chan struct{}, name string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(clientStopTimeout):
		t.Fatalf("%s: not observed in time", name)
	}
}

// TestClientConnectOnce covers the entry guard. Connect starts the
// client exactly once; a later call is refused without dialling again,
// as ErrRunning while the client is live and as ErrClosed once it has
// been shut down, whether or not it ever ran.
func TestClientConnectOnce(t *testing.T) {
	t.Run("while running", runTestClientConnectWhileRunning)
	t.Run("after shutdown", runTestClientConnectAfterShutdown)
	t.Run("never started", runTestClientConnectNeverStarted)
}

// runTestClientConnectNeverStarted refuses a first Connect against a
// client shut down before it ever ran. The guard answers ErrClosed
// before the one-shot check, so no dial is attempted.
func runTestClientConnectNeverStarted(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	dials := new(atomic.Int32)

	c, err := reconnect.New(&reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		OnConnect: func(context.Context, net.Conn) error {
			dials.Add(1)
			return nil
		},
	})
	core.AssertMustNoError(t, err, "New")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")

	core.AssertErrorIs(t, c.Connect(), reconnect.ErrClosed, "Connect")
	core.AssertEqual(t, int32(0), dials.Load(), "dials")
}

// runTestClientConnectWhileRunning refuses a second Connect against a
// live client. The session is held open so the call lands while the
// client is genuinely running, which is what ErrRunning reports.
func runTestClientConnectWhileRunning(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	peerDone := make(chan struct{})
	defer close(peerDone)
	go holdPeer(lsn, peerDone)

	dials := new(atomic.Int32)
	established := make(chan struct{})

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		OnConnect: func(context.Context, net.Conn) error {
			dials.Add(1)
			return nil
		},
		OnSession: func(ctx context.Context) error {
			close(established)
			<-ctx.Done()
			return nil
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	assertClosedWithin(t, established, "session established")

	core.AssertErrorIs(t, c.Connect(), reconnect.ErrRunning, "second Connect")

	// the guard returns before the dial, so the refused call leaves the
	// remote untouched.
	core.AssertEqual(t, int32(1), dials.Load(), "dials")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
}

// runTestClientConnectAfterShutdown refuses a Connect against a client
// that ran and has since stopped. Shutdown waits for the workers, so
// the client is stopped, not stopping, and ErrClosed is the answer
// rather than ErrRunning.
func runTestClientConnectAfterShutdown(t *testing.T) {
	t.Helper()

	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	go acceptAndDrop(lsn)

	dials := new(atomic.Int32)

	cfg := &reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),

		// stop after the first session instead of retrying, so the
		// dial count below cannot grow before Shutdown lands.
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),

		OnConnect: func(context.Context, net.Conn) error {
			dials.Add(1)
			return nil
		},
	}

	c, err := reconnect.New(cfg)
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	ctx, cancel := core.WithTimeout(context.Background(), clientStopTimeout)
	defer cancel()
	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")

	core.AssertErrorIs(t, c.Connect(), reconnect.ErrClosed,
		"Connect after Shutdown")
	core.AssertEqual(t, int32(1), dials.Load(), "dials")
}

// assertLogged fails unless logger recorded a message at level whose
// text contains want.
func assertLogged(t *testing.T, logger *mock.Logger, level slog.LogLevel, want string) {
	t.Helper()

	for _, msg := range logger.GetMessages() {
		if msg.Level == level && strings.Contains(msg.Message, want) {
			return
		}
	}
	t.Errorf("no %v record containing %q", level, want)
}

// TestClientNoSessionHandler covers a client configured without
// OnSession. Nobody owns the connection, so the session is recorded as
// connected at info level and ended at once by the client itself, the
// disconnection recorded the same way; the peer never closes its end.
func TestClientNoSessionHandler(t *testing.T) {
	lsn, err := net.Listen("tcp", addrLoopbackAny)
	core.AssertMustNoError(t, err, "listen")
	defer func() { _ = lsn.Close() }()

	peerDone := make(chan struct{})
	defer close(peerDone)
	go holdPeer(lsn, peerDone)

	logger := mock.NewLogger()
	c, err := reconnect.New(&reconnect.Config{
		Context: context.Background(),
		Remote:  lsn.Addr().String(),
		Logger:  logger,

		// stop after the first session instead of retrying.
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),
	})
	core.AssertMustNoError(t, err, "New")
	core.AssertMustNoError(t, c.Connect(), "Connect")

	// the session ends on its own and the waiter then stops the client.
	assertClosedWithin(t, c.Done(), "client stopped")
	core.AssertNoError(t, c.Wait(), "Wait")

	assertLogged(t, logger, slog.Info, "connected")
	assertLogged(t, logger, slog.Info, "disconnected")
}
