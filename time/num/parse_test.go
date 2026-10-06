package num_test

import (
	"cmp"
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
	_ core.TestCase = parseCase[num.Int32, *num.Int32]{}
	_ core.TestCase = parseCrossCase{}
	_ core.TestCase = parseFloatCase{}
	_ core.TestCase = parseFloatRefusedCase{}
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
	case *num.Uint128:
		*p, err = num.ParseUint128(s)
	case *num.Int128:
		*p, err = num.ParseInt128(s)
	case *num.Milli32:
		*p, err = num.ParseMilli32(s)
	case *num.Milli64:
		*p, err = num.ParseMilli64(s)
	case *num.Atto128:
		*p, err = num.ParseAtto128(s)
	default:
		err = parseFixtureAs(p, s)
	}
	return out, err
}

// parseFixtureAs parses s into p, a pointer to one of the outside
// instantiations these tests define, through the parser named for it.
func parseFixtureAs(p any, s string) error {
	var err error
	switch p := p.(type) {
	case *Centi64:
		*p, err = ParseCenti64(s)
	case *Unit64:
		*p, err = ParseUnit64(s)
	case *Zepto128:
		*p, err = ParseZepto128(s)
	default:
		// every parsed type that parseAs has no arm for has one here.
		err = core.NewUnreachableErrorf(0, nil, "no parser for %T", p)
	}
	return err
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
	case num.Uint128:
		return "ParseUint128"
	case num.Int128:
		return "ParseInt128"
	case num.Milli32:
		return "ParseMilli32"
	case num.Milli64:
		return "ParseMilli64"
	case num.Atto128:
		return "ParseAtto128"
	default:
		return fixtureFuncName(zero)
	}
}

// fixtureFuncName returns the parser name of v, one of the outside
// instantiations these tests define, which its report derives from
// the scaler's own name, num_test.Centi64 giving ParseCenti64.
func fixtureFuncName(v any) string {
	switch v.(type) {
	case Centi64:
		return "ParseCenti64"
	case Unit64:
		return "ParseUnit64"
	case Zepto128:
		return "ParseZepto128"
	default:
		// every parsed type that parseFuncName has no arm for has one
		// here.
		panic(core.NewUnreachableErrorf(0, nil,
			"no parser name for %T", v))
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
	core.AssertTrue(t, strings.HasPrefix(err.Error(), "UnmarshalText: "),
		"UnmarshalText prefix")
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
	seed32   = num.AsInt32(99)
	seed64   = num.AsInt64(99)
	seedU128 = num.AsUint128(99)
	seed128  = num.AsInt128(99)
	seedM32  = num.NewMilli32(9, 99)
	seedM64  = num.NewMilli64(9, 99)
	seedA128 = num.NewAtto128(9, 99)
	seedC64  = NewCenti64(9, 99)
	seedU64  = NewUnit64(99, 0)
	seedZ128 = NewZepto128(9, 99)
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

func parseUint128Cases() []parseCase[num.Uint128, *num.Uint128] {
	return []parseCase[num.Uint128, *num.Uint128]{
		newParseCase("zero", "0", seedU128, num.ZeroUint128),
		newParseCase("leading zeros", "007", seedU128, num.AsUint128(7)),
		newParseCase("low word", "42", seedU128, num.AsUint128(42)),
		newParseCase("max uint64", "18446744073709551615", seedU128,
			num.AsUint128(math.MaxUint64)),
		newParseCase("two to the 64", "18446744073709551616", seedU128,
			num.NewUint128(1, 0)),
		newParseCase("max", "340282366920938463463374607431768211455",
			seedU128, num.MaxUint128),
		newParseCase("forty digits", "0000000000000000000000000000000000000001",
			seedU128, num.AsUint128(1)),
		// past the range the value is the bound, as strconv has it.
		newParseCaseErr("two to the 128",
			"340282366920938463463374607431768211456", seedU128,
			num.MaxUint128, num.ErrRange),
		newParseCaseErr("far past", "1"+strings.Repeat("0", 39), seedU128,
			num.MaxUint128, num.ErrRange),
		// a text both malformed and too long reports the failure strconv
		// meets first, reading one digit at a time: the range once the
		// digits before the bad byte pass 128 bits, whether within the
		// group holding it or in a group before, and the syntax before
		// that.
		newParseCaseErr("letter past range", strings.Repeat("9", 41)+"x",
			seedU128, num.MaxUint128, num.ErrRange),
		newParseCaseErr("letter a group after the overflow",
			"1"+strings.Repeat("0", 57)+"x", seedU128, num.MaxUint128,
			num.ErrRange),
		newParseCaseErr("letter before range", "1x"+strings.Repeat("0", 40),
			seedU128, num.ZeroUint128, num.ErrSyntax),
		newParseCaseErr("plus", "+1", seedU128, num.ZeroUint128, num.ErrSyntax),
		newParseCaseErr("minus", "-1", seedU128, num.ZeroUint128,
			num.ErrSyntax),
		newParseCaseErr("empty", "", seedU128, num.ZeroUint128, num.ErrSyntax),
		newParseCaseErr("point", "1.0", seedU128, num.ZeroUint128,
			num.ErrSyntax),
		newParseCaseErr("underscore", "1_000", seedU128, num.ZeroUint128,
			num.ErrSyntax),
		newParseCaseErr("hex", "0x10", seedU128, num.ZeroUint128,
			num.ErrSyntax),
		newParseCaseErr("leading space", " 1", seedU128, num.ZeroUint128,
			num.ErrSyntax),
		newParseCaseErr("trailing letter", "12a", seedU128, num.ZeroUint128,
			num.ErrSyntax),
	}
}

func parseInt128Cases() []parseCase[num.Int128, *num.Int128] {
	return []parseCase[num.Int128, *num.Int128]{
		newParseCase("zero", "0", seed128, num.ZeroInt128),
		newParseCase("negative zero", "-0", seed128, num.ZeroInt128),
		newParseCase("negative", "-42", seed128, num.AsInt128(-42)),
		newParseCase("plus", "+7", seed128, num.AsInt128(7)),
		newParseCase("leading zeros", "007", seed128, num.AsInt128(7)),
		newParseCase("past int64", "9223372036854775808", seed128,
			num.NewInt128(0, 1<<63)),
		// -(2^63+1) in two's complement: the high word all ones, the low
		// word 2^63-1.
		newParseCase("below int64", "-9223372036854775809", seed128,
			num.NewInt128(math.MaxUint64, math.MaxInt64)),
		newParseCase("max", "170141183460469231731687303715884105727",
			seed128, num.MaxInt128),
		newParseCase("min", "-170141183460469231731687303715884105728",
			seed128, num.MinInt128),
		// past the range the value is the nearest bound, as strconv has it.
		newParseCaseErr("past max", "170141183460469231731687303715884105728",
			seed128, num.MaxInt128, num.ErrRange),
		newParseCaseErr("below min",
			"-170141183460469231731687303715884105729", seed128,
			num.MinInt128, num.ErrRange),
		newParseCaseErr("far past", strings.Repeat("9", 45), seed128,
			num.MaxInt128, num.ErrRange),
		newParseCaseErr("far below", "-"+strings.Repeat("9", 45), seed128,
			num.MinInt128, num.ErrRange),
		// the digits before the letter pass 128 bits, so the range
		// failure comes first, as strconv has it, clamped to the sign.
		newParseCaseErr("letter past range", "-"+strings.Repeat("9", 41)+"x",
			seed128, num.MinInt128, num.ErrRange),
		// the digits before the letter fit 128 bits though not the sign,
		// so the letter comes first, as strconv.ParseInt has it.
		newParseCaseErr("letter past sign bound",
			"170141183460469231731687303715884105728x", seed128,
			num.ZeroInt128, num.ErrSyntax),
		newParseCaseErr("empty", "", seed128, num.ZeroInt128, num.ErrSyntax),
		newParseCaseErr("minus only", "-", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("plus only", "+", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("two signs", "+-1", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("two minuses", "--1", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("point", "1.0", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("underscore", "1_000", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("trailing space", "1 ", seed128, num.ZeroInt128,
			num.ErrSyntax),
		newParseCaseErr("trailing letter", "12a", seed128, num.ZeroInt128,
			num.ErrSyntax),
	}
}

func parseMilli32Cases() []parseCase[num.Milli32, *num.Milli32] {
	return []parseCase[num.Milli32, *num.Milli32]{
		newParseCase("zero", "0", seedM32, num.NewMilli32(0, 0)),
		newParseCase("zero at resolution", "0.000", seedM32,
			num.NewMilli32(0, 0)),
		newParseCase("negative zero", "-0", seedM32, num.NewMilli32(0, 0)),
		newParseCase("whole", "42", seedM32, num.NewMilli32(42, 0)),
		newParseCase("half", "1.5", seedM32, num.NewMilli32(1, 500)),
		newParseCase("at resolution", "1.500", seedM32, num.NewMilli32(1, 500)),
		newParseCase("point on the right", "1.", seedM32, num.NewMilli32(1, 0)),
		newParseCase("point on the left", ".5", seedM32,
			num.NewMilli32(0, 500)),
		newParseCase("negative fraction", "-0.005", seedM32,
			num.NewMilli32(0, -5)),
		newParseCase("plus", "+1.5", seedM32, num.NewMilli32(1, 500)),
		newParseCase("leading zeros", "007.5", seedM32, num.NewMilli32(7, 500)),
		// digits below the resolution drop towards zero, with no error.
		newParseCase("truncated", "1.5009", seedM32, num.NewMilli32(1, 500)),
		newParseCase("truncated to zero", "0.0009", seedM32,
			num.NewMilli32(0, 0)),
		newParseCase("negative truncated to zero", "-0.0009", seedM32,
			num.NewMilli32(0, 0)),
		newParseCase("max", "2147483.647", seedM32,
			num.AsMilli32(math.MaxInt32)),
		newParseCase("min", "-2147483.648", seedM32,
			num.AsMilli32(math.MinInt32)),
		// past the range the value is the nearest bound, as strconv has it.
		newParseCaseErr("past max", "2147483.648", seedM32,
			num.AsMilli32(math.MaxInt32), num.ErrRange),
		newParseCaseErr("below min", "-2147483.649", seedM32,
			num.AsMilli32(math.MinInt32), num.ErrRange),
		newParseCaseErr("whole past max", "2147484", seedM32,
			num.AsMilli32(math.MaxInt32), num.ErrRange),
		newParseCaseErr("far past", strings.Repeat("9", 41), seedM32,
			num.AsMilli32(math.MaxInt32), num.ErrRange),
		newParseCaseErr("far below", "-"+strings.Repeat("9", 41), seedM32,
			num.AsMilli32(math.MinInt32), num.ErrRange),
		newParseCaseErr("empty", "", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("point only", ".", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("sign only", "-", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("sign and point", "-.", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("two points", "1.2.3", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("exponent", "1e3", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("underscore", "1_000.5", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("leading space", " 1.5", seedM32, num.NewMilli32(0, 0),
			num.ErrSyntax),
		newParseCaseErr("trailing letter", "1.5a", seedM32,
			num.NewMilli32(0, 0), num.ErrSyntax),
		// the text is read for its syntax before its size, as ParseFloat
		// has it, so a bad byte past the range is still a syntax failure.
		newParseCaseErr("letter past range", strings.Repeat("9", 41)+"x",
			seedM32, num.NewMilli32(0, 0), num.ErrSyntax),
		newParseCaseErr("letter below the resolution", "1.5009a", seedM32,
			num.NewMilli32(0, 0), num.ErrSyntax),
	}
}

func parseMilli64Cases() []parseCase[num.Milli64, *num.Milli64] {
	return []parseCase[num.Milli64, *num.Milli64]{
		newParseCase("zero", "0", seedM64, num.NewMilli64(0, 0)),
		newParseCase("half", "1.5", seedM64, num.NewMilli64(1, 500)),
		newParseCase("past milli32", "2147484.000", seedM64,
			num.NewMilli64(2147484, 0)),
		newParseCase("max", "9223372036854775.807", seedM64,
			num.AsMilli64(math.MaxInt64)),
		newParseCase("min", "-9223372036854775.808", seedM64,
			num.AsMilli64(math.MinInt64)),
		newParseCaseErr("past max", "9223372036854775.808", seedM64,
			num.AsMilli64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("below min", "-9223372036854775.809", seedM64,
			num.AsMilli64(math.MinInt64), num.ErrRange),
		newParseCaseErr("empty", "", seedM64, num.NewMilli64(0, 0),
			num.ErrSyntax),
		newParseCaseErr("point only", ".", seedM64, num.NewMilli64(0, 0),
			num.ErrSyntax),
	}
}

func parseAtto128Cases() []parseCase[num.Atto128, *num.Atto128] {
	return []parseCase[num.Atto128, *num.Atto128]{
		newParseCase("zero", "0", seedA128, num.NewAtto128(0, 0)),
		newParseCase("half", "1.5", seedA128, num.NewAtto128(1, 500e15)),
		newParseCase("one atto", "0.000000000000000001", seedA128,
			num.NewAtto128(0, 1)),
		newParseCase("truncated below one atto", "0.0000000000000000019",
			seedA128, num.NewAtto128(0, 1)),
		newParseCase("max", "170141183460469231731.687303715884105727",
			seedA128, num.AsAtto128(num.MaxInt128)),
		// the magnitude of the minimum is 2^127, which fits only under
		// the sign.
		newParseCase("min", "-170141183460469231731.687303715884105728",
			seedA128, num.AsAtto128(num.MinInt128)),
		newParseCaseErr("past max",
			"170141183460469231731.687303715884105728", seedA128,
			num.AsAtto128(num.MaxInt128), num.ErrRange),
		newParseCaseErr("below min",
			"-170141183460469231731.687303715884105729", seedA128,
			num.AsAtto128(num.MinInt128), num.ErrRange),
		// 41 digits pass 128 bits before the scale, so the whole count
		// is refused as it is read.
		newParseCaseErr("whole past the width", strings.Repeat("9", 41),
			seedA128, num.AsAtto128(num.MaxInt128), num.ErrRange),
		// 30 digits the whole count holds, and the atto scale then
		// takes past 128 bits, so the multiply is what refuses them.
		newParseCaseErr("count past the width", strings.Repeat("9", 30),
			seedA128, num.AsAtto128(num.MaxInt128), num.ErrRange),
		newParseCaseErr("empty", "", seedA128, num.NewAtto128(0, 0),
			num.ErrSyntax),
	}
}

func TestParse(t *testing.T) {
	t.Run("int32", runTestParseInt32)
	t.Run("int64", runTestParseInt64)
	t.Run("uint128", runTestParseUint128)
	t.Run("int128", runTestParseInt128)
	t.Run("milli32", runTestParseMilli32)
	t.Run("milli64", runTestParseMilli64)
	t.Run("atto128", runTestParseAtto128)
	t.Run("allocs", runTestParseAllocs)
}

// runTestParseAllocs checks a parse that succeeds allocates nothing.
// The report is the only thing a parser builds, so nothing but a
// failure may reach the heap; a Decimal names its own parser from the
// scaler, which is a concatenation, so its name belongs on that path
// alone.
func runTestParseAllocs(t *testing.T) {
	t.Helper()
	assertParseAllocs(t, num.ParseInt32, "-42", num.AsInt32(-42),
		"ParseInt32")
	assertParseAllocs(t, num.ParseInt64, "-42", num.AsInt64(-42),
		"ParseInt64")
	assertParseAllocs(t, num.ParseUint128, "42", num.AsUint128(42),
		"ParseUint128")
	assertParseAllocs(t, num.ParseInt128, "-42", num.AsInt128(-42),
		"ParseInt128")
	assertParseAllocs(t, num.ParseMilli32, "1.5", num.NewMilli32(1, 500),
		"ParseMilli32")
	assertParseAllocs(t, num.ParseMilli64, "1.5", num.NewMilli64(1, 500),
		"ParseMilli64")
	assertParseAllocs(t, num.ParseAtto128, "1.5", num.NewAtto128(1, 500e15),
		"ParseAtto128")
	// a fraction wider than one group of digits.
	assertParseAllocs(t, ParseZepto128, "1.5", newZepto128Milli(1500),
		"ParseZepto128")
}

// assertParseAllocs checks parse reads s as want and allocates nothing
// doing it. The value is asserted as well, so a parser answering the
// wrong thing cheaply still fails.
func assertParseAllocs[T any](t *testing.T, parse func(string) (T, error),
	s string, want T, name string) {
	t.Helper()
	var got T
	allocs := testing.AllocsPerRun(10, func() {
		got, _ = parse(s)
	})
	core.AssertEqual(t, want, got, "%s value", name)
	core.AssertEqual(t, 0, allocs, "%s allocs", name)
}

func runTestParseMilli32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseMilli32Cases())
}

func runTestParseMilli64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseMilli64Cases())
}

func runTestParseAtto128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseAtto128Cases())
}

func runTestParseUint128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseUint128Cases())
}

func runTestParseInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt128Cases())
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
	t.Run("uint128", runTestParseRoundTrip(textUint128Cases,
		roundTripOf[num.Uint128, *num.Uint128]("uint128", seedU128)))
	t.Run("int128", runTestParseRoundTrip(textInt128Cases,
		roundTripOf[num.Int128, *num.Int128]("int128", seed128)))
	t.Run("decimal", runTestParseRoundTrip(textDecimalCases,
		roundTripOf[num.Milli32, *num.Milli32]("milli32", seedM32),
		roundTripOf[num.Milli64, *num.Milli64]("milli64", seedM64),
		roundTripOf[num.Atto128, *num.Atto128]("atto128", seedA128)))
	t.Run("unit64", runTestParseRoundTrip(unit64TextCases,
		roundTripOf[Unit64, *Unit64]("unit64", seedU64)))
	t.Run("zepto128", runTestParseRoundTrip(zepto128TextCases,
		roundTripOf[Zepto128, *Zepto128]("zepto128", seedZ128)))
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

func runTestParseInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt32Cases())
}

func runTestParseInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt64Cases())
}

// parseCrossCase asserts the parsers answer one text as strconv answers
// it: the natives exactly as strconv.ParseInt at the same width, the
// value, the clamp included, and the class of failure alike; the
// 128-bit integers as the 64-bit strconv parser of their kind, ParseUint
// or ParseInt, the same syntax verdict on every text, the same value
// while strconv has one, and on its range failure a value past the
// bound strconv clamps to, on the same side. The corpus holds the forms
// strconv takes and the ones it refuses, so the rows enforce strconv's
// grammar rather than transcribe a reading of it.
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
	tc.testInt32(t)
	tc.testInt64(t)
	tc.testUint128(t)
	tc.testInt128(t)
}

func (tc parseCrossCase) testInt32(t *testing.T) {
	t.Helper()
	want, wantErr := strconv.ParseInt(tc.in, 10, 32)
	got, err := num.ParseInt32(tc.in)
	core.AssertEqual(t, num.AsInt32(int32(want)), got, "int32")
	assertSameFailure(t, wantErr, err, "int32")
}

func (tc parseCrossCase) testInt64(t *testing.T) {
	t.Helper()
	want, wantErr := strconv.ParseInt(tc.in, 10, 64)
	got, err := num.ParseInt64(tc.in)
	core.AssertEqual(t, num.AsInt64(want), got, "int64")
	assertSameFailure(t, wantErr, err, "int64")
}

func (tc parseCrossCase) testUint128(t *testing.T) {
	t.Helper()
	want, wantErr := strconv.ParseUint(tc.in, 10, 64)
	got, err := num.ParseUint128(tc.in)
	assertWiderParse(t, wantErr, err, "uint128")
	switch {
	case wantErr == nil:
		core.AssertEqual(t, num.AsUint128(want), got, "uint128")
	case errors.Is(wantErr, strconv.ErrRange):
		core.AssertEqual(t, cmp.Compare(want, 0),
			got.Cmp(num.AsUint128(want)), "uint128 past uint64")
	default:
		core.AssertEqual(t, num.ZeroUint128, got, "uint128 syntax")
	}
}

func (tc parseCrossCase) testInt128(t *testing.T) {
	t.Helper()
	want, wantErr := strconv.ParseInt(tc.in, 10, 64)
	got, err := num.ParseInt128(tc.in)
	assertWiderParse(t, wantErr, err, "int128")
	switch {
	case wantErr == nil:
		core.AssertEqual(t, num.AsInt128(want), got, "int128")
	case errors.Is(wantErr, strconv.ErrRange):
		core.AssertEqual(t, cmp.Compare(want, 0),
			got.Cmp(num.AsInt128(want)), "int128 past int64")
	default:
		core.AssertEqual(t, num.ZeroInt128, got, "int128 syntax")
	}
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

// assertWiderParse checks the answer of a wider parser against a
// narrower strconv one on the same text: the same syntax verdict, since
// the grammar is the same at any width, and success wherever strconv
// succeeds. A range failure of the narrower parser leaves the wider one
// free to succeed or to fail on its own range.
func assertWiderParse(t *testing.T, want, err error, name string) {
	t.Helper()
	core.AssertEqual(t, errors.Is(want, strconv.ErrSyntax),
		errors.Is(err, strconv.ErrSyntax), "%s syntax", name)
	if want == nil {
		core.AssertNoError(t, err, "%s", name)
	}
}

// parseCorpus is the texts every parser is run against strconv on:
// the forms it takes, the forms it refuses, and the bounds of every
// width and one past them.
func parseCorpus() []string {
	return core.S(
		"0", "-0", "+0", "7", "+7", "-7", "007", "-007", "+007",
		"2147483647", "2147483648", "-2147483648", "-2147483649",
		"9223372036854775807", "9223372036854775808",
		"-9223372036854775808", "-9223372036854775809",
		"18446744073709551615", "18446744073709551616",
		"-18446744073709551616", "99999999999999999999999",
		"170141183460469231731687303715884105727",
		"170141183460469231731687303715884105728",
		"-170141183460469231731687303715884105728",
		"-170141183460469231731687303715884105729",
		"340282366920938463463374607431768211455",
		"340282366920938463463374607431768211456",
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

// parseFloatText is a text a Milli64 and [strconv.ParseFloat], the
// parser of the shape a Decimal takes, are both given, and the name its
// row reports under. Each outcome is a row type embedding it with a
// Test of its own.
type parseFloatText struct {
	in   string
	name string
}

func (tc parseFloatText) Name() string { return tc.name }

// parseFloatCase is a text both parsers read alike. It carries its
// count in thousandths and holds both parsers to it: the Milli64 to the
// count itself, the float to the count divided by a thousand. Both
// sides of that division are exact in a float64 while the count stays
// below 2^53, as every row keeps it, so it rounds once, to the same
// float ParseFloat rounds the text to, and a text with a digit below
// the resolution reads as another value and fails the row.
type parseFloatCase struct {
	parseFloatText
	want int64
}

func newParseFloatCase(in string, want int64) parseFloatCase {
	return parseFloatCase{
		parseFloatText: parseFloatText{in: in, name: quoteName(in)},
		want:           want,
	}
}

func (tc parseFloatCase) Test(t *testing.T) {
	t.Helper()
	f, err := strconv.ParseFloat(tc.in, 64)
	core.AssertMustNoError(t, err, "ParseFloat")
	core.AssertEqual(t, float64(tc.want)/1000, f, "float")
	got, err := num.ParseMilli64(tc.in)
	core.AssertNoError(t, err, "parse")
	core.AssertEqual(t, num.AsMilli64(num.AsInt64(tc.want)), got, "value")
}

// parseFloatRefusedCase is a text this grammar refuses as malformed,
// with what ParseFloat answers for it: strconv.ErrSyntax for a form
// both refuse, and nil for a form ParseFloat takes, asserted on both
// sides so the divergence is pinned rather than assumed.
type parseFloatRefusedCase struct {
	floatErr error
	parseFloatText
}

func newParseFloatRefusedCase(name, in string,
	floatErr error) parseFloatRefusedCase {
	return parseFloatRefusedCase{
		parseFloatText: parseFloatText{in: in, name: name},
		floatErr:       floatErr,
	}
}

// newParseFloatCaseSyntax returns the row of a text both parsers refuse.
func newParseFloatCaseSyntax(in string) parseFloatRefusedCase {
	return newParseFloatRefusedCase(quoteName(in)+" syntax", in,
		strconv.ErrSyntax)
}

// newParseFloatCaseRefused returns the row of a form ParseFloat takes
// and this grammar does not.
func newParseFloatCaseRefused(in string) parseFloatRefusedCase {
	return newParseFloatRefusedCase(quoteName(in)+" refused", in, nil)
}

func (tc parseFloatRefusedCase) Test(t *testing.T) {
	t.Helper()
	_, err := strconv.ParseFloat(tc.in, 64)
	if tc.floatErr == nil {
		core.AssertNoError(t, err, "ParseFloat")
	} else {
		core.AssertErrorIs(t, err, tc.floatErr, "ParseFloat")
	}
	_, err = num.ParseMilli64(tc.in)
	core.AssertErrorIs(t, err, num.ErrSyntax, "refused")
}

// parseFloatCases holds the texts a Decimal and ParseFloat read alike,
// the forms both refuse, and the forms only ParseFloat takes: the
// exponent, the underscores and the hexadecimal mantissa, which this
// grammar leaves out, and the infinities and the not-a-number, which
// name no value this package holds.
func parseFloatCases() []core.TestCase {
	return core.S[core.TestCase](
		newParseFloatCase("0", 0),
		newParseFloatCase("-0", 0),
		newParseFloatCase("+0", 0),
		newParseFloatCase("1", 1000),
		newParseFloatCase("-1", -1000),
		newParseFloatCase("+1", 1000),
		newParseFloatCase("007", 7000),
		newParseFloatCase("1.5", 1500),
		newParseFloatCase("-1.5", -1500),
		newParseFloatCase("+1.5", 1500),
		newParseFloatCase("0.5", 500),
		newParseFloatCase(".5", 500),
		newParseFloatCase("-.5", -500),
		newParseFloatCase("+.5", 500),
		newParseFloatCase("1.", 1000),
		newParseFloatCase("-1.", -1000),
		newParseFloatCase("0.", 0),
		newParseFloatCase(".0", 0),
		newParseFloatCase("1.000", 1000),
		newParseFloatCase("0.001", 1),
		newParseFloatCase("-0.001", -1),
		newParseFloatCase("00.00", 0),
		newParseFloatCase("1234567.89", 1234567890),
		newParseFloatCaseSyntax(""),
		newParseFloatCaseSyntax("+"),
		newParseFloatCaseSyntax("-"),
		newParseFloatCaseSyntax("."),
		newParseFloatCaseSyntax("-."),
		newParseFloatCaseSyntax("+."),
		newParseFloatCaseSyntax("1.2.3"),
		newParseFloatCaseSyntax("1..2"),
		newParseFloatCaseSyntax(" 1"),
		newParseFloatCaseSyntax("1 "),
		newParseFloatCaseSyntax("12a"),
		newParseFloatCaseSyntax("a12"),
		newParseFloatCaseSyntax("1,000"),
		newParseFloatCaseSyntax("١٢"),
		newParseFloatCaseSyntax("∞"),
		newParseFloatCaseSyntax("0x10"),
		newParseFloatCaseSyntax("--1"),
		newParseFloatCaseSyntax("+-1"),
		newParseFloatCaseSyntax("1_"),
		newParseFloatCaseSyntax("_1"),
		// both read the whole text for its syntax before its size, so a
		// bad byte past the range or below the resolution is still a
		// syntax failure.
		newParseFloatCaseSyntax(strings.Repeat("9", 41)+"x"),
		newParseFloatCaseSyntax("1.5009a"),
		newParseFloatCaseRefused("1e3"),
		newParseFloatCaseRefused("1E3"),
		newParseFloatCaseRefused("1e-3"),
		newParseFloatCaseRefused("1.5e2"),
		newParseFloatCaseRefused("1_000"),
		newParseFloatCaseRefused("Inf"),
		newParseFloatCaseRefused("-Inf"),
		newParseFloatCaseRefused("NaN"),
		newParseFloatCaseRefused("0x1p-2"),
	)
}

func TestParseDecimalMatchesParseFloat(t *testing.T) {
	core.RunTestCases(t, parseFloatCases())
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
	t.Run("uint128", runTestUnmarshalTextNil[num.Uint128, *num.Uint128])
	t.Run("int128", runTestUnmarshalTextNil[num.Int128, *num.Int128])
	t.Run("milli32", runTestUnmarshalTextNil[num.Milli32, *num.Milli32])
	t.Run("milli64", runTestUnmarshalTextNil[num.Milli64, *num.Milli64])
	t.Run("atto128", runTestUnmarshalTextNil[num.Atto128, *num.Atto128])
}
