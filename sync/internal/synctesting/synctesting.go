package synctesting

import (
	"time"

	"darvaza.org/core"
)

// PollStep is the default polling cadence used by AssertEventually and
// AssertMustEventually. Callers that need a different cadence should call
// WaitForCond directly with an explicit step.
const PollStep = time.Millisecond

// WaitForCond polls predicate at step intervals until it returns true or
// timeout elapses. Reports whether predicate became true within the budget.
// The predicate is evaluated before the deadline is checked, so a predicate
// that is already true returns true even when the timeout has elapsed.
//
// WaitForCond does not interact with any test framework; use it when a
// caller needs the boolean outcome to feed into a follow-up decision.
func WaitForCond(predicate func() bool, timeout, step time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if predicate() {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(step)
	}
}

// AssertEventually waits up to timeout for predicate to hold, then asserts
// the outcome via core.AssertTrue. Returns true if predicate held within
// timeout, false otherwise.
func AssertEventually(t core.T, predicate func() bool, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	return core.AssertTrue(t,
		WaitForCond(predicate, timeout, PollStep), name, args...)
}

// AssertMustEventually is AssertEventually that aborts the test on failure.
func AssertMustEventually(t core.T, predicate func() bool,
	timeout time.Duration, name string, args ...any) {
	t.Helper()
	if !AssertEventually(t, predicate, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertClosed asserts that ch becomes readable within timeout — either
// closed or receiving a value. Intended for close-to-signal cancellation
// channels, where a successful receive proves the close happened. Returns
// true on success, false on timeout.
func AssertClosed[T any](t core.T, ch <-chan T, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	select {
	case <-ch:
		return core.AssertTrue(t, true, name, args...)
	case <-time.After(timeout):
		return core.AssertTrue(t, false, name, args...)
	}
}

// AssertMustClosed is AssertClosed that aborts the test on failure.
func AssertMustClosed[T any](t core.T, ch <-chan T, timeout time.Duration,
	name string, args ...any) {
	t.Helper()
	if !AssertClosed(t, ch, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertOpen asserts that ch neither closes nor produces a value within
// timeout. Returns true on success (still open at deadline), false on
// premature close or receive.
func AssertOpen[T any](t core.T, ch <-chan T, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	select {
	case <-ch:
		return core.AssertTrue(t, false, name, args...)
	case <-time.After(timeout):
		return core.AssertTrue(t, true, name, args...)
	}
}

// AssertMustOpen is AssertOpen that aborts the test on failure.
func AssertMustOpen[T any](t core.T, ch <-chan T, timeout time.Duration,
	name string, args ...any) {
	t.Helper()
	if !AssertOpen(t, ch, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertReadersReady asserts that n values arrive on ch within timeout.
// Returns true on success, false on timeout or on close-before-n. The
// timeout is shared across the whole collection — it does not reset
// between receives. A closed channel counts as failure: zero values
// received after close are not signals.
//
//revive:disable-next-line:argument-limit
func AssertReadersReady[T any](t core.T, ch <-chan T, n int,
	timeout time.Duration, name string, args ...any) bool {
	t.Helper()
	deadline := time.After(timeout)
	for range n {
		select {
		case _, ok := <-ch:
			if !ok {
				return core.AssertTrue(t, false, name, args...)
			}
		case <-deadline:
			return core.AssertTrue(t, false, name, args...)
		}
	}
	return core.AssertTrue(t, true, name, args...)
}

// AssertMustReadersReady is AssertReadersReady that aborts the test on
// failure.
//
//revive:disable-next-line:argument-limit
func AssertMustReadersReady[T any](t core.T, ch <-chan T, n int,
	timeout time.Duration, name string, args ...any) {
	t.Helper()
	if !AssertReadersReady(t, ch, n, timeout, name, args...) {
		t.FailNow()
	}
}
