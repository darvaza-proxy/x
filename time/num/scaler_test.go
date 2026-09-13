package num_test

// cspell:words centi

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// centi64Scale is a scaler defined outside the package, two fraction
// digits over an Int64 backing. The rows in this file hold only while
// [num.DecimalScaler] asks nothing an outside package cannot supply.
type centi64Scale struct{}

// Scale returns the sub-units in one whole unit, a hundred.
func (centi64Scale) Scale() num.Int64 { return 100 }

// Name returns the type as written from another package, which
// GoString derives the constructor names from.
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

// TestCenti64 runs the shared Decimal suite over the outside
// instantiation, arithmetic, sign rule, division by zero and the
// wrapping quotient included.
func TestCenti64(t *testing.T) {
	runDecimalTests(t, decimalType[Centi64]{
		mk:    NewCenti64,
		as:    func(centi int64) Centi64 { return AsCenti64(num.AsInt64(centi)) },
		scale: 100,
		// 1844674407370956e4 is the first multiple of 10^4 past 2^64.
		big:       1844674407370956,
		wrapWhole: 83,
		wrapFrac:  84,
	})
}

func centi64TextCases() []textCase {
	return []textCase{
		newTextCase("zero", NewCenti64(0, 0), "0.00"),
		newTextCase("one and a half", NewCenti64(1, 50), "1.50"),
		newTextCase("negative fraction", NewCenti64(0, -5), "-0.05"),
		newTextCase("min", AsCenti64(math.MinInt64), "-92233720368547758.08"),
		newTextCase("max", AsCenti64(math.MaxInt64), "92233720368547758.07"),
	}
}

func centi64FormatCases() []formatCase {
	return []formatCase{
		newFormatCase("v", NewCenti64(1, 50), "%v", "1.50"),
		newFormatCase("f", NewCenti64(1, 50), "%f", "1.500000"),
		newFormatCase("f round", NewCenti64(1, 55), "%.1f", "1.6"),
		newFormatCase("f half up", NewCenti64(2, 50), "%.0f", "3"),
		newFormatCase("width", NewCenti64(1, 50), "%+8.3f", "  +1.500"),
		newFormatCase("go syntax", NewCenti64(1, 50), "%#v",
			"num_test.NewCenti64(1, 50)"),
		newFormatCase("bad verb", NewCenti64(1, 50), "%d",
			"%!d(num_test.Centi64=1.50)"),
		newFormatCase("bad verb signed", NewCenti64(1, 50), "%+x",
			"%!x(num_test.Centi64=+1.50)"),
	}
}

func centi64GoStringCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("half", NewCenti64(1, 50), "num_test.NewCenti64(1, 50)"),
		newGoStringCase("negative", NewCenti64(-1, -5),
			"num_test.NewCenti64(-1, -5)"),
		newGoStringCase("negative fraction", NewCenti64(0, -5),
			"num_test.NewCenti64(0, -5)"),
		newGoStringCase("as", AsCenti64(150), "num_test.NewCenti64(1, 50)"),
		newGoStringCase("min", AsCenti64(math.MinInt64),
			"num_test.NewCenti64(-92233720368547758, -8)"),
	}
}

func centi64JSONCases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("half", NewCenti64(1, 50), "1.50"),
		newJSONNumberCase("count below ten to the 15", AsCenti64(999999999999999),
			"9999999999999.99"),
		newJSONStringCase("count at ten to the 15", AsCenti64(1e15),
			"10000000000000.00"),
		newJSONStringCase("min", AsCenti64(math.MinInt64),
			"-92233720368547758.08"),
	}
}

// TestCenti64Text checks the text, format, %#v, JSON and conversion
// surfaces of the outside instantiation, which reach the backing and
// the scaler through the exported interface alone.
func TestCenti64Text(t *testing.T) {
	t.Run("text", runTestCenti64Text)
	t.Run("format", runTestCenti64Format)
	t.Run("go string", runTestCenti64GoString)
	t.Run("json", runTestCenti64JSON)
	t.Run("convert", runTestCenti64Convert)
}

func runTestCenti64Text(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, centi64TextCases())
}

func runTestCenti64Format(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, centi64FormatCases())
}

func runTestCenti64GoString(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, centi64GoStringCases())
}

func runTestCenti64JSON(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, centi64JSONCases())
}

// runTestCenti64Convert checks the value-preserving conversions out of
// the outside instantiation and its count accessors. The conversion
// matrix cannot hold it as a target, since [num.Number] names only the
// seven types of the package, so the rows stand alone.
func runTestCenti64Convert(t *testing.T) {
	t.Helper()
	v := NewCenti64(1, 55)

	m32, ok := v.Milli32()
	core.AssertTrue(t, ok, "milli32 ok")
	core.AssertEqual(t, num.NewMilli32(1, 550), m32, "milli32")

	m64, ok := v.Milli64()
	core.AssertTrue(t, ok, "milli64 ok")
	core.AssertEqual(t, num.NewMilli64(1, 550), m64, "milli64")

	a, ok := v.Atto128()
	core.AssertTrue(t, ok, "atto128 ok")
	core.AssertEqual(t, num.NewAtto128(1, 550e15), a, "atto128")

	// the fraction drops towards zero.
	i, ok := v.Int64()
	core.AssertTrue(t, ok, "int64 ok")
	core.AssertEqual(t, num.AsInt64(1), i, "int64")

	c, ok := v.AsInt64()
	core.AssertTrue(t, ok, "count ok")
	core.AssertEqual(t, num.AsInt64(155), c, "count")
}
