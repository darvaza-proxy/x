package num_test

import (
	"testing"

	"darvaza.org/x/time/num"
)

// BenchmarkConvert measures the conversions between the types of the
// family: an integer into another integer, wider, narrower and
// between signed and unsigned, and into a finer resolution, a Decimal
// into the integers and across resolutions both ways, and
// OutsideCenti, whose backing widens through the fallback.
func BenchmarkConvert(b *testing.B) {
	i32 := num.AsInt32(1234)
	i64 := num.AsInt64(1234)
	i128 := num.AsInt128(1234)
	u128 := num.AsUint128(1234)
	m64 := num.NewMilli64(1234, 567)
	a128 := num.NewAtto128(1234, 567_890_123_456_789_012)
	oc := NewOutsideCenti(1234, 56)

	b.Run("Int32ToInt64", runBenchmarkConvert(i32, num.Int32.Int64))
	b.Run("Int64ToInt32", runBenchmarkConvert(i64, num.Int64.Int32))
	b.Run("Int64ToInt128", runBenchmarkConvert(i64, num.Int64.Int128))
	b.Run("Int128ToInt64", runBenchmarkConvert(i128, num.Int128.Int64))
	b.Run("Int128ToUint128", runBenchmarkConvert(i128, num.Int128.Uint128))
	b.Run("Uint128ToInt128", runBenchmarkConvert(u128, num.Uint128.Int128))
	b.Run("Int64ToMilli64", runBenchmarkConvert(i64, num.Int64.Milli64))
	b.Run("Int64ToAtto128", runBenchmarkConvert(i64, num.Int64.Atto128))
	b.Run("Milli64ToInt64", runBenchmarkConvert(m64, num.Milli64.Int64))
	b.Run("Milli64ToAtto128", runBenchmarkConvert(m64, num.Milli64.Atto128))
	b.Run("Atto128ToInt64", runBenchmarkConvert(a128, num.Atto128.Int64))
	b.Run("Atto128ToMilli64", runBenchmarkConvert(a128, num.Atto128.Milli64))
	b.Run("OutsideCentiToInt64", runBenchmarkConvert(oc, OutsideCenti.Int64))
	b.Run("OutsideCentiToMilli64",
		runBenchmarkConvert(oc, OutsideCenti.Milli64))
	b.Run("OutsideCentiToAtto128",
		runBenchmarkConvert(oc, OutsideCenti.Atto128))
}

// runBenchmarkConvert returns a benchmark of conv on v.
func runBenchmarkConvert[T, U any](v T,
	conv func(T) (U, bool)) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			conv(v)
		}
	}
}
