package num_test

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// assertSignedEqual compares two values through core.AreEqual, which
// honours their Equal method, so the shared suite constrains T only to
// num.Signed and never to comparable.
func assertSignedEqual[T num.Signed[T]](t *testing.T, expected, actual T,
	name string) bool {
	t.Helper()
	ok, _ := core.AreEqual(expected, actual)
	return core.AssertTrue(t, ok, "%s want %v got %v", name, expected, actual)
}

// signedIntType carries what a signed integer type needs to run the
// shared suites, truncated and Euclidean alike: how to build a value
// from an int64, its bounds, and a wide multiply exercising the
// MulDivMod intermediate; for the native integers its product also
// overflows the type, proving the wider path.
type signedIntType[T num.SignedEuclidean[T]] struct {
	mk    func(int64) T
	min   T
	max   T
	wideA int64
	wideB int64
	wideD int64
	wideQ int64
}

func runSignedIntTests[T num.SignedEuclidean[T]](t *testing.T,
	it signedIntType[T]) {
	t.Helper()
	t.Run("div-mod", it.testDivMod)
	t.Run("mul-div-mod", it.testMulDivMod)
	t.Run("unary", it.testUnary)
	t.Run("cmp", it.testCmp)
	t.Run("div-zero", it.testDivZero)
	t.Run("basics", it.testBasics)
}

func (it signedIntType[T]) testDivMod(t *testing.T) {
	core.RunTestCases(t, signedDivModCases(it.mk, it.min))
}

func (it signedIntType[T]) testMulDivMod(t *testing.T) {
	core.RunTestCases(t, signedMulDivModCases(it.mk,
		it.wideA, it.wideB, it.wideD, it.wideQ))
}

func (it signedIntType[T]) testUnary(t *testing.T) {
	core.RunTestCases(t, signedUnaryCases(it.mk, it.min))
}

func (it signedIntType[T]) testCmp(t *testing.T) {
	core.RunTestCases(t, signedCmpCases(it.mk, it.min, it.max))
}

func (it signedIntType[T]) testDivZero(t *testing.T) {
	t.Helper()
	zero, one := it.mk(0), it.mk(1)
	core.AssertPanic(t, func() { one.DivMod(zero) }, num.ErrDivZero,
		"div-mod by zero")
	core.AssertPanic(t, func() { one.Div(zero) }, num.ErrDivZero, "div by zero")
	core.AssertPanic(t, func() { one.Mod(zero) }, num.ErrDivZero, "mod by zero")
	core.AssertPanic(t, func() { one.MulDivMod(one, zero) }, num.ErrDivZero,
		"mul-div-mod by zero")
}

func (it signedIntType[T]) testBasics(t *testing.T) {
	t.Helper()
	core.AssertTrue(t, it.mk(0).IsZero(), "zero")
	core.AssertFalse(t, it.mk(1).IsZero(), "non-zero")
	core.AssertFalse(t, it.min.IsZero(), "min non-zero")
	// 5 - 3 == 2, and 3 - 5 == -2 across zero.
	assertSignedEqual(t, it.mk(2), it.mk(5).Sub(it.mk(3)), "sub positive")
	assertSignedEqual(t, it.mk(-2), it.mk(3).Sub(it.mk(5)), "sub crossing zero")
}

var (
	_ core.TestCase = signedDivModCase[num.Int128]{}
	_ core.TestCase = signedMulDivModCase[num.Int128]{}
	_ core.TestCase = signedUnaryCase[num.Int128]{}
	_ core.TestCase = signedCmpCase[num.Int128]{}
)

// signedDivModCase exercises truncated signed division across the sign
// matrix.
type signedDivModCase[T num.Signed[T]] struct {
	a     T
	b     T
	wantQ T
	wantR T
	name  string
}

func newSignedDivModCase[T num.Signed[T]](name string, a, b, wantQ,
	wantR T) signedDivModCase[T] {
	return signedDivModCase[T]{
		name:  name,
		a:     a,
		b:     b,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc signedDivModCase[T]) Name() string { return tc.name }

func (tc signedDivModCase[T]) Test(t *testing.T) {
	t.Helper()

	q, r := tc.a.DivMod(tc.b)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
	assertSignedEqual(t, tc.wantR, r, "remainder")
	assertSignedEqual(t, tc.wantQ, tc.a.Div(tc.b), "div")
	assertSignedEqual(t, tc.wantR, tc.a.Mod(tc.b), "mod")
	// invariant: a == q*b + r
	assertSignedEqual(t, tc.a, q.Mul(tc.b).Add(r), "identity")
}

func signedDivModCases[T num.Signed[T]](mk func(int64) T,
	minVal T) []signedDivModCase[T] {
	return []signedDivModCase[T]{
		newSignedDivModCase("pos pos", mk(7), mk(2), mk(3), mk(1)),
		newSignedDivModCase("neg pos", mk(-7), mk(2), mk(-3), mk(-1)),
		newSignedDivModCase("pos neg", mk(7), mk(-2), mk(-3), mk(1)),
		newSignedDivModCase("neg neg", mk(-7), mk(-2), mk(3), mk(-1)),
		newSignedDivModCase("exact", mk(6), mk(3), mk(2), mk(0)),
		newSignedDivModCase("divisor larger", mk(3), mk(5), mk(0), mk(3)),
		newSignedDivModCase("neg divisor larger", mk(-3), mk(5), mk(0), mk(-3)),
		// the minimum over -1 overflows and wraps to the minimum, matching Go.
		newSignedDivModCase("min over neg one", minVal, mk(-1), minVal, mk(0)),
	}
}

// signedMulDivModCase exercises the signed wide multiply-then-divide
// across the sign matrix.
type signedMulDivModCase[T num.Signed[T]] struct {
	a     T
	b     T
	d     T
	wantQ T
	name  string
}

func newSignedMulDivModCase[T num.Signed[T]](name string, a, b, d,
	wantQ T) signedMulDivModCase[T] {
	return signedMulDivModCase[T]{name: name, a: a, b: b, d: d, wantQ: wantQ}
}

func (tc signedMulDivModCase[T]) Name() string { return tc.name }

func (tc signedMulDivModCase[T]) Test(t *testing.T) {
	t.Helper()
	q, r := tc.a.MulDivMod(tc.b, tc.d)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
	// the remainder is pinned by the identity a*b == q*d + r, which also
	// fixes its sign.
	assertSignedEqual(t, tc.a.Mul(tc.b), q.Mul(tc.d).Add(r), "identity")
}

func signedMulDivModCases[T num.Signed[T]](mk func(int64) T, wideA, wideB,
	wideD, wideQ int64) []signedMulDivModCase[T] {
	return []signedMulDivModCase[T]{
		newSignedMulDivModCase("pos", mk(7), mk(3), mk(5), mk(4)),
		newSignedMulDivModCase("neg product", mk(-7), mk(3), mk(5), mk(-4)),
		newSignedMulDivModCase("neg divisor", mk(7), mk(3), mk(-5), mk(-4)),
		newSignedMulDivModCase("neg product neg divisor", mk(-7), mk(3), mk(-5),
			mk(4)),
		newSignedMulDivModCase("both operands neg", mk(-7), mk(-3), mk(5),
			mk(4)),
		newSignedMulDivModCase("exact", mk(6), mk(2), mk(3), mk(4)),
		newSignedMulDivModCase("divisor larger", mk(2), mk(3), mk(10), mk(0)),
		// wide product: the product overflows the native width, so a correct
		// quotient proves the wider intermediate.
		newSignedMulDivModCase("wide product", mk(wideA), mk(wideB), mk(wideD),
			mk(wideQ)),
	}
}

// signedUnaryCase exercises Neg, Abs and IsNegative.
type signedUnaryCase[T num.Signed[T]] struct {
	in       T
	wantNeg  T
	wantAbs  T
	name     string
	negative bool
}

func newSignedUnaryCase[T num.Signed[T]](name string, in, wantNeg, wantAbs T,
	negative bool) signedUnaryCase[T] {
	return signedUnaryCase[T]{
		name:     name,
		in:       in,
		wantNeg:  wantNeg,
		wantAbs:  wantAbs,
		negative: negative,
	}
}

func (tc signedUnaryCase[T]) Name() string { return tc.name }

func (tc signedUnaryCase[T]) Test(t *testing.T) {
	t.Helper()
	assertSignedEqual(t, tc.wantNeg, tc.in.Neg(), "neg")
	assertSignedEqual(t, tc.wantAbs, tc.in.Abs(), "abs")
	core.AssertEqual(t, tc.negative, tc.in.IsNegative(), "negative")
}

func signedUnaryCases[T num.Signed[T]](mk func(int64) T,
	minVal T) []signedUnaryCase[T] {
	return []signedUnaryCase[T]{
		newSignedUnaryCase("positive", mk(5), mk(-5), mk(5), false),
		newSignedUnaryCase("negative", mk(-5), mk(5), mk(5), true),
		newSignedUnaryCase("zero", mk(0), mk(0), mk(0), false),
		// the minimum has no positive counterpart: Neg and Abs wrap to it.
		newSignedUnaryCase("min", minVal, minVal, minVal, true),
	}
}

// signedCmpCase exercises signed ordering and Equal.
type signedCmpCase[T num.Signed[T]] struct {
	a    T
	b    T
	name string
	want int
}

func newSignedCmpCase[T num.Signed[T]](name string, a, b T,
	want int) signedCmpCase[T] {
	return signedCmpCase[T]{name: name, a: a, b: b, want: want}
}

func (tc signedCmpCase[T]) Name() string { return tc.name }

func (tc signedCmpCase[T]) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.a.Cmp(tc.b), "cmp")
	core.AssertEqual(t, tc.want == 0, tc.a.Equal(tc.b), "equal")
}

func signedCmpCases[T num.Signed[T]](mk func(int64) T, minVal,
	maxVal T) []signedCmpCase[T] {
	return []signedCmpCase[T]{
		newSignedCmpCase("equal", mk(5), mk(5), 0),
		newSignedCmpCase("less", mk(4), mk(5), -1),
		newSignedCmpCase("greater", mk(6), mk(5), 1),
		newSignedCmpCase("neg less than pos", mk(-1), mk(1), -1),
		newSignedCmpCase("min less than max", minVal, maxVal, -1),
		newSignedCmpCase("max greater than min", maxVal, minVal, 1),
	}
}

func TestInt32(t *testing.T) {
	runSignedIntTests(t, signedIntType[num.Int32]{
		mk:    func(x int64) num.Int32 { return num.Int32(x) },
		wideA: 100000, // 1e5 * 1e5 / 1e3 = 1e7, product overflows int32.
		wideB: 100000,
		wideD: 1000,
		wideQ: 10000000,
		min:   num.Int32(math.MinInt32),
		max:   num.Int32(math.MaxInt32),
	})
}

func TestInt64(t *testing.T) {
	runSignedIntTests(t, signedIntType[num.Int64]{
		mk:    func(x int64) num.Int64 { return num.Int64(x) },
		wideA: 1e12, // 1e12 * 1e12 / 1e6 = 1e18, product overflows int64.
		wideB: 1e12,
		wideD: 1e6,
		wideQ: 1e18,
		min:   num.Int64(math.MinInt64),
		max:   num.Int64(math.MaxInt64),
	})
}

func TestInt128(t *testing.T) {
	runSignedIntTests(t, signedIntType[num.Int128]{
		mk:    num.NewInt128,
		wideA: 1e12, // 1e12 * 1e12 / 1e6 = 1e18, through the mul256 path.
		wideB: 1e12,
		wideD: 1e6,
		wideQ: 1e18,
		min:   num.MinInt128,
		max:   num.MaxInt128,
	})
}
