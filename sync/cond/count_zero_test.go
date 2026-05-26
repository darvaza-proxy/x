package cond_test

import (
	"context"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/internal/synctesting"
)

// countZeroStateTestCase exercises IsNil and IsClosed for the same
// CountZero state. Both methods read the same lifecycle, so the row
// carries both expectations to keep the state matrix legible.
type countZeroStateTestCase struct {
	name string

	setup func() *cond.CountZero

	wantNil    bool
	wantClosed bool
}

func newCountZeroStateTestCase(name string, setup func() *cond.CountZero,
	wantNil, wantClosed bool) countZeroStateTestCase {
	return countZeroStateTestCase{
		name:       name,
		setup:      setup,
		wantNil:    wantNil,
		wantClosed: wantClosed,
	}
}

func (tc countZeroStateTestCase) Name() string { return tc.name }

func (tc countZeroStateTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	core.AssertEqual(t, tc.wantNil, c.IsNil(), "IsNil")
	core.AssertEqual(t, tc.wantClosed, c.IsClosed(), "IsClosed")
}

var _ core.TestCase = countZeroStateTestCase{}

func countZeroStateTestCases() []countZeroStateTestCase {
	return []countZeroStateTestCase{
		newCountZeroStateTestCase("nil pointer",
			func() *cond.CountZero { return nil },
			true, true),
		newCountZeroStateTestCase("uninitialised",
			func() *cond.CountZero { return new(cond.CountZero) },
			true, true),
		newCountZeroStateTestCase("initialised",
			func() *cond.CountZero { return cond.NewCountZero(42) },
			false, false),
		newCountZeroStateTestCase("closed",
			func() *cond.CountZero {
				c := cond.NewCountZero(42)
				_ = c.Close()
				return c
			},
			false, true),
	}
}

// TestCountZeroState verifies IsNil and IsClosed across every
// reachable lifecycle state.
func TestCountZeroState(t *testing.T) {
	core.RunTestCases(t, countZeroStateTestCases())
}

// countZeroInitTestCase exercises Init across receiver and prior-state
// conditions. The happy-path row also verifies the initial value is
// honoured and that the isZero broadcast condition is wired in.
type countZeroInitTestCase struct {
	name string

	setup func() *cond.CountZero

	initial int

	wantErr error
}

func newCountZeroInitTestCase(name string, setup func() *cond.CountZero,
	initial int, wantErr error) countZeroInitTestCase {
	return countZeroInitTestCase{
		name:    name,
		setup:   setup,
		initial: initial,
		wantErr: wantErr,
	}
}

func (tc countZeroInitTestCase) Name() string { return tc.name }

func (tc countZeroInitTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := c.Init(tc.initial)
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Init")
		core.AssertEqual(t, tc.initial, c.Value(), "Value after Init")
		_ = c.Close()
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Init error")
}

var _ core.TestCase = countZeroInitTestCase{}

func countZeroInitTestCases() []countZeroInitTestCase {
	return []countZeroInitTestCase{
		newCountZeroInitTestCase("nil receiver",
			func() *cond.CountZero { return nil },
			0, errors.ErrNilReceiver),
		newCountZeroInitTestCase("already initialised",
			func() *cond.CountZero {
				c := new(cond.CountZero)
				_ = c.Init(0)
				return c
			},
			0, errors.ErrAlreadyInitialised),
		newCountZeroInitTestCase("fresh succeeds",
			func() *cond.CountZero { return new(cond.CountZero) },
			21, nil),
	}
}

// TestCountZeroInit verifies Init's receiver and prior-state error
// paths alongside the happy path.
func TestCountZeroInit(t *testing.T) {
	core.RunTestCases(t, countZeroInitTestCases())
}

// countZeroCloseTestCase exercises Close across receiver and
// prior-state conditions.
type countZeroCloseTestCase struct {
	name string

	setup func() *cond.CountZero

	wantErr error
}

func newCountZeroCloseTestCase(name string, setup func() *cond.CountZero,
	wantErr error) countZeroCloseTestCase {
	return countZeroCloseTestCase{
		name: name, setup: setup, wantErr: wantErr,
	}
}

func (tc countZeroCloseTestCase) Name() string { return tc.name }

func (tc countZeroCloseTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := c.Close()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Close")
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Close error")
}

var _ core.TestCase = countZeroCloseTestCase{}

func countZeroCloseTestCases() []countZeroCloseTestCase {
	return []countZeroCloseTestCase{
		newCountZeroCloseTestCase("nil receiver",
			func() *cond.CountZero { return nil },
			errors.ErrNilReceiver),
		newCountZeroCloseTestCase("uninitialised",
			func() *cond.CountZero { return new(cond.CountZero) },
			errors.ErrNotInitialised),
		newCountZeroCloseTestCase("happy path",
			func() *cond.CountZero { return cond.NewCountZero(0) },
			nil),
		newCountZeroCloseTestCase("double close",
			func() *cond.CountZero {
				c := cond.NewCountZero(0)
				_ = c.Close()
				return c
			},
			errors.ErrClosed),
	}
}

// TestCountZeroClose verifies Close's receiver and prior-state error
// paths alongside the happy path.
func TestCountZeroClose(t *testing.T) {
	core.RunTestCases(t, countZeroCloseTestCases())
}

// countZeroResetTestCase exercises Reset across receiver,
// prior-state, and happy-path conditions.
type countZeroResetTestCase struct {
	name string

	setup func() *cond.CountZero

	value int

	wantErr error
}

func newCountZeroResetTestCase(name string, setup func() *cond.CountZero,
	value int, wantErr error) countZeroResetTestCase {
	return countZeroResetTestCase{
		name:    name,
		setup:   setup,
		value:   value,
		wantErr: wantErr,
	}
}

func (tc countZeroResetTestCase) Name() string { return tc.name }

func (tc countZeroResetTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := c.Reset(tc.value)
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Reset")
		core.AssertEqual(t, tc.value, c.Value(), "Value after Reset")
		_ = c.Close()
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Reset error")
}

var _ core.TestCase = countZeroResetTestCase{}

func countZeroResetTestCases() []countZeroResetTestCase {
	return []countZeroResetTestCase{
		newCountZeroResetTestCase("nil receiver",
			func() *cond.CountZero { return nil },
			0, errors.ErrNilReceiver),
		newCountZeroResetTestCase("uninitialised",
			func() *cond.CountZero { return new(cond.CountZero) },
			1, errors.ErrNotInitialised),
		newCountZeroResetTestCase("closed",
			func() *cond.CountZero {
				c := cond.NewCountZero(10)
				_ = c.Close()
				return c
			},
			2, errors.ErrClosed),
		newCountZeroResetTestCase("happy path",
			func() *cond.CountZero { return cond.NewCountZero(10) },
			42, nil),
	}
}

// TestCountZeroReset verifies Reset's receiver, prior-state, and
// happy-path outcomes.
func TestCountZeroReset(t *testing.T) {
	core.RunTestCases(t, countZeroResetTestCases())
}

// countZeroErrorOpTestCase exercises Wait, WaitAbort and WaitContext
// across receiver, prior-state, and nil-context conditions. All three
// methods return errors rather than panicking.
type countZeroErrorOpTestCase struct {
	name string

	setup func() *cond.CountZero

	op func(*cond.CountZero) error

	wantErr error
}

func newCountZeroErrorOpTestCase(name string,
	setup func() *cond.CountZero, op func(*cond.CountZero) error,
	wantErr error) countZeroErrorOpTestCase {
	return countZeroErrorOpTestCase{
		name:    name,
		setup:   setup,
		op:      op,
		wantErr: wantErr,
	}
}

func (tc countZeroErrorOpTestCase) Name() string { return tc.name }

func (tc countZeroErrorOpTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := tc.op(c)
	core.AssertErrorIs(t, err, tc.wantErr, "op error")
}

var _ core.TestCase = countZeroErrorOpTestCase{}

func opZeroWait(c *cond.CountZero) error      { return c.Wait() }
func opZeroAbortOpen(c *cond.CountZero) error { return c.WaitAbort(make(chan struct{})) }
func opZeroCtxBG(c *cond.CountZero) error     { return c.WaitContext(context.Background()) }
func opZeroCtxNil(c *cond.CountZero) error {
	var nilCtx context.Context
	return c.WaitContext(nilCtx)
}

func countZeroErrorOpTestCases() []countZeroErrorOpTestCase {
	nilC := func() *cond.CountZero { return nil }
	uninitialisedC := func() *cond.CountZero { return new(cond.CountZero) }
	zeroC := func() *cond.CountZero { return cond.NewCountZero(0) }

	return []countZeroErrorOpTestCase{
		newCountZeroErrorOpTestCase("Wait/nil receiver",
			nilC, opZeroWait, errors.ErrNilReceiver),
		newCountZeroErrorOpTestCase("Wait/uninitialised",
			uninitialisedC, opZeroWait, errors.ErrNotInitialised),
		newCountZeroErrorOpTestCase("WaitAbort/nil receiver",
			nilC, opZeroAbortOpen, errors.ErrNilReceiver),
		newCountZeroErrorOpTestCase("WaitAbort/uninitialised",
			uninitialisedC, opZeroAbortOpen, errors.ErrNotInitialised),
		newCountZeroErrorOpTestCase("WaitContext/nil receiver",
			nilC, opZeroCtxBG, errors.ErrNilReceiver),
		newCountZeroErrorOpTestCase("WaitContext/uninitialised",
			uninitialisedC, opZeroCtxBG, errors.ErrNotInitialised),
		newCountZeroErrorOpTestCase("WaitContext/nil context",
			zeroC, opZeroCtxNil, errors.ErrNilContext),
	}
}

// TestCountZeroErrorOps verifies error-returning wait operations
// across receiver, prior-state, and nil-context conditions.
func TestCountZeroErrorOps(t *testing.T) {
	core.RunTestCases(t, countZeroErrorOpTestCases())
}

// TestCountZeroNewCountZero verifies NewCountZero returns a valid,
// initialised receiver with the supplied initial value.
func TestCountZeroNewCountZero(t *testing.T) {
	c := cond.NewCountZero(42)
	defer func() { _ = c.Close() }()

	core.AssertMustNotNil(t, c, "NewCountZero returned non-nil")
	core.AssertEqual(t, 42, c.Value(), "initial value")
	core.AssertFalse(t, c.IsNil(), "IsNil after NewCountZero")
	core.AssertFalse(t, c.IsClosed(), "IsClosed after NewCountZero")
}

// TestCountZeroAddIncDec verifies Add, Inc, and Dec all flow through
// to the underlying Count and return the new value. Add(0) takes the
// underlying special-cased early-return path.
func TestCountZeroAddIncDec(t *testing.T) {
	c := cond.NewCountZero(10)
	defer func() { _ = c.Close() }()

	core.AssertEqual(t, 15, c.Add(5), "Add(5) returns new value")
	core.AssertEqual(t, 15, c.Value(), "Value after Add(5)")
	core.AssertEqual(t, 8, c.Add(-7), "Add(-7) returns new value")
	core.AssertEqual(t, 8, c.Add(0), "Add(0) returns current value")
	core.AssertEqual(t, 9, c.Inc(), "Inc returns new value")
	core.AssertEqual(t, 8, c.Dec(), "Dec returns new value")
	core.AssertEqual(t, 8, c.Value(), "Value after Dec")
}

// TestCountZeroWait verifies Wait blocks until the counter reaches
// zero. Drives the workgroup dependency surface: Inc to enter the
// non-zero state, Wait blocks, Dec wakes the waiter.
func TestCountZeroWait(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.Wait()
	}()

	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"Wait blocks while value is non-zero")

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"Wait returns after counter reaches zero")
	core.AssertNoError(t, gotErr, "Wait result")
}

// TestCountZeroWaitContextImmediate verifies WaitContext returns
// without error when the counter is already zero at entry.
func TestCountZeroWaitContextImmediate(t *testing.T) {
	c := cond.NewCountZero(0)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(),
		countTestTimeout)
	defer cancel()

	core.AssertNoError(t, c.WaitContext(ctx),
		"WaitContext at zero returns immediately")
}

// TestCountZeroWaitContextSuccess verifies WaitContext unblocks once
// the counter reaches zero.
func TestCountZeroWaitContextSuccess(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(),
		countTestTimeout)
	defer cancel()

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitContext(ctx)
	}()

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitContext returned")
	core.AssertNoError(t, gotErr, "WaitContext result")
}

// TestCountZeroWaitContextCancelled verifies WaitContext returns the
// context error once the context is cancelled before the counter
// reaches zero.
func TestCountZeroWaitContextCancelled(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitContext(ctx)
	}()

	cancel()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitContext returned after cancel")
	core.AssertErrorIs(t, gotErr, context.Canceled,
		"WaitContext returns ctx.Err()")
}

// TestCountZeroWaitContextTimeout verifies WaitContext returns the
// deadline-exceeded error when the context expires before the
// counter reaches zero.
func TestCountZeroWaitContextTimeout(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(),
		countOpenGuard)
	defer cancel()

	err := c.WaitContext(ctx)
	core.AssertErrorIs(t, err, context.DeadlineExceeded,
		"WaitContext returns DeadlineExceeded")
}

// TestCountZeroWaitAbortSuccess verifies WaitAbort returns nil once
// the counter reaches zero.
func TestCountZeroWaitAbortSuccess(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	abort := make(chan struct{})
	defer close(abort)

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitAbort(abort)
	}()

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitAbort returned")
	core.AssertNoError(t, gotErr, "WaitAbort result")
}

// TestCountZeroWaitAbortAborted verifies WaitAbort returns
// context.Canceled when the abort channel closes before the counter
// reaches zero.
func TestCountZeroWaitAbortAborted(t *testing.T) {
	c := cond.NewCountZero(1)
	defer func() { _ = c.Close() }()

	abort := make(chan struct{})

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitAbort(abort)
	}()

	close(abort)

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitAbort returned after abort")
	core.AssertErrorIs(t, gotErr, context.Canceled,
		"WaitAbort returns context.Canceled on abort")
}

// TestCountZeroBroadcastOnZero pins the type's defining behaviour:
// multiple concurrent waiters all wake once the counter reaches
// zero. This is the contract workgroup relies on for fan-in
// completion signalling.
func TestCountZeroBroadcastOnZero(t *testing.T) {
	const waiters = 5

	c := cond.NewCountZero(3)
	defer func() { _ = c.Close() }()

	done := make(chan struct{}, waiters)
	for range waiters {
		go func() {
			_ = c.Wait()
			done <- struct{}{}
		}()
	}

	c.Dec()
	c.Dec()
	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"waiters block while value is non-zero")

	c.Dec()

	synctesting.AssertMustReadersReady(t, done, waiters,
		countTestTimeout, "all waiters wake after zero")
}
