package num_test

import (
	"encoding"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = parseAllocsCase[num.Int32, *num.Int32]{}
	_ core.TestCase = parseCase[num.Int32, *num.Int32]{}
	_ core.TestCase = parseCrossCase{}
	_ core.TestCase = unmarshalTextNilCase[num.Int32, *num.Int32]{}
)

// textUnmarshaler is the pointer side of a parsed type: the pointer
// to T that reads the text T's parser takes.
type textUnmarshaler[T any] interface {
	*T
	encoding.TextUnmarshaler
}

// parseAs parses s as T through the function named for T, the one
// grammar a row pins. T is one of the parsed types of the family, so
// the switch always finds its arm. Each arm converts the value back
// into a T, which is the arm's own type so the conversion cannot fail,
// rather than storing through a pointer to it, so nothing escapes and
// a parse allocates no more than the parser does.
func parseAs[T num.Number[T]](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case num.Int32:
		x, err := num.ParseInt32(s)
		return core.MustOK(core.As[num.Int32, T](x)), err
	case num.Int64:
		x, err := num.ParseInt64(s)
		return core.MustOK(core.As[num.Int64, T](x)), err
	default:
		// T is one of the parsed types, each with an arm above.
		return zero, core.NewUnreachableErrorf(0, nil, "no parser for %s",
			core.TypeName[T]())
	}
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
		// T is one of the parsed types, each with an arm above.
		panic(core.NewUnreachableErrorf(0, nil, "no parser name for %s",
			core.TypeName[T]()))
	}
}

// parseCase pins one text against the parser of T and against
// UnmarshalText on *T: the value, and on failure the sentinel, the
// ParseError naming the parser and quoting the text, and the
// receiver left as it was. A range failure keeps strconv's clamped
// value on the parser's side, so the rows declare it. Every row also
// declares the value the receiver holds before the call, chosen apart
// from every value a row expects, so a store, or a missing one, shows.
type parseCase[T num.Number[T], PT textUnmarshaler[T]] struct {
	wantErr error
	seed    T
	want    T
	in      string
	name    string
}

func newParseCase[T num.Number[T], PT textUnmarshaler[T]](name, in string,
	seed, want T) parseCase[T, PT] {
	return parseCase[T, PT]{name: name, in: in, seed: seed, want: want}
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
	core.AssertMustNotEqual(t, tc.want, tc.seed, "seed")
	got := tc.seed
	err := PT(&got).UnmarshalText([]byte(tc.in))
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "UnmarshalText")
		core.AssertEqual(t, tc.want, got, "UnmarshalText value")
		return
	}
	core.AssertMustError(t, err, "UnmarshalText")
	core.AssertContains(t, err.Error(), "UnmarshalText", "method name")
	tc.assertError(t, err)
	core.AssertEqual(t, tc.seed, got, "UnmarshalText untouched")
}

// assertError checks err against the row: nil for a good text, and
// otherwise a ParseError whose cause is the sentinel, naming the
// parser and quoting the text.
func (tc parseCase[T, PT]) assertError(t *testing.T, err error) {
	t.Helper()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "error")
		return
	}
	p := core.AssertMustErrorAs[*num.ParseError](t, err, "report")
	core.AssertEqual(t, parseFuncName[T](), p.Func, "func")
	core.AssertEqual(t, tc.in, p.Num, "num")
	core.AssertSame(t, tc.wantErr, p.Unwrap(), "cause")
}

// The seeds every row starts the receiver from, a value no row
// expects.
var (
	seed32 = num.AsInt32(99)
	seed64 = num.AsInt64(99)
)

func parseInt32Cases() []parseCase[num.Int32, *num.Int32] {
	return []parseCase[num.Int32, *num.Int32]{
		newParseCase("zero", "0", seed32, num.AsInt32(0)),
		newParseCase("negative zero", "-0", seed32, num.AsInt32(0)),
		newParseCase("negative", "-42", seed32, num.AsInt32(-42)),
		newParseCase("plus", "+7", seed32, num.AsInt32(7)),
		newParseCase("leading zeros", "007", seed32, num.AsInt32(7)),
		newParseCase("max", "2147483647", seed32, num.AsInt32(math.MaxInt32)),
		newParseCase("min", "-2147483648", seed32, num.AsInt32(math.MinInt32)),
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
		newParseCase("zero", "0", seed64, num.AsInt64(0)),
		newParseCase("negative", "-42", seed64, num.AsInt64(-42)),
		newParseCase("past int32", "2147483648", seed64, num.AsInt64(1<<31)),
		newParseCase("max", "9223372036854775807", seed64,
			num.AsInt64(math.MaxInt64)),
		newParseCase("min", "-9223372036854775808", seed64,
			num.AsInt64(math.MinInt64)),
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
	t.Run("allocs", runTestParseAllocs)
}

func runTestParseInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt32Cases())
}

func runTestParseInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt64Cases())
}

// parseAllocsCase pins that a text which parses allocates nothing, read
// by the parser of T and by UnmarshalText on *T. The report is the only
// thing a parser builds, so nothing but a failure may reach the heap.
// The error and the value are asserted as well, so a parser answering
// the wrong thing cheaply still fails, and a failure at a bound, which
// returns that bound, fails on its error.
type parseAllocsCase[T num.Number[T], PT textUnmarshaler[T]] struct {
	want T
	in   string
	name string
}

func newParseAllocsCase[T num.Number[T], PT textUnmarshaler[T]](name,
	in string, want T) parseAllocsCase[T, PT] {
	return parseAllocsCase[T, PT]{name: name, in: in, want: want}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc parseAllocsCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc parseAllocsCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	tc.countParse(t)
	tc.countUnmarshalText(t)
}

// countParse reads the text through the parser of T.
func (tc parseAllocsCase[T, PT]) countParse(t *testing.T) {
	t.Helper()
	var got T
	var err error
	allocs := testing.AllocsPerRun(10, func() {
		got, err = parseAs[T](tc.in)
	})
	core.AssertNoError(t, err, "parse error")
	core.AssertEqual(t, tc.want, got, "parse value")
	core.AssertEqual(t, 0, allocs, "parse allocs")
}

// countUnmarshalText reads the text through UnmarshalText on *T, from
// bytes made before the count starts.
func (tc parseAllocsCase[T, PT]) countUnmarshalText(t *testing.T) {
	t.Helper()
	b := []byte(tc.in)
	var got T
	var err error
	allocs := testing.AllocsPerRun(10, func() {
		err = PT(&got).UnmarshalText(b)
	})
	core.AssertNoError(t, err, "UnmarshalText error")
	core.AssertEqual(t, tc.want, got, "UnmarshalText value")
	core.AssertEqual(t, 0, allocs, "UnmarshalText allocs")
}

// parseAllocsCases returns the allocation rows of T on text, the
// longest its width prints, which reads as want: as printed, and padded
// with zeros to 40 bytes, past the 32 the runtime converts from bytes
// to a string on the stack.
func parseAllocsCases[T num.Number[T], PT textUnmarshaler[T]](text string,
	want T) []parseAllocsCase[T, PT] {
	return []parseAllocsCase[T, PT]{
		newParseAllocsCase[T, PT]("longest", text, want),
		newParseAllocsCase[T, PT]("past the stack", zeroPad(text), want),
	}
}

// runTestParseAllocs checks a parse that succeeds allocates nothing on
// each type's longest text.
func runTestParseAllocs(t *testing.T) {
	t.Helper()
	int32s := parseAllocsCases[num.Int32, *num.Int32]("-2147483648",
		num.AsInt32(math.MinInt32))
	int64s := parseAllocsCases[num.Int64, *num.Int64]("-9223372036854775808",
		num.AsInt64(math.MinInt64))
	t.Run("int32", func(t *testing.T) { core.RunTestCases(t, int32s) })
	t.Run("int64", func(t *testing.T) { core.RunTestCases(t, int64s) })
}

// zeroPad returns s, a sign and its digits, with zeros after the sign
// to make it 40 bytes long, past the 32 bytes the runtime converts on
// the stack.
func zeroPad(s string) string {
	return s[:1] + strings.Repeat("0", 40-len(s)) + s[1:]
}

// textParseCases turns the rows of a text table whose value is a T into
// parse rows seeded with seed, so the parser must read back the value
// the writer printed for it; rows of the other types in the table are
// left out.
func textParseCases[T num.Number[T], PT textUnmarshaler[T]](
	cases []textCase, seed T) []parseCase[T, PT] {
	out := make([]parseCase[T, PT], 0, len(cases))
	for _, tc := range cases {
		if v, ok := tc.in.(T); ok {
			out = append(out, newParseCase[T, PT](tc.name, tc.want, seed, v))
		}
	}
	return out
}

func TestParseRoundTrip(t *testing.T) {
	t.Run("int", runTestParseRoundTrip(textIntCases,
		roundTripOf[num.Int32, *num.Int32]("int32", seed32),
		roundTripOf[num.Int64, *num.Int64]("int64", seed64)))
}

// roundTrip reads back the rows of a text table whose value is of one
// type, returning how many it took.
type roundTrip func(t *testing.T, cases []textCase) int

// runTestParseRoundTrip reads a text table back through the parsers
// sharing it, which together must take every row, so no value in the
// table prints without reading back.
func runTestParseRoundTrip(table func() []textCase,
	parsers ...roundTrip) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		cases := table()
		taken := 0
		for _, p := range parsers {
			taken += p(t, cases)
		}
		core.AssertEqual(t, len(cases), taken, "rows read back")
	}
}

// roundTripOf returns the roundTrip of T, running under name the rows
// of the table whose value is a T, of which there must be some.
func roundTripOf[T num.Number[T], PT textUnmarshaler[T]](name string,
	seed T) roundTrip {
	return func(t *testing.T, cases []textCase) int {
		t.Helper()
		rows := textParseCases[T, PT](cases, seed)
		t.Run(name, func(t *testing.T) {
			core.AssertMustNotEqual(t, 0, len(rows), "rows")
			core.RunTestCases(t, rows)
		})
		return len(rows)
	}
}

// parseCrossCase asserts the native parsers answer one text exactly as
// strconv.ParseInt answers it at the same width: the same value, the
// clamp included, and the same class of failure. Each text is run,
// through the parser and through UnmarshalText on the pointer, as the
// parseCase row built from strconv's answer, so the whole corpus
// reaches the bytes path too. The corpus holds the forms strconv takes
// and the ones it refuses, so the rows enforce strconv's grammar rather
// than transcribe a reading of it.
type parseCrossCase struct {
	in   string
	name string
}

func newParseCrossCase(in string) parseCrossCase {
	return parseCrossCase{name: quoteName(in), in: in}
}

// quoteName returns a text as the Go literal a row is named by, its
// spaces escaped too: a sub-test name turns a space into an underscore,
// so " 1" would otherwise report under the name of "_1".
func quoteName(s string) string {
	return strings.ReplaceAll(strconv.Quote(s), " ", `\x20`)
}

func (tc parseCrossCase) Name() string { return tc.name }

func (tc parseCrossCase) Test(t *testing.T) {
	t.Helper()
	want32, err32 := strconv.ParseInt(tc.in, 10, 32)
	t.Run("int32", newParseCaseFromStrconv(tc.name, tc.in, seed32,
		num.AsInt32(int32(want32)), err32).Test)

	want64, err64 := strconv.ParseInt(tc.in, 10, 64)
	t.Run("int64", newParseCaseFromStrconv(tc.name, tc.in, seed64,
		num.AsInt64(want64), err64).Test)
}

// newParseCaseFromStrconv returns the parseCase row of strconv's answer
// on in: its value, and on a failure the package's sentinel of the
// same class.
func newParseCaseFromStrconv[T num.Number[T], PT textUnmarshaler[T]](name,
	in string, seed, want T, err error) parseCase[T, PT] {
	switch {
	case err == nil:
		return newParseCase[T, PT](name, in, seed, want)
	case errors.Is(err, strconv.ErrSyntax):
		return newParseCaseErr[T, PT](name, in, seed, want, num.ErrSyntax)
	case errors.Is(err, strconv.ErrRange):
		return newParseCaseErr[T, PT](name, in, seed, want, num.ErrRange)
	default:
		// strconv in base 10 at a valid width fails on syntax or range.
		panic(core.NewUnreachableError(0, err, "strconv failure"))
	}
}

// parseCorpus is the texts every parser is run against strconv on:
// the forms it takes, the bounds of both native widths and one past
// them, the edges of the digit run, and the forms it refuses.
func parseCorpus() []string {
	return core.S(
		"0",
		"-0",
		"+0",
		"7",
		"+7",
		"-7",
		"007",
		"-007",
		"+007",

		"2147483647",
		"2147483648",
		"-2147483648",
		"-2147483649",
		"9223372036854775807",
		"9223372036854775808",
		"-9223372036854775808",
		"-9223372036854775809",
		"18446744073709551616",
		"99999999999999999999999",
		"0000000000000000000000000000000000000001",

		// a group of 19 digits and one past it, the largest magnitude
		// of each width's run and one past it, a group boundary
		// reached after leading zeros, and a stray byte before an
		// overflowing digit and after one.
		"1234567890123456789",
		"-1234567890123456789",
		"12345678901234567890",
		"-12345678901234567890",
		"4294967295",
		"4294967296",
		"18446744073709551615",
		"-18446744073709551615",
		"-18446744073709551616",
		"000000000000000000018446744073709551615",
		"00000000001234567890123456789",
		"1234567890123456789x",
		"429496729x6",
		"4294967296x",
		"1844674407370955161x6",
		"18446744073709551615x",
		"18446744073709551616x",

		"",
		"+",
		"-",
		"+-1",
		"-+1",
		"--1",
		"++1",
		" 1",
		"1 ",
		"1 2",
		"\t1",
		"1\n",
		"1.0",
		".5",
		"1.",
		"1e3",
		"1E3",
		"1e-3",
		"1_000",
		"_1",
		"1_",
		"0x10",
		"0X10",
		"0b1",
		"0o7",
		"0_7",
		"1,000",
		"12a",
		"a12",
		"١٢",
		"Inf",
		"-Inf",
		"NaN",
		"∞",
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

// unmarshalTextNilCase pins one text read by UnmarshalText on a nil
// receiver: the one cause the row expects, of core.ErrNilReceiver and
// ErrSyntax, and the method's name in front. The text is parsed first,
// so a text that parses fails with core.ErrNilReceiver, and one that
// does not reports its own failure.
type unmarshalTextNilCase[T any, PT textUnmarshaler[T]] struct {
	wantErr error
	in      string
	name    string
}

// newUnmarshalTextNilCase returns a row on a text that parses, failing
// on the nil receiver.
func newUnmarshalTextNilCase[T any, PT textUnmarshaler[T]](name,
	in string) unmarshalTextNilCase[T, PT] {
	return unmarshalTextNilCase[T, PT]{name: name, in: in,
		wantErr: core.ErrNilReceiver}
}

// newUnmarshalTextNilCaseSyntax returns a row on a text that does not
// parse, failing on the text before the nil receiver.
func newUnmarshalTextNilCaseSyntax[T any, PT textUnmarshaler[T]](name,
	in string) unmarshalTextNilCase[T, PT] {
	return unmarshalTextNilCase[T, PT]{name: name, in: in,
		wantErr: num.ErrSyntax}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc unmarshalTextNilCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc unmarshalTextNilCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	var p PT
	err := p.UnmarshalText([]byte(tc.in))
	core.AssertMustError(t, err, "UnmarshalText")
	core.AssertErrorIs(t, err, tc.wantErr, "cause")
	core.AssertNotEqual(t, errors.Is(err, core.ErrNilReceiver),
		errors.Is(err, num.ErrSyntax), "one cause")
	core.AssertContains(t, err.Error(), "UnmarshalText", "method name")
}

// unmarshalTextNilCases returns the nil-receiver rows every type reads
// alike: a text that parses and one that does not.
func unmarshalTextNilCases[T any,
	PT textUnmarshaler[T]]() []unmarshalTextNilCase[T, PT] {
	return []unmarshalTextNilCase[T, PT]{
		newUnmarshalTextNilCase[T, PT]("good text", "1"),
		newUnmarshalTextNilCaseSyntax[T, PT]("bad text", "x"),
	}
}

func TestUnmarshalTextNilReceiver(t *testing.T) {
	int32s := unmarshalTextNilCases[num.Int32, *num.Int32]()
	int64s := unmarshalTextNilCases[num.Int64, *num.Int64]()
	t.Run("int32", func(t *testing.T) { core.RunTestCases(t, int32s) })
	t.Run("int64", func(t *testing.T) { core.RunTestCases(t, int64s) })
}
