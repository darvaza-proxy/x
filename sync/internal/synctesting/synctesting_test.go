package synctesting_test

import (
	"fmt"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/internal/synctesting"
)

const (
	// shortBudgetMS caps a "should already be true" path quickly.
	shortBudgetMS = 50
	// tinyBudgetMS caps a "should time out" path while keeping tests fast.
	tinyBudgetMS = 10
	// pacedBudgetMS gives the shared-deadline rows a ±20 ms margin around
	// the deadline so the property test isn't flaky under CI load.
	pacedBudgetMS = 100
)

// Compile-time assertions that the generic channel helpers instantiate
// for element types other than struct{}. The body never inspects T, so
// these references are sufficient to prove the surface compiles.
var (
	_ = synctesting.AssertClosed[int]
	_ = synctesting.AssertMustClosed[error]
	_ = synctesting.AssertOpen[*time.Time]
	_ = synctesting.AssertMustOpen[bool]
	_ = synctesting.AssertReadersReady[string]
	_ = synctesting.AssertMustReadersReady[byte]
)

// alwaysTrue returns a predicate that is true from the first call.
func alwaysTrue() func() bool { return func() bool { return true } }

// alwaysFalse returns a predicate that never becomes true.
func alwaysFalse() func() bool { return func() bool { return false } }

// flagFlip returns a predicate that becomes true once a couple of poll
// steps have elapsed since the predicate's construction. Time-based, so no
// goroutine, no shared state, and no scheduler jitter on the transition.
func flagFlip() func() bool {
	target := time.Now().Add(2 * synctesting.PollStep)
	return func() bool {
		return !time.Now().Before(target)
	}
}

// chanLeaveOpen leaves ch untouched.
func chanLeaveOpen(_ chan struct{}) {}

// chanClose closes ch.
func chanClose(ch chan struct{}) { close(ch) }

// chanSendOnce sends a single value into ch; the channel must have cap ≥ 1.
func chanSendOnce(ch chan struct{}) { ch <- struct{}{} }

// chanWithSends returns a setup that pre-fills a buffered channel with n
// values synchronously, then returns it for the helper to drain.
func chanWithSends(n int) func() <-chan struct{} {
	return func() <-chan struct{} {
		ch := make(chan struct{}, n)
		for range n {
			ch <- struct{}{}
		}
		return ch
	}
}

// chanClosedEmpty returns an already-closed channel with no values sent.
// Reads return the zero value immediately and signal close via the
// two-value receive form.
func chanClosedEmpty() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// chanWithPacedSends returns a setup that spawns a goroutine pushing n
// values into a buffered channel, sleeping stepMS milliseconds before each
// send. The channel buffer matches n so the sender never blocks even if
// the helper exits early.
func chanWithPacedSends(n, stepMS int) func() <-chan struct{} {
	step := time.Duration(stepMS) * time.Millisecond
	return func() <-chan struct{} {
		ch := make(chan struct{}, n)
		go func() {
			for range n {
				time.Sleep(step)
				ch <- struct{}{}
			}
		}()
		return ch
	}
}

// nameProbe is the literal name passed to every Assert* call in this file.
// On failure, core.doMessage formats the recorded error as
// "<name>: <body>", so the probe substring appears iff the helper forwarded
// name through to core.AssertTrue. A regression that drops or empties name
// would yield a message starting with "expected ..., got ..." instead.
const nameProbe = "test"

// argsFormat and argsValue drive the args ...any propagation rows: passing
// argsFormat as name and argsValue via args ...any should produce a prefix
// matching argsExpected. A regression that dropped args would record the
// literal format string instead, missing the digits.
const (
	argsFormat = "probe %d"
	argsValue  = 42
)

// argsExpected is the formatted prefix the helpers should record when
// argsFormat and argsValue propagate correctly. Computed once at init so
// it stays in lock-step with argsFormat/argsValue.
var argsExpected = fmt.Sprintf(argsFormat, argsValue)

// checkAssertOutcome validates a bare Assert* helper's outcome against the
// row's wantOK expectation: ok must match, mock errors are recorded
// exactly when failure was expected, and any recorded error must carry
// the name argument forwarded from the helper's call site.
func checkAssertOutcome(t *testing.T, mock *core.MockT, ok, wantOK bool) {
	t.Helper()
	core.AssertEqual(t, wantOK, ok, "result")
	core.AssertEqual(t, !wantOK, mock.HasErrors(), "errors")
	checkNameForwarded(t, mock)
}

// runMustCase drives an AssertMust* helper through MockT.Run and checks
// the inner-abort vs. mock-failed outcomes against wantOK. Both
// mock.Failed() (set by FailNow) and mock.HasErrors() (set by Errorf in
// the bare assertion before the abort) must agree, and any recorded
// error must carry the name argument.
func runMustCase(t *testing.T, mock *core.MockT, wantOK bool,
	body func(core.T)) {
	t.Helper()
	ok := mock.Run("inner", body)
	core.AssertEqual(t, wantOK, ok, "inner result")
	core.AssertEqual(t, !wantOK, mock.Failed(), "mock failed")
	core.AssertEqual(t, !wantOK, mock.HasErrors(), "mock errors")
	checkNameForwarded(t, mock)
}

// checkNameForwarded asserts that the most recent recorded error, if any,
// contains nameProbe — proof that the Assert* helper threaded its name
// argument through to core.AssertTrue. On success rows the mock has no
// errors and the check is a silent no-op; on failure rows the recorded
// message must start with "<nameProbe>: " (the format produced by
// core.doMessage when a prefix is provided).
func checkNameForwarded(t *testing.T, mock *core.MockT) {
	t.Helper()
	msg, found := mock.LastError()
	if !found {
		return
	}
	core.AssertContains(t, msg, nameProbe, "name forwarded")
}

// waitForCondTestCase exercises the pure WaitForCond primitive.
type waitForCondTestCase struct {
	predicate func() func() bool
	name      string
	budgetMS  int
	want      bool
}

func newWaitForCondTestCase(name string, predicate func() func() bool,
	budgetMS int, want bool) waitForCondTestCase {
	return waitForCondTestCase{
		name:      name,
		predicate: predicate,
		budgetMS:  budgetMS,
		want:      want,
	}
}

func (tc waitForCondTestCase) Name() string { return tc.name }

func (tc waitForCondTestCase) Test(t *testing.T) {
	t.Helper()
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	got := synctesting.WaitForCond(tc.predicate(), budget, synctesting.PollStep)
	core.AssertEqual(t, tc.want, got, "result")
}

var _ core.TestCase = waitForCondTestCase{}

func TestWaitForCond(t *testing.T) {
	// The "zero timeout but predicate true" row pins the predicate-
	// before-deadline-check ordering: a regression that consulted the
	// deadline first would return false with timeout=0 even when the
	// predicate is true.
	core.RunTestCases(t, []waitForCondTestCase{
		newWaitForCondTestCase("true immediately", alwaysTrue, shortBudgetMS, true),
		newWaitForCondTestCase("becomes true within budget", flagFlip, shortBudgetMS, true),
		newWaitForCondTestCase("never true within budget", alwaysFalse, tinyBudgetMS, false),
		newWaitForCondTestCase("zero timeout", alwaysFalse, 0, false),
		newWaitForCondTestCase("zero timeout but predicate true", alwaysTrue, 0, true),
	})
}

// assertEventuallyTestCase exercises the bare AssertEventually helper.
type assertEventuallyTestCase struct {
	predicate func() func() bool
	name      string
	budgetMS  int
	wantOK    bool
}

func newAssertEventuallyTestCase(name string, predicate func() func() bool,
	budgetMS int, wantOK bool) assertEventuallyTestCase {
	return assertEventuallyTestCase{
		name:      name,
		predicate: predicate,
		budgetMS:  budgetMS,
		wantOK:    wantOK,
	}
}

func (tc assertEventuallyTestCase) Name() string { return tc.name }

func (tc assertEventuallyTestCase) Test(t *testing.T) {
	t.Helper()
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	ok := synctesting.AssertEventually(mock, tc.predicate(), budget, nameProbe)
	checkAssertOutcome(t, mock, ok, tc.wantOK)
}

var _ core.TestCase = assertEventuallyTestCase{}

func TestAssertEventually(t *testing.T) {
	core.RunTestCases(t, []assertEventuallyTestCase{
		newAssertEventuallyTestCase("success", alwaysTrue, shortBudgetMS, true),
		newAssertEventuallyTestCase("becomes true within budget", flagFlip, shortBudgetMS, true),
		newAssertEventuallyTestCase("timeout", alwaysFalse, tinyBudgetMS, false),
	})
}

// assertMustEventuallyTestCase exercises AssertMustEventually via MockT.Run.
type assertMustEventuallyTestCase struct {
	predicate func() func() bool
	name      string
	budgetMS  int
	wantOK    bool
}

func newAssertMustEventuallyTestCase(name string,
	predicate func() func() bool, budgetMS int,
	wantOK bool) assertMustEventuallyTestCase {
	return assertMustEventuallyTestCase{
		name:      name,
		predicate: predicate,
		budgetMS:  budgetMS,
		wantOK:    wantOK,
	}
}

func (tc assertMustEventuallyTestCase) Name() string { return tc.name }

func (tc assertMustEventuallyTestCase) Test(t *testing.T) {
	t.Helper()
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	runMustCase(t, mock, tc.wantOK, func(tt core.T) {
		synctesting.AssertMustEventually(tt, tc.predicate(), budget, nameProbe)
	})
}

var _ core.TestCase = assertMustEventuallyTestCase{}

func TestAssertMustEventually(t *testing.T) {
	core.RunTestCases(t, []assertMustEventuallyTestCase{
		newAssertMustEventuallyTestCase("success continues", alwaysTrue, shortBudgetMS, true),
		newAssertMustEventuallyTestCase("becomes true within budget continues",
			flagFlip, shortBudgetMS, true),
		newAssertMustEventuallyTestCase("timeout aborts", alwaysFalse, tinyBudgetMS, false),
	})
}

// assertClosedTestCase exercises the bare AssertClosed helper.
type assertClosedTestCase struct {
	setup    func(chan struct{})
	name     string
	budgetMS int
	wantOK   bool
}

func newAssertClosedTestCase(name string, setup func(chan struct{}),
	budgetMS int, wantOK bool) assertClosedTestCase {
	return assertClosedTestCase{
		name:     name,
		setup:    setup,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertClosedTestCase) Name() string { return tc.name }

func (tc assertClosedTestCase) Test(t *testing.T) {
	t.Helper()
	ch := make(chan struct{}, 1)
	tc.setup(ch)
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	ok := synctesting.AssertClosed(mock, ch, budget, nameProbe)
	checkAssertOutcome(t, mock, ok, tc.wantOK)
}

var _ core.TestCase = assertClosedTestCase{}

func TestAssertClosed(t *testing.T) {
	core.RunTestCases(t, []assertClosedTestCase{
		newAssertClosedTestCase("closed", chanClose, shortBudgetMS, true),
		newAssertClosedTestCase("value sent", chanSendOnce, shortBudgetMS, true),
		newAssertClosedTestCase("timeout", chanLeaveOpen, tinyBudgetMS, false),
	})
}

// assertMustClosedTestCase exercises AssertMustClosed via MockT.Run.
type assertMustClosedTestCase struct {
	setup    func(chan struct{})
	name     string
	budgetMS int
	wantOK   bool
}

func newAssertMustClosedTestCase(name string, setup func(chan struct{}),
	budgetMS int, wantOK bool) assertMustClosedTestCase {
	return assertMustClosedTestCase{
		name:     name,
		setup:    setup,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertMustClosedTestCase) Name() string { return tc.name }

func (tc assertMustClosedTestCase) Test(t *testing.T) {
	t.Helper()
	ch := make(chan struct{}, 1)
	tc.setup(ch)
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	runMustCase(t, mock, tc.wantOK, func(tt core.T) {
		synctesting.AssertMustClosed(tt, ch, budget, nameProbe)
	})
}

var _ core.TestCase = assertMustClosedTestCase{}

func TestAssertMustClosed(t *testing.T) {
	core.RunTestCases(t, []assertMustClosedTestCase{
		newAssertMustClosedTestCase("closed continues", chanClose, shortBudgetMS, true),
		newAssertMustClosedTestCase("value sent continues", chanSendOnce, shortBudgetMS, true),
		newAssertMustClosedTestCase("timeout aborts", chanLeaveOpen, tinyBudgetMS, false),
	})
}

// assertOpenTestCase exercises the bare AssertOpen helper.
type assertOpenTestCase struct {
	setup    func(chan struct{})
	name     string
	budgetMS int
	wantOK   bool
}

func newAssertOpenTestCase(name string, setup func(chan struct{}),
	budgetMS int, wantOK bool) assertOpenTestCase {
	return assertOpenTestCase{
		name:     name,
		setup:    setup,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertOpenTestCase) Name() string { return tc.name }

func (tc assertOpenTestCase) Test(t *testing.T) {
	t.Helper()
	ch := make(chan struct{}, 1)
	tc.setup(ch)
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	ok := synctesting.AssertOpen(mock, ch, budget, nameProbe)
	checkAssertOutcome(t, mock, ok, tc.wantOK)
}

var _ core.TestCase = assertOpenTestCase{}

func TestAssertOpen(t *testing.T) {
	core.RunTestCases(t, []assertOpenTestCase{
		newAssertOpenTestCase("stays open", chanLeaveOpen, tinyBudgetMS, true),
		newAssertOpenTestCase("closed fails", chanClose, shortBudgetMS, false),
		newAssertOpenTestCase("value sent fails", chanSendOnce, shortBudgetMS, false),
	})
}

// assertMustOpenTestCase exercises AssertMustOpen via MockT.Run.
type assertMustOpenTestCase struct {
	setup    func(chan struct{})
	name     string
	budgetMS int
	wantOK   bool
}

func newAssertMustOpenTestCase(name string, setup func(chan struct{}),
	budgetMS int, wantOK bool) assertMustOpenTestCase {
	return assertMustOpenTestCase{
		name:     name,
		setup:    setup,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertMustOpenTestCase) Name() string { return tc.name }

func (tc assertMustOpenTestCase) Test(t *testing.T) {
	t.Helper()
	ch := make(chan struct{}, 1)
	tc.setup(ch)
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	runMustCase(t, mock, tc.wantOK, func(tt core.T) {
		synctesting.AssertMustOpen(tt, ch, budget, nameProbe)
	})
}

var _ core.TestCase = assertMustOpenTestCase{}

func TestAssertMustOpen(t *testing.T) {
	core.RunTestCases(t, []assertMustOpenTestCase{
		newAssertMustOpenTestCase("stays open", chanLeaveOpen, tinyBudgetMS, true),
		newAssertMustOpenTestCase("close aborts", chanClose, shortBudgetMS, false),
		newAssertMustOpenTestCase("value sent aborts", chanSendOnce, shortBudgetMS, false),
	})
}

// assertReadersReadyTestCase exercises the bare AssertReadersReady helper.
type assertReadersReadyTestCase struct {
	setup    func() <-chan struct{}
	name     string
	expectN  int
	budgetMS int
	wantOK   bool
}

func newAssertReadersReadyTestCase(name string, setup func() <-chan struct{},
	expectN, budgetMS int, wantOK bool) assertReadersReadyTestCase {
	return assertReadersReadyTestCase{
		name:     name,
		setup:    setup,
		expectN:  expectN,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertReadersReadyTestCase) Name() string { return tc.name }

func (tc assertReadersReadyTestCase) Test(t *testing.T) {
	t.Helper()
	ch := tc.setup()
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	ok := synctesting.AssertReadersReady(mock, ch, tc.expectN, budget, nameProbe)
	checkAssertOutcome(t, mock, ok, tc.wantOK)
}

var _ core.TestCase = assertReadersReadyTestCase{}

func TestAssertReadersReady(t *testing.T) {
	// The "paced exceeds budget" row exercises the shared-deadline
	// property: four sends spaced 40 ms apart against a 100 ms budget —
	// a per-receive timeout would let all four through (each gap below
	// budget), the shared timeout cuts after two. The 2.5×pacing budget
	// gives ±20 ms margin around the deadline so CI load spikes don't
	// push the test onto an adjacent path.
	core.RunTestCases(t, []assertReadersReadyTestCase{
		newAssertReadersReadyTestCase("immediate full", chanWithSends(4), 4, shortBudgetMS, true),
		newAssertReadersReadyTestCase("immediate partial", chanWithSends(3), 4, tinyBudgetMS, false),
		newAssertReadersReadyTestCase("paced within budget", chanWithPacedSends(3, 2), 3, shortBudgetMS, true),
		newAssertReadersReadyTestCase("paced exceeds budget", chanWithPacedSends(4, 40), 4, pacedBudgetMS, false),
		newAssertReadersReadyTestCase("closed before n", chanClosedEmpty, 3, shortBudgetMS, false),
		newAssertReadersReadyTestCase("zero expected", chanWithSends(0), 0, tinyBudgetMS, true),
	})
}

// assertMustReadersReadyTestCase exercises AssertMustReadersReady via
// MockT.Run.
type assertMustReadersReadyTestCase struct {
	setup    func() <-chan struct{}
	name     string
	expectN  int
	budgetMS int
	wantOK   bool
}

func newAssertMustReadersReadyTestCase(name string,
	setup func() <-chan struct{}, expectN, budgetMS int,
	wantOK bool) assertMustReadersReadyTestCase {
	return assertMustReadersReadyTestCase{
		name:     name,
		setup:    setup,
		expectN:  expectN,
		budgetMS: budgetMS,
		wantOK:   wantOK,
	}
}

func (tc assertMustReadersReadyTestCase) Name() string { return tc.name }

func (tc assertMustReadersReadyTestCase) Test(t *testing.T) {
	t.Helper()
	ch := tc.setup()
	budget := time.Duration(tc.budgetMS) * time.Millisecond
	mock := &core.MockT{}
	runMustCase(t, mock, tc.wantOK, func(tt core.T) {
		synctesting.AssertMustReadersReady(tt, ch, tc.expectN, budget, nameProbe)
	})
}

var _ core.TestCase = assertMustReadersReadyTestCase{}

func TestAssertMustReadersReady(t *testing.T) {
	core.RunTestCases(t, []assertMustReadersReadyTestCase{
		newAssertMustReadersReadyTestCase("immediate full continues", chanWithSends(4), 4, shortBudgetMS, true),
		newAssertMustReadersReadyTestCase("immediate partial aborts", chanWithSends(3), 4, tinyBudgetMS, false),
		newAssertMustReadersReadyTestCase("paced within budget continues",
			chanWithPacedSends(3, 2), 3, shortBudgetMS, true),
		newAssertMustReadersReadyTestCase("paced exceeds budget aborts",
			chanWithPacedSends(4, 40), 4, pacedBudgetMS, false),
		newAssertMustReadersReadyTestCase("closed before n aborts",
			chanClosedEmpty, 3, shortBudgetMS, false),
		newAssertMustReadersReadyTestCase("zero expected continues", chanWithSends(0), 0, tinyBudgetMS, true),
	})
}

// argsPropagationTestCase exercises the args ...any propagation path of a
// bare Assert* helper. invoke calls the helper in a configuration that
// guarantees failure (so a message is recorded); the case verifies the
// formatted prefix landed in the recorded message.
type argsPropagationTestCase struct {
	invoke func(core.T) bool
	name   string
}

func newArgsPropagationTestCase(name string,
	invoke func(core.T) bool) argsPropagationTestCase {
	return argsPropagationTestCase{name: name, invoke: invoke}
}

func (tc argsPropagationTestCase) Name() string { return tc.name }

func (tc argsPropagationTestCase) Test(t *testing.T) {
	t.Helper()
	mock := &core.MockT{}
	ok := tc.invoke(mock)
	core.AssertEqual(t, false, ok, "result")
	msg, found := mock.LastError()
	core.AssertEqual(t, true, found, "error recorded")
	core.AssertContains(t, msg, argsExpected, "args forwarded")
}

var _ core.TestCase = argsPropagationTestCase{}

func TestAssertArgsPropagation(t *testing.T) {
	tiny := time.Duration(tinyBudgetMS) * time.Millisecond
	core.RunTestCases(t, []argsPropagationTestCase{
		newArgsPropagationTestCase("AssertEventually", func(tt core.T) bool {
			return synctesting.AssertEventually(tt, alwaysFalse(), tiny,
				argsFormat, argsValue)
		}),
		newArgsPropagationTestCase("AssertClosed", func(tt core.T) bool {
			ch := make(chan struct{}, 1)
			return synctesting.AssertClosed(tt, ch, tiny, argsFormat, argsValue)
		}),
		newArgsPropagationTestCase("AssertOpen", func(tt core.T) bool {
			ch := make(chan struct{}, 1)
			close(ch)
			return synctesting.AssertOpen(tt, ch, tiny, argsFormat, argsValue)
		}),
		newArgsPropagationTestCase("AssertReadersReady", func(tt core.T) bool {
			ch := make(chan struct{}, 1)
			return synctesting.AssertReadersReady(tt, ch, 1, tiny,
				argsFormat, argsValue)
		}),
	})
}

// argsPropagationMustTestCase exercises the args ...any propagation path of
// an AssertMust* helper. invoke calls the helper in a guaranteed-failing
// configuration; the case drives the abort through MockT.Run and verifies
// the formatted prefix landed in the recorded message.
type argsPropagationMustTestCase struct {
	invoke func(core.T)
	name   string
}

func newArgsPropagationMustTestCase(name string,
	invoke func(core.T)) argsPropagationMustTestCase {
	return argsPropagationMustTestCase{name: name, invoke: invoke}
}

func (tc argsPropagationMustTestCase) Name() string { return tc.name }

func (tc argsPropagationMustTestCase) Test(t *testing.T) {
	t.Helper()
	mock := &core.MockT{}
	ok := mock.Run("inner", tc.invoke)
	core.AssertEqual(t, false, ok, "inner result")
	core.AssertEqual(t, true, mock.Failed(), "mock failed")
	core.AssertEqual(t, true, mock.HasErrors(), "mock errors")
	msg, found := mock.LastError()
	core.AssertEqual(t, true, found, "error recorded")
	core.AssertContains(t, msg, argsExpected, "args forwarded")
}

var _ core.TestCase = argsPropagationMustTestCase{}

func TestAssertMustArgsPropagation(t *testing.T) {
	tiny := time.Duration(tinyBudgetMS) * time.Millisecond
	core.RunTestCases(t, []argsPropagationMustTestCase{
		newArgsPropagationMustTestCase("AssertMustEventually", func(tt core.T) {
			synctesting.AssertMustEventually(tt, alwaysFalse(), tiny,
				argsFormat, argsValue)
		}),
		newArgsPropagationMustTestCase("AssertMustClosed", func(tt core.T) {
			ch := make(chan struct{}, 1)
			synctesting.AssertMustClosed(tt, ch, tiny, argsFormat, argsValue)
		}),
		newArgsPropagationMustTestCase("AssertMustOpen", func(tt core.T) {
			ch := make(chan struct{}, 1)
			close(ch)
			synctesting.AssertMustOpen(tt, ch, tiny, argsFormat, argsValue)
		}),
		newArgsPropagationMustTestCase("AssertMustReadersReady", func(tt core.T) {
			ch := make(chan struct{}, 1)
			synctesting.AssertMustReadersReady(tt, ch, 1, tiny,
				argsFormat, argsValue)
		}),
	})
}

// metricRecorder captures ReportMetric calls keyed by unit label.
type metricRecorder map[string]float64

func (r metricRecorder) ReportMetric(n float64, unit string) { r[unit] = n }

var _ synctesting.MetricReporter = metricRecorder{}

// reportTryMetricsTestCase exercises ReportTryMetrics across its guard
// clauses. The unit noun is fixed to "lock": it only feeds string
// concatenation, never a branch. want lists every metric the call must
// emit; absent keys must not be emitted at all.
type reportTryMetricsTestCase struct {
	want map[string]float64

	name string

	elapsedMS int

	attempts int32
	count    int32
}

func newReportTryMetricsTestCase(name string, attempts, count int32,
	elapsedMS int, want map[string]float64) reportTryMetricsTestCase {
	return reportTryMetricsTestCase{
		name:      name,
		attempts:  attempts,
		count:     count,
		elapsedMS: elapsedMS,
		want:      want,
	}
}

func (tc reportTryMetricsTestCase) Name() string { return tc.name }

func (tc reportTryMetricsTestCase) Test(t *testing.T) {
	t.Helper()
	got := metricRecorder{}
	elapsed := time.Duration(tc.elapsedMS) * time.Millisecond
	synctesting.ReportTryMetrics(got, tc.attempts, tc.count, elapsed, "lock")
	core.AssertEqual(t, len(tc.want), len(got), "metric count")
	for unit, want := range tc.want {
		core.AssertEqual(t, want, got[unit], unit)
	}
}

var _ core.TestCase = reportTryMetricsTestCase{}

func TestReportTryMetrics(t *testing.T) {
	core.RunTestCases(t, []reportTryMetricsTestCase{
		newReportTryMetricsTestCase("no acquisitions", 10, 0, 1000,
			map[string]float64{}),
		newReportTryMetricsTestCase("no attempts", 0, 5, 1000,
			map[string]float64{}),
		newReportTryMetricsTestCase("zero elapsed", 10, 5, 0,
			map[string]float64{"attempts/lock": 2}),
		newReportTryMetricsTestCase("happy path", 10, 5, 2000,
			map[string]float64{
				"attempts/lock": 2,
				"locks/sec":     2.5,
				"ns/attempt":    2e8,
			}),
	})
}
