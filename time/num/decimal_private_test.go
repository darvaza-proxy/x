package num

// cspell:words zepto

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = scaleFormatCase{}

// unitScale is a scaler at 10^0 over an Int64 backing, so the whole
// count of a unitDecimal is the backing itself, its minimum included,
// and its fraction has no digits. No Decimal of the package has a
// scale of one.
type unitScale struct{}

func (unitScale) Scale() Int64 {
	return 1
}

func (unitScale) name() string {
	return "Unit64"
}

func (unitScale) asInt128(v Int64) Int128 {
	return v.asInt128()
}

// unitDecimal is a Decimal at a scale of one.
type unitDecimal = Decimal[Int64, unitScale]

// asUnit takes x as a count of whole units.
func asUnit(x int64) unitDecimal {
	return unitDecimal{Int64(x)}
}

// zeptoScale is a scaler at 10^21 over an Int128 backing, a scale and
// a fraction past the int64 range. No Decimal of the package has a
// scale past 10^18.
type zeptoScale struct{}

func (zeptoScale) Scale() Int128 {
	return Int128{hi: 0x36, lo: 0x35c9adc5dea00000} // 10^21
}

func (zeptoScale) name() string {
	return "Zepto128"
}

func (zeptoScale) asInt128(v Int128) Int128 {
	return v
}

// zeptoDecimal is a Decimal at a scale of 10^21.
type zeptoDecimal = Decimal[Int128, zeptoScale]

// newZepto builds a zeptoDecimal from a whole-unit count and a
// zepto-unit fraction.
func newZepto(whole, zepto int64) zeptoDecimal {
	return newDecimal[Int128, zeptoScale](AsInt128(whole), AsInt128(zepto))
}

// newZeptoMilli builds a zeptoDecimal from a count of thousandths, for
// the rows whose fraction passes an int64 of zepto-units.
func newZeptoMilli(milli int64) zeptoDecimal {
	return zeptoDecimal{AsInt128(milli).Mul(AsInt128(1e18))}
}

// scaleFormatCase pins the text one format string produces for one
// Decimal at a scale of the white-box scalers, through fmt.
type scaleFormatCase struct {
	in     any
	name   string
	format string
	want   string
}

func newScaleFormatCase(name string, in any, format,
	want string) scaleFormatCase {
	return scaleFormatCase{
		in:     in,
		name:   name,
		format: format,
		want:   want,
	}
}

func (tc scaleFormatCase) Name() string { return tc.name }

func (tc scaleFormatCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, fmt.Sprintf(tc.format, tc.in), "text")
}

func TestUnitDecimalFormat(t *testing.T) {
	core.RunTestCases(t, []scaleFormatCase{
		newScaleFormatCase("v", asUnit(-5), "%v", "-5"),
		newScaleFormatCase("v min", asUnit(math.MinInt64), "%v",
			"-9223372036854775808"),
		newScaleFormatCase("v max", asUnit(math.MaxInt64), "%v",
			"9223372036854775807"),
		// a precision zero-fills a fraction of no digits.
		newScaleFormatCase("f", asUnit(5), "%.2f", "5.00"),
		newScaleFormatCase("f min", asUnit(math.MinInt64), "%f",
			"-9223372036854775808.000000"),
	})
}

func TestZeptoDecimalFormat(t *testing.T) {
	core.RunTestCases(t, []scaleFormatCase{
		newScaleFormatCase("v small fraction", newZepto(1, 5), "%v",
			"1.000000000000000000005"),
		newScaleFormatCase("v negative fraction", newZepto(0, -5), "%v",
			"-0.000000000000000000005"),
		// the fraction passes an int64 from here on.
		newScaleFormatCase("v one and a half", newZeptoMilli(1500), "%v",
			"1.500000000000000000000"),
		newScaleFormatCase("v min", zeptoDecimal{MinInt128}, "%v",
			"-170141183460469231.731687303715884105728"),
		// rounding to one digit or none divides by 10^20 or 10^21, past
		// an int64.
		newScaleFormatCase("f round", newZeptoMilli(1950), "%.1f", "2.0"),
		newScaleFormatCase("f half up", newZeptoMilli(2500), "%.0f", "3"),
		newScaleFormatCase("f past resolution", newZepto(1, 5), "%.23f",
			"1.00000000000000000000500"),
		newScaleFormatCase("go string", newZepto(1, 5), "%#v",
			"num.NewZepto128(1, 5)"),
		// a fraction past int64 cannot be an argument of the parts
		// constructor, so the backing's words form takes over.
		newScaleFormatCase("go string min", zeptoDecimal{MinInt128}, "%#v",
			"num.AsZepto128(num.NewInt128(0x8000000000000000, 0x0))"),
	})
}
