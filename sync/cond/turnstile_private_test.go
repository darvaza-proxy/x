package cond

import (
	"context"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// turnstileReadyRounds is how many times wait runs with its token free
// and another of its cases ready. select picks between them at random,
// so each round reaches the token case with probability 1/2, and the
// chance that no round does is 2^-256.
const turnstileReadyRounds = 256

func newInitialisedTurnstile(tb testing.TB) *Turnstile {
	tb.Helper()
	ts := &Turnstile{}
	if err := ts.lazyInit(); err != nil {
		tb.Fatalf("lazyInit returned error: %v", err)
	}
	return ts
}

func TestTurnstileDoInit(t *testing.T) {
	ts := newInitialisedTurnstile(t)
	core.AssertEqual(t, ts.b.Token(), ts.done, "close signal is the token")
	lock, done := ts.b.Acquire(), ts.done

	// doInit on an initialised Turnstile finds the work done, as a
	// goroutine that lost the race to initialise it would.
	core.AssertNoPanic(t, ts.doInit, "doInit")
	core.AssertEqual(t, lock, ts.b.Acquire(), "lock barrier kept")
	core.AssertEqual(t, done, ts.done, "close signal kept")
}

// assertWaitGivesBack runs wait on ts with its token free while another
// of its cases is ready, and states that it fails with want and leaves
// the token behind, whichever case select picks.
func assertWaitGivesBack(t *testing.T, ts *Turnstile, abort <-chan struct{}, want error) {
	t.Helper()
	for range turnstileReadyRounds {
		tok, err := ts.wait(abort)
		core.AssertErrorIs(t, err, want, "wait")
		core.AssertNil(t, tok, "token")

		tok, ok := ts.b.TryAcquire()
		core.AssertMustTrue(t, ok, "token left behind")
		ts.b.Release(tok)
	}
}

func TestTurnstileWaitClosed(t *testing.T) {
	ts := newInitialisedTurnstile(t)
	core.AssertMustNoError(t, ts.Close(), "Close")
	assertWaitGivesBack(t, ts, nil, errors.ErrClosed)
}

func TestTurnstileWaitCancelled(t *testing.T) {
	ts := newInitialisedTurnstile(t)
	abort := make(chan struct{})
	close(abort)
	assertWaitGivesBack(t, ts, abort, context.Canceled)
}

// BenchmarkTurnstileLazyInit measures lazyInit on a Turnstile that is
// already initialised, the state every call after the first finds it
// in.
func BenchmarkTurnstileLazyInit(b *testing.B) {
	b.Run("serial", runBenchmarkTurnstileLazyInitSerial)
	b.Run("parallel", runBenchmarkTurnstileLazyInitParallel)
}

func runBenchmarkTurnstileLazyInitSerial(b *testing.B) {
	ts := newInitialisedTurnstile(b)

	for b.Loop() {
		if err := ts.lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

func runBenchmarkTurnstileLazyInitParallel(b *testing.B) {
	ts := newInitialisedTurnstile(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := ts.lazyInit(); err != nil {
				b.Errorf("lazyInit returned error: %v", err)
				return
			}
		}
	})
}
