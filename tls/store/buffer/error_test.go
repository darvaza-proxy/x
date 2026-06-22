package buffer_test

import (
	"errors"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/store/buffer"
)

// errBoom is a sentinel base error for the anonymous-source wrap cases.
var errBoom = errors.New("boom")

// newErrorTestCase exercises SourceName.NewError for anonymous sources
// (FileName == ""); named sources route through NewPathError, covered in
// source_test.go. Every row expects a non-nil error whose text carries the
// annotation token; the nil result is checked in TestSourceNameNewErrorNil.
type newErrorTestCase struct {
	base     error
	name     string
	op       string
	note     string
	wantText string
}

var _ core.TestCase = newErrorTestCase{}

func newNewErrorTestCase(name string, base error, op, note,
	wantText string) newErrorTestCase {
	return newErrorTestCase{
		base:     base,
		name:     name,
		op:       op,
		note:     note,
		wantText: wantText,
	}
}

func (tc newErrorTestCase) Name() string { return tc.name }

func (tc newErrorTestCase) Test(t *testing.T) {
	t.Helper()

	sn := buffer.NewSourceName(nil, "")
	err := sn.NewError(tc.base, tc.op, tc.note)

	core.AssertMustNotNil(t, err, "err")
	core.AssertContains(t, err.Error(), tc.wantText, "annotation")
	if tc.base != nil {
		core.AssertErrorIs(t, err, tc.base, "wraps base")
	}
}

func newErrorTestCases() []newErrorTestCase {
	return []newErrorTestCase{
		// err == nil: doNewError2 synthesises a fresh error.
		newNewErrorTestCase("op and note", nil, "read", "bad", "read: bad"),
		newNewErrorTestCase("note only", nil, "", "bad", "bad"),
		newNewErrorTestCase("op only", nil, "read", "", "read"),
		// err != nil: the annotation must wrap the base. F52 regression: the
		// single-annotation arms wrapped with the empty variable, so
		// core.Wrap returned the bare base and the annotation was lost.
		newNewErrorTestCase("wrap pass-through", errBoom, "", "", "boom"),
		newNewErrorTestCase("wrap note only", errBoom, "", "label", "label"),
		newNewErrorTestCase("wrap op only", errBoom, "verb", "", "verb"),
		newNewErrorTestCase("wrap op and note", errBoom, "verb", "label",
			"verb: label"),
	}
}

func TestSourceNameNewError(t *testing.T) {
	core.RunTestCases(t, newErrorTestCases())
}

// TestSourceNameNewErrorNil confirms an anonymous source with neither error
// nor annotation yields nil.
func TestSourceNameNewErrorNil(t *testing.T) {
	sn := buffer.NewSourceName(nil, "")
	core.AssertNil(t, sn.NewError(nil, "", ""), "nil")
}

// TestSourceNameNewErrorf confirms the formatted annotation reaches the
// wrapped error.
func TestSourceNameNewErrorf(t *testing.T) {
	sn := buffer.NewSourceName(nil, "")
	err := sn.NewErrorf(errBoom, "op", "count=%d", 3)

	core.AssertMustNotNil(t, err, "err")
	core.AssertContains(t, err.Error(), "count=3", "formatted note")
	core.AssertErrorIs(t, err, errBoom, "wraps base")
}
