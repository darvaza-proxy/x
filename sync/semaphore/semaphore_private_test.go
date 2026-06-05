package semaphore

import (
	"testing"

	"darvaza.org/core"
)

// TestSemaphore_Init exercises the unexported lazyInit and the channel
// fields it populates. It lives in the white-box test package because
// neither lazyInit nor the global/readers channels are part of the
// public surface.
func TestSemaphore_Init(t *testing.T) {
	t.Run("nil receiver", runTestInitNilReceiver)
	t.Run("initialise channels", runTestInitChannels)
	t.Run("idempotent", runTestInitIdempotent)
}

func runTestInitNilReceiver(t *testing.T) {
	t.Helper()
	var s *Semaphore
	core.AssertError(t, s.lazyInit(), "lazyInit on nil receiver")
}

func runTestInitChannels(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")
	core.AssertNotNil(t, s.global, "global channel")
	core.AssertNotNil(t, s.readers, "readers channel")
}

func runTestInitIdempotent(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "first lazyInit")

	global, readers := s.global, s.readers

	core.AssertMustNoError(t, s.lazyInit(), "second lazyInit")
	core.AssertSame(t, global, s.global, "global channel unchanged")
	core.AssertSame(t, readers, s.readers, "readers channel unchanged")
}

// cancelledRLockRounds is the number of cancelled unsafeRLock attempts run
// per scenario. select chooses uniformly between the ready acquire case and
// the already-closed abort, so each post-acquire rollback arm is reached with
// probability 1/2 per round; enough rounds drive the miss probability
// (2^-rounds) to nothing without depending on a single lucky scheduling.
const cancelledRLockRounds = 256

// TestSemaphore_UnsafeRLockCancelled pins the two post-acquire rollback arms
// of unsafeRLock: a reader that wins its slot and only then observes the
// abort must release what it took. These live in the white-box test package
// because unsafeRLock and the channel fields are unexported, and because the
// arms are unreachable from the public RLockContext tests, which always hold
// a lock before cancelling and so never present an idle or reader-occupied
// slot as the ready acquire case.
func TestSemaphore_UnsafeRLockCancelled(t *testing.T) {
	t.Run("first reader", runTestUnsafeRLockCancelledFirst)
	t.Run("subsequent reader", runTestUnsafeRLockCancelledSubsequent)
}

// runTestUnsafeRLockCancelledFirst exercises the "won the global slot as
// first reader, then noticed the abort" arm. On an idle semaphore both the
// global send and the closed abort are ready; whichever wins, unsafeRLock
// reports cancellation and leaves the slot empty, so the receiver stays idle
// across rounds.
func runTestUnsafeRLockCancelledFirst(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")

	abort := closedAbort()
	for range cancelledRLockRounds {
		core.AssertMustEqual(t, true, s.unsafeRLock(abort), "cancelled")
		core.AssertMustEqual(t, 0, len(s.global), "global released")
	}
}

// runTestUnsafeRLockCancelledSubsequent exercises the "took the next reader's
// slot, then noticed the abort" arm. With one reader already holding, the
// readers channel — not the global send — is the ready acquire case, so the
// rollback puts the count back. Each round restores the held reader, so the
// receiver state is stable across rounds.
func runTestUnsafeRLockCancelledSubsequent(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")

	// Seed one reader so the readers channel is the ready acquire case.
	core.AssertMustEqual(t, false, s.unsafeRLock(nil), "seed reader")

	abort := closedAbort()
	for range cancelledRLockRounds {
		core.AssertMustEqual(t, true, s.unsafeRLock(abort), "cancelled")
		core.AssertMustEqual(t, 1, len(s.readers), "reader count restored")
		core.AssertMustEqual(t, 1, len(s.global), "global still held")
	}
}

// closedAbort returns an already-cancelled abort channel: every receive on it
// succeeds immediately, so isCancelled reports true.
func closedAbort() <-chan struct{} {
	abort := make(chan struct{})
	close(abort)
	return abort
}
