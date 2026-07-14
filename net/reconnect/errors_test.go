package reconnect_test

import (
	"context"
	"errors"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = isNotConnectedTestCase{}

// isNotConnectedTestCase tests the IsNotConnected predicate.
type isNotConnectedTestCase struct {
	err  error
	name string
	want bool
}

func newIsNotConnectedTestCase(name string, err error, want bool) isNotConnectedTestCase {
	return isNotConnectedTestCase{
		err:  err,
		name: name,
		want: want,
	}
}

func (tc isNotConnectedTestCase) Name() string {
	return tc.name
}

func (tc isNotConnectedTestCase) Test(t *testing.T) {
	t.Helper()

	got := reconnect.IsNotConnected(tc.err)
	core.AssertEqual(t, tc.want, got, "IsNotConnected")
}

func makeIsNotConnectedTestCases() []core.TestCase {
	return []core.TestCase{
		newIsNotConnectedTestCase("nil", nil, false),
		newIsNotConnectedTestCase("not connected", reconnect.ErrNotConnected, true),
		newIsNotConnectedTestCase("wrapped not connected",
			core.Wrap(reconnect.ErrNotConnected, "getSession"), true),
		// ErrNotConnected wraps ErrClosed, not the reverse: a bare closed
		// client must not report as not-connected.
		newIsNotConnectedTestCase("closed only", reconnect.ErrClosed, false),
		newIsNotConnectedTestCase("unrelated", core.ErrInvalid, false),
	}
}

func TestIsNotConnected(t *testing.T) {
	core.RunTestCases(t, makeIsNotConnectedTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = isClosedTestCase{}

// isClosedTestCase tests the IsClosed predicate.
type isClosedTestCase struct {
	err  error
	name string
	want bool
}

func newIsClosedTestCase(name string, err error, want bool) isClosedTestCase {
	return isClosedTestCase{
		err:  err,
		name: name,
		want: want,
	}
}

func (tc isClosedTestCase) Name() string {
	return tc.name
}

func (tc isClosedTestCase) Test(t *testing.T) {
	t.Helper()

	got := reconnect.IsClosed(tc.err)
	core.AssertEqual(t, tc.want, got, "IsClosed")
}

func makeIsClosedTestCases() []core.TestCase {
	return []core.TestCase{
		newIsClosedTestCase("nil", nil, false),
		newIsClosedTestCase("closed", reconnect.ErrClosed, true),
		// ErrNotConnected wraps ErrClosed, so IsClosed covers it too.
		newIsClosedTestCase("not connected", reconnect.ErrNotConnected, true),
		newIsClosedTestCase("wrapped closed",
			core.Wrap(reconnect.ErrClosed, "shutdown"), true),
		// ErrRunning wraps EBUSY, not ErrClosed: a running client is not closed.
		newIsClosedTestCase("running", reconnect.ErrRunning, false),
		newIsClosedTestCase("unrelated", core.ErrInvalid, false),
	}
}

func TestIsClosed(t *testing.T) {
	core.RunTestCases(t, makeIsClosedTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = invalidFamilyTestCase{}

// invalidFamilyTestCase pins which sentinels extend core.ErrInvalid, so a
// consumer matching the invalid-argument family with errors.Is keeps
// catching them.
type invalidFamilyTestCase struct {
	err  error
	name string
	want bool
}

func newInvalidFamilyTestCase(name string, err error, want bool) invalidFamilyTestCase {
	return invalidFamilyTestCase{
		err:  err,
		name: name,
		want: want,
	}
}

func (tc invalidFamilyTestCase) Name() string {
	return tc.name
}

func (tc invalidFamilyTestCase) Test(t *testing.T) {
	t.Helper()

	got := errors.Is(tc.err, core.ErrInvalid)
	core.AssertEqual(t, tc.want, got, "extends core.ErrInvalid")
}

func makeInvalidFamilyTestCases() []core.TestCase {
	return []core.TestCase{
		newInvalidFamilyTestCase("name empty", reconnect.ErrNameEmpty, true),
		newInvalidFamilyTestCase("name too long", reconnect.ErrNameTooLong, true),
		// Connection and control errors stay out of the invalid family.
		newInvalidFamilyTestCase("not connected", reconnect.ErrNotConnected, false),
		newInvalidFamilyTestCase("do not reconnect", reconnect.ErrDoNotReconnect, false),
	}
}

func TestInvalidFamily(t *testing.T) {
	core.RunTestCases(t, makeInvalidFamilyTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = isFatalTestCase{}

// isFatalTestCase tests the IsFatal predicate.
type isFatalTestCase struct {
	err  error
	name string
	want bool
}

func newIsFatalTestCase(name string, err error, want bool) isFatalTestCase {
	return isFatalTestCase{
		err:  err,
		name: name,
		want: want,
	}
}

func (tc isFatalTestCase) Name() string {
	return tc.name
}

func (tc isFatalTestCase) Test(t *testing.T) {
	t.Helper()

	got := reconnect.IsFatal(tc.err)
	core.AssertEqual(t, tc.want, got, "IsFatal")
}

func makeIsFatalTestCases() []core.TestCase {
	return []core.TestCase{
		// nil never reaches the loop's classifier, but the public
		// entry point must still report it as non-fatal.
		newIsFatalTestCase("nil", nil, false),
		newIsFatalTestCase("do not reconnect", reconnect.ErrDoNotReconnect, true),
		newIsFatalTestCase("wrapped do not reconnect",
			core.Wrap(reconnect.ErrDoNotReconnect, "waiter"), true),
		// Connection errors are recoverable, so the loop retries them.
		newIsFatalTestCase("not connected", reconnect.ErrNotConnected, false),
		newIsFatalTestCase("unrelated", core.ErrInvalid, false),
	}
}

func TestIsFatal(t *testing.T) {
	core.RunTestCases(t, makeIsFatalTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = isNonErrorTestCase{}

// isNonErrorTestCase tests the IsNonError predicate.
type isNonErrorTestCase struct {
	err  error
	name string
	want bool
}

func newIsNonErrorTestCase(name string, err error, want bool) isNonErrorTestCase {
	return isNonErrorTestCase{
		err:  err,
		name: name,
		want: want,
	}
}

func (tc isNonErrorTestCase) Name() string {
	return tc.name
}

func (tc isNonErrorTestCase) Test(t *testing.T) {
	t.Helper()

	got := reconnect.IsNonError(tc.err)
	core.AssertEqual(t, tc.want, got, "IsNonError")
}

func makeIsNonErrorTestCases() []core.TestCase {
	return []core.TestCase{
		// filterNonError only ever passes a non-nil error, so the
		// nil short-circuit is reachable solely through the public API.
		newIsNonErrorTestCase("nil", nil, true),
		newIsNonErrorTestCase("canceled", context.Canceled, true),
		newIsNonErrorTestCase("do not reconnect", reconnect.ErrDoNotReconnect, true),
		newIsNonErrorTestCase("wrapped canceled",
			core.Wrap(context.Canceled, "shutdown"), true),
		// A genuine failure must not be filtered out as a clean stop.
		newIsNonErrorTestCase("real error", reconnect.ErrNotConnected, false),
	}
}

func TestIsNonError(t *testing.T) {
	core.RunTestCases(t, makeIsNonErrorTestCases())
}
