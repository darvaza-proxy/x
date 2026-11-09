package tai_test

import (
	"fmt"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
	"darvaza.org/x/time/tai"
)

var _ core.TestCase = stringRangeTestCase{}

// stringRangeTestCase checks the shared Stringer contract on counts that cannot
// be constructed through label parsing, including the full internal limits.
type stringRangeTestCase struct {
	value      fmt.Stringer
	name, want string
}

func newStringRangeTestCase(name string, value fmt.Stringer, want string) stringRangeTestCase {
	return stringRangeTestCase{name: name, value: value, want: want}
}

func (tc stringRangeTestCase) Name() string { return tc.name }

func (tc stringRangeTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, tc.value.String(), "diagnostic")
	core.AssertEqual(t, tc.want, fmt.Sprint(tc.value), "fmt.Sprint")
}

func taiStringRangeTestCases() []stringRangeTestCase {
	return core.S(
		newStringRangeTestCase("TAI negative second", tai.NewTAIForTest(num.AsInt128(-1)), "TAI(-1s)"),
		newStringRangeTestCase("TAI seconds exceed uint64", tai.NewTAIForTest(num.NewInt128(1, 0)),
			"TAI(18446744073709551616s)"),
		newStringRangeTestCase("TAI minimum count", tai.NewTAIForTest(num.MinInt128),
			"TAI(-170141183460469231731687303715884105728s)"),
		newStringRangeTestCase("TAI maximum count", tai.NewTAIForTest(num.MaxInt128),
			"TAI(170141183460469231731687303715884105727s)"),
	)
}

func tainStringRangeTestCases() []stringRangeTestCase {
	tooLarge := num.NewInt128(1, 0).Mul(num.AsInt128(tai.AttosPerSec / tai.AttosPerNano))
	return core.S(
		newStringRangeTestCase("TAIN negative nanosecond", tai.NewTAINForTest(num.AsInt128(-1)), "TAIN(-1ns)"),
		newStringRangeTestCase("TAIN seconds exceed uint64", tai.NewTAINForTest(tooLarge),
			"TAIN(18446744073709551616000000000ns)"),
		newStringRangeTestCase("TAIN minimum count", tai.NewTAINForTest(num.MinInt128),
			"TAIN(-170141183460469231731687303715884105728ns)"),
		newStringRangeTestCase("TAIN maximum count", tai.NewTAINForTest(num.MaxInt128),
			"TAIN(170141183460469231731687303715884105727ns)"),
	)
}

func taiaStringRangeTestCases() []stringRangeTestCase {
	tooLarge := num.NewInt128(1, 0).Mul(num.AsInt128(tai.AttosPerSec))
	return core.S(
		newStringRangeTestCase("TAIA negative attosecond", tai.NewTAIAForTest(num.AsInt128(-1)), "TAIA(-1as)"),
		newStringRangeTestCase("TAIA seconds exceed uint64", tai.NewTAIAForTest(tooLarge),
			"TAIA(18446744073709551616000000000000000000as)"),
		newStringRangeTestCase("TAIA minimum count", tai.NewTAIAForTest(num.MinInt128),
			"TAIA(-170141183460469231731687303715884105728as)"),
		newStringRangeTestCase("TAIA maximum count", tai.NewTAIAForTest(num.MaxInt128),
			"TAIA(170141183460469231731687303715884105727as)"),
	)
}

func TestStringOutsideLabelRange(t *testing.T) {
	cases := append(taiStringRangeTestCases(), tainStringRangeTestCases()...)
	core.RunTestCases(t, append(cases, taiaStringRangeTestCases()...))
}
