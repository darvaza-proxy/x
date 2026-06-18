package reconnect_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

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
	name    string
	session *reconnect.StreamSession[string, string]
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
	got, ok := s.Next()
	core.AssertMustTrue(t, ok, "Next")
	core.AssertEqual(t, "ping", got, "echo")

	// clean shutdown stops the workers and reports the cancellation
	// cause; Wait then returns nil for a user-initiated stop.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	core.AssertErrorIs(t, s.Shutdown(ctx), context.Canceled, "Shutdown")
	core.AssertNoError(t, s.Wait(), "Wait")

	// the inbound channel is closed once the reader stops.
	_, ok = s.Next()
	core.AssertFalse(t, ok, "Next after shutdown")
}

// assertWaitReturns fails if the session does not finish winding down
// within a short deadline, turning a regression that would otherwise
// hang the suite into a clean failure.
func assertWaitReturns(t *testing.T, s *reconnect.StreamSession[string, string],
	name string) {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- s.Wait() }()

	select {
	case err := <-done:
		core.AssertNoError(t, err, name)
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: did not return", name)
	}
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
	_, ok := s.Next()
	core.AssertFalse(t, ok, "Next after EOF")

	// ...and the session winds down without hanging.
	assertWaitReturns(t, s, "Wait after EOF")
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
	<-reading

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	core.AssertErrorIs(t, s.Shutdown(ctx), context.Canceled, "Shutdown")
	assertWaitReturns(t, s, "Wait after shutdown")
}
