package num_test

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = euclideanDivModCase[num.Int32]{}
	_ core.TestCase = euclideanMulDivModCase[num.Int32]{}
	_ core.TestCase = euclideanDecimalMulDivModCase[num.Milli32]{}
)

// assertEuclideanRange checks the range half of the Euclidean
// contract: the remainder is non-negative and, whenever |w| is
// representable, below it.
func assertEuclideanRange[T num.SignedEuclidean[T]](t *testing.T, w, r T) {
	t.Helper()
	core.AssertFalse(t, r.IsNegative(), "remainder sign")
	if abs := w.Abs(); !abs.IsNegative() {
		core.AssertTrue(t, r.Cmp(abs) < 0, "remainder below |w|")
	}
}

// The Euclidean suite runs off the same signedIntType configuration as
// the truncated one, sharing the builder, bounds and wide
// multiply-divide triple.
func runEuclideanTests[T num.SignedEuclidean[T]](t *testing.T,
	it signedIntType[T]) {
	t.Helper()
	t.Run("div-mod", it.testEuclideanDivMod)
	t.Run("mul-div-mod", it.testEuclideanMulDivMod)
	t.Run("div-zero", it.testEuclideanDivZero)
}

func (it signedIntType[T]) testEuclideanDivMod(t *testing.T) {
	core.RunTestCases(t, euclideanDivModCases(it.mk, it.min, it.max))
}

func (it signedIntType[T]) testEuclideanMulDivMod(t *testing.T) {
	core.RunTestCases(t, euclideanMulDivModCases(it.mk,
		it.wideA, it.wideB, it.wideD, it.wideQ))
}

func (it signedIntType[T]) testEuclideanDivZero(t *testing.T) {
	t.Helper()
	zero, one := it.mk(0), it.mk(1)
	core.AssertPanic(t, func() { num.EuclideanDivMod(one, zero) },
		num.ErrDivZero, "div-mod by zero")
	core.AssertPanic(t, func() { num.EuclideanMulDivMod(one, one, zero) },
		num.ErrDivZero, "mul-div-mod by zero")
}

// euclideanDivModCase exercises EuclideanDivMod across the sign
// matrix, for the integer types and the fixed-point instantiations
// alike.
type euclideanDivModCase[T num.SignedEuclidean[T]] struct {
	v     T
	w     T
	wantQ T
	wantR T
	name  string
}

func newEuclideanDivModCase[T num.SignedEuclidean[T]](name string, v, w,
	wantQ, wantR T) euclideanDivModCase[T] {
	return euclideanDivModCase[T]{
		name:  name,
		v:     v,
		w:     w,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc euclideanDivModCase[T]) Name() string { return tc.name }

func (tc euclideanDivModCase[T]) Test(t *testing.T) {
	t.Helper()
	q, r := num.EuclideanDivMod(tc.v, tc.w)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
	assertSignedEqual(t, tc.wantR, r, "remainder")
	assertEuclideanRange(t, tc.w, r)
	// invariant: v == q*w + r
	assertSignedEqual(t, tc.v, q.Mul(tc.w).Add(r), "identity")
}

func euclideanDivModCases[T num.SignedEuclidean[T]](mk func(int64) T,
	minVal, maxVal T) []euclideanDivModCase[T] {
	return []euclideanDivModCase[T]{
		newEuclideanDivModCase("pos pos", mk(7), mk(2), mk(3), mk(1)),
		newEuclideanDivModCase("neg pos", mk(-7), mk(2), mk(-4), mk(1)),
		newEuclideanDivModCase("pos neg", mk(7), mk(-2), mk(-3), mk(1)),
		newEuclideanDivModCase("neg neg", mk(-7), mk(-2), mk(4), mk(1)),
		newEuclideanDivModCase("exact", mk(6), mk(3), mk(2), mk(0)),
		newEuclideanDivModCase("exact neg", mk(-6), mk(3), mk(-2), mk(0)),
		newEuclideanDivModCase("divisor larger", mk(3), mk(5), mk(0), mk(3)),
		newEuclideanDivModCase("neg divisor larger", mk(-3), mk(5), mk(-1),
			mk(2)),
		// |min| exceeds the positive range, yet -1 == 1*min + max still
		// lands on a representable remainder.
		newEuclideanDivModCase("min divisor", mk(-1), minVal, mk(1), maxVal),
		// min over -1 wraps like the truncated form; the remainder is
		// zero, so no adjustment applies.
		newEuclideanDivModCase("min over neg one", minVal, mk(-1), minVal,
			mk(0)),
	}
}

// euclideanMulDivModCase exercises EuclideanMulDivMod on the integer
// types across the sign matrix.
type euclideanMulDivModCase[T num.SignedEuclidean[T]] struct {
	v     T
	w     T
	d     T
	wantQ T
	wantR T
	name  string
}

//revive:disable-next-line:argument-limit
func newEuclideanMulDivModCase[T num.SignedEuclidean[T]](name string, v, w,
	d, wantQ, wantR T) euclideanMulDivModCase[T] {
	return euclideanMulDivModCase[T]{
		name:  name,
		v:     v,
		w:     w,
		d:     d,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc euclideanMulDivModCase[T]) Name() string { return tc.name }

func (tc euclideanMulDivModCase[T]) Test(t *testing.T) {
	t.Helper()
	q, r := num.EuclideanMulDivMod(tc.v, tc.w, tc.d)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
	assertSignedEqual(t, tc.wantR, r, "remainder")
	assertEuclideanRange(t, tc.d, r)
	// the identity holds against the wrapped product, as in the
	// truncated suite.
	assertSignedEqual(t, tc.v.Mul(tc.w), q.Mul(tc.d).Add(r), "identity")
}

func euclideanMulDivModCases[T num.SignedEuclidean[T]](mk func(int64) T,
	wideA, wideB, wideD, wideQ int64) []euclideanMulDivModCase[T] {
	return []euclideanMulDivModCase[T]{
		newEuclideanMulDivModCase("pos", mk(7), mk(3), mk(5), mk(4), mk(1)),
		newEuclideanMulDivModCase("neg product", mk(-7), mk(3), mk(5),
			mk(-5), mk(4)),
		newEuclideanMulDivModCase("neg divisor", mk(7), mk(3), mk(-5),
			mk(-4), mk(1)),
		newEuclideanMulDivModCase("neg product neg divisor", mk(-7), mk(3),
			mk(-5), mk(5), mk(4)),
		newEuclideanMulDivModCase("both operands neg", mk(-7), mk(-3),
			mk(5), mk(4), mk(1)),
		newEuclideanMulDivModCase("exact neg", mk(-6), mk(2), mk(3),
			mk(-4), mk(0)),
		// the product overflows the native width, so a correct quotient
		// proves the wider intermediate; both divisions are exact.
		newEuclideanMulDivModCase("wide product", mk(wideA), mk(wideB),
			mk(wideD), mk(wideQ), mk(0)),
		newEuclideanMulDivModCase("neg wide product", mk(-wideA), mk(wideB),
			mk(wideD), mk(-wideQ), mk(0)),
	}
}

// euclideanDecimalSuite carries what a fixed-point instantiation needs
// to run the Decimal Euclidean suite: how to build a value from a
// whole count and a sub-unit fraction, and its resolution.
type euclideanDecimalSuite[D num.SignedEuclidean[D]] struct {
	mk    func(whole, frac int64) D
	scale int64
}

func runEuclideanDecimalTests[D num.SignedEuclidean[D]](t *testing.T,
	s euclideanDecimalSuite[D]) {
	t.Helper()
	t.Run("div-mod", s.testDivMod)
	t.Run("mul-div-mod", s.testMulDivMod)
	t.Run("div-zero", s.testDivZero)
}

func (s euclideanDecimalSuite[D]) testDivMod(t *testing.T) {
	core.RunTestCases(t, euclideanDecimalDivModCases(s.mk, s.scale))
}

func (s euclideanDecimalSuite[D]) testMulDivMod(t *testing.T) {
	core.RunTestCases(t, euclideanDecimalMulDivModCases(s.mk, s.scale))
}

func (s euclideanDecimalSuite[D]) testDivZero(t *testing.T) {
	t.Helper()
	zero, one := s.mk(0, 0), s.mk(1, 0)
	core.AssertPanic(t, func() { num.EuclideanDivMod(one, zero) },
		num.ErrDivZero, "div-mod by zero")
	core.AssertPanic(t, func() { num.EuclideanMulDivMod(one, one, zero) },
		num.ErrDivZero, "mul-div-mod by zero")
}

func euclideanDecimalDivModCases[D num.SignedEuclidean[D]](
	mk func(whole, frac int64) D, scale int64) []euclideanDivModCase[D] {
	half, tenth := scale/2, scale/10
	return []euclideanDivModCase[D]{
		// 4.5 over 2.1: the truncated remainder is already positive.
		newEuclideanDivModCase("pos", mk(4, half), mk(2, tenth),
			mk(2, 0), mk(0, 3*tenth)),
		// -4.5 over 2.1 = (-3, 1.8): the motivating case.
		newEuclideanDivModCase("neg dividend", mk(-4, -half), mk(2, tenth),
			mk(-3, 0), mk(1, 8*tenth)),
		newEuclideanDivModCase("neg divisor", mk(4, half), mk(-2, -tenth),
			mk(-2, 0), mk(0, 3*tenth)),
		newEuclideanDivModCase("both neg", mk(-4, -half), mk(-2, -tenth),
			mk(3, 0), mk(1, 8*tenth)),
		newEuclideanDivModCase("exact neg", mk(-4, -2*tenth), mk(2, tenth),
			mk(-2, 0), mk(0, 0)),
	}
}

// euclideanDecimalMulDivModCase exercises EuclideanMulDivMod on the
// fixed-point instantiations, where the identity lives in backing
// sub-units and the explicit wantR pins the remainder instead.
type euclideanDecimalMulDivModCase[D num.SignedEuclidean[D]] struct {
	v     D
	w     D
	d     D
	wantQ D
	wantR D
	name  string
}

//revive:disable-next-line:argument-limit
func newEuclideanDecimalMulDivModCase[D num.SignedEuclidean[D]](name string,
	v, w, d, wantQ, wantR D) euclideanDecimalMulDivModCase[D] {
	return euclideanDecimalMulDivModCase[D]{
		name:  name,
		v:     v,
		w:     w,
		d:     d,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc euclideanDecimalMulDivModCase[D]) Name() string { return tc.name }

func (tc euclideanDecimalMulDivModCase[D]) Test(t *testing.T) {
	t.Helper()
	q, r := num.EuclideanMulDivMod(tc.v, tc.w, tc.d)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
	assertSignedEqual(t, tc.wantR, r, "remainder")
	assertEuclideanRange(t, tc.d, r)
}

func euclideanDecimalMulDivModCases[D num.SignedEuclidean[D]](
	mk func(whole, frac int64) D,
	scale int64) []euclideanDecimalMulDivModCase[D] {
	// every scale is a power of ten, so scale ≡ 1 (mod 3): 1.0*1.0/3.0
	// leaves a backing remainder of one whole unit, and correcting its
	// negation moves the quotient by a single sub-unit.
	third, half := scale/3, scale/2
	return []euclideanDecimalMulDivModCase[D]{
		newEuclideanDecimalMulDivModCase("pos", mk(1, 0), mk(1, 0),
			mk(3, 0), mk(0, third), mk(1, 0)),
		newEuclideanDecimalMulDivModCase("neg product", mk(-1, 0), mk(1, 0),
			mk(3, 0), mk(0, -(third+1)), mk(2, 0)),
		newEuclideanDecimalMulDivModCase("neg divisor", mk(1, 0), mk(1, 0),
			mk(-3, 0), mk(0, -third), mk(1, 0)),
		newEuclideanDecimalMulDivModCase("neg product neg divisor",
			mk(-1, 0), mk(1, 0), mk(-3, 0), mk(0, third+1), mk(2, 0)),
		newEuclideanDecimalMulDivModCase("exact", mk(1, half), mk(2, 0),
			mk(3, 0), mk(1, 0), mk(0, 0)),
		newEuclideanDecimalMulDivModCase("exact neg", mk(-1, -half),
			mk(2, 0), mk(3, 0), mk(-1, 0), mk(0, 0)),
	}
}

func TestEuclideanInt32(t *testing.T) {
	runEuclideanTests(t, signedIntType[num.Int32]{
		mk:    func(x int64) num.Int32 { return num.Int32(x) },
		wideA: 100000, // 1e5 * 1e5 / 1e3 = 1e7, product overflows int32.
		wideB: 100000,
		wideD: 1000,
		wideQ: 10000000,
		min:   num.Int32(math.MinInt32),
		max:   num.Int32(math.MaxInt32),
	})
}

func TestEuclideanInt64(t *testing.T) {
	runEuclideanTests(t, signedIntType[num.Int64]{
		mk:    func(x int64) num.Int64 { return num.Int64(x) },
		wideA: 1e12, // 1e12 * 1e12 / 1e6 = 1e18, product overflows int64.
		wideB: 1e12,
		wideD: 1e6,
		wideQ: 1e18,
		min:   num.Int64(math.MinInt64),
		max:   num.Int64(math.MaxInt64),
	})
}

func TestEuclideanInt128(t *testing.T) {
	runEuclideanTests(t, signedIntType[num.Int128]{
		mk:    num.NewInt128,
		wideA: 1e12, // 1e12 * 1e12 / 1e6 = 1e18, wide beyond 64 bits.
		wideB: 1e12,
		wideD: 1e6,
		wideQ: 1e18,
		min:   num.MinInt128,
		max:   num.MaxInt128,
	})
}

// TestEuclideanUint128 pins the unsigned no-op: with no negative
// remainder possible the correction never fires, so the Euclidean
// helpers must match the plain truncated forms.
func TestEuclideanUint128(t *testing.T) {
	q, r := num.EuclideanDivMod(u(100), u(7))
	core.AssertEqual(t, u(14), q, "quotient")
	core.AssertEqual(t, u(2), r, "remainder")

	q, r = num.EuclideanMulDivMod(u(6), u(7), u(5))
	core.AssertEqual(t, u(8), q, "quotient")
	core.AssertEqual(t, u(2), r, "remainder")
}

func TestEuclideanMilli32(t *testing.T) {
	runEuclideanDecimalTests(t, euclideanDecimalSuite[num.Milli32]{
		mk: func(whole, frac int64) num.Milli32 {
			return num.NewMilli32(int32(whole), int32(frac))
		},
		scale: 1e3,
	})
}

func TestEuclideanMilli64(t *testing.T) {
	runEuclideanDecimalTests(t, euclideanDecimalSuite[num.Milli64]{
		mk:    num.NewMilli64,
		scale: 1e3,
	})
}

func TestEuclideanAtto128(t *testing.T) {
	runEuclideanDecimalTests(t, euclideanDecimalSuite[num.Atto128]{
		mk:    num.NewAtto128,
		scale: 1e18,
	})
}
