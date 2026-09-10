package num_test

import (
	"strconv"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var _ core.TestCase = sentinelTestCase{}

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

func sentinelTestCases() []sentinelTestCase {
	syntax := &num.ParseError{Func: "Parse", Num: "x", Err: num.ErrSyntax}
	rng := &num.ParseError{Func: "Parse", Num: "x", Err: num.ErrRange}
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
	err := error(&num.ParseError{Func: "ParseInt128", Num: "abc",
		Err: num.ErrSyntax})
	core.AssertEqual(t, `ParseInt128: parsing "abc": num: invalid syntax`,
		err.Error(), "text")

	p, ok := core.AssertErrorAs[*num.ParseError](t, err, "as")
	core.AssertMustTrue(t, ok, "as")
	ne := *p
	core.AssertEqual(t, "ParseInt128", ne.Func, "func")
	core.AssertEqual(t, "abc", ne.Num, "num")
	core.AssertSame(t, num.ErrSyntax, ne.Unwrap(), "unwrap")
}
