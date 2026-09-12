package num_test

// cspell:words femto

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var _ core.TestCase = goStringCase{}

// goStringCase pins the %#v form of a value. Each row builds its value
// with the constructor the output names, so the row shows the round
// trip, and the same text must come back through fmt.
type goStringCase struct {
	in   fmt.GoStringer
	want string
	name string
}

func newGoStringCase(name string, in fmt.GoStringer,
	want string) goStringCase {
	return goStringCase{name: name, in: in, want: want}
}

func (tc goStringCase) Name() string { return tc.name }

func (tc goStringCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.in.GoString(), "GoString")
	core.AssertEqual(t, tc.want, fmt.Sprintf("%#v", tc.in), "fmt")
}

func goStringIntCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("int32 zero", num.AsInt32(0), "num.AsInt32(0)"),
		newGoStringCase("int32 negative", num.AsInt32(-5), "num.AsInt32(-5)"),
		newGoStringCase("int32 min", num.AsInt32(math.MinInt32),
			"num.AsInt32(-2147483648)"),
		newGoStringCase("int64 positive", num.AsInt64(42), "num.AsInt64(42)"),
		newGoStringCase("int64 min", num.AsInt64(math.MinInt64),
			"num.AsInt64(-9223372036854775808)"),
	}
}

func goStringUint128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("uint128 zero", num.ZeroUint128, "num.AsUint128(0)"),
		newGoStringCase("uint128 low word", num.AsUint128(5),
			"num.AsUint128(5)"),
		newGoStringCase("uint128 max word", num.AsUint128(maxWord),
			"num.AsUint128(18446744073709551615)"),
		// the high word set switches to the words form, in hex.
		newGoStringCase("uint128 high word", num.NewUint128(1, 0),
			"num.NewUint128(0x1, 0x0)"),
		newGoStringCase("uint128 max", num.MaxUint128,
			"num.NewUint128(0xffffffffffffffff, 0xffffffffffffffff)"),
	}
}

func goStringInt128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("int128 zero", num.ZeroInt128, "num.AsInt128(0)"),
		newGoStringCase("int128 positive", num.AsInt128(42),
			"num.AsInt128(42)"),
		newGoStringCase("int128 negative", num.AsInt128(-42),
			"num.AsInt128(-42)"),
		// the int64 bounds still fit: the words are all ones for -1 and
		// the sign extension of the top bit for the minimum.
		newGoStringCase("int128 minus one", num.AsInt128(-1),
			"num.AsInt128(-1)"),
		newGoStringCase("int128 max int64", num.AsInt128(math.MaxInt64),
			"num.AsInt128(9223372036854775807)"),
		newGoStringCase("int128 min int64", num.AsInt128(math.MinInt64),
			"num.AsInt128(-9223372036854775808)"),
		// past int64 the words form takes over: 2^63 has the low sign bit
		// set with a zero high word, -2^64 the reverse.
		newGoStringCase("int128 two to the 63", num.NewInt128(0, signBit),
			"num.NewInt128(0x0, 0x8000000000000000)"),
		newGoStringCase("int128 minus two to the 64",
			num.NewInt128(maxWord, 0), "num.NewInt128(0xffffffffffffffff, 0x0)"),
		newGoStringCase("int128 min", num.MinInt128,
			"num.NewInt128(0x8000000000000000, 0x0)"),
		newGoStringCase("int128 max", num.MaxInt128,
			"num.NewInt128(0x7fffffffffffffff, 0xffffffffffffffff)"),
	}
}

func goStringMilliCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("milli32 zero", num.NewMilli32(0, 0),
			"num.NewMilli32(0, 0)"),
		newGoStringCase("milli32 half", num.NewMilli32(1, 500),
			"num.NewMilli32(1, 500)"),
		// both parts carry the sign, so the call rebuilds the value under
		// either sign rule of the constructor.
		newGoStringCase("milli32 negative", num.NewMilli32(-1, -500),
			"num.NewMilli32(-1, -500)"),
		newGoStringCase("milli32 negative fraction", num.NewMilli32(0, -500),
			"num.NewMilli32(0, -500)"),
		newGoStringCase("milli32 as", num.AsMilli32(1500),
			"num.NewMilli32(1, 500)"),
		newGoStringCase("milli32 min", num.AsMilli32(math.MinInt32),
			"num.NewMilli32(-2147483, -648)"),
		newGoStringCase("milli64 day", num.NewMilli64(86400, 5),
			"num.NewMilli64(86400, 5)"),
		newGoStringCase("milli64 min", num.AsMilli64(math.MinInt64),
			"num.NewMilli64(-9223372036854775, -808)"),
	}
}

func goStringAtto128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("atto128 zero", num.NewAtto128(0, 0),
			"num.NewAtto128(0, 0)"),
		// the fraction is grouped in thousands from four digits up.
		newGoStringCase("atto128 one atto", num.NewAtto128(0, 1),
			"num.NewAtto128(0, 1)"),
		newGoStringCase("atto128 four digits", num.NewAtto128(2, 1234),
			"num.NewAtto128(2, 1_234)"),
		newGoStringCase("atto128 femto", num.NewAtto128(0, 1e15),
			"num.NewAtto128(0, 1_000_000_000_000_000)"),
		newGoStringCase("atto128 half", num.NewAtto128(1, 500e15),
			"num.NewAtto128(1, 500_000_000_000_000_000)"),
		newGoStringCase("atto128 negative", num.NewAtto128(-1, -500e15),
			"num.NewAtto128(-1, -500_000_000_000_000_000)"),
		newGoStringCase("atto128 negative fraction",
			num.NewAtto128(0, -123e15),
			"num.NewAtto128(0, -123_000_000_000_000_000)"),
		newGoStringCase("atto128 max int64 whole",
			num.NewAtto128(math.MaxInt64, 0),
			"num.NewAtto128(9223372036854775807, 0)"),
		// a whole count past int64 falls back to the backing's words
		// form: 2^63 whole units is 10^18 shifted left by 63, and the
		// minimum backing is the most negative Int128.
		newGoStringCase("atto128 whole past int64",
			num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0)),
			"num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0x0))"),
		newGoStringCase("atto128 min", num.AsAtto128(num.MinInt128),
			"num.AsAtto128(num.NewInt128(0x8000000000000000, 0x0))"),
	}
}

// TestAtto128GoStringFallback pins the fallback row's value: the words
// above are 2^63 whole units, one past the largest int64 whole count.
func TestAtto128GoStringFallback(t *testing.T) {
	past := num.NewAtto128(math.MaxInt64, 0).Add(num.NewAtto128(1, 0))
	core.AssertEqual(t, num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0)),
		past, "2^63 whole")
}

func TestGoString(t *testing.T) {
	core.RunTestCases(t, goStringIntCases())
	core.RunTestCases(t, goStringUint128Cases())
	core.RunTestCases(t, goStringInt128Cases())
	core.RunTestCases(t, goStringMilliCases())
	core.RunTestCases(t, goStringAtto128Cases())
}
