package reconnect

import (
	"errors"
	"net"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/workgroup"
)

// waitTimeout bounds every wait on the code under test in the white-box
// tests, so a regression that never answers fails the test instead of
// hanging the suite.
const waitTimeout = 2 * time.Second

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
	errs := core.AssertMustReceives(t, got, 1, waitTimeout, "OnError")
	core.AssertErrorIs(t, errs[0], wantErr, "OnError cause")

	// the inbound stream is closed on the never-started reader's behalf,
	// with nothing delivered: the first thing a consumer sees is the end.
	// The goroutine bounds the receive, which blocks until that end
	// arrives.
	ended := make(chan bool, 1)
	go func() {
		_, ok := <-s.Recv()
		ended <- ok
	}()
	gotOK := core.AssertMustReceives(t, ended, 1, waitTimeout,
		"Recv returned")
	core.AssertFalse(t, gotOK[0], "Recv after refused Spawn")

	done := make(chan error, 1)
	go func() { done <- s.Wait() }()
	core.AssertMustReceives(t, done, 1, waitTimeout, "Wait")
}
