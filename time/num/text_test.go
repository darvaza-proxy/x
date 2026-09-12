package num_test

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var _ core.TestCase = textCase{}

// textCase pins the String form of a value, which is the text %v
// prints, so the same text must come back through fmt.
type textCase struct {
	in   fmt.Stringer
	want string
	name string
}

func newTextCase(name string, in fmt.Stringer, want string) textCase {
	return textCase{name: name, in: in, want: want}
}

func (tc textCase) Name() string { return tc.name }

func (tc textCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.in.String(), "String")
	core.AssertEqual(t, tc.want, fmt.Sprint(tc.in), "fmt")
}

func textIntCases() []textCase {
	return []textCase{
		newTextCase("int32 zero", num.AsInt32(0), "0"),
		newTextCase("int32 negative", num.AsInt32(-42), "-42"),
		newTextCase("int32 min", num.AsInt32(math.MinInt32), "-2147483648"),
		newTextCase("int64 positive", num.AsInt64(42), "42"),
		newTextCase("int64 min", num.AsInt64(math.MinInt64),
			"-9223372036854775808"),
	}
}

func textUint128Cases() []textCase {
	return []textCase{
		newTextCase("uint128 zero", num.ZeroUint128, "0"),
		newTextCase("uint128 low word", u(42), "42"),
		newTextCase("uint128 two to the 64", num.NewUint128(1, 0),
			"18446744073709551616"),
		newTextCase("uint128 max", num.MaxUint128,
			"340282366920938463463374607431768211455"),
	}
}

func textInt128Cases() []textCase {
	return []textCase{
		newTextCase("int128 zero", num.ZeroInt128, "0"),
		newTextCase("int128 negative", num.AsInt128(-42), "-42"),
		newTextCase("int128 max", num.MaxInt128,
			"170141183460469231731687303715884105727"),
		newTextCase("int128 min", num.MinInt128,
			"-170141183460469231731687303715884105728"),
	}
}

func textDecimalCases() []textCase {
	return []textCase{
		newTextCase("milli32 zero", num.NewMilli32(0, 0), "0.000"),
		newTextCase("milli32", num.NewMilli32(1, 500), "1.500"),
		newTextCase("milli32 negative fraction", num.NewMilli32(0, -5),
			"-0.005"),
		newTextCase("milli32 min", num.AsMilli32(math.MinInt32),
			"-2147483.648"),
		newTextCase("milli64", num.NewMilli64(86400, 5), "86400.005"),
		newTextCase("milli64 min", num.AsMilli64(math.MinInt64),
			"-9223372036854775.808"),
		newTextCase("atto128", num.NewAtto128(1, 500e15),
			"1.500000000000000000"),
		newTextCase("atto128 one atto", num.NewAtto128(0, 1),
			"0.000000000000000001"),
		newTextCase("atto128 min", num.AsAtto128(num.MinInt128),
			"-170141183460469231731.687303715884105728"),
	}
}

func TestText(t *testing.T) {
	t.Run("int", runTestTextInt)
	t.Run("uint128", runTestTextUint128)
	t.Run("int128", runTestTextInt128)
	t.Run("decimal", runTestTextDecimal)
}

func runTestTextInt(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, textIntCases())
}

func runTestTextUint128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, textUint128Cases())
}

func runTestTextInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, textInt128Cases())
}

func runTestTextDecimal(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, textDecimalCases())
}
