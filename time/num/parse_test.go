package num_test

import (
	"encoding"
	"errors"
	"math"
	"math/big"
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
	_ core.TestCase = parseFloatCase{}
	_ core.TestCase = parseFloatRefusedCase{}
	_ core.TestCase = unmarshalTextNilCase[num.Int32, *num.Int32]{}
)

// textUnmarshaler is the pointer side of a parsed type: the pointer
// to T that reads the text T's parser takes.
type textUnmarshaler[T any] interface {
	*T
	encoding.TextUnmarshaler
}

// parseAs parses s as T through the function named for T, the one
// grammar a row pins. A parsed type of the package has its arm here,
// and an outside instantiation these tests define has its own in
// parseFixtureAs. Each arm converts the value back into a T, which is
// the arm's own type so the conversion cannot fail, rather than storing
// through a pointer to it, so nothing escapes and a parse allocates no
// more than the parser does.
func parseAs[T num.Number[T]](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case num.Int32:
		x, err := num.ParseInt32(s)
		return core.MustOK(core.As[num.Int32, T](x)), err
	case num.Int64:
		x, err := num.ParseInt64(s)
		return core.MustOK(core.As[num.Int64, T](x)), err
	case num.Uint128:
		x, err := num.ParseUint128(s)
		return core.MustOK(core.As[num.Uint128, T](x)), err
	case num.Int128:
		x, err := num.ParseInt128(s)
		return core.MustOK(core.As[num.Int128, T](x)), err
	case num.Milli32:
		x, err := num.ParseMilli32(s)
		return core.MustOK(core.As[num.Milli32, T](x)), err
	case num.Milli64:
		x, err := num.ParseMilli64(s)
		return core.MustOK(core.As[num.Milli64, T](x)), err
	case num.Atto128:
		x, err := num.ParseAtto128(s)
		return core.MustOK(core.As[num.Atto128, T](x)), err
	default:
		return parseFixtureAs[T](s)
	}
}

// parseFixtureAs parses s as T, one of the outside instantiations
// these tests define, through the parser named for it, converting the
// value back as parseAs does.
func parseFixtureAs[T num.Number[T]](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case Centi64:
		x, err := ParseCenti64(s)
		return core.MustOK(core.As[Centi64, T](x)), err
	case Unit64:
		x, err := ParseUnit64(s)
		return core.MustOK(core.As[Unit64, T](x)), err
	case Unit128:
		x, err := ParseUnit128(s)
		return core.MustOK(core.As[Unit128, T](x)), err
	case Zepto128:
		x, err := ParseZepto128(s)
		return core.MustOK(core.As[Zepto128, T](x)), err
	case OutsideCenti:
		x, err := ParseOutsideCenti(s)
		return core.MustOK(core.As[OutsideCenti, T](x)), err
	case OutsidePlainCenti:
		x, err := ParseOutsidePlainCenti(s)
		return core.MustOK(core.As[OutsidePlainCenti, T](x)), err
	default:
		// every parsed type that parseAs has no arm for has one here.
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
		return fixtureFuncName[T]()
	}
}

// fixtureFuncName returns the parser name of T, one of the outside
// instantiations these tests define, which its report derives from
// the scaler's own name, num_test.Centi64 giving ParseCenti64.
func fixtureFuncName[T num.Number[T]]() string {
	var zero T
	switch any(zero).(type) {
	case Centi64:
		return "ParseCenti64"
	case Unit64:
		return "ParseUnit64"
	case Unit128:
		return "ParseUnit128"
	case Zepto128:
		return "ParseZepto128"
	case OutsideCenti:
		return "ParseOutsideCenti"
	case OutsidePlainCenti:
		return "ParseOutsidePlainCenti"
	default:
		// every parsed type that parseFuncName has no arm for has one
		// here.
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
	assertParseReport[T](t, err, tc.in, tc.wantErr)
}

// assertParseReport checks err holds a ParseError whose cause is
// wantErr, naming the parser of T and quoting text, the text it read.
func assertParseReport[T num.Number[T]](t *testing.T, err error,
	text string, wantErr error) {
	t.Helper()
	p := core.AssertMustErrorAs[*num.ParseError](t, err, "report")
	core.AssertEqual(t, parseFuncName[T](), p.Func, "func")
	core.AssertEqual(t, text, p.Num, "num")
	core.AssertSame(t, wantErr, p.Unwrap(), "cause")
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
	seedOC   = NewOutsideCenti(9, 99)
	seedOPC  = NewOutsidePlainCenti(9, 99)

	// the Unit128 seed takes the long name, seedU128 being Uint128's.
	seedUnit128 = NewUnit128(99, 0)
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
		newParseCase("max", maxUint128Text, seedU128, num.MaxUint128),
		newParseCase("forty digits", "0000000000000000000000000000000000000001",
			seedU128, num.AsUint128(1)),
		// past the range the value is the bound, as strconv has it.
		newParseCaseErr("two to the 128",
			"340282366920938463463374607431768211456", seedU128,
			num.MaxUint128, num.ErrRange),
		newParseCaseErr("far past", "1"+strings.Repeat("0", 39), seedU128,
			num.MaxUint128, num.ErrRange),
		// 2^128 * 10^18, three full groups of 19 digits, wraps to zero
		// past 128 bits, so the group after it must not read as a fit.
		newParseCaseErr("past the wrap to zero",
			"340282366920938463463374607431768211456"+
				strings.Repeat("0", 18)+"1", seedU128,
			num.MaxUint128, num.ErrRange),
		// a text both malformed and too long fails on range once the
		// digits before the bad byte pass 128 bits, and on syntax while
		// they fit, as strconv has it, reading one digit at a time.
		newParseCaseErr("letter after max", maxUint128Text+"x", seedU128,
			num.ZeroUint128, num.ErrSyntax),
		newParseCaseErr("letter after two to the 128",
			"340282366920938463463374607431768211456x", seedU128,
			num.MaxUint128, num.ErrRange),
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
		newParseCase("max", maxInt128Text, seed128, num.MaxInt128),
		newParseCase("min", minInt128Text, seed128, num.MinInt128),
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
		newParseCaseErr("letter after two to the 128",
			"340282366920938463463374607431768211456x", seed128,
			num.MaxInt128, num.ErrRange),
		newParseCaseErr("letter below minus two to the 128",
			"-340282366920938463463374607431768211456x", seed128,
			num.MinInt128, num.ErrRange),
		// the digits before the letter fit 128 bits, within the sign's
		// bound or past it, so the letter comes first, as
		// strconv.ParseInt has it.
		newParseCaseErr("letter past sign bound",
			"170141183460469231731687303715884105728x", seed128,
			num.ZeroInt128, num.ErrSyntax),
		newParseCaseErr("letter after min", minInt128Text+"x", seed128,
			num.ZeroInt128, num.ErrSyntax),
		newParseCaseErr("letter below sign bound",
			"-170141183460469231731687303715884105729x", seed128,
			num.ZeroInt128, num.ErrSyntax),
		newParseCaseErr("letter after minus max uint128",
			"-340282366920938463463374607431768211455x", seed128,
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
		// the fraction is still read past a whole that passed 64 bits.
		newParseCaseErr("far past with a fraction",
			strings.Repeat("9", 41)+".5", seedM32,
			num.AsMilli32(math.MaxInt32), num.ErrRange),
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
		newParseCaseErr("whole past max", "9223372036854776", seedM64,
			num.AsMilli64(math.MaxInt64), num.ErrRange),
		// a whole within 64 bits whose count at the resolution passes
		// them, and one whose count the fraction carries past them.
		newParseCaseErr("count past 64 bits", "18446744073709552", seedM64,
			num.AsMilli64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("count below 64 bits", "-18446744073709552", seedM64,
			num.AsMilli64(math.MinInt64), num.ErrRange),
		newParseCaseErr("count carried past 64 bits", "18446744073709551.999",
			seedM64, num.AsMilli64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("far past", strings.Repeat("9", 41), seedM64,
			num.AsMilli64(math.MaxInt64), num.ErrRange),
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
		// the largest whole the atto scale takes, the whole of MaxInt128,
		// and one past it on either side, which no fraction brings back.
		newParseCase("whole at the bound", "170141183460469231731", seedA128,
			num.AsAtto128(num.MaxInt128.Sub(num.AsInt128(687303715884105727)))),
		newParseCaseErr("whole past the bound", "170141183460469231732",
			seedA128, num.AsAtto128(num.MaxInt128), num.ErrRange),
		newParseCaseErr("whole below the bound", "-170141183460469231732",
			seedA128, num.AsAtto128(num.MinInt128), num.ErrRange),
		// 41 digits pass 128 bits, so the scanner caps the whole at
		// MaxUint128, past the largest whole the atto scale takes.
		newParseCaseErr("whole past the width", strings.Repeat("9", 41),
			seedA128, num.AsAtto128(num.MaxInt128), num.ErrRange),
		// 30 digits fit in 128 bits but pass the largest whole the atto
		// scale takes, which atScale checks before it multiplies.
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

func runTestParseInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt32Cases())
}

func runTestParseInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, parseInt64Cases())
}

func runTestParseMilli32(t *testing.T) {
	core.RunTestCases(t, parseMilli32Cases())
}

func runTestParseMilli64(t *testing.T) {
	core.RunTestCases(t, parseMilli64Cases())
}

func runTestParseAtto128(t *testing.T) {
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

// parseAllocsCase pins that a text which parses allocates nothing, read
// by the parser of T and by UnmarshalText on *T. The error and the
// value are asserted as well, so a parser answering the wrong thing
// cheaply still fails, and a failure at a bound, which returns that
// bound, fails on its error.
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
// each type's longest text: those of the natives, Milli32 and Milli64
// as printed and padded, and those of the 128-bit integers, Atto128 and
// Zepto128 as printed alone, already past the 32 bytes, 39 for
// MaxUint128, 40 for MinInt128 and 41 for the minimum of Atto128 and of
// Zepto128, the latter's fraction wider than one group of digits. A
// Decimal names its own parser from the scaler, which is a
// concatenation, so its name belongs on the failure path alone.
func runTestParseAllocs(t *testing.T) {
	t.Helper()
	int32s := parseAllocsCases[num.Int32, *num.Int32]("-2147483648",
		num.AsInt32(math.MinInt32))
	int64s := parseAllocsCases[num.Int64, *num.Int64]("-9223372036854775808",
		num.AsInt64(math.MinInt64))
	uint128s := core.S(newParseAllocsCase[num.Uint128, *num.Uint128]("longest",
		maxUint128Text, num.MaxUint128))
	int128s := core.S(newParseAllocsCase[num.Int128, *num.Int128]("longest",
		minInt128Text, num.MinInt128))
	milli32s := parseAllocsCases[num.Milli32, *num.Milli32]("-2147483.648",
		num.AsMilli32(math.MinInt32))
	milli64s := parseAllocsCases[num.Milli64, *num.Milli64](
		"-9223372036854775.808", num.AsMilli64(math.MinInt64))
	atto128s := core.S(newParseAllocsCase[num.Atto128, *num.Atto128]("longest",
		minAtto128Text, num.AsAtto128(num.MinInt128)))
	zepto128s := core.S(newParseAllocsCase[Zepto128, *Zepto128]("longest",
		minZepto128Text, AsZepto128(num.MinInt128)))
	t.Run("int32", func(t *testing.T) { core.RunTestCases(t, int32s) })
	t.Run("int64", func(t *testing.T) { core.RunTestCases(t, int64s) })
	t.Run("uint128", func(t *testing.T) { core.RunTestCases(t, uint128s) })
	t.Run("int128", func(t *testing.T) { core.RunTestCases(t, int128s) })
	t.Run("milli32", func(t *testing.T) { core.RunTestCases(t, milli32s) })
	t.Run("milli64", func(t *testing.T) { core.RunTestCases(t, milli64s) })
	t.Run("atto128", func(t *testing.T) { core.RunTestCases(t, atto128s) })
	t.Run("zepto128", func(t *testing.T) { core.RunTestCases(t, zepto128s) })
}

// The texts of the 128-bit bounds; those of MaxUint128 and of the
// minimum of the others are the widest each type prints.
const (
	maxUint128Text  = "340282366920938463463374607431768211455"
	maxInt128Text   = "170141183460469231731687303715884105727"
	minInt128Text   = "-170141183460469231731687303715884105728"
	minAtto128Text  = "-170141183460469231731.687303715884105728"
	minZepto128Text = "-170141183460469231.731687303715884105728"
)

// zeroPad returns s, a sign and the number after it, with zeros after
// the sign to make it 40 bytes long, past the 32 bytes the runtime
// converts on the stack.
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
	t.Run("outside centi", runTestParseRoundTrip(outsideCentiTextCases,
		roundTripOf[OutsideCenti, *OutsideCenti]("outside centi", seedOC)))
	t.Run("outside plain centi", runTestParseRoundTrip(
		outsidePlainCentiTextCases,
		roundTripOf[OutsidePlainCenti, *OutsidePlainCenti](
			"outside plain centi", seedOPC)))
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

// parseCrossCase asserts the parsers answer one text as strconv
// answers it. The natives answer exactly as strconv.ParseInt at the
// same width: the same value, the clamp included, and the same class
// of failure. The 128-bit integers answer as the 64-bit strconv parser
// of their kind, ParseUint or ParseInt, wherever its answer holds at
// any width: a success with the same value, and a syntax failure, its
// stray byte coming before any overflow, with zero; past strconv's
// range they read on where strconv stops, their answer read from the
// text itself, as runTestPastStrconv checks. Each text is run, through
// the parser and through UnmarshalText on the pointer, as the
// parseCase row built from strconv's answer, so the whole corpus
// reaches the bytes path too. The corpus holds the forms strconv takes
// and the ones it refuses, so the rows enforce strconv's grammar
// rather than transcribe a reading of it.
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

	t.Run("uint128", tc.testUint128)
	t.Run("int128", tc.testInt128)
}

// testUint128 checks ParseUint128 against strconv.ParseUint at 64 bits,
// past whose range the bound is MaxUint128.
func (tc parseCrossCase) testUint128(t *testing.T) {
	t.Helper()
	want, err := strconv.ParseUint(tc.in, 10, 64)
	if !errors.Is(err, strconv.ErrRange) {
		newParseCaseFromStrconv(tc.name, tc.in, seedU128, num.AsUint128(want),
			err).Test(t)
		return
	}
	runTestPastStrconv(t, tc, seedU128, maxUint128Text)
}

// testInt128 checks ParseInt128 against strconv.ParseInt at 64 bits,
// past whose range the bound is the one on the side strconv clamps to.
func (tc parseCrossCase) testInt128(t *testing.T) {
	t.Helper()
	want, err := strconv.ParseInt(tc.in, 10, 64)
	if !errors.Is(err, strconv.ErrRange) {
		newParseCaseFromStrconv(tc.name, tc.in, seed128, num.AsInt128(want),
			err).Test(t)
		return
	}
	bound := maxInt128Text
	if want < 0 {
		bound = minInt128Text
	}
	runTestPastStrconv(t, tc, seed128, bound)
}

// runTestPastStrconv checks the 128-bit parser of T on a text strconv
// refuses as past its 64-bit range, bound being the text of the
// parser's own bound on the side strconv clamps to. The answer is read
// from the text through math/big, from the left as strconv reads it: a
// run of digits past 128 bits fails on range with the bound, whatever
// follows it; any other run fails on syntax with zero when a byte
// follows it, on range with the bound when it passes the magnitude of
// the bound, and otherwise parses to the value it names under the
// sign. The parseCase row of that answer checks the report and
// UnmarshalText as on any other row.
func runTestPastStrconv[T num.Number[T], PT textUnmarshaler[T]](t *testing.T,
	tc parseCrossCase, seed T, bound string) {
	t.Helper()
	var zero T
	neg, run, rest := splitDigitRun(tc.in)
	u, limit := bigOf(t, run), bigOf(t, bound)
	clamp := fromBig[T](limit)
	switch {
	case u.Cmp(bigOf(t, maxUint128Text)) > 0:
		newParseCaseErr[T, PT](tc.name, tc.in, seed, clamp, num.ErrRange).Test(t)
	case rest != "":
		newParseCaseErr[T, PT](tc.name, tc.in, seed, zero, num.ErrSyntax).Test(t)
	case u.CmpAbs(limit) > 0:
		newParseCaseErr[T, PT](tc.name, tc.in, seed, clamp, num.ErrRange).Test(t)
	case neg:
		want := fromBig[T](u.Neg(u))
		newParseCase[T, PT](tc.name, tc.in, seed, want).Test(t)
	default:
		newParseCase[T, PT](tc.name, tc.in, seed, fromBig[T](u)).Test(t)
	}
}

// splitDigitRun splits a text into whether its sign is a minus, the
// run of digits after the sign and the rest from the first byte that
// is not a digit.
func splitDigitRun(s string) (neg bool, run, rest string) {
	neg = strings.HasPrefix(s, "-")
	if neg || strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	rest = strings.TrimLeft(s, "0123456789")
	return neg, s[:len(s)-len(rest)], rest
}

// bigOf returns the value of a decimal text, an optional sign and the
// digits, as math/big reads it.
func bigOf(t *testing.T, s string) *big.Int {
	t.Helper()
	v, ok := new(big.Int).SetString(s, 10)
	core.AssertMustTrue(t, ok, "decimal %q", s)
	return v
}

// fromBig returns v, which fits T, as T, one of the 128-bit integers,
// built from its two words, a negative v taken in two's complement as
// an Int128 holds it.
func fromBig[T num.Number[T]](v *big.Int) T {
	w := new(big.Int).Set(v)
	if w.Sign() < 0 {
		w.Add(w, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	lo := new(big.Int).And(w, new(big.Int).SetUint64(math.MaxUint64)).Uint64()
	hi := new(big.Int).Rsh(w, 64).Uint64()
	var zero T
	switch any(zero).(type) {
	case num.Uint128:
		return core.MustOK(core.As[num.Uint128, T](num.NewUint128(hi, lo)))
	case num.Int128:
		return core.MustOK(core.As[num.Int128, T](num.NewInt128(hi, lo)))
	default:
		// T is one of the 128-bit integers, each with an arm above.
		panic(core.NewUnreachableErrorf(0, nil, "no words for %s",
			core.TypeName[T]()))
	}
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
// the forms it takes, the bounds of every width and one past them, the
// edges of the digit run, and the forms it refuses.
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
		"170141183460469231731687303715884105727",
		"170141183460469231731687303715884105728",
		"-170141183460469231731687303715884105728",
		"-170141183460469231731687303715884105729",
		"340282366920938463463374607431768211455",
		"340282366920938463463374607431768211456",
		"0000000000000000000000000000000000000001",

		// a group of 19 digits and one past it, the largest magnitude
		// of each width's run and one past it, a group boundary
		// reached after leading zeros, and a stray byte before an
		// overflowing digit and after one, past the sign's bound and
		// past the unsigned width.
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
		"170141183460469231731687303715884105728x",
		"-170141183460469231731687303715884105729x",
		"340282366920938463463374607431768211456x",

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
		newParseFloatCaseSyntax(strings.Repeat("9", 41)+".5x"),
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
	uint128s := unmarshalTextNilCases[num.Uint128, *num.Uint128]()
	int128s := unmarshalTextNilCases[num.Int128, *num.Int128]()
	milli32s := unmarshalTextNilCases[num.Milli32, *num.Milli32]()
	milli64s := unmarshalTextNilCases[num.Milli64, *num.Milli64]()
	atto128s := unmarshalTextNilCases[num.Atto128, *num.Atto128]()
	t.Run("int32", func(t *testing.T) { core.RunTestCases(t, int32s) })
	t.Run("int64", func(t *testing.T) { core.RunTestCases(t, int64s) })
	t.Run("uint128", func(t *testing.T) { core.RunTestCases(t, uint128s) })
	t.Run("int128", func(t *testing.T) { core.RunTestCases(t, int128s) })
	t.Run("milli32", func(t *testing.T) { core.RunTestCases(t, milli32s) })
	t.Run("milli64", func(t *testing.T) { core.RunTestCases(t, milli64s) })
	t.Run("atto128", func(t *testing.T) { core.RunTestCases(t, atto128s) })
}
