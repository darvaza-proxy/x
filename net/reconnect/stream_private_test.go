package reconnect

import (
	"errors"
	"net"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/workgroup"
)

// streamStopTimeout bounds every wait in this file, so a session that
// never answers fails the test instead of hanging the suite.
const streamStopTimeout = 2 * time.Second

// TestStreamSessionOnErrorFiresOnInitWindowCancel guards the handler
// ordering: if onCancel is wired after setDefaults realises the workgroup
// context, a cancellation landing in between makes markCancelled read a
// nil OnCancel, so the cause never reaches OnError and the session is
// never released. The streamInitCancelHook seam fires wg.Cancel at exactly
// that point — the moment the context is live; Cancel blocks until the
// OnCancel handler has started, so whether the cause reaches OnError is
// decided before Cancel returns, deterministically rather than as a race.
//
// A group cancelled before the workers are enrolled cannot start them, so
// Spawn reports ErrClosed, and the inbound stream it closes on their
// behalf lets a consumer observe the end instead of parking. The test
// complements TestStreamSessionOnErrorFiresOnParentCancel, which covers
// the post-Spawn cancellation where the handler is already in place.
func TestStreamSessionOnErrorFiresOnInitWindowCancel(t *testing.T) {
	wantErr := errors.New("cancelled in init window")

	streamInitCancelHook = func(wg *workgroup.Group) {
		wg.Cancel(wantErr)
	}
	t.Cleanup(func() { streamInitCancelHook = nil })

	c1, c2 := net.Pipe()
	defer func() { _ = c2.Close() }()

	got := make(chan error, 1)
	s := &StreamSession[string, string]{
		Conn:      c1,
		Marshal:   func(v string) ([]byte, error) { return []byte(v + "\n"), nil },
		Unmarshal: func(b []byte) (string, error) { return string(b), nil },
		OnError: func(err error) {
			select {
			case got <- err:
			default:
			}
		},
	}
	core.AssertErrorIs(t, s.Spawn(), ErrClosed, "Spawn")

	// the cancellation observed during init must still reach OnError.
	select {
	case err := <-got:
		core.AssertErrorIs(t, err, wantErr, "OnError cause")
	case <-time.After(streamStopTimeout):
		t.Fatal("OnError did not fire for an in-init-window cancellation")
	}

	// the inbound stream is closed on the never-started reader's behalf,
	// and the session winds down without hanging.
	select {
	case _, ok := <-s.Recv():
		core.AssertFalse(t, ok, "Recv after refused Spawn")
	case <-time.After(streamStopTimeout):
		t.Fatal("Recv did not close after a refused Spawn")
	}

	done := make(chan error, 1)
	go func() { done <- s.Wait() }()
	select {
	case <-done:
	case <-time.After(streamStopTimeout):
		t.Fatal("Wait did not return")
	}
}
