package num_test

import (
	"encoding"
	"fmt"
	"testing"

	"darvaza.org/x/time/num"
)

// BenchmarkDecimalAppendText measures AppendText into a buffer with
// room on each Decimal instantiation of the package, and on
// OutsideCenti, whose backing widens through the fallback: the
// primitive under String, MarshalText and %v, which allocates nothing.
func BenchmarkDecimalAppendText(b *testing.B) {
	b.Run("Milli32", runBenchmarkAppendText(num.NewMilli32(1234, 567)))
	b.Run("Milli64", runBenchmarkAppendText(num.NewMilli64(1234, 567)))
	b.Run("Atto128", runBenchmarkAppendText(
		num.NewAtto128(1234, 567_890_123_456_789_012)))
	b.Run("OutsideCenti", runBenchmarkAppendText(NewOutsideCenti(1234, 56)))
}

// BenchmarkDecimalFormat measures %.2f through fmt on each Decimal
// instantiation of the package, which rounds the fraction, and on
// OutsideCenti, whose backing widens through the fallback and whose
// two digits need no rounding.
func BenchmarkDecimalFormat(b *testing.B) {
	b.Run("Milli32", runBenchmarkFormat(num.NewMilli32(1234, 567)))
	b.Run("Milli64", runBenchmarkFormat(num.NewMilli64(1234, 567)))
	b.Run("Atto128", runBenchmarkFormat(
		num.NewAtto128(1234, 567_890_123_456_789_012)))
	b.Run("OutsideCenti", runBenchmarkFormat(NewOutsideCenti(1234, 56)))
}

// BenchmarkDecimalEuclideanMulDivMod measures EuclideanMulDivMod on each
// Decimal instantiation of the package, and on OutsideCenti, whose ULP
// comes from the scale divided by itself, over a product whose
// truncated remainder is negative, so the correction runs.
func BenchmarkDecimalEuclideanMulDivMod(b *testing.B) {
	b.Run("Milli32", runBenchmarkEuclideanMulDivMod(
		num.NewMilli32(-1234, -567), num.NewMilli32(3, 0),
		num.NewMilli32(11, 0)))
	b.Run("Milli64", runBenchmarkEuclideanMulDivMod(
		num.NewMilli64(-1234, -567), num.NewMilli64(3, 0),
		num.NewMilli64(11, 0)))
	b.Run("Atto128", runBenchmarkEuclideanMulDivMod(
		num.NewAtto128(-1234, -567_890_123_456_789_012),
		num.NewAtto128(3, 0), num.NewAtto128(11, 0)))
	b.Run("OutsideCenti", runBenchmarkEuclideanMulDivMod(
		NewOutsideCenti(-1234, -56), NewOutsideCenti(3, 0),
		NewOutsideCenti(11, 0)))
}

// runBenchmarkAppendText returns a benchmark of AppendText on v into a
// buffer with room.
func runBenchmarkAppendText(v encoding.TextAppender) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 64)
		for b.Loop() {
			buf, _ = v.AppendText(buf[:0])
		}
	}
}

// runBenchmarkFormat returns a benchmark of v under %.2f through fmt,
// into a buffer with room.
func runBenchmarkFormat(v fmt.Formatter) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 64)
		for b.Loop() {
			buf = fmt.Appendf(buf[:0], "%.2f", v)
		}
	}
}

// runBenchmarkEuclideanMulDivMod returns a benchmark of
// EuclideanMulDivMod on v*w/d.
func runBenchmarkEuclideanMulDivMod[T num.Euclidean[T]](v, w,
	d T) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			num.EuclideanMulDivMod(v, w, d)
		}
	}
}
