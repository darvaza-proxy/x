package mutex_test

import (
	"context"
	"sync"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/mutex"
)

// errKind classifies the error a Safe* operation is expected to return,
// keeping the test tables free of per-row assertion logic.
type errKind int

const (
	errNone       errKind = iota // no error
	errNilMutex                  // wraps errors.ErrNilMutex
	errNilContext                // wraps errors.ErrNilContext
	errCancelled                 // wraps context.Canceled
	errPanic                     // a caught panic, reported as *core.PanicError
)

// assertErrKind asserts err matches the expected classification. The errPanic
// arm also pins the typed-nil contract: a caught nil-pointer dereference is a
// PanicError, never the ErrNilMutex sentinel reserved for interface nils.
func assertErrKind(t *testing.T, err error, kind errKind, name string) {
	t.Helper()
	switch kind {
	case errNone:
		core.AssertNoError(t, err, name)
	case errNilMutex:
		core.AssertErrorIs(t, err, errors.ErrNilMutex, name)
	case errNilContext:
		core.AssertErrorIs(t, err, errors.ErrNilContext, name)
	case errCancelled:
		core.AssertErrorIs(t, err, context.Canceled, name)
	case errPanic:
		_, _ = core.AssertTypeIs[*core.PanicError](t, err, name)
		core.AssertFalse(t, core.IsError(err, errors.ErrNilMutex),
			name+" not ErrNilMutex")
	default: // any other non-nil error
		core.AssertError(t, err, name)
	}
}

// ----- mutex fixtures -----

func newStdMutex() mutex.Mutex     { return &sync.Mutex{} }
func newRWMutex() mutex.Mutex      { return &sync.RWMutex{} }
func nilMutex() mutex.Mutex        { return nil }
func typedNilMutex() mutex.Mutex   { return (*sync.Mutex)(nil) }
func typedNilRWMutex() mutex.Mutex { return (*sync.RWMutex)(nil) }

func newPanicOnLock() mutex.Mutex    { return &panicOnLockMutex{} }
func newPanicOnTryLock() mutex.Mutex { return &panicOnTryLockMutex{} }
func newPanicOnUnlock() mutex.Mutex  { return &panicOnUnlockMutex{} }

func lockedStdMutex() mutex.Mutex {
	m := &sync.Mutex{}
	m.Lock()
	return m
}

func readLockedRWMutex() mutex.Mutex {
	m := &sync.RWMutex{}
	m.RLock()
	return m
}

// writeLockedRWMutex returns a write-locked RWMutex. It exists only to make
// SafeTryRLock fail without blocking; the lock is never released, so it must
// not be handed to a release helper.
func writeLockedRWMutex() mutex.Mutex {
	m := &sync.RWMutex{}
	m.Lock()
	return m
}

// ----- Safe acquire -----

func opSafeLock(mu mutex.Mutex) (bool, error)     { return mutex.SafeLock(mu) }
func opSafeTryLock(mu mutex.Mutex) (bool, error)  { return mutex.SafeTryLock(mu) }
func opSafeRLock(mu mutex.Mutex) (bool, error)    { return mutex.SafeRLock(mu) }
func opSafeTryRLock(mu mutex.Mutex) (bool, error) { return mutex.SafeTryRLock(mu) }

// safeAcquireTestCase exercises the (bool, error)-returning Safe acquirers.
// op selects the operation; newMu builds the receiver in the desired state.
// wantOK is independent of want: the non-blocking TryLock/TryRLock acquirers
// return (false, nil) on a contended lock, so a held row pairs wantOK false
// with want errNone — the blocking acquirers cannot reach that combination.
type safeAcquireTestCase struct {
	name string

	op    func(mutex.Mutex) (bool, error)
	newMu func() mutex.Mutex

	want   errKind
	wantOK bool
}

func newSafeAcquireTestCase(name string, op func(mutex.Mutex) (bool, error),
	newMu func() mutex.Mutex, wantOK bool, want errKind) safeAcquireTestCase {
	return safeAcquireTestCase{
		name:   name,
		op:     op,
		newMu:  newMu,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc safeAcquireTestCase) Name() string { return tc.name }

func (tc safeAcquireTestCase) Test(t *testing.T) {
	t.Helper()
	ok, err := tc.op(tc.newMu())
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	assertErrKind(t, err, tc.want, "err")
}

var _ core.TestCase = safeAcquireTestCase{}

func safeAcquireTestCases() []safeAcquireTestCase {
	return []safeAcquireTestCase{
		newSafeAcquireTestCase("SafeLock free", opSafeLock, newStdMutex, true, errNone),
		newSafeAcquireTestCase("SafeLock interface-nil", opSafeLock, nilMutex, false, errNilMutex),
		newSafeAcquireTestCase("SafeLock typed-nil", opSafeLock, typedNilMutex, false, errPanic),
		newSafeAcquireTestCase("SafeLock panic", opSafeLock, newPanicOnLock, false, errPanic),

		newSafeAcquireTestCase("SafeTryLock free", opSafeTryLock, newStdMutex, true, errNone),
		newSafeAcquireTestCase("SafeTryLock held", opSafeTryLock, lockedStdMutex, false, errNone),
		newSafeAcquireTestCase("SafeTryLock interface-nil", opSafeTryLock, nilMutex, false, errNilMutex),
		newSafeAcquireTestCase("SafeTryLock typed-nil", opSafeTryLock, typedNilMutex, false, errPanic),
		newSafeAcquireTestCase("SafeTryLock panic", opSafeTryLock, newPanicOnTryLock, false, errPanic),

		newSafeAcquireTestCase("SafeRLock RWMutex", opSafeRLock, newRWMutex, true, errNone),
		newSafeAcquireTestCase("SafeRLock Mutex", opSafeRLock, newStdMutex, true, errNone),
		newSafeAcquireTestCase("SafeRLock interface-nil", opSafeRLock, nilMutex, false, errNilMutex),
		newSafeAcquireTestCase("SafeRLock typed-nil", opSafeRLock, typedNilRWMutex, false, errPanic),

		newSafeAcquireTestCase("SafeTryRLock RWMutex", opSafeTryRLock, newRWMutex, true, errNone),
		newSafeAcquireTestCase("SafeTryRLock Mutex", opSafeTryRLock, newStdMutex, true, errNone),
		newSafeAcquireTestCase("SafeTryRLock write-locked", opSafeTryRLock, writeLockedRWMutex, false, errNone),
		newSafeAcquireTestCase("SafeTryRLock interface-nil", opSafeTryRLock, nilMutex, false, errNilMutex),
		newSafeAcquireTestCase("SafeTryRLock typed-nil", opSafeTryRLock, typedNilRWMutex, false, errPanic),
	}
}

func TestSafeAcquire(t *testing.T) {
	core.RunTestCases(t, safeAcquireTestCases())
}

// TestSafeTryLockHeld verifies SafeTryLock's acquisition has real effect:
// a successful acquire makes a second attempt on the same mutex fail
// without blocking, and the lock releases cleanly afterwards. The table's
// "SafeTryLock held" row covers the (false, nil) outcome on a pre-locked
// mutex; this pins that SafeTryLock itself is what holds it.
func TestSafeTryLockHeld(t *testing.T) {
	mu := &sync.Mutex{}
	ok, err := mutex.SafeTryLock[mutex.Mutex](mu)
	core.AssertMustEqual(t, true, ok, "first acquire")
	core.AssertNoError(t, err, "first acquire")

	ok, err = mutex.SafeTryLock[mutex.Mutex](mu)
	core.AssertEqual(t, false, ok, "second acquire blocked")
	core.AssertNoError(t, err, "second acquire")

	mu.Unlock()
}

// ----- Safe release -----

func opSafeUnlock(mu mutex.Mutex) error  { return mutex.SafeUnlock(mu) }
func opSafeRUnlock(mu mutex.Mutex) error { return mutex.SafeRUnlock(mu) }

// safeReleaseTestCase exercises the error-returning Safe releasers against a
// receiver pre-arranged by newMu.
type safeReleaseTestCase struct {
	name string

	op    func(mutex.Mutex) error
	newMu func() mutex.Mutex

	want errKind
}

func newSafeReleaseTestCase(name string, op func(mutex.Mutex) error,
	newMu func() mutex.Mutex, want errKind) safeReleaseTestCase {
	return safeReleaseTestCase{
		name:  name,
		op:    op,
		newMu: newMu,
		want:  want,
	}
}

func (tc safeReleaseTestCase) Name() string { return tc.name }

func (tc safeReleaseTestCase) Test(t *testing.T) {
	t.Helper()
	assertErrKind(t, tc.op(tc.newMu()), tc.want, "err")
}

var _ core.TestCase = safeReleaseTestCase{}

func safeReleaseTestCases() []safeReleaseTestCase {
	return []safeReleaseTestCase{
		newSafeReleaseTestCase("SafeUnlock locked", opSafeUnlock, lockedStdMutex, errNone),
		newSafeReleaseTestCase("SafeUnlock interface-nil", opSafeUnlock, nilMutex, errNilMutex),
		newSafeReleaseTestCase("SafeUnlock typed-nil", opSafeUnlock, typedNilMutex, errPanic),
		newSafeReleaseTestCase("SafeUnlock panic", opSafeUnlock, newPanicOnUnlock, errPanic),

		newSafeReleaseTestCase("SafeRUnlock read-locked RWMutex", opSafeRUnlock, readLockedRWMutex, errNone),
		newSafeReleaseTestCase("SafeRUnlock locked Mutex", opSafeRUnlock, lockedStdMutex, errNone),
		newSafeReleaseTestCase("SafeRUnlock interface-nil", opSafeRUnlock, nilMutex, errNilMutex),
		newSafeReleaseTestCase("SafeRUnlock typed-nil", opSafeRUnlock, typedNilRWMutex, errPanic),
	}
}

func TestSafeRelease(t *testing.T) {
	core.RunTestCases(t, safeReleaseTestCases())
}

// ----- Safe context acquire -----

func opSafeLockContext(ctx context.Context, mu mutex.MutexContext) (bool, error) {
	return mutex.SafeLockContext(ctx, mu)
}

func opSafeRLockContext(ctx context.Context, mu mutex.MutexContext) (bool, error) {
	return mutex.SafeRLockContext(ctx, mu)
}

func bgCtx() context.Context  { return context.Background() }
func nilCtx() context.Context { return nil }

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func newMutexContext() mutex.MutexContext        { return &testMutexContext{} }
func newRWMutexContext() mutex.MutexContext      { return &testRWMutexContext{} }
func nilMutexContext() mutex.MutexContext        { return nil }
func typedNilMutexContext() mutex.MutexContext   { return (*testMutexContext)(nil) }
func typedNilRWMutexContext() mutex.MutexContext { return (*testRWMutexContext)(nil) }
func newPanicLockContext() mutex.MutexContext    { return &panicOnLockContextMutex{} }
func newPanicRLockContext() mutex.MutexContext   { return &panicOnRLockContextMutex{} }

// safeContextTestCase exercises the context-aware Safe acquirers across
// context and receiver states.
type safeContextTestCase struct {
	name string

	op     func(context.Context, mutex.MutexContext) (bool, error)
	newCtx func() context.Context
	newMu  func() mutex.MutexContext

	want errKind
}

func newSafeContextTestCase(name string,
	op func(context.Context, mutex.MutexContext) (bool, error),
	newCtx func() context.Context, newMu func() mutex.MutexContext,
	want errKind) safeContextTestCase {
	return safeContextTestCase{
		name:   name,
		op:     op,
		newCtx: newCtx,
		newMu:  newMu,
		want:   want,
	}
}

func (tc safeContextTestCase) Name() string { return tc.name }

func (tc safeContextTestCase) Test(t *testing.T) {
	t.Helper()
	ok, err := tc.op(tc.newCtx(), tc.newMu())
	// The context acquirers have no non-blocking path, so ok is exactly
	// err == nil. Assert that invariant on the returned values rather than
	// carrying a wantOK column that would merely restate want.
	core.AssertEqual(t, err == nil, ok, "ok tracks err")
	assertErrKind(t, err, tc.want, "err")
}

var _ core.TestCase = safeContextTestCase{}

func safeContextTestCases() []safeContextTestCase {
	return []safeContextTestCase{
		newSafeContextTestCase("SafeLockContext valid", opSafeLockContext,
			bgCtx, newMutexContext, errNone),
		newSafeContextTestCase("SafeLockContext cancelled", opSafeLockContext,
			cancelledCtx, newMutexContext, errCancelled),
		newSafeContextTestCase("SafeLockContext nil ctx", opSafeLockContext,
			nilCtx, newMutexContext, errNilContext),
		newSafeContextTestCase("SafeLockContext nil mu", opSafeLockContext,
			bgCtx, nilMutexContext, errNilMutex),
		newSafeContextTestCase("SafeLockContext typed-nil mu", opSafeLockContext,
			bgCtx, typedNilMutexContext, errPanic),
		newSafeContextTestCase("SafeLockContext panic", opSafeLockContext,
			bgCtx, newPanicLockContext, errPanic),

		newSafeContextTestCase("SafeRLockContext RWMutex", opSafeRLockContext,
			bgCtx, newRWMutexContext, errNone),
		newSafeContextTestCase("SafeRLockContext Mutex", opSafeRLockContext,
			bgCtx, newMutexContext, errNone),
		newSafeContextTestCase("SafeRLockContext nil ctx", opSafeRLockContext,
			nilCtx, newRWMutexContext, errNilContext),
		newSafeContextTestCase("SafeRLockContext nil mu", opSafeRLockContext,
			bgCtx, nilMutexContext, errNilMutex),
		newSafeContextTestCase("SafeRLockContext typed-nil mu", opSafeRLockContext,
			bgCtx, typedNilRWMutexContext, errPanic),
		newSafeContextTestCase("SafeRLockContext panic", opSafeRLockContext,
			bgCtx, newPanicRLockContext, errPanic),
	}
}

func TestSafeContext(t *testing.T) {
	core.RunTestCases(t, safeContextTestCases())
}

// TestNewSafeLockContext verifies the context-bound factory yields a working
// acquirer.
func TestNewSafeLockContext(t *testing.T) {
	fn := mutex.NewSafeLockContext[mutex.MutexContext](context.Background())
	core.AssertMustNotNil(t, fn, "lock fn")

	ok, err := fn(&testMutexContext{})
	core.AssertTrue(t, ok, "ok")
	core.AssertNoError(t, err, "err")
}

// TestNewSafeRLockContext verifies the read-side context-bound factory yields
// a working acquirer.
func TestNewSafeRLockContext(t *testing.T) {
	fn := mutex.NewSafeRLockContext[mutex.RWMutexContext](context.Background())
	core.AssertMustNotNil(t, fn, "rlock fn")

	ok, err := fn(&testRWMutexContext{})
	core.AssertTrue(t, ok, "ok")
	core.AssertNoError(t, err, "err")
}

// ----- ReverseUnlock -----

func TestReverseUnlock(t *testing.T) {
	t.Run("reverse order", runTestReverseUnlockOrder)
	t.Run("nil function", runTestReverseUnlockNilFn)
	t.Run("error function", runTestReverseUnlockErrorFn)
	t.Run("nil mutex in list", runTestReverseUnlockNilMutex)
	t.Run("empty locks", runTestReverseUnlockEmpty)
}

func runTestReverseUnlockOrder(t *testing.T) {
	t.Helper()
	m1, m2, m3 := &sync.Mutex{}, &sync.Mutex{}, &sync.Mutex{}
	m1.Lock()
	m2.Lock()
	m3.Lock()

	order := make([]int, 0, 3)
	unlockFn := func(mu mutex.Mutex) error {
		order = append(order, reverseUnlockID(mu, m1, m2, m3))
		mu.Unlock()
		return nil
	}

	core.AssertNoError(t, mutex.ReverseUnlock[mutex.Mutex](unlockFn, m1, m2, m3),
		"reverse unlock")
	core.AssertSliceEqual(t, core.S(3, 2, 1), order, "reverse order")
}

func reverseUnlockID(mu, m1, m2, m3 mutex.Mutex) int {
	switch mu {
	case m1:
		return 1
	case m2:
		return 2
	case m3:
		return 3
	default:
		return 0
	}
}

func runTestReverseUnlockNilFn(t *testing.T) {
	t.Helper()
	err := mutex.ReverseUnlock[mutex.Mutex](nil, &sync.Mutex{})
	core.AssertError(t, err, "nil function")
	core.AssertContains(t, err.Error(), "unlock function is nil", "message")
}

func runTestReverseUnlockErrorFn(t *testing.T) {
	t.Helper()
	errFn := func(mutex.Mutex) error { return errors.ErrNilMutex }
	m := &sync.Mutex{}
	m.Lock()
	core.AssertError(t, mutex.ReverseUnlock[mutex.Mutex](errFn, m),
		"error function")
}

func runTestReverseUnlockNilMutex(t *testing.T) {
	t.Helper()
	m1, m2 := &sync.Mutex{}, &sync.Mutex{}
	m1.Lock()
	m2.Lock()
	core.AssertError(t,
		mutex.ReverseUnlock[mutex.Mutex](mutex.SafeUnlock, m1, nil, m2),
		"nil mutex in list")
}

func runTestReverseUnlockEmpty(t *testing.T) {
	t.Helper()
	core.AssertNoError(t, mutex.ReverseUnlock[mutex.Mutex](mutex.SafeUnlock),
		"empty locks")
}
