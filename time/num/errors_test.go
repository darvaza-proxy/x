package num_test

import (
	"errors"
	"strconv"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = parseErrorCase{}
	_ core.TestCase = sentinelTestCase{}
)

// sentinelTestCase pins an error the package returns, a sentinel or a
// ParseError carrying one, to one of the errors it must match under
// errors.Is, and to the package prefix on its text.
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

// newParseError builds the report the sentinel rows and the shape test
// share, under a neutral function name: the sentinels belong to no one
// parser.
func newParseError(text string, cause error) *num.ParseError {
	return &num.ParseError{Func: "Parse", Num: text, Err: cause}
}

func sentinelTestCases() []sentinelTestCase {
	syntax := newParseError("x", num.ErrSyntax)
	rng := newParseError("x", num.ErrRange)
	return []sentinelTestCase{
		newSentinelTestCase("div zero is invalid", num.ErrDivZero,
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

// TestParseError pins the report's shape: the function, the quoted input
// and the sentinel's text, with the sentinel as the unwrapped cause and
// the type reachable through errors.As.
func TestParseError(t *testing.T) {
	err := error(newParseError("abc", num.ErrSyntax))
	core.AssertEqual(t, `Parse: parsing "abc": num: invalid syntax`,
		err.Error(), "text")

	p, ok := core.AssertErrorAs[*num.ParseError](t, err, "as")
	core.AssertMustTrue(t, ok, "as")
	ne := *p
	core.AssertEqual(t, "Parse", ne.Func, "func")
	core.AssertEqual(t, "abc", ne.Num, "num")
	core.AssertSame(t, num.ErrSyntax, ne.Unwrap(), "unwrap")
}

// parseErrorCase pins what AsParseError reports for one error:
// nil for a nil, typed or not, and otherwise a ParseError naming the
// function and the text given, or those of the strconv.NumError or
// ParseError in when none were, its cause reaching both the error the
// row names and core.ErrInvalid.
type parseErrorCase struct {
	err       error
	wantCause error
	fn        string
	s         string
	wantFunc  string
	wantNum   string
	name      string
}

//revive:disable-next-line:argument-limit
func newParseErrorCase(name, fn, s string, err error,
	wantFunc, wantNum string, wantCause error) parseErrorCase {
	return parseErrorCase{
		name:      name,
		fn:        fn,
		s:         s,
		err:       err,
		wantFunc:  wantFunc,
		wantNum:   wantNum,
		wantCause: wantCause,
	}
}

// newParseErrorCaseNil pins an error AsParseError answers with
// nil, the function and text given as every parser gives them.
func newParseErrorCaseNil(name string, err error) parseErrorCase {
	return newParseErrorCase(name, "ParseInt32", "1", err, "", "", nil)
}

func (tc parseErrorCase) Name() string { return tc.name }

func (tc parseErrorCase) Test(t *testing.T) {
	t.Helper()
	err := num.AsParseError(tc.fn, tc.s, tc.err)
	if tc.wantCause == nil {
		core.AssertNoError(t, err, "error")
		return
	}
	pp, ok := core.AssertErrorAs[*num.ParseError](t, err, "report")
	core.AssertMustTrue(t, ok, "report")
	p := *pp
	core.AssertEqual(t, tc.wantFunc, p.Func, "func")
	core.AssertEqual(t, tc.wantNum, p.Num, "num")
	core.AssertErrorIs(t, p.Err, tc.wantCause, "cause")
	core.AssertErrorIs(t, p.Err, core.ErrInvalid, "invalid")
}

// newNumError builds strconv's own report, as its parsers return it.
func newNumError(fn, s string, err error) *strconv.NumError {
	return &strconv.NumError{Func: fn, Num: s, Err: err}
}

func parseErrorCases() []parseErrorCase {
	errUnknown := errors.New("unknown")
	report := &num.ParseError{Func: "ParseInt32", Num: "x", Err: num.ErrSyntax}
	return []parseErrorCase{
		newParseErrorCaseNil("nil", nil),
		newParseErrorCaseNil("typed nil num error",
			(*strconv.NumError)(nil)),
		newParseErrorCaseNil("typed nil parse error",
			(*num.ParseError)(nil)),
		newParseErrorCase("strconv syntax", "ParseInt32", "x",
			newNumError("ParseInt", "x", strconv.ErrSyntax),
			"ParseInt32", "x", num.ErrSyntax),
		newParseErrorCase("strconv range", "ParseInt32", "9",
			newNumError("ParseInt", "9", strconv.ErrRange),
			"ParseInt32", "9", num.ErrRange),
		newParseErrorCase("strconv parts kept", "", "",
			newNumError("ParseInt", "9", strconv.ErrRange),
			"ParseInt", "9", num.ErrRange),
		newParseErrorCase("bare strconv range", "ParseInt64", "9",
			strconv.ErrRange, "ParseInt64", "9", num.ErrRange),
		newParseErrorCase("bare strconv syntax", "ParseInt64", "x",
			strconv.ErrSyntax, "ParseInt64", "x", num.ErrSyntax),
		newParseErrorCase("own range", "ParseInt64", "9",
			num.ErrRange, "ParseInt64", "9", num.ErrRange),
		newParseErrorCase("own syntax", "ParseInt64", "x",
			num.ErrSyntax, "ParseInt64", "x", num.ErrSyntax),
		newParseErrorCase("invalid", "ParseInt64", "x",
			core.ErrInvalid, "ParseInt64", "x", core.ErrInvalid),
		newParseErrorCase("parse error parts kept", "", "",
			report, "ParseInt32", "x", num.ErrSyntax),
		newParseErrorCase("parse error renamed", "Parse", "x",
			report, "Parse", "x", num.ErrSyntax),
		newParseErrorCase("unknown cause", "ParseInt32", "x",
			errUnknown, "ParseInt32", "x", errUnknown),
	}
}

func TestAsParseError(t *testing.T) {
	core.RunTestCases(t, parseErrorCases())
}
