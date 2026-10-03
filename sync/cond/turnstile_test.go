package cond_test

import (
	"context"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
)

// turnstileTestTimeout caps each wait for a call to return.
//
// turnstileQuietGuard is how long a call must stay blocked for the
// tests to accept that it is waiting.
const (
	turnstileTestTimeout = time.Second

	turnstileQuietGuard = 20 * time.Millisecond
)

func newHeldTurnstile(t *testing.T) *cond.Turnstile {
	t.Helper()
	ts := &cond.Turnstile{}
	core.AssertMustTrue(t, ts.TryLock(), "TryLock")
	return ts
}

func newClosedTurnstile(t *testing.T) *cond.Turnstile {
	t.Helper()
	ts := &cond.Turnstile{}
	core.AssertMustNoError(t, ts.Close(), "Close")
	return ts
}

func newHeldClosedTurnstile(t *testing.T) *cond.Turnstile {
	t.Helper()
	ts := newHeldTurnstile(t)
	core.AssertMustNoError(t, ts.Close(), "Close")
	return ts
}

func newUsedTurnstile(t *testing.T) *cond.Turnstile {
	t.Helper()
	ts := newHeldTurnstile(t)
	ts.Unlock()
	return ts
}

// goWait runs fn on a new goroutine and delivers its result on the
// returned channel.
func goWait(fn func() error) <-chan error {
	done := make(chan error, 1)
	go func() { done <- fn() }()
	return done
}

// assertWaiting states that the call behind done is still blocked.
func assertWaiting(t *testing.T, done <-chan error, name string) {
	t.Helper()
	core.AssertMustQuiet(t, done, turnstileQuietGuard, "%s waiting", name)
}

// assertReturns waits for the call behind done and states that it
// failed with want, or succeeded when want is nil.
func assertReturns(t *testing.T, done <-chan error, want error, name string) {
	t.Helper()
	got := core.AssertMustReceives(t, done, 1, turnstileTestTimeout, name)
	if want == nil {
		core.AssertNoError(t, got[0], name)
	} else {
		core.AssertErrorIs(t, got[0], want, name)
	}
}

// turnstileTryTestCase exercises TryLock and TryPass on a Turnstile in a
// given state. Both go through exactly when the Turnstile is free and
// open, so one expectation covers both.
type turnstileTryTestCase struct {
	setup func(*testing.T) *cond.Turnstile

	name string

	want bool
}

func newTurnstileTryTestCase(name string,
	setup func(*testing.T) *cond.Turnstile, want bool) turnstileTryTestCase {
	return turnstileTryTestCase{
		name:  name,
		setup: setup,
		want:  want,
	}
}

func (tc turnstileTryTestCase) Name() string { return tc.name }

func (tc turnstileTryTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.setup(t).TryLock(), "TryLock")
	core.AssertEqual(t, tc.want, tc.setup(t).TryPass(), "TryPass")
}

var _ core.TestCase = turnstileTryTestCase{}

func turnstileTryTestCases() []turnstileTryTestCase {
	return core.S(
		newTurnstileTryTestCase("zero value",
			func(*testing.T) *cond.Turnstile { return &cond.Turnstile{} },
			true),
		newTurnstileTryTestCase("unlocked", newUsedTurnstile, true),
		newTurnstileTryTestCase("held", newHeldTurnstile, false),
		newTurnstileTryTestCase("closed", newClosedTurnstile, false),
		newTurnstileTryTestCase("held and closed", newHeldClosedTurnstile, false),
	)
}

func TestTurnstileTry(t *testing.T) {
	core.RunTestCases(t, turnstileTryTestCases())
}

func TestTurnstileFree(t *testing.T) {
	t.Run("lock", runTestTurnstileFreeLock)
	t.Run("pass", runTestTurnstileFreePass)
}

func runTestTurnstileFreeLock(t *testing.T) {
	t.Helper()
	ts := &cond.Turnstile{}
	ts.Lock()
	core.AssertFalse(t, ts.TryPass(), "TryPass after Lock")
	ts.Unlock()

	core.AssertNoError(t, ts.LockContext(t.Context()), "LockContext")
	core.AssertFalse(t, ts.TryPass(), "TryPass after LockContext")
	ts.Unlock()
}

func runTestTurnstileFreePass(t *testing.T) {
	t.Helper()
	ts := &cond.Turnstile{}
	core.AssertNoPanic(t, ts.Pass, "Pass")
	core.AssertNoError(t, ts.PassContext(t.Context()), "PassContext")
	core.AssertNoError(t, ts.PassAbort(nil), "PassAbort")
	core.AssertTrue(t, ts.TryLock(), "TryLock after passing")
}

func TestTurnstileWaitsForHolder(t *testing.T) {
	t.Run("pass", runTestTurnstilePassWaits)
	t.Run("lock", runTestTurnstileLockWaits)
	t.Run("pass behind lock", runTestTurnstilePassBehindLock)
}

func runTestTurnstilePassWaits(t *testing.T) {
	t.Helper()
	ts := newHeldTurnstile(t)
	done := goWait(func() error { return ts.PassContext(t.Context()) })

	assertWaiting(t, done, "PassContext")
	ts.Unlock()
	assertReturns(t, done, nil, "PassContext")
	core.AssertTrue(t, ts.TryLock(), "TryLock after pass")
}

func runTestTurnstileLockWaits(t *testing.T) {
	t.Helper()
	ts := newHeldTurnstile(t)
	done := goWait(func() error { return ts.LockContext(t.Context()) })

	assertWaiting(t, done, "LockContext")
	ts.Unlock()
	assertReturns(t, done, nil, "LockContext")
	core.AssertFalse(t, ts.TryPass(), "TryPass while held")

	ts.Unlock()
	core.AssertTrue(t, ts.TryPass(), "TryPass after unlock")
}

func runTestTurnstilePassBehindLock(t *testing.T) {
	t.Helper()
	ts := newHeldTurnstile(t)
	locking := goWait(func() error { return ts.LockContext(t.Context()) })
	assertWaiting(t, locking, "LockContext")
	passing := goWait(func() error { return ts.PassContext(t.Context()) })
	assertWaiting(t, passing, "PassContext")

	ts.Unlock()
	assertReturns(t, locking, nil, "LockContext")
	assertWaiting(t, passing, "PassContext behind the holder")

	ts.Unlock()
	assertReturns(t, passing, nil, "PassContext")
}

func TestTurnstileCancel(t *testing.T) {
	t.Run("lock", runTestTurnstileCancelLock)
	t.Run("pass", runTestTurnstileCancelPass)
	t.Run("abort", runTestTurnstileCancelAbort)
	t.Run("lock deadline", runTestTurnstileDeadlineLock)
	t.Run("pass deadline", runTestTurnstileDeadlinePass)
	t.Run("cancelled before", runTestTurnstileCancelledBefore)
	t.Run("nil context", runTestTurnstileNilContext)
}

// assertCancelWaiting starts a wait on a held Turnstile, cancels it,
// and states that the wait fails with context.Canceled and leaves the
// Turnstile to its holder.
func assertCancelWaiting(t *testing.T,
	wait func(*cond.Turnstile, context.Context) error, name string) {
	t.Helper()
	ts := newHeldTurnstile(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := goWait(func() error { return wait(ts, ctx) })

	assertWaiting(t, done, name)
	cancel()
	assertReturns(t, done, context.Canceled, name)

	ts.Unlock()
	core.AssertTrue(t, ts.TryLock(), "TryLock after cancel")
}

func runTestTurnstileCancelLock(t *testing.T) {
	t.Helper()
	assertCancelWaiting(t, (*cond.Turnstile).LockContext, "LockContext")
}

func runTestTurnstileCancelPass(t *testing.T) {
	t.Helper()
	assertCancelWaiting(t, (*cond.Turnstile).PassContext, "PassContext")
}

func runTestTurnstileCancelAbort(t *testing.T) {
	t.Helper()
	ts := newHeldTurnstile(t)
	abort := make(chan struct{})
	done := goWait(func() error { return ts.PassAbort(abort) })

	assertWaiting(t, done, "PassAbort")
	close(abort)
	assertReturns(t, done, context.Canceled, "PassAbort")

	ts.Unlock()
	core.AssertTrue(t, ts.TryLock(), "TryLock after abort")
}

// assertDeadlineWaiting waits on a held Turnstile until a deadline
// passes, and states that the wait fails with the context's error.
func assertDeadlineWaiting(t *testing.T,
	wait func(*cond.Turnstile, context.Context) error, name string) {
	t.Helper()
	ts := newHeldTurnstile(t)
	ctx, cancel := context.WithTimeout(t.Context(), turnstileQuietGuard)
	defer cancel()

	core.AssertErrorIs(t, wait(ts, ctx), context.DeadlineExceeded, name)
}

func runTestTurnstileDeadlineLock(t *testing.T) {
	t.Helper()
	assertDeadlineWaiting(t, (*cond.Turnstile).LockContext, "LockContext")
}

func runTestTurnstileDeadlinePass(t *testing.T) {
	t.Helper()
	assertDeadlineWaiting(t, (*cond.Turnstile).PassContext, "PassContext")
}

func runTestTurnstileCancelledBefore(t *testing.T) {
	t.Helper()
	ts := &cond.Turnstile{}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	core.AssertErrorIs(t, ts.LockContext(ctx), context.Canceled, "LockContext")
	core.AssertErrorIs(t, ts.PassContext(ctx), context.Canceled, "PassContext")
	core.AssertTrue(t, ts.TryLock(), "TryLock after cancelled calls")
}

func runTestTurnstileNilContext(t *testing.T) {
	t.Helper()
	var nilCtx context.Context
	ts := &cond.Turnstile{}
	core.AssertErrorIs(t, ts.LockContext(nilCtx), errors.ErrNilContext, "LockContext")
	core.AssertErrorIs(t, ts.PassContext(nilCtx), errors.ErrNilContext, "PassContext")
}

func TestTurnstileClose(t *testing.T) {
	t.Run("wakes waiters", runTestTurnstileCloseWakes)
	t.Run("later calls", runTestTurnstileCloseLater)
	t.Run("twice", runTestTurnstileCloseTwice)
	t.Run("nil receiver", runTestTurnstileCloseNil)
}

func runTestTurnstileCloseWakes(t *testing.T) {
	t.Helper()
	ts := newHeldTurnstile(t)
	locking := goWait(func() error { return ts.LockContext(t.Context()) })
	passing := goWait(func() error { return ts.PassContext(t.Context()) })

	assertWaiting(t, locking, "LockContext")
	assertWaiting(t, passing, "PassContext")
	core.AssertNoError(t, ts.Close(), "Close")
	assertReturns(t, locking, errors.ErrClosed, "LockContext")
	assertReturns(t, passing, errors.ErrClosed, "PassContext")

	core.AssertNoPanic(t, ts.Unlock, "Unlock after Close")
}

func runTestTurnstileCloseLater(t *testing.T) {
	t.Helper()
	ts := newClosedTurnstile(t)
	core.AssertErrorIs(t, ts.LockContext(t.Context()), errors.ErrClosed, "LockContext")
	core.AssertErrorIs(t, ts.PassContext(t.Context()), errors.ErrClosed, "PassContext")
	core.AssertErrorIs(t, ts.PassAbort(nil), errors.ErrClosed, "PassAbort")
	core.AssertPanic(t, ts.Lock, errors.ErrClosed, "Lock")
	core.AssertPanic(t, ts.Pass, errors.ErrClosed, "Pass")
}

func runTestTurnstileCloseTwice(t *testing.T) {
	t.Helper()
	ts := newClosedTurnstile(t)
	core.AssertErrorIs(t, ts.Close(), errors.ErrClosed, "second Close")
}

func runTestTurnstileCloseNil(t *testing.T) {
	t.Helper()
	var ts *cond.Turnstile
	core.AssertErrorIs(t, ts.Close(), errors.ErrNilReceiver, "Close")
}

func TestTurnstileMisuse(t *testing.T) {
	t.Run("unlock of unlocked", runTestTurnstileUnlockUnlocked)
	t.Run("nil receiver", runTestTurnstileNilReceiver)
}

func runTestTurnstileUnlockUnlocked(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, (&cond.Turnstile{}).Unlock, errors.ErrNotLocked, "zero value")
	core.AssertPanic(t, newUsedTurnstile(t).Unlock, errors.ErrNotLocked, "unlocked")
}

func runTestTurnstileNilReceiver(t *testing.T) {
	t.Helper()
	var ts *cond.Turnstile
	core.AssertPanic(t, ts.Lock, errors.ErrNilReceiver, "Lock")
	core.AssertPanic(t, ts.Unlock, errors.ErrNilReceiver, "Unlock")
	core.AssertPanic(t, ts.Pass, errors.ErrNilReceiver, "Pass")
	core.AssertPanic(t, func() { ts.TryLock() }, errors.ErrNilReceiver, "TryLock")
	core.AssertPanic(t, func() { ts.TryPass() }, errors.ErrNilReceiver, "TryPass")
	core.AssertErrorIs(t, ts.LockContext(t.Context()), errors.ErrNilReceiver, "LockContext")
	core.AssertErrorIs(t, ts.PassContext(t.Context()), errors.ErrNilReceiver, "PassContext")
	core.AssertErrorIs(t, ts.PassAbort(nil), errors.ErrNilReceiver, "PassAbort")
}

// callLock and the functions after it call one method each, so the stack
// of a panic that starts at the method's caller starts at a function
// [core.AssertTopFrame] can name; a closure is named by its generated
// funcN.
func callLock(ts *cond.Turnstile)    { ts.Lock() }
func callPass(ts *cond.Turnstile)    { ts.Pass() }
func callTryLock(ts *cond.Turnstile) { ts.TryLock() }
func callTryPass(ts *cond.Turnstile) { ts.TryPass() }
func callUnlock(ts *cond.Turnstile)  { ts.Unlock() }

// catchCall calls fn with ts and returns the panic it raises.
func catchCall(fn func(*cond.Turnstile), ts *cond.Turnstile) error {
	return core.Catch(func() error {
		fn(ts)
		return nil
	})
}

func TestTurnstilePanicStack(t *testing.T) {
	var nilTS *cond.Turnstile
	closed := newClosedTurnstile(t)
	unlocked := &cond.Turnstile{}

	core.AssertTopFrame(t, catchCall(callLock, closed), "callLock", "Lock")
	core.AssertTopFrame(t, catchCall(callPass, closed), "callPass", "Pass")
	core.AssertTopFrame(t, catchCall(callTryLock, nilTS), "callTryLock", "TryLock")
	core.AssertTopFrame(t, catchCall(callTryPass, nilTS), "callTryPass", "TryPass")
	core.AssertTopFrame(t, catchCall(callUnlock, nilTS), "callUnlock", "Unlock")
	core.AssertTopFrame(t, catchCall(callUnlock, unlocked), "callUnlock", "Unlock of unlocked")
}
