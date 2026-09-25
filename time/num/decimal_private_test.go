package num

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = unitFormatCase{}

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

func (unitScale) asInt64(v Int64) (int64, bool) {
	return v.sys(), true
}

func (unitScale) asInt128(v Int64) Int128 {
	return AsInt128(v.sys())
}

// unitDecimal is a Decimal at a scale of one.
type unitDecimal = Decimal[Int64, unitScale]

// unitFormatCase pins the text one format string produces for one
// unitDecimal, through fmt.
type unitFormatCase struct {
	name   string
	format string
	want   string
	in     unitDecimal
}

func newUnitFormatCase(name string, in int64, format,
	want string) unitFormatCase {
	return unitFormatCase{
		name:   name,
		format: format,
		want:   want,
		in:     unitDecimal{Int64(in)},
	}
}

func (tc unitFormatCase) Name() string { return tc.name }

func (tc unitFormatCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, fmt.Sprintf(tc.format, tc.in), "text")
}

func TestUnitDecimalFormat(t *testing.T) {
	core.RunTestCases(t, []unitFormatCase{
		newUnitFormatCase("v", -5, "%v", "-5"),
		newUnitFormatCase("v min", math.MinInt64, "%v",
			"-9223372036854775808"),
		newUnitFormatCase("v max", math.MaxInt64, "%v",
			"9223372036854775807"),
		// a precision zero-fills a fraction of no digits.
		newUnitFormatCase("f", 5, "%.2f", "5.00"),
		newUnitFormatCase("f min", math.MinInt64, "%f",
			"-9223372036854775808.000000"),
	})
}
