package num

import (
	"errors"
	"strconv"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = parseCauseTestCase{}

// parseCauseTestCase pins the very error parseCause answers with for
// one in, by identity: nil and the errors it passes through, already
// matching core.ErrInvalid at any depth, and the package sentinel each
// strconv sentinel becomes.
type parseCauseTestCase struct {
	err  error
	want error
	name string
}

func newParseCauseTestCase(name string, err, want error) parseCauseTestCase {
	return parseCauseTestCase{name: name, err: err, want: want}
}

func (tc parseCauseTestCase) Name() string { return tc.name }

func (tc parseCauseTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertSame(t, tc.want, parseCause(tc.err), "cause")
}

func parseCauseTestCases() []parseCauseTestCase {
	wrapped := core.Wrap(ErrRange, "outer")
	return []parseCauseTestCase{
		newParseCauseTestCase("nil", nil, nil),
		newParseCauseTestCase("own range", ErrRange, ErrRange),
		newParseCauseTestCase("own syntax", ErrSyntax, ErrSyntax),
		newParseCauseTestCase("invalid", core.ErrInvalid, core.ErrInvalid),
		newParseCauseTestCase("invalid wrapped", ErrDivZero, ErrDivZero),
		newParseCauseTestCase("own range wrapped", wrapped, wrapped),
		newParseCauseTestCase("strconv range", strconv.ErrRange, ErrRange),
		newParseCauseTestCase("strconv syntax", strconv.ErrSyntax, ErrSyntax),
	}
}

func TestParseCause(t *testing.T) {
	core.RunTestCases(t, parseCauseTestCases())
}

// TestParseCauseUnknown pins the answer for an error parseCause does
// not know: a new error reaching it and core.ErrInvalid alike, so the
// cause is kept and the report still reads as invalid.
func TestParseCauseUnknown(t *testing.T) {
	errUnknown := errors.New("unknown")
	err := parseCause(errUnknown)
	core.AssertNotSame(t, errUnknown, err, "compound")
	core.AssertErrorIs(t, err, errUnknown, "cause")
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid")
}
