package reconnect_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs"

	"darvaza.org/x/net/reconnect"
)

// Compile-time verification that test case types implement TestCase.
var _ core.TestCase = streamSpawnTestCase{}

func newlineMarshal(s string) ([]byte, error) {
	return []byte(s + "\n"), nil
}

func identityUnmarshal(b []byte) (string, error) {
	return string(b), nil
}

func newStringSession(conn net.Conn) *reconnect.StreamSession[string, string] {
	return &reconnect.StreamSession[string, string]{
		Conn:      conn,
		Marshal:   newlineMarshal,
		Unmarshal: identityUnmarshal,
	}
}

// dummyConn is a non-nil connection for cases where Spawn fails before
// the connection is ever touched.
type dummyConn struct{}

func (dummyConn) Read([]byte) (int, error)    { return 0, io.EOF }
func (dummyConn) Write(b []byte) (int, error) { return len(b), nil }
func (dummyConn) Close() error                { return nil }

// streamSpawnTestCase checks the error Spawn returns for a given
// session state: an incomplete configuration rejected before any
// worker starts, or an already-started session.
type streamSpawnTestCase struct {
	wantErr error
	session *reconnect.StreamSession[string, string]
	name    string
	started bool
}

func newStreamSpawnTestCase(name string,
	session *reconnect.StreamSession[string, string],
	wantErr error) streamSpawnTestCase {
	return streamSpawnTestCase{
		name:    name,
		session: session,
		wantErr: wantErr,
	}
}

// newStreamSpawnTestCaseStarted builds a fully configured session that
// is Spawned once before the asserted Spawn, so the second call hits
// init's "already started" arm and returns fs.ErrExist.
func newStreamSpawnTestCaseStarted(name string,
	session *reconnect.StreamSession[string, string]) streamSpawnTestCase {
	return streamSpawnTestCase{
		name:    name,
		session: session,
		wantErr: fs.ErrExist,
		started: true,
	}
}

func (tc streamSpawnTestCase) Name() string {
	return tc.name
}

func (tc streamSpawnTestCase) Test(t *testing.T) {
	t.Helper()

	if tc.started {
		core.AssertMustNoError(t, tc.session.Spawn(), "initial Spawn")
		t.Cleanup(func() {
			_ = tc.session.Close()
			_ = tc.session.Wait()
		})
	}

	core.AssertErrorIs(t, tc.session.Spawn(), tc.wantErr, "Spawn")
}

func streamSpawnTestCases() []streamSpawnTestCase {
	return []streamSpawnTestCase{
		newStreamSpawnTestCase("missing Conn",
			&reconnect.StreamSession[string, string]{
				Unmarshal: identityUnmarshal,
				Marshal:   newlineMarshal,
			}, fs.ErrInvalid),
		newStreamSpawnTestCase("missing Unmarshal",
			&reconnect.StreamSession[string, string]{
				Conn:    dummyConn{},
				Marshal: newlineMarshal,
			}, fs.ErrInvalid),
		newStreamSpawnTestCase("missing Marshal and MarshalTo",
			&reconnect.StreamSession[string, string]{
				Conn:      dummyConn{},
				Unmarshal: identityUnmarshal,
			}, fs.ErrInvalid),
		newStreamSpawnTestCaseStarted("already started",
			&reconnect.StreamSession[string, string]{
				Conn:      dummyConn{},
				Unmarshal: identityUnmarshal,
				Marshal:   newlineMarshal,
			}),
	}
}

func TestStreamSessionSpawn(t *testing.T) {
	core.RunTestCases(t, streamSpawnTestCases())
}

// TestStreamSessionBeforeSpawn verifies every guarded method panics
// via mustStarted rather than dereferencing the nil work group or
// blocking on a nil channel when called before Spawn.
func TestStreamSessionBeforeSpawn(t *testing.T) {
	fresh := func() *reconnect.StreamSession[string, string] {
		return &reconnect.StreamSession[string, string]{}
	}

	core.AssertPanic(t, func() { fresh().Go(nil) }, fs.ErrInvalid, "Go")
	core.AssertPanic(t, func() { fresh().GoCatch(nil, nil) }, fs.ErrInvalid, "GoCatch")
	core.AssertPanic(t, func() { _ = fresh().Close() }, fs.ErrInvalid, "Close")
	core.AssertPanic(t, func() { _ = fresh().Shutdown(context.Background()) },
		fs.ErrInvalid, "Shutdown")
	core.AssertPanic(t, func() { _ = fresh().Wait() }, fs.ErrInvalid, "Wait")
	core.AssertPanic(t, func() { _ = fresh().Done() }, fs.ErrInvalid, "Done")
	core.AssertPanic(t, func() { _ = fresh().Err() }, fs.ErrInvalid, "Err")
	core.AssertPanic(t, func() { _ = fresh().Send("x") }, fs.ErrInvalid, "Send")
	core.AssertPanic(t, func() { _ = fresh().Recv() }, fs.ErrInvalid, "Recv")
	core.AssertPanic(t, func() { fresh().Next() }, fs.ErrInvalid, "Next")
}

// nextWithin bounds a blocking Next with a deadline, turning a
// regression that would otherwise hang the suite into a clean failure.
// It returns Next's result; asserting on ok stays with the caller.
func nextWithin(t *testing.T, s *reconnect.StreamSession[string, string],
	name string) (string, bool) {
	t.Helper()

	type result struct {
		value string
		ok    bool
	}

	res := make(chan result, 1)
	go func() {
		v, ok := s.Next()
		res <- result{value: v, ok: ok}
	}()

	got := core.AssertMustReceives(t, res, 1, waitTimeout, name)
	return got[0].value, got[0].ok
}

// echoPeer copies conn's inbound byte stream straight back to it until
// conn is closed, echoing every frame the session writes.
func echoPeer(conn net.Conn) {
	_, _ = io.Copy(conn, conn)
}

// TestStreamSessionEcho drives a full round trip through an echo peer
// and a clean shutdown, exercising the migrated workgroup.Group
// lifecycle end to end.
func TestStreamSessionEcho(t *testing.T) {
	c1, c2 := net.Pipe()
	defer func() { _ = c2.Close() }()

	go echoPeer(c2)

	s := newStringSession(c1)
	core.AssertMustNoError(t, s.Spawn(), "Spawn")

	// a second Spawn is rejected.
	core.AssertErrorIs(t, s.Spawn(), fs.ErrExist, "second Spawn")

	// round trip: the peer echoes what we send back to us.
	core.AssertMustNoError(t, s.Send("ping"), "Send")
	got, ok := nextWithin(t, s, "Next")
	core.AssertMustTrue(t, ok, "Next")
	core.AssertEqual(t, "ping", got, "echo")

	// clean shutdown stops the workers and reports the cancellation
	// cause; Wait then returns nil for a user-initiated stop.
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	core.AssertErrorIs(t, s.Shutdown(ctx), context.Canceled, "Shutdown")
	core.AssertNoError(t, s.Wait(), "Wait")

	// the inbound channel is closed once the reader stops.
	_, ok = s.Next()
	core.AssertFalse(t, ok, "Next after shutdown")

	// Send after shutdown fails with the package sentinel.
	core.AssertErrorIs(t, s.Send("late"), reconnect.ErrClosed,
		"Send after shutdown")
}

// waitWithin bounds a blocking Wait with a deadline, turning a
// regression that would otherwise hang the suite into a clean failure.
// It returns Wait's result; asserting on it stays with the caller.
func waitWithin(t *testing.T, s *reconnect.StreamSession[string, string],
	name string) error {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- s.Wait() }()

	got := core.AssertMustReceives(t, done, 1, waitTimeout, name)
	return got[0]
}

// assertWaitReturns fails if the session does not finish winding down
// cleanly within a short deadline.
func assertWaitReturns(t *testing.T, s *reconnect.StreamSession[string, string],
	name string) {
	t.Helper()

	core.AssertNoError(t, waitWithin(t, s, name), name)
}

// shutdownWithin shuts s down, bounded by a short deadline, and asserts
// the cause it reports is the user-initiated cancellation.
func shutdownWithin(t *testing.T, s *reconnect.StreamSession[string, string]) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	core.AssertErrorIs(t, s.Shutdown(ctx), context.Canceled, "Shutdown")
}

// newSpawnedSession returns a spawned session against a peer that holds
// its end open until cleanup, and the channel its OnError reports on.
func newSpawnedSession(t *testing.T) (*reconnect.StreamSession[string, string],
	<-chan error) {
	t.Helper()

	c1, c2 := net.Pipe()
	t.Cleanup(func() { _ = c2.Close() })

	got := make(chan error, 1)
	s := newStringSession(c1)
	s.OnError = func(err error) {
		select {
		case got <- err:
		default:
		}
	}
	core.AssertMustNoError(t, s.Spawn(), "Spawn")
	return s, got
}

// TestStreamSessionGo covers the session's own Go and GoCatch on a
// live session: the worker runs under the session's context, and its
// error ends the session unless a catch absorbs it.
func TestStreamSessionGo(t *testing.T) {
	t.Run("worker runs", runTestStreamSessionGoRuns)
	t.Run("worker error ends the session", runTestStreamSessionGoFails)
	t.Run("catch absorbs the error", runTestStreamSessionGoCatchAbsorbs)
}

// runTestStreamSessionGoRuns runs a worker under Go and checks it
// received a live context, then shuts the session down cleanly.
func runTestStreamSessionGoRuns(t *testing.T) {
	t.Helper()

	s, _ := newSpawnedSession(t)

	ran := make(chan error, 1)
	s.Go(func(ctx context.Context) error {
		ran <- ctx.Err()
		return nil
	})

	errs := core.AssertMustReceives(t, ran, 1, waitTimeout, "worker ran")
	core.AssertNoError(t, errs[0], "worker context")

	shutdownWithin(t, s)
	assertWaitReturns(t, s, "Wait")
}

// runTestStreamSessionGoFails runs a worker under Go that fails. With
// no catch the error cancels the session: OnError receives it and
// Wait reports it as the cause.
func runTestStreamSessionGoFails(t *testing.T) {
	t.Helper()

	errWorker := errors.New("worker failed")
	s, got := newSpawnedSession(t)

	s.Go(func(context.Context) error {
		return errWorker
	})

	errs := core.AssertMustReceives(t, got, 1, waitTimeout, "OnError")
	core.AssertErrorIs(t, errs[0], errWorker, "OnError cause")

	core.AssertErrorIs(t, waitWithin(t, s, "Wait"), errWorker, "Wait cause")
}

// runTestStreamSessionGoCatchAbsorbs runs a failing worker under
// GoCatch with a catch that absorbs the error. The catch sees it, the
// session stays live, and the later shutdown is clean.
func runTestStreamSessionGoCatchAbsorbs(t *testing.T) {
	t.Helper()

	errWorker := errors.New("worker failed")
	s, got := newSpawnedSession(t)

	caught := make(chan error, 1)
	s.GoCatch(func(context.Context) error {
		return errWorker
	}, func(_ context.Context, err error) error {
		caught <- err
		return nil
	})

	errs := core.AssertMustReceives(t, caught, 1, waitTimeout, "catch ran")
	core.AssertErrorIs(t, errs[0], errWorker, "caught")
	core.AssertNoError(t, s.Err(), "Err")

	shutdownWithin(t, s)
	assertWaitReturns(t, s, "Wait")

	// OnError saw the shutdown, not the absorbed error.
	errs = core.AssertMustReceives(t, got, 1, waitTimeout, "OnError")
	core.AssertNotErrorIs(t, errs[0], errWorker, "OnError cause")
}

// TestStreamSessionReaderEOF verifies a clean remote EOF ends the
// session: the reader winds the group down so Wait returns and Recv
// closes, instead of leaving the writer and kill watchers parked.
func TestStreamSessionReaderEOF(t *testing.T) {
	c1, c2 := net.Pipe()
	s := newStringSession(c1)
	core.AssertMustNoError(t, s.Spawn(), "Spawn")

	// the peer disconnects; the reader observes a clean EOF.
	core.AssertMustNoError(t, c2.Close(), "peer Close")

	// the inbound channel closes...
	_, ok := nextWithin(t, s, "Next after EOF")
	core.AssertFalse(t, ok, "Next after EOF")

	// ...and the session winds down without hanging.
	assertWaitReturns(t, s, "Wait after EOF")
}

// TestStreamSessionOnErrorFiresOnParentCancel verifies the OnError
// bridge wired in init: the workgroup enrols OnCancel as a task driven
// by context.AfterFunc, so cancelling the parent context forwards the
// cancellation cause to the configured OnError hook.
func TestStreamSessionOnErrorFiresOnParentCancel(t *testing.T) {
	c1, c2 := net.Pipe()
	defer func() { _ = c2.Close() }()

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	got := make(chan error, 1)
	s := newStringSession(c1)
	s.Context = ctx
	s.OnError = func(err error) {
		select {
		case got <- err:
		default:
		}
	}
	core.AssertMustNoError(t, s.Spawn(), "Spawn")

	// cancelling the parent context drives the group's OnCancel
	// transition, which forwards the cause to OnError.
	wantErr := errors.New("parent cancelled")
	cancel(wantErr)

	errs := core.AssertMustReceives(t, got, 1, waitTimeout, "OnError")
	core.AssertErrorIs(t, errs[0], wantErr, "OnError cause")

	// the cause recorded by the parent cancellation is what Wait
	// surfaces once the workers wind down.
	core.AssertErrorIs(t, waitWithin(t, s, "Wait"), wantErr, "Wait cause")
}

// TestStreamSessionShutdownUnblocksReader verifies a shutdown frees a
// reader parked on the unbuffered inbound channel. The peer sends a
// frame the consumer never drains; closing the connection alone would
// not release the pending send, so the reader must also observe the
// cancelled context.
func TestStreamSessionShutdownUnblocksReader(t *testing.T) {
	c1, c2 := net.Pipe()
	defer func() { _ = c2.Close() }()

	// UnsetReadDeadline fires after a frame is scanned and before it is
	// delivered, so it signals that the reader is about to block on the
	// undrained channel.
	reading := make(chan struct{}, 1)
	s := newStringSession(c1)
	s.UnsetReadDeadline = func() error {
		select {
		case reading <- struct{}{}:
		default:
		}
		return nil
	}
	core.AssertMustNoError(t, s.Spawn(), "Spawn")

	go func() { _, _ = c2.Write([]byte("stuck\n")) }()
	core.AssertMustReceives(t, reading, 1, waitTimeout,
		"reader at delivery point")

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	core.AssertErrorIs(t, s.Shutdown(ctx), context.Canceled, "Shutdown")
	assertWaitReturns(t, s, "Wait after shutdown")
}
