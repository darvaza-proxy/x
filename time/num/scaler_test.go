package num_test

import (
	"math"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// centi64Scale is a scaler defined outside the package, two fraction
// digits over an Int64 backing.
type centi64Scale struct{}

// Scale returns the sub-units in one whole unit, a hundred.
func (centi64Scale) Scale() num.Int64 { return 100 }

// Name returns the type as written from another package, which
// GoString derives the constructor names from and ParseDecimal the
// parser name its report carries.
func (centi64Scale) Name() string { return "num_test.Centi64" }

var _ num.DecimalScaler[num.Int64] = centi64Scale{}

// Centi64 is a signed fixed-point number with two fraction digits,
// backed by an Int64 counting centi-units (10^-2): an instantiation of
// [num.Decimal] outside its package.
type Centi64 = num.Decimal[num.Int64, centi64Scale]

var _ num.Number[Centi64] = Centi64{}

// NewCenti64 builds a Centi64 from a whole-unit count and a centi-unit
// fraction, under the sign rule of the package's own parts
// constructors; it is the constructor GoString names.
func NewCenti64(whole, centi int64) Centi64 {
	return num.NewDecimal[num.Int64, centi64Scale](num.AsInt64(whole),
		num.AsInt64(centi))
}

// AsCenti64 takes an Int64 as a count of centi-units, so AsCenti64(150)
// is 1.5.
func AsCenti64(centi num.Int64) Centi64 {
	return num.AsDecimal[num.Int64, centi64Scale](centi)
}

// ParseCenti64 reads a Centi64 from its decimal text as
// [num.ParseDecimal] reads the instantiation; it is the parser a
// report names.
func ParseCenti64(s string) (Centi64, error) {
	return num.ParseDecimal[num.Int64, centi64Scale](s)
}

// TestCenti64 shows the outside instantiation built through its own
// constructors, printed, parsed, read back as a count, and built from
// whole units through the Decimal factory.
func TestCenti64(t *testing.T) {
	core.RunTestCases(t, core.S[core.TestCase](
		newTextCase("text", NewCenti64(1, 50), "1.50"),
		newParseCase("parse", "1.5", seedC64, NewCenti64(1, 50)),
		newParseCase("parse truncated", "-1.559", seedC64,
			NewCenti64(-1, -55)),
		newParseCaseErr("parse past max", "92233720368547758.08", seedC64,
			AsCenti64(math.MaxInt64), num.ErrRange),
		newGoStringCase("go string", NewCenti64(-1, -5),
			"num_test.NewCenti64(-1, -5)"),
		newGoStringCase("go string of as", AsCenti64(150),
			"num_test.NewCenti64(1, 50)"),
		newCountCase("count", NewCenti64(1, 55), num.AsInt64(155)),
		newDecimalFromInt128Case[num.Int64, centi64Scale]("from int128",
			num.AsInt128(-5), NewCenti64(-5, 0)),
		newDecimalFromInt128CaseRange[num.Int64, centi64Scale](
			"from int128 past whole", num.AsInt128(math.MaxInt64/100+1),
			AsCenti64(math.MaxInt64)),
	))
}

// unit64Scale is the scaler of Unit64, unit resolution over an Int64
// backing: 10^0, the smallest power of ten, so the whole count spans
// the backing, its minimum included.
type unit64Scale struct{}

// Scale returns the sub-units in one whole unit, one.
func (unit64Scale) Scale() num.Int64 { return 1 }

// Name returns the type as written from another package, which
// GoString derives the constructor names from.
func (unit64Scale) Name() string { return "num_test.Unit64" }

var _ num.DecimalScaler[num.Int64] = unit64Scale{}

// Unit64 is a signed fixed-point number with no fraction digits, backed
// by an Int64 counting whole units: an instantiation of [num.Decimal]
// at the smallest scale.
type Unit64 = num.Decimal[num.Int64, unit64Scale]

var _ num.Number[Unit64] = Unit64{}

// NewUnit64 builds a Unit64 from a whole-unit count and a sub-unit
// count, the same unit at this resolution, under the sign rule of the
// package's own parts constructors; it is the constructor GoString
// names.
func NewUnit64(whole, frac int64) Unit64 {
	return num.NewDecimal[num.Int64, unit64Scale](num.AsInt64(whole),
		num.AsInt64(frac))
}

// AsUnit64 takes an Int64 as a count of whole units.
func AsUnit64(units num.Int64) Unit64 {
	return num.AsDecimal[num.Int64, unit64Scale](units)
}

// ParseUnit64 reads a Unit64 from its decimal text as
// [num.ParseDecimal] reads the instantiation; it is the parser a
// report names.
func ParseUnit64(s string) (Unit64, error) {
	return num.ParseDecimal[num.Int64, unit64Scale](s)
}

func unit64TextCases() []textCase {
	return []textCase{
		newTextCase("zero", NewUnit64(0, 0), "0"),
		newTextCase("negative", NewUnit64(-5, 0), "-5"),
		// the whole count is the backing itself, so it reaches the
		// minimum, whose magnitude no Int64 holds.
		newTextCase("min", AsUnit64(math.MinInt64), "-9223372036854775808"),
		newTextCase("max", AsUnit64(math.MaxInt64), "9223372036854775807"),
	}
}

func unit64FormatCases() []formatCase {
	return []formatCase{
		newFormatCase("v min", AsUnit64(math.MinInt64), "%v",
			"-9223372036854775808"),
		// a precision zero-fills a fraction of no digits.
		newFormatCase("f", NewUnit64(5, 0), "%.2f", "5.00"),
		newFormatCase("f min", AsUnit64(math.MinInt64), "%f",
			"-9223372036854775808.000000"),
	}
}

func unit64GoStringCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("five", NewUnit64(5, 0), "num_test.NewUnit64(5, 0)"),
		newGoStringCase("min", AsUnit64(math.MinInt64),
			"num_test.NewUnit64(-9223372036854775808, 0)"),
	}
}

func unit64JSONCases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("five", NewUnit64(5, 0), "5"),
		newJSONStringCase("min", AsUnit64(math.MinInt64),
			"-9223372036854775808"),
	}
}

func unit64ParseCases() []parseCase[Unit64, *Unit64] {
	return []parseCase[Unit64, *Unit64]{
		newParseCase("zero", "0", seedU64, NewUnit64(0, 0)),
		newParseCase("whole", "42", seedU64, NewUnit64(42, 0)),
		newParseCase("negative", "-5", seedU64, NewUnit64(-5, 0)),
		newParseCase("plus", "+5", seedU64, NewUnit64(5, 0)),
		newParseCase("point on the right", "5.", seedU64, NewUnit64(5, 0)),
		// at unit resolution every fraction digit is below it, so the
		// fraction drops towards zero.
		newParseCase("truncated", "1.9", seedU64, NewUnit64(1, 0)),
		newParseCase("negative truncated", "-1.9", seedU64,
			NewUnit64(-1, 0)),
		newParseCase("point on the left", ".5", seedU64, NewUnit64(0, 0)),
		newParseCase("negative truncated to zero", "-0.5", seedU64,
			NewUnit64(0, 0)),
		newParseCase("max", "9223372036854775807", seedU64,
			AsUnit64(math.MaxInt64)),
		newParseCase("min", "-9223372036854775808", seedU64,
			AsUnit64(math.MinInt64)),
		newParseCase("max truncated", "9223372036854775807.9", seedU64,
			AsUnit64(math.MaxInt64)),
		newParseCaseErr("past max", "9223372036854775808", seedU64,
			AsUnit64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("below min", "-9223372036854775809", seedU64,
			AsUnit64(math.MinInt64), num.ErrRange),
		newParseCaseErr("far past", strings.Repeat("9", 41), seedU64,
			AsUnit64(math.MaxInt64), num.ErrRange),
		newParseCaseErr("empty", "", seedU64, NewUnit64(0, 0), num.ErrSyntax),
		newParseCaseErr("point only", ".", seedU64, NewUnit64(0, 0),
			num.ErrSyntax),
		newParseCaseErr("exponent", "1e3", seedU64, NewUnit64(0, 0),
			num.ErrSyntax),
		// the dropped digits are still read for their syntax.
		newParseCaseErr("letter below the resolution", "1.5a", seedU64,
			NewUnit64(0, 0), num.ErrSyntax),
	}
}

// TestUnit64Text checks the text, format, %#v and JSON surfaces and the
// parser at the smallest scale, where the whole count spans the
// backing.
func TestUnit64Text(t *testing.T) {
	t.Run("text", runTestUnit64Text)
	t.Run("format", runTestUnit64Format)
	t.Run("go string", runTestUnit64GoString)
	t.Run("json", runTestUnit64JSON)
	t.Run("parse", runTestUnit64Parse)
}

func runTestUnit64Text(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, unit64TextCases())
}

func runTestUnit64Format(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, unit64FormatCases())
}

func runTestUnit64GoString(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, unit64GoStringCases())
}

func runTestUnit64JSON(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, unit64JSONCases())
}

func runTestUnit64Parse(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, unit64ParseCases())
}

// zeptoScale returns the number of zepto-units in one whole unit,
// 10^21, a count past the int64 range. It is built on each call rather
// than held in a package variable: the compiler sees no dependency of
// a package variable built through the scaler, such as a seed, on one
// Scale reads, since Scale is reached through a type argument, so the
// seed could be initialised first and read a zero scale.
func zeptoScale() num.Int128 {
	return core.MustOK(num.Pow10(21).Int128())
}

// zepto128Scale is the scaler of Zepto128, zepto resolution over an
// Int128 backing: 10^21, a power of ten no int64 holds, so the scale
// and the fraction reach past 64 bits.
type zepto128Scale struct{}

// Scale returns the sub-units in one whole unit, 10^21.
func (zepto128Scale) Scale() num.Int128 { return zeptoScale() }

// Name returns the type as written from another package, which
// GoString derives the constructor names from.
func (zepto128Scale) Name() string { return "num_test.Zepto128" }

var _ num.DecimalScaler[num.Int128] = zepto128Scale{}

// Zepto128 is a signed fixed-point number with 21 fraction digits,
// backed by an Int128 counting zepto-units (10^-21): an instantiation
// of [num.Decimal] at a scale past 10^18.
type Zepto128 = num.Decimal[num.Int128, zepto128Scale]

var _ num.Number[Zepto128] = Zepto128{}

// NewZepto128 builds a Zepto128 from a whole-unit count and a
// zepto-unit fraction, under the sign rule of the package's own parts
// constructors; it is the constructor GoString names.
func NewZepto128(whole, zepto int64) Zepto128 {
	return num.NewDecimal[num.Int128, zepto128Scale](num.AsInt128(whole),
		num.AsInt128(zepto))
}

// AsZepto128 takes an Int128 as a count of zepto-units.
func AsZepto128(zepto num.Int128) Zepto128 {
	return num.AsDecimal[num.Int128, zepto128Scale](zepto)
}

// ParseZepto128 reads a Zepto128 from its decimal text as
// [num.ParseDecimal] reads the instantiation; it is the parser a
// report names.
func ParseZepto128(s string) (Zepto128, error) {
	return num.ParseDecimal[num.Int128, zepto128Scale](s)
}

// newZepto128Milli returns a Zepto128 from a count of thousandths, for
// the rows whose fraction passes an int64 of zepto-units.
func newZepto128Milli(milli int64) Zepto128 {
	return AsZepto128(num.AsInt128(milli).Mul(num.AsInt128(1e18)))
}

func zepto128TextCases() []textCase {
	return []textCase{
		newTextCase("zero", NewZepto128(0, 0), "0.000000000000000000000"),
		newTextCase("small fraction", NewZepto128(1, 5),
			"1.000000000000000000005"),
		newTextCase("negative fraction", NewZepto128(0, -5),
			"-0.000000000000000000005"),
		// the fraction passes an int64 from here on.
		newTextCase("one and a half", newZepto128Milli(1500),
			"1.500000000000000000000"),
		newTextCase("min", AsZepto128(num.MinInt128),
			"-170141183460469231.731687303715884105728"),
		newTextCase("max", AsZepto128(num.MaxInt128),
			"170141183460469231.731687303715884105727"),
	}
}

func zepto128FormatCases() []formatCase {
	return []formatCase{
		newFormatCase("v", NewZepto128(1, 5), "%v", "1.000000000000000000005"),
		newFormatCase("f", newZepto128Milli(1500), "%f", "1.500000"),
		// rounding to one digit or none divides by 10^20 or 10^21, past
		// an int64.
		newFormatCase("f round", newZepto128Milli(1950), "%.1f", "2.0"),
		newFormatCase("f below half", NewZepto128(2, 5), "%.0f", "2"),
		newFormatCase("f half up", newZepto128Milli(2500), "%.0f", "3"),
		newFormatCase("f past resolution", NewZepto128(1, 5), "%.23f",
			"1.00000000000000000000500"),
	}
}

func zepto128GoStringCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("small fraction", NewZepto128(1, 5),
			"num_test.NewZepto128(1, 5)"),
		newGoStringCase("negative", NewZepto128(-1, -5),
			"num_test.NewZepto128(-1, -5)"),
		newGoStringCase("grouped", NewZepto128(0, 1234567),
			"num_test.NewZepto128(0, 1_234_567)"),
		// a fraction past int64 cannot be an argument of the parts
		// constructor, so the backing's words form takes over.
		newGoStringCase("min", AsZepto128(num.MinInt128),
			"num_test.AsZepto128(num.NewInt128(0x8000000000000000, 0x0))"),
		newGoStringCase("max", AsZepto128(num.MaxInt128),
			"num_test.AsZepto128(num.NewInt128(0x7fffffffffffffff, 0xffffffffffffffff))"),
	}
}

func zepto128JSONCases() []jsonCase {
	return []jsonCase{
		// the scale is past 10^15, so every value is a string.
		newJSONStringCase("zero", NewZepto128(0, 0), "0.000000000000000000000"),
		newJSONStringCase("small fraction", NewZepto128(2, 5),
			"2.000000000000000000005"),
		newJSONStringCase("min", AsZepto128(num.MinInt128),
			"-170141183460469231.731687303715884105728"),
	}
}

func zepto128ParseCases() []parseCase[Zepto128, *Zepto128] {
	return []parseCase[Zepto128, *Zepto128]{
		newParseCase("zero", "0", seedZ128, NewZepto128(0, 0)),
		newParseCase("whole", "42", seedZ128, NewZepto128(42, 0)),
		newParseCase("one zepto", "0.000000000000000000001", seedZ128,
			NewZepto128(0, 1)),
		newParseCase("small fraction", "1.000000000000000000005", seedZ128,
			NewZepto128(1, 5)),
		newParseCase("negative fraction", "-0.000000000000000000005",
			seedZ128, NewZepto128(0, -5)),
		// the fraction passes a uint64 of zepto-units from here on.
		newParseCase("half", "1.5", seedZ128, newZepto128Milli(1500)),
		newParseCase("every digit", "0.999999999999999999999", seedZ128,
			AsZepto128(zeptoScale().Sub(num.AsInt128(1)))),
		newParseCase("truncated", "1.0000000000000000000059", seedZ128,
			NewZepto128(1, 5)),
		newParseCase("max", "170141183460469231.731687303715884105727",
			seedZ128, AsZepto128(num.MaxInt128)),
		newParseCase("min", "-170141183460469231.731687303715884105728",
			seedZ128, AsZepto128(num.MinInt128)),
		newParseCaseErr("past max",
			"170141183460469231.731687303715884105728", seedZ128,
			AsZepto128(num.MaxInt128), num.ErrRange),
		newParseCaseErr("below min",
			"-170141183460469231.731687303715884105729", seedZ128,
			AsZepto128(num.MinInt128), num.ErrRange),
		// 20 digits the whole count holds, and the zepto scale then
		// takes past 128 bits.
		newParseCaseErr("count past the width", strings.Repeat("9", 20),
			seedZ128, AsZepto128(num.MaxInt128), num.ErrRange),
		newParseCaseErr("empty", "", seedZ128, NewZepto128(0, 0),
			num.ErrSyntax),
		newParseCaseErr("letter in the fraction", "1.5a", seedZ128,
			NewZepto128(0, 0), num.ErrSyntax),
	}
}

// TestZepto128Text checks the text, format, %#v and JSON surfaces and
// the parser at a scale past 10^18, whose fraction reaches past 64
// bits.
func TestZepto128Text(t *testing.T) {
	t.Run("text", runTestZepto128Text)
	t.Run("format", runTestZepto128Format)
	t.Run("go string", runTestZepto128GoString)
	t.Run("json", runTestZepto128JSON)
	t.Run("parse", runTestZepto128Parse)
}

func runTestZepto128Text(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, zepto128TextCases())
}

func runTestZepto128Format(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, zepto128FormatCases())
}

func runTestZepto128GoString(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, zepto128GoStringCases())
}

func runTestZepto128JSON(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, zepto128JSONCases())
}

func runTestZepto128Parse(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, zepto128ParseCases())
}
