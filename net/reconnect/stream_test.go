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
// via mustStarted rather than dereferencing the nil work group when
// called before Spawn. Send is intentionally excluded: it is not
// guarded (a known bug) and would block on a nil channel.
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
