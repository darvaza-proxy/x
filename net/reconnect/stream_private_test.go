package reconnect

import (
	"errors"
	"net"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/workgroup"
)

// TestStreamSessionOnErrorFiresOnInitWindowCancel guards the bridge-
// ordering fix: if the OnError->OnCancel bridge is wired after setDefaults
// realises the workgroup context, a cancellation landing in between makes
// markCancelled read a nil OnCancel and the cause never reaches OnError. The
// streamInitCancelHook seam fires wg.Cancel at exactly that point — the
// moment the context is live; Cancel blocks until the OnCancel handler has
// started, so whether the cause reaches OnError is decided before Cancel
// returns, deterministically rather than as a race.
//
// It fails while the bridge is wired after setDefaults and passes once the
// bridge moves ahead of it, complementing
// TestStreamSessionOnErrorFiresOnParentCancel, which covers the post-Spawn
// cancellation where the bridge is already in place.
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
	core.AssertMustNoError(t, s.Spawn(), "Spawn")

	// the cancellation observed during init must still reach OnError.
	select {
	case err := <-got:
		core.AssertErrorIs(t, err, wantErr, "OnError cause")
	case <-time.After(2 * time.Second):
		t.Fatal("OnError did not fire for an in-init-window cancellation")
	}

	// the session winds down without hanging.
	done := make(chan error, 1)
	go func() { done <- s.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return")
	}
}
