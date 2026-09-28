package num_test

import (
	"math"
	"testing"

	"darvaza.org/x/time/num"
)

// BenchmarkConvert measures the conversions between the types of the
// family: an integer into another integer, wider, narrower and between
// signed and unsigned, or into a finer resolution; a Decimal into the
// integers and across resolutions both ways; OutsideCenti, whose
// backing widens through the fallback; and whole units through the
// factories, NewFromInt128 into the types of the package and
// NewDecimalFromInt128 into the outside instantiations, whose backings
// narrow one bit at a time, a value past the bound included.
func BenchmarkConvert(b *testing.B) {
	i32 := num.AsInt32(1234)
	i64 := num.AsInt64(1234)
	i128 := num.AsInt128(1234)
	u128 := num.AsUint128(1234)
	m64 := num.NewMilli64(1234, 567)
	a128 := num.NewAtto128(1234, 567_890_123_456_789_012)
	oc := NewOutsideCenti(1234, 56)
	// whole units one past what an OutsideCenti holds.
	past := num.AsInt128(math.MaxInt64/100 + 1)
	newOC := num.NewDecimalFromInt128[outsideSigned, outsideCentiScale]
	newOPC := num.NewDecimalFromInt128[outsidePlain, outsidePlainCentiScale]

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
	b.Run("NewFromInt128ToInt64",
		runBenchmarkConvert(i128, num.NewFromInt128[num.Int64]))
	b.Run("NewFromInt128ToMilli64",
		runBenchmarkConvert(i128, num.NewFromInt128[num.Milli64]))
	b.Run("NewFromInt128ToAtto128",
		runBenchmarkConvert(i128, num.NewFromInt128[num.Atto128]))
	b.Run("NewDecimalFromInt128ToOutsideCenti",
		runBenchmarkConvert(i128, newOC))
	b.Run("NewDecimalFromInt128ToOutsidePlainCenti",
		runBenchmarkConvert(i128, newOPC))
	b.Run("NewDecimalFromInt128PastOutsideCenti",
		runBenchmarkConvert(past, newOC))
}

// runBenchmarkConvert returns a benchmark of conv on v, a conversion
// method answering whether the value fitted or a factory answering an
// error.
func runBenchmarkConvert[T, U, R any](v T,
	conv func(T) (U, R)) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			conv(v)
		}
	}
}
