package num_test

import (
	"errors"
	"strconv"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = parseErrorTextCase{}
	_ core.TestCase = sentinelDistinctTestCase{}
	_ core.TestCase = sentinelTestCase{}
)

// sentinelTestCase pins an error the package reports, a sentinel or a
// ParseError carrying one, to one of the errors it must match under
// errors.Is, and to the package prefix in its text.
type sentinelTestCase struct {
	err    error
	target error
	name   string
}

func newSentinelTestCase(name string, err, target error) sentinelTestCase {
	return sentinelTestCase{name: name, err: err, target: target}
}

func (tc sentinelTestCase) Name() string { return tc.name }

func (tc sentinelTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertErrorIs(t, tc.err, tc.target, "matches")
	core.AssertContains(t, tc.err.Error(), "num: ", "text")
}

// newNeutralParseError builds the report the tests of this file share,
// under a neutral function name: the sentinels belong to no one parser.
func newNeutralParseError(text string, cause error) *num.ParseError {
	return &num.ParseError{Func: "Parse", Num: text, Err: cause}
}

func sentinelTestCases() []sentinelTestCase {
	syntax := newNeutralParseError("x", num.ErrSyntax)
	rng := newNeutralParseError("x", num.ErrRange)
	return []sentinelTestCase{
		newSentinelTestCase("div zero is invalid", num.ErrDivZero,
			core.ErrInvalid),
		newSentinelTestCase("pow10 range is invalid", num.ErrPow10Range,
			core.ErrInvalid),
		newSentinelTestCase("syntax is strconv", num.ErrSyntax,
			strconv.ErrSyntax),
		newSentinelTestCase("syntax is invalid", num.ErrSyntax,
			core.ErrInvalid),
		newSentinelTestCase("range is strconv", num.ErrRange,
			strconv.ErrRange),
		newSentinelTestCase("range is invalid", num.ErrRange,
			core.ErrInvalid),
		newSentinelTestCase("num error is syntax", syntax, num.ErrSyntax),
		newSentinelTestCase("num error syntax is strconv", syntax,
			strconv.ErrSyntax),
		newSentinelTestCase("num error syntax is invalid", syntax,
			core.ErrInvalid),
		newSentinelTestCase("num error is range", rng, num.ErrRange),
		newSentinelTestCase("num error range is strconv", rng,
			strconv.ErrRange),
		newSentinelTestCase("num error range is invalid", rng,
			core.ErrInvalid),
	}
}

func TestSentinels(t *testing.T) {
	core.RunTestCases(t, sentinelTestCases())
}

// sentinelDistinctTestCase pins an error the package reports, a
// sentinel or a ParseError carrying one, apart from one of the errors
// the other parse sentinel matches, so that errors.Is tells a syntax
// failure from a range one.
type sentinelDistinctTestCase struct {
	err   error
	other error
	name  string
}

func newSentinelDistinctTestCase(name string,
	err, other error) sentinelDistinctTestCase {
	return sentinelDistinctTestCase{name: name, err: err, other: other}
}

func (tc sentinelDistinctTestCase) Name() string { return tc.name }

func (tc sentinelDistinctTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertNotErrorIs(t, tc.err, tc.other, "distinct")
}

func sentinelDistinctTestCases() []sentinelDistinctTestCase {
	syntax := newNeutralParseError("x", num.ErrSyntax)
	rng := newNeutralParseError("x", num.ErrRange)
	return []sentinelDistinctTestCase{
		newSentinelDistinctTestCase("syntax is not range", num.ErrSyntax,
			num.ErrRange),
		newSentinelDistinctTestCase("syntax is not strconv range",
			num.ErrSyntax, strconv.ErrRange),
		newSentinelDistinctTestCase("range is not syntax", num.ErrRange,
			num.ErrSyntax),
		newSentinelDistinctTestCase("range is not strconv syntax",
			num.ErrRange, strconv.ErrSyntax),
		newSentinelDistinctTestCase("num error syntax is not range",
			syntax, num.ErrRange),
		newSentinelDistinctTestCase("num error syntax is not strconv range",
			syntax, strconv.ErrRange),
		newSentinelDistinctTestCase("num error range is not syntax", rng,
			num.ErrSyntax),
		newSentinelDistinctTestCase("num error range is not strconv syntax",
			rng, strconv.ErrSyntax),
	}
}

func TestSentinelsDistinct(t *testing.T) {
	core.RunTestCases(t, sentinelDistinctTestCases())
}

// TestParseError pins how the report is matched: errors.As finds it as
// a *ParseError holding the function and the text it was given, and
// not as the strconv type it is defined over.
func TestParseError(t *testing.T) {
	err := error(newNeutralParseError("abc", num.ErrSyntax))
	p := core.AssertMustErrorAs[*num.ParseError](t, err, "as")
	core.AssertEqual(t, "Parse", p.Func, "func")
	core.AssertEqual(t, "abc", p.Num, "num")

	var ne *strconv.NumError
	core.AssertFalse(t, errors.As(err, &ne), "as strconv")
}

// parseErrorTextCase pins what a ParseError prints and unwraps to: a
// nil one prints nothing and unwraps to nil, one with no cause prints
// the function and the text alone and unwraps to nil, and one with a
// cause prints it after them and unwraps to it.
type parseErrorTextCase struct {
	err       *num.ParseError
	wantCause error
	want      string
	name      string
}

func newParseErrorTextCase(name string, err *num.ParseError,
	want string, wantCause error) parseErrorTextCase {
	return parseErrorTextCase{
		name:      name,
		err:       err,
		want:      want,
		wantCause: wantCause,
	}
}

func (tc parseErrorTextCase) Name() string { return tc.name }

func (tc parseErrorTextCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.err.Error(), "text")
	core.AssertSame(t, tc.wantCause, tc.err.Unwrap(), "unwrap")
}

func parseErrorTextCases() []parseErrorTextCase {
	return []parseErrorTextCase{
		newParseErrorTextCase("nil", nil, "", nil),
		newParseErrorTextCase("no cause",
			newNeutralParseError("x", nil),
			`Parse: parsing "x"`, nil),
		newParseErrorTextCase("cause",
			newNeutralParseError("x", num.ErrRange),
			`Parse: parsing "x": num: value out of range`, num.ErrRange),
	}
}

func TestParseErrorText(t *testing.T) {
	core.RunTestCases(t, parseErrorTextCases())
}
