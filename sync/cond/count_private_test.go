package cond

import (
	"context"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/internal/synctesting"
)

// countTestTimeout and countOpenGuard mirror the bounds declared in
// count_test.go; constants cannot be shared across the black-box and
// white-box test packages.
const (
	countTestTimeout = time.Second

	countOpenGuard = 20 * time.Millisecond
)

// waitResultTestCase exercises the IsCancelled and IsContinue accessors
// for every waitResult value, so any future variant added to the enum
// surfaces as an unhandled row.
type waitResultTestCase struct {
	name string

	value waitResult

	wantCancelled bool
	wantContinue  bool
}

func newWaitResultTestCase(name string, value waitResult,
	wantCancelled, wantContinue bool) waitResultTestCase {
	return waitResultTestCase{
		name:          name,
		value:         value,
		wantCancelled: wantCancelled,
		wantContinue:  wantContinue,
	}
}

func (tc waitResultTestCase) Name() string { return tc.name }

func (tc waitResultTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.wantCancelled, tc.value.IsCancelled(),
		"IsCancelled")
	core.AssertEqual(t, tc.wantContinue, tc.value.IsContinue(),
		"IsContinue")
}

var _ core.TestCase = waitResultTestCase{}

func waitResultTestCases() []waitResultTestCase {
	return []waitResultTestCase{
		newWaitResultTestCase("waitContinue",
			waitContinue, false, true),
		newWaitResultTestCase("waitSuccess",
			waitSuccess, false, false),
		newWaitResultTestCase("waitCancelled",
			waitCancelled, true, false),
	}
}

// TestWaitResult verifies the waitResult accessors across every value.
func TestWaitResult(t *testing.T) {
	core.RunTestCases(t, waitResultTestCases())
}

// TestCountWaitFnAbortDuringAcquire verifies WaitFnAbort returns
// context.Canceled when the abort channel closes while the waiter is
// blocked at the barrier Acquire receive. The outer select arm in
// doWaitFnPass1 cannot be reached through the public API alone: the
// happy path always receives a token before observing abort. Holding
// the token externally via TryAcquire on the unexported barrier field
// pins the waiter at the receive long enough for close(abort) to
// drive that arm deterministically.
func TestCountWaitFnAbortDuringAcquire(t *testing.T) {
	c := NewCount(1)
	defer func() { _ = c.Close() }()

	tok, ok := c.b.TryAcquire()
	core.AssertMustTrue(t, ok, "external TryAcquire")
	defer c.b.Release(tok)

	abort := make(chan struct{})
	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitFnAbort(abort, nil)
	}()

	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"WaitFnAbort blocks while token is held externally")

	close(abort)

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFnAbort returned after abort during Acquire")
	core.AssertErrorIs(t, gotErr, context.Canceled,
		"WaitFnAbort returns context.Canceled on abort during Acquire")
}
