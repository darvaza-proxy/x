package cond_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/atomic"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/internal/synctesting"
)

// countTestTimeout caps each foreground synchronisation step. Generous
// enough to absorb scheduler jitter on loaded CI workers, tight enough
// that a hung waiter fails the run instead of stalling.
//
// countOpenGuard bounds negative-case assertions ("did not signal"):
// long enough to catch a spurious early wake, short enough to keep the
// test responsive.
const (
	countTestTimeout = time.Second

	countOpenGuard = 20 * time.Millisecond
)

// countStateTestCase exercises IsNil and IsClosed for the same Count
// state. Both methods read the same lifecycle, so the row carries both
// expectations to keep the state matrix legible.
type countStateTestCase struct {
	name string

	setup func() *cond.Count

	wantNil    bool
	wantClosed bool
}

func newCountStateTestCase(name string, setup func() *cond.Count,
	wantNil, wantClosed bool) countStateTestCase {
	return countStateTestCase{
		name:       name,
		setup:      setup,
		wantNil:    wantNil,
		wantClosed: wantClosed,
	}
}

func (tc countStateTestCase) Name() string { return tc.name }

func (tc countStateTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	core.AssertEqual(t, tc.wantNil, c.IsNil(), "IsNil")
	core.AssertEqual(t, tc.wantClosed, c.IsClosed(), "IsClosed")
}

var _ core.TestCase = countStateTestCase{}

func countStateTestCases() []countStateTestCase {
	return []countStateTestCase{
		newCountStateTestCase("nil pointer",
			func() *cond.Count { return nil },
			true, true),
		newCountStateTestCase("uninitialised",
			func() *cond.Count { return new(cond.Count) },
			true, true),
		newCountStateTestCase("initialised",
			func() *cond.Count { return cond.NewCount(42) },
			false, false),
		newCountStateTestCase("closed",
			func() *cond.Count {
				c := cond.NewCount(42)
				_ = c.Close()
				return c
			},
			false, true),
	}
}

// TestCountState verifies IsNil and IsClosed across every reachable
// lifecycle state.
func TestCountState(t *testing.T) {
	core.RunTestCases(t, countStateTestCases())
}

// countInitTestCase exercises Init across receiver and prior-state
// conditions. The happy-path row also verifies the initial value is
// honoured.
type countInitTestCase struct {
	name string

	setup func() *cond.Count

	initial int

	wantErr error
}

func newCountInitTestCase(name string, setup func() *cond.Count,
	initial int, wantErr error) countInitTestCase {
	return countInitTestCase{
		name:    name,
		setup:   setup,
		initial: initial,
		wantErr: wantErr,
	}
}

func (tc countInitTestCase) Name() string { return tc.name }

func (tc countInitTestCase) Test(t *testing.T) {
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

var _ core.TestCase = countInitTestCase{}

func countInitTestCases() []countInitTestCase {
	return []countInitTestCase{
		newCountInitTestCase("nil receiver",
			func() *cond.Count { return nil },
			0, errors.ErrNilReceiver),
		newCountInitTestCase("already initialised",
			func() *cond.Count {
				c := new(cond.Count)
				_ = c.Init(0)
				return c
			},
			0, errors.ErrAlreadyInitialised),
		newCountInitTestCase("fresh succeeds",
			func() *cond.Count { return new(cond.Count) },
			21, nil),
	}
}

// TestCountInit verifies Init's receiver and prior-state error paths
// alongside the happy path.
func TestCountInit(t *testing.T) {
	core.RunTestCases(t, countInitTestCases())
}

// countCloseTestCase exercises Close across receiver and prior-state
// conditions.
type countCloseTestCase struct {
	name string

	setup func() *cond.Count

	wantErr error
}

func newCountCloseTestCase(name string, setup func() *cond.Count,
	wantErr error) countCloseTestCase {
	return countCloseTestCase{name: name, setup: setup, wantErr: wantErr}
}

func (tc countCloseTestCase) Name() string { return tc.name }

func (tc countCloseTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := c.Close()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Close")
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Close error")
}

var _ core.TestCase = countCloseTestCase{}

func countCloseTestCases() []countCloseTestCase {
	return []countCloseTestCase{
		newCountCloseTestCase("nil receiver",
			func() *cond.Count { return nil },
			errors.ErrNilReceiver),
		newCountCloseTestCase("uninitialised",
			func() *cond.Count { return new(cond.Count) },
			errors.ErrNotInitialised),
		newCountCloseTestCase("happy path",
			func() *cond.Count { return cond.NewCount(0) },
			nil),
		newCountCloseTestCase("double close",
			func() *cond.Count {
				c := cond.NewCount(0)
				_ = c.Close()
				return c
			},
			errors.ErrClosed),
	}
}

// TestCountClose verifies Close's receiver and prior-state error paths
// alongside the happy path.
func TestCountClose(t *testing.T) {
	core.RunTestCases(t, countCloseTestCases())
}

// countResetTestCase exercises Reset across receiver, prior-state, and
// happy-path conditions. The waiter-wake path is exercised by
// TestCountResetWithWaiters.
type countResetTestCase struct {
	name string

	setup func() *cond.Count

	value int

	wantErr error
}

func newCountResetTestCase(name string, setup func() *cond.Count, value int,
	wantErr error) countResetTestCase {
	return countResetTestCase{
		name:    name,
		setup:   setup,
		value:   value,
		wantErr: wantErr,
	}
}

func (tc countResetTestCase) Name() string { return tc.name }

func (tc countResetTestCase) Test(t *testing.T) {
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

var _ core.TestCase = countResetTestCase{}

func countResetTestCases() []countResetTestCase {
	return []countResetTestCase{
		newCountResetTestCase("nil receiver",
			func() *cond.Count { return nil },
			0, errors.ErrNilReceiver),
		newCountResetTestCase("uninitialised",
			func() *cond.Count { return new(cond.Count) },
			1, errors.ErrNotInitialised),
		newCountResetTestCase("closed",
			func() *cond.Count {
				c := cond.NewCount(10)
				_ = c.Close()
				return c
			},
			2, errors.ErrClosed),
		newCountResetTestCase("happy path",
			func() *cond.Count { return cond.NewCount(10) },
			42, nil),
	}
}

// TestCountReset verifies Reset's receiver, prior-state, and happy-path
// outcomes.
func TestCountReset(t *testing.T) {
	core.RunTestCases(t, countResetTestCases())
}

// countErrorOpTestCase exercises error-returning operations that share
// the same nil-receiver and not-initialised handling. WaitFnContext
// also has a nil-context fast-path.
type countErrorOpTestCase struct {
	name string

	setup func() *cond.Count

	op func(*cond.Count) error

	wantErr error
}

func newCountErrorOpTestCase(name string, setup func() *cond.Count,
	op func(*cond.Count) error, wantErr error) countErrorOpTestCase {
	return countErrorOpTestCase{
		name:    name,
		setup:   setup,
		op:      op,
		wantErr: wantErr,
	}
}

func (tc countErrorOpTestCase) Name() string { return tc.name }

func (tc countErrorOpTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	err := tc.op(c)
	core.AssertErrorIs(t, err, tc.wantErr, "op error")
}

var _ core.TestCase = countErrorOpTestCase{}

func opCtxBG(c *cond.Count) error {
	return c.WaitFnContext(context.Background(), nil)
}

func opCtxNil(c *cond.Count) error {
	var nilCtx context.Context
	return c.WaitFnContext(nilCtx, nil)
}

func opAbortOpen(c *cond.Count) error {
	return c.WaitFnAbort(make(chan struct{}), nil)
}

func countErrorOpTestCases() []countErrorOpTestCase {
	nilC := func() *cond.Count { return nil }
	uninitialisedC := func() *cond.Count { return new(cond.Count) }
	zeroC := func() *cond.Count { return cond.NewCount(0) }

	return []countErrorOpTestCase{
		newCountErrorOpTestCase("WaitFnContext/nil receiver",
			nilC, opCtxBG, errors.ErrNilReceiver),
		newCountErrorOpTestCase("WaitFnContext/uninitialised",
			uninitialisedC, opCtxBG, errors.ErrNotInitialised),
		newCountErrorOpTestCase("WaitFnContext/nil context",
			zeroC, opCtxNil, errors.ErrNilContext),
		newCountErrorOpTestCase("WaitFnAbort/nil receiver",
			nilC, opAbortOpen, errors.ErrNilReceiver),
		newCountErrorOpTestCase("WaitFnAbort/uninitialised",
			uninitialisedC, opAbortOpen, errors.ErrNotInitialised),
	}
}

// TestCountErrorOps verifies error-returning wait operations across
// receiver, prior-state, and nil-context conditions.
func TestCountErrorOps(t *testing.T) {
	core.RunTestCases(t, countErrorOpTestCases())
}

// countPanicTestCase exercises methods that panic via the check() guard
// when called on a nil or uninitialised Count. Each row pins both the
// trigger state and the expected wrapped error so a future change that
// silently returns or panics with a different error fails loudly.
type countPanicTestCase struct {
	name string

	setup func() *cond.Count

	op func(*cond.Count)

	wantErr error
}

func newCountPanicTestCase(name string, setup func() *cond.Count,
	op func(*cond.Count), wantErr error) countPanicTestCase {
	return countPanicTestCase{
		name:    name,
		setup:   setup,
		op:      op,
		wantErr: wantErr,
	}
}

func (tc countPanicTestCase) Name() string { return tc.name }

func (tc countPanicTestCase) Test(t *testing.T) {
	t.Helper()
	c := tc.setup()
	core.AssertPanic(t, func() { tc.op(c) }, tc.wantErr, "panic error")
}

var _ core.TestCase = countPanicTestCase{}

func opWait(c *cond.Count)      { c.Wait() }
func opWaitFn(c *cond.Count)    { c.WaitFn(nil) }
func opIsZero(c *cond.Count)    { _ = c.IsZero() }
func opMatch(c *cond.Count)     { _ = c.Match(nil) }
func opSignal(c *cond.Count)    { _ = c.Signal() }
func opBroadcast(c *cond.Count) { c.Broadcast() }

func countPanicTestCases() []countPanicTestCase {
	nilC := func() *cond.Count { return nil }
	uninitialisedC := func() *cond.Count { return new(cond.Count) }

	return []countPanicTestCase{
		newCountPanicTestCase("Wait/nil",
			nilC, opWait, errors.ErrNilReceiver),
		newCountPanicTestCase("Wait/uninitialised",
			uninitialisedC, opWait, errors.ErrNotInitialised),
		newCountPanicTestCase("WaitFn/nil",
			nilC, opWaitFn, errors.ErrNilReceiver),
		newCountPanicTestCase("WaitFn/uninitialised",
			uninitialisedC, opWaitFn, errors.ErrNotInitialised),
		newCountPanicTestCase("IsZero/nil",
			nilC, opIsZero, errors.ErrNilReceiver),
		newCountPanicTestCase("IsZero/uninitialised",
			uninitialisedC, opIsZero, errors.ErrNotInitialised),
		newCountPanicTestCase("Match/nil",
			nilC, opMatch, errors.ErrNilReceiver),
		newCountPanicTestCase("Match/uninitialised",
			uninitialisedC, opMatch, errors.ErrNotInitialised),
		newCountPanicTestCase("Signal/nil",
			nilC, opSignal, errors.ErrNilReceiver),
		newCountPanicTestCase("Signal/uninitialised",
			uninitialisedC, opSignal, errors.ErrNotInitialised),
		newCountPanicTestCase("Broadcast/nil",
			nilC, opBroadcast, errors.ErrNilReceiver),
		newCountPanicTestCase("Broadcast/uninitialised",
			uninitialisedC, opBroadcast, errors.ErrNotInitialised),
	}
}

// TestCountPanics verifies every method that panics via check() does so
// for both nil and uninitialised receivers.
func TestCountPanics(t *testing.T) {
	core.RunTestCases(t, countPanicTestCases())
}

// TestCountNewCount verifies the initial value and live state of a
// freshly-constructed Count.
func TestCountNewCount(t *testing.T) {
	c := cond.NewCount(42)
	defer func() { _ = c.Close() }()

	core.AssertMustNotNil(t, c, "NewCount returned non-nil")
	core.AssertEqual(t, 42, c.Value(), "initial value")
	core.AssertFalse(t, c.IsNil(), "IsNil after NewCount")
	core.AssertFalse(t, c.IsClosed(), "IsClosed after NewCount")
}

// TestCountAdd verifies Add applies positive, negative, and zero
// increments. Add(0) takes the special-cased early-return path that
// skips the broadcast machinery.
func TestCountAdd(t *testing.T) {
	c := cond.NewCount(10)
	defer func() { _ = c.Close() }()

	core.AssertEqual(t, 15, c.Add(5), "Add(5) returns new value")
	core.AssertEqual(t, 15, c.Value(), "Value after Add(5)")
	core.AssertEqual(t, 8, c.Add(-7), "Add(-7) returns new value")
	core.AssertEqual(t, 8, c.Value(), "Value after Add(-7)")
	core.AssertEqual(t, 8, c.Add(0), "Add(0) returns current value")
}

// TestCountIncDec verifies Inc and Dec each return the new value and
// move the counter by one.
func TestCountIncDec(t *testing.T) {
	c := cond.NewCount(41)
	defer func() { _ = c.Close() }()

	core.AssertEqual(t, 42, c.Inc(), "Inc returns new value")
	core.AssertEqual(t, 42, c.Value(), "Value after Inc")
	core.AssertEqual(t, 41, c.Dec(), "Dec returns new value")
	core.AssertEqual(t, 41, c.Value(), "Value after Dec")
}

// TestCountIsZero verifies IsZero across Inc/Dec transitions.
func TestCountIsZero(t *testing.T) {
	c := cond.NewCount(0)
	defer func() { _ = c.Close() }()

	core.AssertTrue(t, c.IsZero(), "IsZero at zero")
	c.Inc()
	core.AssertFalse(t, c.IsZero(), "IsZero after Inc")
	c.Dec()
	core.AssertTrue(t, c.IsZero(), "IsZero after Dec")
}

// TestCountMatch verifies Match with explicit and nil predicates. A nil
// predicate is equivalent to IsZero.
func TestCountMatch(t *testing.T) {
	c := cond.NewCount(42)
	defer func() { _ = c.Close() }()

	isFortyTwo := func(v int32) bool { return v == 42 }
	core.AssertTrue(t, c.Match(isFortyTwo), "Match isFortyTwo at 42")

	c.Inc()
	core.AssertFalse(t, c.Match(isFortyTwo), "Match isFortyTwo at 43")

	core.AssertMustNoError(t, c.Reset(0), "Reset to zero")
	core.AssertTrue(t, c.Match(nil), "Match nil at zero")
}

// TestCountWait verifies Wait blocks until the counter reaches zero.
func TestCountWait(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	go func() {
		c.Wait()
		close(done)
	}()

	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"Wait blocks while value is non-zero")

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"Wait returns after counter reaches zero")
}

// TestCountWaitFn verifies WaitFn blocks until the supplied predicate
// is satisfied.
func TestCountWaitFn(t *testing.T) {
	c := cond.NewCount(5)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	go func() {
		c.WaitFn(func(v int32) bool { return v >= 10 })
		close(done)
	}()

	c.Add(3) // 8 — predicate still false
	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"WaitFn blocks while predicate is false")

	c.Add(2) // 10 — predicate now true
	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFn returns once predicate holds")
}

// TestCountWaitFnContextImmediate verifies WaitFnContext returns
// without error when the predicate is already satisfied at entry.
func TestCountWaitFnContextImmediate(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(),
		countTestTimeout)
	defer cancel()

	err := c.WaitFnContext(ctx, func(v int32) bool { return v == 1 })
	core.AssertNoError(t, err,
		"WaitFnContext, predicate already true")

	core.AssertMustNoError(t, c.Reset(0), "Reset to zero")

	err = c.WaitFnContext(ctx, nil)
	core.AssertNoError(t, err,
		"WaitFnContext nil predicate at zero")
}

// TestCountWaitFnContextSuccess verifies WaitFnContext unblocks once
// the predicate becomes true.
func TestCountWaitFnContextSuccess(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(),
		countTestTimeout)
	defer cancel()

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitFnContext(ctx, func(v int32) bool {
			return v == 0
		})
	}()

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFnContext returned")
	core.AssertNoError(t, gotErr, "WaitFnContext result")
}

// TestCountWaitFnContextCancelled verifies WaitFnContext returns the
// context error once the context is cancelled.
func TestCountWaitFnContextCancelled(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitFnContext(ctx, nil)
	}()

	cancel()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFnContext returned after cancel")
	core.AssertErrorIs(t, gotErr, context.Canceled,
		"WaitFnContext returns ctx.Err()")
}

// TestCountWaitFnContextTimeout verifies WaitFnContext returns the
// deadline-exceeded error when the context expires before the
// predicate is satisfied.
func TestCountWaitFnContextTimeout(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), countOpenGuard)
	defer cancel()

	err := c.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
	core.AssertErrorIs(t, err, context.DeadlineExceeded,
		"WaitFnContext returns DeadlineExceeded")
}

// TestCountWaitFnAbortSuccess verifies WaitFnAbort returns nil once the
// predicate becomes true.
func TestCountWaitFnAbortSuccess(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	abort := make(chan struct{})
	defer close(abort)

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitFnAbort(abort, func(v int32) bool {
			return v == 0
		})
	}()

	c.Dec()

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFnAbort returned")
	core.AssertNoError(t, gotErr, "WaitFnAbort result")
}

// TestCountWaitFnAbortAborted verifies WaitFnAbort returns
// context.Canceled when the abort channel closes before the predicate
// is satisfied.
func TestCountWaitFnAbortAborted(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	abort := make(chan struct{})

	done := make(chan struct{})
	var gotErr error
	go func() {
		defer close(done)
		gotErr = c.WaitFnAbort(abort, nil)
	}()

	close(abort)

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"WaitFnAbort returned after abort")
	core.AssertErrorIs(t, gotErr, context.Canceled,
		"WaitFnAbort returns context.Canceled on abort")
}

// TestCountSignal verifies Signal returns false when no waiter is
// parked, wakes a single waiter when one is parked, and returns false
// again once the Count is closed.
func TestCountSignal(t *testing.T) {
	c := cond.NewCount(1)

	core.AssertFalse(t, c.Signal(), "Signal without waiter")

	var doneClosed sync.Once
	done := make(chan struct{})

	go func() {
		c.WaitFn(func(_ int32) bool {
			doneClosed.Do(func() { close(done) })
			return false
		})
	}()

	signaled := synctesting.WaitForCond(c.Signal, countTestTimeout,
		synctesting.PollStep)
	core.AssertTrue(t, signaled,
		"Signal eventually wakes the parked waiter")

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"predicate ran after Signal")

	_ = c.Close()
	core.AssertFalse(t, c.Signal(), "Signal after Close")
}

// broadcastWaiter is one waiter goroutine for TestCountBroadcast. It
// parks inside WaitFn and, on each wake-up, sets its assigned bit in
// the shared mask; the goroutine that completes the all-bits-set
// transition closes allNotified.
//
//revive:disable-next-line:argument-limit
func broadcastWaiter(c *cond.Count, bit uint32, allMask uint32,
	notified *atomic.Uint32, ready chan<- struct{},
	allNotified chan struct{}) {
	ready <- struct{}{}
	c.WaitFn(func(_ int32) bool {
		next, changed := atomic.BitmaskOr(notified, bit)
		if changed && next == allMask {
			close(allNotified)
		}
		return false
	})
}

// TestCountBroadcast verifies Broadcast wakes every parked waiter.
// Each waiter sets its own bit in a shared mask; the test waits for
// the all-bits-set transition rather than polling counts.
func TestCountBroadcast(t *testing.T) {
	c := cond.NewCount(1)
	defer func() { _ = c.Close() }()

	const numWaiters = 3
	const allMask uint32 = (1 << numWaiters) - 1

	var notified atomic.Uint32
	allNotified := make(chan struct{})
	ready := make(chan struct{}, numWaiters)

	for i := range numWaiters {
		bit := uint32(1) << i
		go broadcastWaiter(c, bit, allMask, &notified, ready, allNotified)
	}

	synctesting.AssertMustReadersReady(t, ready, numWaiters,
		countTestTimeout, "all waiters ready")

	// Let the waiters park inside WaitFn before broadcasting; the
	// ready signal fires before the receive on the barrier.
	time.Sleep(countOpenGuard)

	c.Broadcast()

	synctesting.AssertMustClosed(t, allNotified, countTestTimeout,
		"all waiters notified")
	core.AssertEqual(t, allMask, notified.Load(),
		"all bits set in notified mask")
}

// TestCountBroadcastCondition verifies that custom broadcast conditions
// gate when waiters wake: an Inc that does not match the condition
// leaves the waiter parked, while one that does match wakes it.
func TestCountBroadcastCondition(t *testing.T) {
	broadcastOn10 := func(v int32) bool { return v%10 == 0 }
	c := cond.NewCount(0, broadcastOn10)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	go func() {
		c.WaitFn(func(v int32) bool { return v >= 5 })
		close(done)
	}()

	c.Inc() // 1 — 1 % 10 != 0, no broadcast
	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"waiter not woken by non-broadcast Inc")

	c.Add(9) // 10 — 10 % 10 == 0, broadcast
	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"waiter woken by broadcast condition")
}

// broadcastConditionStep names a value transition expected to fire a
// broadcast under the multi-condition setup.
type broadcastConditionStep struct {
	name string
	op   func(*cond.Count)
}

// TestCountMultipleBroadcastConditions verifies a Count constructed
// with several broadcast predicates fires on each matching transition.
// The waiter's predicate always returns false to keep it re-arming
// across broadcasts; notified counts wake-ups via an atomic counter.
func TestCountMultipleBroadcastConditions(t *testing.T) {
	isZero := func(v int32) bool { return v == 0 }
	isPositive := func(v int32) bool { return v > 0 }
	isNegative := func(v int32) bool { return v < 0 }

	c := cond.NewCount(0, isZero, isPositive, isNegative)

	var notified atomic.Uint32

	var ready sync.Once
	readyCh := make(chan struct{})
	waiter := make(chan struct{})
	go func() {
		defer close(waiter)
		c.WaitFn(func(_ int32) bool {
			ready.Do(func() { close(readyCh) })
			notified.Add(1)
			return false
		})
	}()

	synctesting.AssertMustClosed(t, readyCh, countTestTimeout,
		"waiter entered predicate")
	// Let the waiter park at the barrier receive between predicate
	// re-runs before firing the first transition.
	time.Sleep(countOpenGuard)

	steps := []broadcastConditionStep{
		{"Add(-1) → -1 (isNegative)", func(c *cond.Count) { c.Add(-1) }},
		{"Inc → 0 (isZero)", func(c *cond.Count) { c.Inc() }},
		{"Inc → 1 (isPositive)", func(c *cond.Count) { c.Inc() }},
		{"Add(-2) → -1 (isNegative)", func(c *cond.Count) { c.Add(-2) }},
	}
	for _, step := range steps {
		prev := notified.Load()
		step.op(c)
		synctesting.AssertMustEventually(t, func() bool {
			return notified.Load() > prev
		}, countTestTimeout, "notified incremented after "+step.name)
	}

	_ = c.Close()
	synctesting.AssertMustClosed(t, waiter, countTestTimeout,
		"waiter exited on Close")
}

// TestCountNoBroadcastConditions verifies that a Count constructed with
// an empty broadcast-conditions slice falls back to the default
// "broadcast on every change" behaviour.
func TestCountNoBroadcastConditions(t *testing.T) {
	c := cond.NewCount(0, []func(int32) bool{}...)

	var notified atomic.Uint32

	var ready sync.Once
	readyCh := make(chan struct{})
	waiter := make(chan struct{})
	go func() {
		defer close(waiter)
		c.WaitFn(func(_ int32) bool {
			ready.Do(func() { close(readyCh) })
			notified.Add(1)
			return false
		})
	}()

	synctesting.AssertMustClosed(t, readyCh, countTestTimeout,
		"waiter entered predicate")
	time.Sleep(countOpenGuard)

	prev := notified.Load()
	c.Inc()
	synctesting.AssertMustEventually(t, func() bool {
		return notified.Load() > prev
	}, countTestTimeout,
		"notified incremented after Inc with no broadcast conditions")

	_ = c.Close()
	synctesting.AssertMustClosed(t, waiter, countTestTimeout,
		"waiter exited on Close")
}

// TestCountResetWithWaiters verifies Reset wakes any parked waiter
// whose predicate is satisfied by the new value.
func TestCountResetWithWaiters(t *testing.T) {
	c := cond.NewCount(5)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	go func() {
		c.WaitFn(func(v int32) bool { return v == 0 })
		close(done)
	}()

	synctesting.AssertMustOpen(t, done, countOpenGuard,
		"WaitFn blocks before Reset")

	core.AssertMustNoError(t, c.Reset(0), "Reset to zero")

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"waiter woken after Reset to zero")
}

func incLoop(c *cond.Count, n int, done chan<- struct{}) {
	for range n {
		c.Inc()
	}
	done <- struct{}{}
}

func decLoop(c *cond.Count, n int, done chan<- struct{}) {
	for range n {
		c.Dec()
	}
	done <- struct{}{}
}

// TestCountConcurrentIncDec verifies the counter survives concurrent
// Inc/Dec from many goroutines and lands back at zero when the
// increments balance the decrements.
func TestCountConcurrentIncDec(t *testing.T) {
	c := cond.NewCount(0)
	defer func() { _ = c.Close() }()

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan struct{}, numGoroutines*2)

	for range numGoroutines {
		go incLoop(c, numOperations, done)
	}
	for range numGoroutines {
		go decLoop(c, numOperations, done)
	}

	synctesting.AssertMustReadersReady(t, done, numGoroutines*2,
		countTestTimeout, "all goroutines finished")

	core.AssertEqual(t, 0, c.Value(),
		"value returns to zero after balanced Inc/Dec")
}

// TestCountConcurrentReadWriteSignal verifies that interleaved
// reader, writer, matcher, and signaller operations do not trigger
// the race detector.
func TestCountConcurrentReadWriteSignal(t *testing.T) {
	c := cond.NewCount(0)
	defer func() { _ = c.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 100 {
			c.Inc()
			_ = c.Value()
			c.Dec()
		}
	}()

	for range 100 {
		_ = c.Value()
		_ = c.Match(func(v int32) bool { return v >= 0 })
		_ = c.Signal()
	}

	synctesting.AssertMustClosed(t, done, countTestTimeout,
		"concurrent writer completed")
}

// stressUpdater runs numUpdates Inc operations or exits when ctx is
// cancelled. Split out from the stress test body to keep the
// orchestration loop within the cognitive-complexity budget.
func stressUpdater(ctx context.Context, c *cond.Count, numUpdates int,
	done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	for range numUpdates {
		if ctx.Err() != nil {
			return
		}
		c.Inc()
		time.Sleep(time.Microsecond)
	}
}

// stressWaiter watches for the counter to cross successive 100-step
// targets until ctx is cancelled. Targets re-arm to a higher value
// each pass so the waiter keeps engaging the barrier.
func stressWaiter(ctx context.Context, c *cond.Count, target int32,
	done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	for ctx.Err() == nil {
		err := c.WaitFnContext(ctx, func(v int32) bool {
			return v >= target
		})
		if err != nil {
			return
		}
		target += 100
	}
}

// TestCountStress runs many waiters and updaters against the same
// Count for a bounded time and verifies the final value matches the
// total number of Inc operations performed. Skipped under -short.
func TestCountStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	c := cond.NewCount(0)
	defer func() { _ = c.Close() }()

	const (
		numWaiters  = 20
		numUpdaters = 10
		numUpdates  = 1000
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{}, numWaiters+numUpdaters)

	for i := range numWaiters {
		target := int32((i % 10) * 100)
		go stressWaiter(ctx, c, target, done)
	}
	for range numUpdaters {
		go stressUpdater(ctx, c, numUpdates, done)
	}

	synctesting.AssertMustReadersReady(t, done, numWaiters+numUpdaters,
		6*time.Second, "all goroutines finished")

	core.AssertEqual(t, numUpdaters*numUpdates, c.Value(),
		"final value matches total updates")
}
