package num_test

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = convCase[num.Int32, num.Int64]{}
	_ core.TestCase = countCase[num.Int32]{}
)

// convertTo converts v to D through the method named for D, the one
// cell of the conversion matrix a row pins. D is one of the seven
// types of the family, so the switch always finds its arm.
func convertTo[D num.Number[D], S num.Number[S]](v S) (D, bool) {
	var out D
	var ok bool
	switch p := any(&out).(type) {
	case *num.Int32:
		*p, ok = v.Int32()
	case *num.Int64:
		*p, ok = v.Int64()
	case *num.Int128:
		*p, ok = v.Int128()
	case *num.Uint128:
		*p, ok = v.Uint128()
	case *num.Milli32:
		*p, ok = v.Milli32()
	case *num.Milli64:
		*p, ok = v.Milli64()
	case *num.Atto128:
		*p, ok = v.Atto128()
	default:
		// D is one of the seven, so no arm is left over.
	}
	return out, ok
}

// convCase pins one cell of the conversion matrix: in converted to D
// through the method named for it, both the value and the ok flag
// declared. An exact row keeps the value, so the way back recovers in;
// a truncated row drops fraction digits, so the way back stops short
// of in; an overflow row has ok false and keeps the low bits.
type convCase[S num.Number[S], D num.Number[D]] struct {
	in     S
	want   D
	name   string
	wantOK bool
	exact  bool
}

func newConvCase[S num.Number[S], D num.Number[D]](name string, in S,
	want D) convCase[S, D] {
	return convCase[S, D]{name: name, in: in, want: want, wantOK: true,
		exact: true}
}

func newConvCaseTruncated[S num.Number[S], D num.Number[D]](name string, in S,
	want D) convCase[S, D] {
	return convCase[S, D]{name: name, in: in, want: want, wantOK: true}
}

func newConvCaseOverflow[S num.Number[S], D num.Number[D]](name string, in S,
	want D) convCase[S, D] {
	return convCase[S, D]{name: name, in: in, want: want}
}

func (tc convCase[S, D]) Name() string { return tc.name }

func (tc convCase[S, D]) Test(t *testing.T) {
	t.Helper()
	got, ok := convertTo[D](tc.in)
	core.AssertEqual(t, tc.want, got, "value")
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	if !tc.wantOK {
		return
	}
	// a value that fitted fits its own type on the way back.
	back, ok := convertTo[S](got)
	core.AssertTrue(t, ok, "back ok")
	if tc.exact {
		core.AssertEqual(t, tc.in, back, "round trip")
	} else {
		core.AssertNotEqual(t, tc.in, back, "truncated")
	}
}

func convInt32Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCase("self", num.AsInt32(-5), num.AsInt32(-5)),
		newConvCase("to int64", num.AsInt32(math.MinInt32),
			num.AsInt64(math.MinInt32)),
		newConvCase("to int128", num.AsInt32(-5), num.AsInt128(-5)),
		newConvCase("to uint128", num.AsInt32(5), u(5)),
		// the bit pattern comes back as -1, but the value never fitted.
		newConvCaseOverflow("negative to uint128", num.AsInt32(-1),
			num.MaxUint128),
		newConvCase("to milli32", num.AsInt32(-5), num.NewMilli32(-5, 0)),
		newConvCase("milli32 whole max", num.AsInt32(2147483),
			num.NewMilli32(2147483, 0)),
		newConvCaseOverflow("past milli32 whole", num.AsInt32(2147484),
			num.AsMilli32(-2147483296)),
		newConvCaseOverflow("max to milli32", num.AsInt32(math.MaxInt32),
			num.NewMilli32(-1, 0)),
		newConvCase("min to milli64", num.AsInt32(math.MinInt32),
			num.NewMilli64(math.MinInt32, 0)),
		newConvCase("to atto128", num.AsInt32(-5), num.NewAtto128(-5, 0)),
	)
}

func convInt64Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCase("to int32", num.AsInt64(math.MinInt32),
			num.AsInt32(math.MinInt32)),
		newConvCaseOverflow("past int32", num.AsInt64(math.MaxInt32+1),
			num.AsInt32(math.MinInt32)),
		newConvCaseOverflow("high bits to int32", num.AsInt64(1<<40),
			num.AsInt32(0)),
		newConvCase("self", num.AsInt64(-5), num.AsInt64(-5)),
		newConvCase("min to int128", num.AsInt64(math.MinInt64),
			num.AsInt128(math.MinInt64)),
		newConvCase("to uint128", num.AsInt64(5), u(5)),
		newConvCaseOverflow("negative to uint128", num.AsInt64(-5),
			num.NewUint128(maxWord, maxWord-4)),
		newConvCase("to milli32", num.AsInt64(-2147483),
			num.NewMilli32(-2147483, 0)),
		newConvCaseOverflow("two to the 31 to milli32", num.AsInt64(1<<31),
			num.NewMilli32(0, 0)),
		newConvCase("milli64 whole max", num.AsInt64(9223372036854775),
			num.NewMilli64(9223372036854775, 0)),
		newConvCaseOverflow("past milli64 whole", num.AsInt64(9223372036854776),
			num.AsMilli64(-9223372036854775616)),
		newConvCaseOverflow("max to milli64", num.AsInt64(math.MaxInt64),
			num.NewMilli64(-1, 0)),
		// 9223372036854775807e18, below the 128-bit range.
		newConvCase("max to atto128", num.AsInt64(math.MaxInt64),
			num.AsAtto128(num.NewInt128(0x6f05b59d3b1ffff, 0xf21f494c589c0000))),
	)
}

func convInt128Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCase("to int32", num.AsInt128(-5), num.AsInt32(-5)),
		newConvCaseOverflow("max to int32", num.MaxInt128, num.AsInt32(-1)),
		newConvCase("to int64", num.AsInt128(math.MaxInt64),
			num.AsInt64(math.MaxInt64)),
		newConvCaseOverflow("max to int64", num.MaxInt128, num.AsInt64(-1)),
		newConvCaseOverflow("min to int64", num.MinInt128, num.AsInt64(0)),
		newConvCase("self", num.MinInt128, num.MinInt128),
		newConvCase("max to uint128", num.MaxInt128,
			num.NewUint128(signBit-1, maxWord)),
		newConvCaseOverflow("min to uint128", num.MinInt128,
			num.NewUint128(signBit, 0)),
		newConvCase("to milli32", num.AsInt128(5), num.NewMilli32(5, 0)),
		newConvCase("to milli64", num.AsInt128(-5), num.NewMilli64(-5, 0)),
		newConvCaseOverflow("max to milli64", num.MaxInt128, num.NewMilli64(-1, 0)),
		// 170141183460469231731 is the last whole count an Atto128 holds.
		newConvCase("atto128 whole max", num.NewInt128(9, 0x392ee8e921d5d073),
			num.AsAtto128(num.NewInt128(0x7fffffffffffffff, 0xf67634e971ec0000))),
		newConvCaseOverflow("past atto128 whole",
			num.NewInt128(9, 0x392ee8e921d5d074),
			num.AsAtto128(num.NewInt128(0x8000000000000000, 0x456eb9d19500000))),
		newConvCaseOverflow("max to atto128", num.MaxInt128, num.NewAtto128(-1, 0)),
		newConvCaseOverflow("min to atto128", num.MinInt128, num.NewAtto128(0, 0)),
	)
}

func convUint128Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCase("to int32", u(5), num.AsInt32(5)),
		newConvCaseOverflow("max to int32", num.MaxUint128, num.AsInt32(-1)),
		newConvCase("to int64", u(math.MaxInt64), num.AsInt64(math.MaxInt64)),
		newConvCaseOverflow("two to the 64 to int64", num.NewUint128(1, 0),
			num.AsInt64(0)),
		newConvCase("to int128", num.NewUint128(signBit-1, maxWord),
			num.MaxInt128),
		newConvCaseOverflow("two to the 127 to int128", num.NewUint128(signBit, 0),
			num.MinInt128),
		newConvCaseOverflow("max to int128", num.MaxUint128, num.AsInt128(-1)),
		newConvCase("self", num.MaxUint128, num.MaxUint128),
		newConvCase("to milli32", u(5), num.NewMilli32(5, 0)),
		// the bits read as -1, which fits a Milli32; the value did not.
		newConvCaseOverflow("max to milli32", num.MaxUint128, num.NewMilli32(-1, 0)),
		newConvCase("to milli64", u(5), num.NewMilli64(5, 0)),
		newConvCase("two to the 64 to atto128", num.NewUint128(1, 0),
			num.AsAtto128(num.NewInt128(1e18, 0))),
		newConvCaseOverflow("two to the 127 to atto128", num.NewUint128(signBit, 0),
			num.NewAtto128(0, 0)),
	)
}

func convMilli32Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCaseTruncated("to int32", num.NewMilli32(1, 500), num.AsInt32(1)),
		newConvCaseTruncated("negative to int32", num.NewMilli32(-1, 500),
			num.AsInt32(-1)),
		newConvCase("whole to int32", num.NewMilli32(-2147483, 0),
			num.AsInt32(-2147483)),
		newConvCaseTruncated("min to int64", num.AsMilli32(math.MinInt32),
			num.AsInt64(-2147483)),
		newConvCase("to int128", num.NewMilli32(5, 0), num.AsInt128(5)),
		newConvCaseTruncated("to uint128", num.NewMilli32(1, 500), u(1)),
		// the fraction drops before the range check, and zero fits.
		newConvCaseTruncated("negative fraction to uint128", num.NewMilli32(0, -5),
			u(0)),
		newConvCaseOverflow("negative to uint128", num.NewMilli32(-1, 500),
			num.MaxUint128),
		newConvCase("self", num.NewMilli32(1, 500), num.NewMilli32(1, 500)),
		newConvCase("min to milli64", num.AsMilli32(math.MinInt32),
			num.NewMilli64(-2147483, 648)),
		newConvCase("to atto128", num.NewMilli32(1, 500),
			num.NewAtto128(1, 500e15)),
	)
}

func convMilli64Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCaseTruncated("to int32", num.NewMilli64(86400, 5),
			num.AsInt32(86400)),
		newConvCaseOverflow("past int32", num.NewMilli64(1<<31, 0),
			num.AsInt32(math.MinInt32)),
		newConvCase("to int64", num.NewMilli64(86400, 0), num.AsInt64(86400)),
		newConvCaseTruncated("max to int64", num.AsMilli64(math.MaxInt64),
			num.AsInt64(9223372036854775)),
		newConvCase("to int128", num.NewMilli64(-5, 0), num.AsInt128(-5)),
		newConvCase("to uint128", num.NewMilli64(5, 0), u(5)),
		newConvCase("to milli32", num.NewMilli64(-2147483, 648),
			num.AsMilli32(math.MinInt32)),
		newConvCaseOverflow("past milli32", num.NewMilli64(2147483, 648),
			num.AsMilli32(math.MinInt32)),
		newConvCase("self", num.NewMilli64(86400, 5), num.NewMilli64(86400, 5)),
		newConvCase("max to atto128", num.AsMilli64(math.MaxInt64),
			num.NewAtto128(9223372036854775, 807e15)),
	)
}

func convAtto128Cases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCaseTruncated("to int32", num.NewAtto128(1, 500e15), num.AsInt32(1)),
		newConvCaseOverflow("max to int32", num.AsAtto128(num.MaxInt128),
			num.AsInt32(567660659)),
		newConvCase("to int64", num.NewAtto128(-5, 0), num.AsInt64(-5)),
		// 170141183460469231731 whole units, past the 64-bit range.
		newConvCaseOverflow("max to int64", num.AsAtto128(num.MaxInt128),
			num.AsInt64(4120486797083267187)),
		newConvCaseOverflow("min to int64", num.AsAtto128(num.MinInt128),
			num.AsInt64(-4120486797083267187)),
		newConvCaseTruncated("to int128", num.NewAtto128(-1, 500e15),
			num.AsInt128(-1)),
		newConvCase("to uint128", num.NewAtto128(5, 0), u(5)),
		newConvCaseOverflow("negative to uint128", num.NewAtto128(-5, 0),
			num.NewUint128(maxWord, maxWord-4)),
		newConvCaseTruncated("to milli32", num.NewAtto128(1, 234500e12),
			num.NewMilli32(1, 234)),
		newConvCaseTruncated("negative to milli32", num.NewAtto128(-1, 234500e12),
			num.NewMilli32(-1, 234)),
		newConvCaseTruncated("one atto to milli32", num.NewAtto128(0, 1),
			num.NewMilli32(0, 0)),
		newConvCase("to milli64", num.NewAtto128(1, 500e15),
			num.NewMilli64(1, 500)),
		newConvCaseOverflow("max to milli64", num.AsAtto128(num.MaxInt128),
			num.AsMilli64(6862868646037177319)),
		newConvCase("self", num.NewAtto128(1, 500e15), num.NewAtto128(1, 500e15)),
	)
}

func TestConvert(t *testing.T) {
	t.Run("int32", runTestConvertInt32)
	t.Run("int64", runTestConvertInt64)
	t.Run("int128", runTestConvertInt128)
	t.Run("uint128", runTestConvertUint128)
	t.Run("milli32", runTestConvertMilli32)
	t.Run("milli64", runTestConvertMilli64)
	t.Run("atto128", runTestConvertAtto128)
}

func runTestConvertInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convInt32Cases())
}

func runTestConvertInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convInt64Cases())
}

func runTestConvertInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convInt128Cases())
}

func runTestConvertUint128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convUint128Cases())
}

func runTestConvertMilli32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convMilli32Cases())
}

func runTestConvertMilli64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convMilli64Cases())
}

func runTestConvertAtto128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convAtto128Cases())
}

// counted is the count surface of a Decimal: its backing integer read
// as each of the signed integers, with a size check.
type counted interface {
	AsInt32() (num.Int32, bool)
	AsInt64() (num.Int64, bool)
	AsInt128() (num.Int128, bool)
}

// asCount reads the count of v as D, one of the three signed integers,
// so the switch always finds its arm.
func asCount[D num.Number[D]](v counted) (D, bool) {
	var out D
	var ok bool
	switch p := any(&out).(type) {
	case *num.Int32:
		*p, ok = v.AsInt32()
	case *num.Int64:
		*p, ok = v.AsInt64()
	case *num.Int128:
		*p, ok = v.AsInt128()
	default:
		// D is one of the three, so no arm is left over.
	}
	return out, ok
}

// countCase pins a count accessor of a Decimal: the backing integer as
// D, narrowed with a size check that keeps the low bits on failure, or
// widened.
type countCase[D num.Number[D]] struct {
	in     counted
	want   D
	name   string
	wantOK bool
}

func newCountCase[D num.Number[D]](name string, in counted,
	want D) countCase[D] {
	return countCase[D]{name: name, in: in, want: want, wantOK: true}
}

func newCountCaseOverflow[D num.Number[D]](name string, in counted,
	want D) countCase[D] {
	return countCase[D]{name: name, in: in, want: want}
}

func (tc countCase[D]) Name() string { return tc.name }

func (tc countCase[D]) Test(t *testing.T) {
	t.Helper()
	got, ok := asCount[D](tc.in)
	core.AssertEqual(t, tc.want, got, "count")
	core.AssertEqual(t, tc.wantOK, ok, "ok")
}

func countCases() []core.TestCase {
	return core.S[core.TestCase](
		newCountCase("milli32 as int32", num.NewMilli32(1, 500),
			num.AsInt32(1500)),
		newCountCase("milli32 min as int32", num.AsMilli32(math.MinInt32),
			num.AsInt32(math.MinInt32)),
		newCountCase("milli32 as int64", num.NewMilli32(-1, 500),
			num.AsInt64(-1500)),
		newCountCase("milli32 as int128", num.NewMilli32(1, 500),
			num.AsInt128(1500)),
		newCountCaseOverflow("milli64 past int32", num.AsMilli64(1<<31),
			num.AsInt32(math.MinInt32)),
		newCountCase("milli64 as int64", num.AsMilli64(1<<31),
			num.AsInt64(1<<31)),
		newCountCase("milli64 min as int128", num.AsMilli64(math.MinInt64),
			num.AsInt128(math.MinInt64)),
		newCountCaseOverflow("atto128 one as int32", num.NewAtto128(1, 0),
			num.AsInt32(-1486618624)),
		newCountCase("atto128 one atto as int32", num.NewAtto128(0, -1),
			num.AsInt32(-1)),
		newCountCase("atto128 one as int64", num.NewAtto128(1, 0),
			num.AsInt64(1e18)),
		newCountCaseOverflow("atto128 max as int64", num.AsAtto128(num.MaxInt128),
			num.AsInt64(-1)),
		newCountCase("atto128 max as int128", num.AsAtto128(num.MaxInt128),
			num.MaxInt128),
	)
}

func TestCount(t *testing.T) {
	core.RunTestCases(t, countCases())
}
