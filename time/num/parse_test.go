package num_test

import (
	"encoding"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = parseCase[num.Int32, *num.Int32]{}
	_ core.TestCase = parseCrossCase{}
)

// textUnmarshaler is the pointer side of a parsed type: the pointer
// to T that reads the text T's parser takes.
type textUnmarshaler[T any] interface {
	*T
	encoding.TextUnmarshaler
}

// parseAs parses s as T through the function named for T, the one
// grammar a row pins. T is one of the parsed types of the family, so
// the switch always finds its arm.
func parseAs[T num.Number[T]](s string) (T, error) {
	var out T
	var err error
	switch p := any(&out).(type) {
	case *num.Int32:
		*p, err = num.ParseInt32(s)
	case *num.Int64:
		*p, err = num.ParseInt64(s)
	default:
		// T is one of the parsed types, so no arm is left over.
	}
	return out, err
}

// parseFuncName returns the name a ParseError reports for the parser
// of T, the bare strconv shape.
func parseFuncName[T num.Number[T]]() string {
	var zero T
	switch any(zero).(type) {
	case num.Int32:
		return "ParseInt32"
	case num.Int64:
		return "ParseInt64"
	default:
		// T is one of the parsed types, so no arm is left over.
		return ""
	}
}

// parseCase pins one text against the parser of T and against
// UnmarshalText on *T: the value, and on failure the sentinel, the
// ParseError naming the parser and quoting the text, and the
// receiver left as it was. A range failure keeps strconv's clamped
// value on the parser's side, so the rows declare it; a failure row
// also declares the value the receiver holds before the call, chosen
// apart from both zero and the clamp so a store of either would show.
type parseCase[T num.Number[T], PT textUnmarshaler[T]] struct {
	wantErr error
	seed    T
	want    T
	in      string
	name    string
}

func newParseCase[T num.Number[T], PT textUnmarshaler[T]](name, in string,
	want T) parseCase[T, PT] {
	return parseCase[T, PT]{name: name, in: in, want: want}
}

func newParseCaseErr[T num.Number[T], PT textUnmarshaler[T]](name, in string,
	seed, want T, wantErr error) parseCase[T, PT] {
	return parseCase[T, PT]{name: name, in: in, seed: seed, want: want,
		wantErr: wantErr}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc parseCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc parseCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	got, err := parseAs[T](tc.in)
	core.AssertEqual(t, tc.want, got, "value")
	tc.assertError(t, err)
	tc.testUnmarshalText(t)
}

// testUnmarshalText reads the text through the pointer side, which
// stores the value on success and, on failure, wraps the parser's
// report in its own name and leaves the seeded receiver alone.
func (tc parseCase[T, PT]) testUnmarshalText(t *testing.T) {
	t.Helper()
	got := tc.seed
	err := PT(&got).UnmarshalText([]byte(tc.in))
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "UnmarshalText")
		core.AssertEqual(t, tc.want, got, "UnmarshalText value")
		return
	}
	core.AssertMustError(t, err, "UnmarshalText")
	core.AssertTrue(t, strings.HasPrefix(err.Error(), "UnmarshalText: "),
		"UnmarshalText prefix")
	tc.assertError(t, err)
	core.AssertEqual(t, tc.seed, got, "UnmarshalText untouched")
}

// assertError checks err against the row: nil for a good text, and
// otherwise a ParseError reaching the sentinel, naming the parser and
// quoting the text.
func (tc parseCase[T, PT]) assertError(t *testing.T, err error) {
	t.Helper()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "error")
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "sentinel")
	p, ok := core.AssertErrorAs[*num.ParseError](t, err, "report")
	core.AssertMustTrue(t, ok, "report")
	report := *p
	core.AssertEqual(t, parseFuncName[T](), report.Func, "func")
	core.AssertEqual(t, tc.in, report.Num, "num")
	core.AssertSame(t, tc.wantErr, report.Unwrap(), "cause")
}

// The seeds a failure row leaves in the receiver: apart from zero, the
// syntax answer, and from every bound, the range answers.
var (
	seed32 = num.AsInt32(99)
	seed64 = num.AsInt64(99)
)

func parseInt32Cases() []parseCase[num.Int32, *num.Int32] {
	return []parseCase[num.Int32, *num.Int32]{
		newParseCase("zero", "0", num.AsInt32(0)),
		newParseCase("negative zero", "-0", num.AsInt32(0)),
		newParseCase("negative", "-42", num.AsInt32(-42)),
		newParseCase("plus", "+7", num.AsInt32(7)),
		newParseCase("leading zeros", "007", num.AsInt32(7)),
		newParseCase("max", "2147483647", num.AsInt32(math.MaxInt32)),
		newParseCase("min", "-2147483648", num.AsInt32(math.MinInt32)),
		// past the range the value is the nearest bound, as strconv has it.
		newParseCaseErr("past max", "2147483648", seed32,
			num.AsInt32(math.MaxInt32), num.ErrRange),
		newParseCaseErr("below min", "-2147483649", seed32,
			num.AsInt32(math.MinInt32), num.ErrRange),
		newParseCaseErr("far past", "99999999999999999999", seed32,
			num.AsInt32(math.MaxInt32), num.ErrRange),
		newParseCaseErr("empty", "", seed32, num.AsInt32(0), num.ErrSyntax),
		newParseCaseErr("sign only", "-", seed32, num.AsInt32(0), num.ErrSyntax),
		newParseCaseErr("two signs", "+-1", seed32, num.AsInt32(0),
			num.ErrSyntax),
		newParseCaseErr("point", "1.0", seed32, num.AsInt32(0), num.ErrSyntax),
		newParseCaseErr("exponent", "1e3", seed32, num.AsInt32(0),
			num.ErrSyntax),
		newParseCaseErr("underscore", "1_000", seed32, num.AsInt32(0),
			num.ErrSyntax),
		newParseCaseErr("hex", "0x10", seed32, num.AsInt32(0), num.ErrSyntax),
		newParseCaseErr("leading space", " 1", seed32, num.AsInt32(0),
			num.ErrSyntax),
		newParseCaseErr("trailing space", "1 ", seed32, num.AsInt32(0),
			num.ErrSyntax),
		newParseCaseErr("trailing letter", "12a", seed32, num.AsInt32(0),
			num.ErrSyntax),
	}
}

func parseInt64Cases() []parseCase[num.Int64, *num.Int64] {
	return []parseCase[num.Int64, *num.Int64]{
		newParseCase("zero", "0", num.AsInt64(0)),
		newParseCase("negative", "-42", num.AsInt64(-42)),
		newParseCase("past int32", "2147483648", num.AsInt64(1<<31)),
		newParseCase("max", "9223372036854775807", num.AsInt64(math.MaxInt64)),
		newParseCase("min", "-9223372036854775808", num.AsInt64(math.MinInt64)),
		newParseCaseErr("past max", "9223372036854775808", seed64,
			num.AsInt64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("below min", "-9223372036854775809", seed64,
			num.AsInt64(math.MinInt64), num.ErrRange),
		newParseCaseErr("two to the 64", "18446744073709551616", seed64,
			num.AsInt64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("empty", "", seed64, num.AsInt64(0), num.ErrSyntax),
		newParseCaseErr("point", "1.0", seed64, num.AsInt64(0), num.ErrSyntax),
	}
}

func TestParse(t *testing.T) {
	t.Run("int32", runTestParseInt32)
	t.Run("int64", runTestParseInt64)
}

func runTestParseInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt32Cases())
}

func runTestParseInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt64Cases())
}

// parseCrossCase asserts the native parsers answer one text exactly as
// strconv.ParseInt answers it at the same width: the same value, the
// clamp included, and the same class of failure. The corpus holds the
// forms strconv takes and the ones it refuses, so the rows enforce
// strconv's grammar rather than transcribe my reading of it.
type parseCrossCase struct {
	in   string
	name string
}

func newParseCrossCase(in string) parseCrossCase {
	return parseCrossCase{name: fmt.Sprintf("%q", in), in: in}
}

func (tc parseCrossCase) Name() string { return tc.name }

func (tc parseCrossCase) Test(t *testing.T) {
	t.Helper()
	want32, wantErr32 := strconv.ParseInt(tc.in, 10, 32)
	got32, err32 := num.ParseInt32(tc.in)
	core.AssertEqual(t, num.AsInt32(int32(want32)), got32, "int32")
	assertSameFailure(t, wantErr32, err32, "int32")

	want64, wantErr64 := strconv.ParseInt(tc.in, 10, 64)
	got64, err64 := num.ParseInt64(tc.in)
	core.AssertEqual(t, num.AsInt64(want64), got64, "int64")
	assertSameFailure(t, wantErr64, err64, "int64")
}

// assertSameFailure checks err fails when want does and matches the
// same strconv sentinel, syntax or range.
func assertSameFailure(t *testing.T, want, err error, name string) {
	t.Helper()
	core.AssertEqual(t, want == nil, err == nil, "%s failed", name)
	core.AssertEqual(t, errors.Is(want, strconv.ErrSyntax),
		errors.Is(err, strconv.ErrSyntax), "%s syntax", name)
	core.AssertEqual(t, errors.Is(want, strconv.ErrRange),
		errors.Is(err, strconv.ErrRange), "%s range", name)
}

// parseCorpus is the texts every parser is run against strconv on:
// the forms it takes, the forms it refuses, and the bounds of both
// native widths and one past them.
func parseCorpus() []string {
	return core.S(
		"0", "-0", "+0", "7", "+7", "-7", "007", "-007", "+007",
		"2147483647", "2147483648", "-2147483648", "-2147483649",
		"9223372036854775807", "9223372036854775808",
		"-9223372036854775808", "-9223372036854775809",
		"18446744073709551616", "99999999999999999999999",
		"0000000000000000000000000000000000000001",
		"", "+", "-", "+-1", "-+1", "--1", "++1",
		" 1", "1 ", "1 2", "\t1", "1\n",
		"1.0", ".5", "1.", "1e3", "1E3", "1e-3",
		"1_000", "_1", "1_", "0x10", "0X10", "0b1", "0o7", "0_7",
		"1,000", "12a", "a12", "١٢", "Inf", "-Inf", "NaN", "∞",
	)
}

func parseCrossCases() []parseCrossCase {
	corpus := parseCorpus()
	cases := make([]parseCrossCase, 0, len(corpus))
	for _, in := range corpus {
		cases = append(cases, newParseCrossCase(in))
	}
	return cases
}

func TestParseMatchesStrconv(t *testing.T) {
	core.RunTestCases(t, parseCrossCases())
}

// runTestUnmarshalTextNil checks the nil receiver: the text is parsed
// first, so a bad text reports its own failure, and a good one reports
// core.ErrNilReceiver behind the method's name.
func runTestUnmarshalTextNil[T any, PT textUnmarshaler[T]](t *testing.T) {
	t.Helper()
	var p PT
	err := p.UnmarshalText([]byte("1"))
	core.AssertMustError(t, err, "nil receiver")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "nil receiver")
	core.AssertTrue(t, strings.HasPrefix(err.Error(), "UnmarshalText: "),
		"nil receiver prefix")

	err = p.UnmarshalText([]byte("x"))
	core.AssertErrorIs(t, err, num.ErrSyntax, "syntax first")
	core.AssertNotErrorIs(t, err, core.ErrNilReceiver, "not nil receiver")
}

func TestUnmarshalTextNilReceiver(t *testing.T) {
	t.Run("int32", runTestUnmarshalTextNil[num.Int32, *num.Int32])
	t.Run("int64", runTestUnmarshalTextNil[num.Int64, *num.Int64])
}
