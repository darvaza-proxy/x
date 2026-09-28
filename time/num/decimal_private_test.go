package num

import (
	"math"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = narrowBitsTestCase[Int64, milli64Scale]{}

// narrowBitsTestCase pins narrowBits, the narrowing into a backing other
// than the package's integers, against the concrete arm of narrow for
// one of them on the same count: the same value, the bound included,
// and the same error.
type narrowBitsTestCase[T Signed[T], S DecimalScaler[T]] struct {
	name string
	in   Int128
}

func newNarrowBitsTestCase[T Signed[T], S DecimalScaler[T]](
	in Int128) narrowBitsTestCase[T, S] {
	return narrowBitsTestCase[T, S]{name: in.String(), in: in}
}

func (tc narrowBitsTestCase[T, S]) Name() string { return tc.name }

func (tc narrowBitsTestCase[T, S]) Test(t *testing.T) {
	t.Helper()
	var d Decimal[T, S]
	want, wantErr := d.narrow(tc.in)
	got, err := d.narrowBits(tc.in)
	core.AssertEqual(t, want, got, "value")
	core.AssertSame(t, wantErr, err, "error")
}

// narrowBitsCases runs every count of narrowCorpus through the
// Decimal over T and S.
func narrowBitsCases[T Signed[T], S DecimalScaler[T]]() []narrowBitsTestCase[T, S] {
	corpus := narrowCorpus()
	out := make([]narrowBitsTestCase[T, S], 0, len(corpus))
	for _, c := range corpus {
		out = append(out, newNarrowBitsTestCase[T, S](c))
	}
	return out
}

// narrowCorpus is the counts narrowBits is run on: zero and the unit,
// a count past the low word, and the bounds of every backing of the
// package and one past them.
func narrowCorpus() []Int128 {
	one := AsInt128(1)
	return core.S(
		AsInt128(0), one, one.Neg(),
		NewInt128(1, 0), NewInt128(1, 0).Neg(),
		AsInt128(math.MaxInt32), AsInt128(math.MaxInt32+1),
		AsInt128(math.MinInt32), AsInt128(math.MinInt32-1),
		AsInt128(math.MaxInt64), AsInt128(math.MaxInt64).Add(one),
		AsInt128(math.MinInt64), AsInt128(math.MinInt64).Sub(one),
		MaxInt128, MinInt128, MinInt128.Add(one),
	)
}

func TestNarrowBits(t *testing.T) {
	t.Run("int32", runTestNarrowBitsInt32)
	t.Run("int64", runTestNarrowBitsInt64)
	t.Run("int128", runTestNarrowBitsInt128)
}

func runTestNarrowBitsInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, narrowBitsCases[Int32, milli32Scale]())
}

func runTestNarrowBitsInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, narrowBitsCases[Int64, milli64Scale]())
}

func runTestNarrowBitsInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, narrowBitsCases[Int128, atto128Scale]())
}
