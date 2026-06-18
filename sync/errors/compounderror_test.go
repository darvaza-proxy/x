package errors_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/sync/errors"
)

var (
	errOne = errors.New("error one")
	errTwo = errors.New("error two")
)

// newLoadedCompoundError returns a CompoundError pre-loaded with errs.
func newLoadedCompoundError(errs ...error) *errors.CompoundError {
	ce := new(errors.CompoundError)
	_ = ce.AppendError(errs...)
	return ce
}

// compoundErrorOKTestCase exercises CompoundError.OK across receiver states.
type compoundErrorOKTestCase struct {
	ce     *errors.CompoundError
	name   string
	expect bool
}

var _ core.TestCase = compoundErrorOKTestCase{}

func newCompoundErrorOKTestCase(name string, ce *errors.CompoundError,
	expect bool) compoundErrorOKTestCase {
	return compoundErrorOKTestCase{
		ce:     ce,
		name:   name,
		expect: expect,
	}
}

func (tc compoundErrorOKTestCase) Name() string { return tc.name }

func (tc compoundErrorOKTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.expect, tc.ce.OK(), "OK")
}

func compoundErrorOKTestCases() []compoundErrorOKTestCase {
	return []compoundErrorOKTestCase{
		newCompoundErrorOKTestCase("nil receiver", nil, true),
		newCompoundErrorOKTestCase("zero value",
			new(errors.CompoundError), true),
		newCompoundErrorOKTestCase("single error",
			newLoadedCompoundError(errOne), false),
		newCompoundErrorOKTestCase("multiple errors",
			newLoadedCompoundError(errOne, errTwo), false),
	}
}

func TestCompoundErrorOK(t *testing.T) {
	core.RunTestCases(t, compoundErrorOKTestCases())
}

// appendErrorTestCase exercises CompoundError.AppendError flattening and
// nil-skipping, and the chaining return value.
type appendErrorTestCase struct {
	name   string
	input  []error
	expect []error
}

var _ core.TestCase = appendErrorTestCase{}

func newAppendErrorTestCase(name string, input,
	expect []error) appendErrorTestCase {
	return appendErrorTestCase{
		name:   name,
		input:  input,
		expect: expect,
	}
}

func (tc appendErrorTestCase) Name() string { return tc.name }

func (tc appendErrorTestCase) Test(t *testing.T) {
	t.Helper()

	ce := new(errors.CompoundError)
	got := ce.AppendError(tc.input...)
	core.AssertSame(t, ce, got, "receiver")
	core.AssertSliceEqual(t, tc.expect, ce.Errors(), "errors")
}

func appendErrorTestCases() []appendErrorTestCase {
	compound := &core.CompoundError{Errs: core.S(errOne, errTwo)}
	return []appendErrorTestCase{
		newAppendErrorTestCase("preserves order",
			core.S(errOne, errTwo), core.S(errOne, errTwo)),
		newAppendErrorTestCase("skips nils",
			core.S(nil, errOne, nil, errTwo), core.S(errOne, errTwo)),
		newAppendErrorTestCase("flattens compound",
			core.S[error](compound), core.S(errOne, errTwo)),
		// empty results return a nil slice (slices.Clone of nil).
		newAppendErrorTestCase("all nil", core.S[error](nil, nil), nil),
		newAppendErrorTestCase("empty input", core.S[error](), nil),
	}
}

func TestCompoundErrorAppendError(t *testing.T) {
	core.RunTestCases(t, appendErrorTestCases())
}

func TestCompoundErrorAppend(t *testing.T) {
	t.Run("note only", runTestAppendNoteOnly)
	t.Run("wrap", runTestAppendWrap)
	t.Run("no-op", runTestAppendNoOp)
}

func runTestAppendNoteOnly(t *testing.T) {
	t.Helper()
	ce := new(errors.CompoundError)
	_ = ce.Append(nil, "boom %d", 42)

	core.AssertFalse(t, ce.OK(), "OK")
	core.AssertContains(t, ce.Error(), "boom 42", "Error")
}

func runTestAppendWrap(t *testing.T) {
	t.Helper()
	ce := new(errors.CompoundError)
	_ = ce.Append(errOne, "context")

	core.AssertErrorIs(t, ce.AsError(), errOne, "wrapped")
	core.AssertContains(t, ce.Error(), "context", "Error")
}

func runTestAppendNoOp(t *testing.T) {
	t.Helper()
	ce := new(errors.CompoundError)
	_ = ce.Append(nil, "")

	core.AssertTrue(t, ce.OK(), "OK")
}

// TestCompoundErrorErrorsSnapshot verifies Errors hands out an independent
// copy: appends after the snapshot do not mutate it.
func TestCompoundErrorErrorsSnapshot(t *testing.T) {
	ce := newLoadedCompoundError(errOne)
	snapshot := ce.Errors()
	_ = ce.AppendError(errTwo)

	core.AssertEqual(t, 1, len(snapshot), "snapshot len")
	core.AssertEqual(t, 2, len(ce.Errors()), "live len")
}

func TestCompoundErrorAsError(t *testing.T) {
	t.Run("empty is nil", runTestAsErrorEmpty)
	t.Run("non-empty is receiver", runTestAsErrorReceiver)
	t.Run("traversable", runTestAsErrorTraversable)
}

func runTestAsErrorEmpty(t *testing.T) {
	t.Helper()
	ce := new(errors.CompoundError)
	core.AssertNil(t, ce.AsError(), "AsError")
}

func runTestAsErrorReceiver(t *testing.T) {
	t.Helper()
	ce := newLoadedCompoundError(errOne)
	core.AssertSame(t, ce, ce.AsError(), "AsError receiver")
}

func runTestAsErrorTraversable(t *testing.T) {
	t.Helper()
	ce := newLoadedCompoundError(errOne, errTwo)
	core.AssertErrorIs(t, ce.AsError(), errTwo, "AsError Is")
}

func TestCompoundErrorError(t *testing.T) {
	t.Run("empty", runTestErrorEmpty)
	t.Run("joined", runTestErrorJoined)
}

func runTestErrorEmpty(t *testing.T) {
	t.Helper()
	ce := new(errors.CompoundError)
	core.AssertEqual(t, "", ce.Error(), "Error")
}

func runTestErrorJoined(t *testing.T) {
	t.Helper()
	ce := newLoadedCompoundError(errOne, errTwo)
	core.AssertEqual(t, "error one\nerror two", ce.Error(), "Error")
}

// TestCompoundErrorConcurrent appends from many goroutines at once; the
// race detector and the final count guard the locking.
func TestCompoundErrorConcurrent(t *testing.T) {
	const workers = 64

	ce := new(errors.CompoundError)
	err := core.RunConcurrentTest(t, workers, func(int) error {
		_ = ce.AppendError(errOne)
		return nil
	})

	core.AssertNoError(t, err, "concurrent")
	core.AssertEqual(t, workers, len(ce.Errors()), "count")
}

// TestCompoundErrorNilReceiver verifies every method is safe on a nil
// receiver, exercising the nil guards.
func TestCompoundErrorNilReceiver(t *testing.T) {
	var ce *errors.CompoundError

	core.AssertTrue(t, ce.OK(), "OK")
	core.AssertNil(t, ce.Errors(), "Errors")
	core.AssertNil(t, ce.Unwrap(), "Unwrap")
	core.AssertEqual(t, "", ce.Error(), "Error")
	core.AssertNil(t, ce.AsError(), "AsError")
	core.AssertNil(t, ce.AppendError(errOne), "AppendError")
	core.AssertNil(t, ce.Append(errOne, "note"), "Append")
	core.AssertTrue(t, ce.OK(), "OK after no-op appends")
}
