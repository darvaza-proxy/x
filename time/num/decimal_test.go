package num_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = decimalSignCase[num.Atto128]{}
	_ core.TestCase = decimalMulCase[num.Atto128]{}
	_ core.TestCase = decimalDivCase[num.Atto128]{}
	_ core.TestCase = decimalDivModCase[num.Atto128]{}
	_ core.TestCase = decimalMulDivModCase[num.Atto128]{}
)

// decimalType carries what a fixed-point [num.Decimal] instantiation
// needs to run the shared suite: how to build a value from a whole count
// and a sub-unit fraction, its resolution (sub-units per whole unit), and
// a whole count large enough that Div overflows the backing width. The
// fraction rows are written as exact divisions of the scale, so one set
// of cases fits every resolution.
type decimalType[D num.Signed[D]] struct {
	mk    func(whole, frac int64) D
	scale int64
	big   int64
}

func runDecimalTests[D num.Signed[D]](t *testing.T, dt decimalType[D]) {
	t.Helper()
	t.Run("sign", dt.testSign)
	t.Run("mul", dt.testMul)
	t.Run("div", dt.testDiv)
	t.Run("div-mod", dt.testDivMod)
	t.Run("mul-div-mod", dt.testMulDivMod)
	t.Run("basics", dt.testBasics)
	t.Run("div-zero", dt.testDivZero)
	t.Run("overflow", dt.testOverflow)
}

func (dt decimalType[D]) testSign(t *testing.T) {
	core.RunTestCases(t, decimalSignCases(dt.mk, dt.scale))
}

func (dt decimalType[D]) testMul(t *testing.T) {
	core.RunTestCases(t, decimalMulCases(dt.mk, dt.scale))
}

func (dt decimalType[D]) testDiv(t *testing.T) {
	core.RunTestCases(t, decimalDivCases(dt.mk, dt.scale))
}

func (dt decimalType[D]) testDivMod(t *testing.T) {
	core.RunTestCases(t, decimalDivModCases(dt.mk, dt.scale))
}

func (dt decimalType[D]) testMulDivMod(t *testing.T) {
	core.RunTestCases(t, decimalMulDivModCases(dt.mk, dt.scale))
}

func (dt decimalType[D]) testBasics(t *testing.T) {
	t.Helper()
	half := dt.scale / 2
	core.AssertTrue(t, dt.mk(0, 0).IsZero(), "zero")
	core.AssertFalse(t, dt.mk(1, 0).IsZero(), "non-zero")
	core.AssertTrue(t, dt.mk(1, half).Equal(dt.mk(1, half)), "equal")
	core.AssertFalse(t, dt.mk(1, 0).Equal(dt.mk(2, 0)), "not equal")
	// 4.0 - 2.5 == 1.5
	assertSignedEqual(t, dt.mk(1, half), dt.mk(4, 0).Sub(dt.mk(2, half)), "sub")
	core.AssertEqual(t, -1, dt.mk(1, 0).Cmp(dt.mk(2, 0)), "cmp less")
	core.AssertEqual(t, 1, dt.mk(2, 0).Cmp(dt.mk(1, 0)), "cmp greater")
	core.AssertEqual(t, 0, dt.mk(1, 0).Cmp(dt.mk(1, 0)), "cmp equal")
	assertSignedEqual(t, dt.mk(-1, half), dt.mk(1, half).Neg(), "neg")
	// a fraction beyond one whole unit carries into the whole part.
	assertSignedEqual(t, dt.mk(3, 0), dt.mk(1, 2*dt.scale), "carry")
}

func (dt decimalType[D]) testDivZero(t *testing.T) {
	t.Helper()
	one, zero := dt.mk(1, 0), dt.mk(0, 0)
	core.AssertPanic(t, func() { one.Div(zero) }, num.ErrDivZero, "div by zero")
	core.AssertPanic(t, func() { one.DivMod(zero) }, num.ErrDivZero,
		"div-mod by zero")
	core.AssertPanic(t, func() { one.Mod(zero) }, num.ErrDivZero, "mod by zero")
	core.AssertPanic(t, func() { one.MulDivMod(one, zero) }, num.ErrDivZero,
		"mul-div-mod by zero")
}

// testOverflow drives a quotient past the backing width, which wraps
// rather than panicking, matching the policy of Add and Mul.
func (dt decimalType[D]) testOverflow(t *testing.T) {
	t.Helper()
	huge := dt.mk(dt.big, 0)
	tiny := dt.mk(0, 1) // one sub-unit
	core.AssertNoPanic(t, func() { huge.Div(tiny) }, "div overflow")
}

// decimalSignCase checks how the constructor assigns the sign: from the
// whole part, or from the fraction when the whole is zero.
type decimalSignCase[D num.Signed[D]] struct {
	in      D
	wantAbs D
	name    string
	wantNeg bool
}

func newDecimalSignCase[D num.Signed[D]](name string, in D, wantNeg bool,
	wantAbs D) decimalSignCase[D] {
	return decimalSignCase[D]{
		name:    name,
		in:      in,
		wantAbs: wantAbs,
		wantNeg: wantNeg,
	}
}

func (tc decimalSignCase[D]) Name() string { return tc.name }

func (tc decimalSignCase[D]) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.wantNeg, tc.in.IsNegative(), "negative")
	assertSignedEqual(t, tc.wantAbs, tc.in.Abs(), "magnitude")
}

func decimalSignCases[D num.Signed[D]](mk func(whole, frac int64) D,
	scale int64) []decimalSignCase[D] {
	half := scale / 2
	return []decimalSignCase[D]{
		newDecimalSignCase("whole only", mk(5, 0), false, mk(5, 0)),
		newDecimalSignCase("neg whole", mk(-5, 0), true, mk(5, 0)),
		// whole is zero: the sign comes from the fraction.
		newDecimalSignCase("frac only", mk(0, half), false, mk(0, half)),
		newDecimalSignCase("neg frac only", mk(0, -half), true, mk(0, half)),
		// whole is non-zero: the fraction is a magnitude, its sign ignored.
		newDecimalSignCase("neg whole pos frac", mk(-1, half), true, mk(1, half)),
		newDecimalSignCase("neg whole neg frac", mk(-1, -half), true, mk(1, half)),
		newDecimalSignCase("pos whole neg frac", mk(1, -half), false, mk(1, half)),
	}
}

// decimalMulCase exercises the fixed-point product.
type decimalMulCase[D num.Signed[D]] struct {
	a    D
	b    D
	want D
	name string
}

func newDecimalMulCase[D num.Signed[D]](name string, a, b,
	want D) decimalMulCase[D] {
	return decimalMulCase[D]{name: name, a: a, b: b, want: want}
}

func (tc decimalMulCase[D]) Name() string { return tc.name }

func (tc decimalMulCase[D]) Test(t *testing.T) {
	t.Helper()
	assertSignedEqual(t, tc.want, tc.a.Mul(tc.b), "product")
}

func decimalMulCases[D num.Signed[D]](mk func(whole, frac int64) D,
	scale int64) []decimalMulCase[D] {
	half, quarter := scale/2, scale/4
	return []decimalMulCase[D]{
		newDecimalMulCase("whole", mk(2, 0), mk(3, 0), mk(6, 0)),
		newDecimalMulCase("halves", mk(0, half), mk(0, half), mk(0, quarter)),
		newDecimalMulCase("mixed", mk(1, half), mk(2, 0), mk(3, 0)),
		newDecimalMulCase("neg times pos", mk(-1, half), mk(2, 0), mk(-3, 0)),
		newDecimalMulCase("neg times neg", mk(-1, half), mk(-2, 0), mk(3, 0)),
	}
}

// decimalDivCase exercises the fractional quotient (truncated at the
// resolution).
type decimalDivCase[D num.Signed[D]] struct {
	a    D
	b    D
	want D
	name string
}

func newDecimalDivCase[D num.Signed[D]](name string, a, b,
	want D) decimalDivCase[D] {
	return decimalDivCase[D]{name: name, a: a, b: b, want: want}
}

func (tc decimalDivCase[D]) Name() string { return tc.name }

func (tc decimalDivCase[D]) Test(t *testing.T) {
	t.Helper()
	assertSignedEqual(t, tc.want, tc.a.Div(tc.b), "quotient")
}

func decimalDivCases[D num.Signed[D]](mk func(whole, frac int64) D,
	scale int64) []decimalDivCase[D] {
	half, quarter, third := scale/2, scale/4, scale/3
	return []decimalDivCase[D]{
		newDecimalDivCase("exact half", mk(1, 0), mk(2, 0), mk(0, half)),
		newDecimalDivCase("whole result", mk(6, 0), mk(2, 0), mk(3, 0)),
		newDecimalDivCase("quarter", mk(1, 0), mk(4, 0), mk(0, quarter)),
		// 1/3 truncates at the resolution.
		newDecimalDivCase("third", mk(1, 0), mk(3, 0), mk(0, third)),
		newDecimalDivCase("neg over pos", mk(-1, 0), mk(2, 0), mk(0, -half)),
	}
}

// decimalDivModCase exercises whole-multiple reduction, where the
// quotient is an integer count and the remainder is the leftover.
type decimalDivModCase[D num.Signed[D]] struct {
	a     D
	b     D
	wantQ D
	wantR D
	name  string
}

func newDecimalDivModCase[D num.Signed[D]](name string, a, b, wantQ,
	wantR D) decimalDivModCase[D] {
	return decimalDivModCase[D]{
		name:  name,
		a:     a,
		b:     b,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc decimalDivModCase[D]) Name() string { return tc.name }

func (tc decimalDivModCase[D]) Test(t *testing.T) {
	t.Helper()
	q, r := tc.a.DivMod(tc.b)
	assertSignedEqual(t, tc.wantQ, q, "count")
	assertSignedEqual(t, tc.wantR, r, "remainder")
	assertSignedEqual(t, tc.wantR, tc.a.Mod(tc.b), "mod")
	// invariant: a == q*b + r (q*b is the fractional product).
	assertSignedEqual(t, tc.a, q.Mul(tc.b).Add(r), "identity")
}

func decimalDivModCases[D num.Signed[D]](mk func(whole, frac int64) D,
	scale int64) []decimalDivModCase[D] {
	half, tenth := scale/2, scale/10
	return []decimalDivModCase[D]{
		// 4.5 DivMod 2.1 = (2, 0.3): the motivating case.
		newDecimalDivModCase("leftover", mk(4, half), mk(2, tenth), mk(2, 0),
			mk(0, 3*tenth)),
		newDecimalDivModCase("exact", mk(6, 0), mk(2, 0), mk(3, 0), mk(0, 0)),
		newDecimalDivModCase("count less than one", mk(2, 0), mk(4, half),
			mk(0, 0), mk(2, 0)),
		newDecimalDivModCase("neg dividend", mk(-4, -half), mk(2, tenth),
			mk(-2, 0), mk(0, -3*tenth)),
		newDecimalDivModCase("neg divisor", mk(4, half), mk(-2, -tenth),
			mk(-2, 0), mk(0, 3*tenth)),
		newDecimalDivModCase("both neg", mk(-4, -half), mk(-2, -tenth), mk(2, 0),
			mk(0, -3*tenth)),
	}
}

// decimalMulDivModCase exercises the scale-preserving fused
// multiply-divide. The two scale factors of the product against the one
// of the divisor leave the quotient at the original resolution.
type decimalMulDivModCase[D num.Signed[D]] struct {
	a     D
	b     D
	d     D
	wantQ D
	name  string
}

func newDecimalMulDivModCase[D num.Signed[D]](name string, a, b, d,
	wantQ D) decimalMulDivModCase[D] {
	return decimalMulDivModCase[D]{name: name, a: a, b: b, d: d, wantQ: wantQ}
}

func (tc decimalMulDivModCase[D]) Name() string { return tc.name }

func (tc decimalMulDivModCase[D]) Test(t *testing.T) {
	t.Helper()
	q, _ := tc.a.MulDivMod(tc.b, tc.d)
	assertSignedEqual(t, tc.wantQ, q, "quotient")
}

func decimalMulDivModCases[D num.Signed[D]](mk func(whole, frac int64) D,
	scale int64) []decimalMulDivModCase[D] {
	half, third := scale/2, scale/3
	return []decimalMulDivModCase[D]{
		// 1.5 * 2.0 / 3.0 = 1.0
		newDecimalMulDivModCase("value", mk(1, half), mk(2, 0), mk(3, 0),
			mk(1, 0)),
		// 6.0 * 2.0 / 4.0 = 3.0
		newDecimalMulDivModCase("whole", mk(6, 0), mk(2, 0), mk(4, 0), mk(3, 0)),
		// -1.5 * 2.0 / 3.0 = -1.0
		newDecimalMulDivModCase("neg product", mk(-1, half), mk(2, 0), mk(3, 0),
			mk(-1, 0)),
		// 1.0 * 1.0 / 3.0 truncates at the resolution.
		newDecimalMulDivModCase("truncated", mk(1, 0), mk(1, 0), mk(3, 0),
			mk(0, third)),
	}
}

func TestAtto128(t *testing.T) {
	runDecimalTests(t, decimalType[num.Atto128]{
		mk:    num.NewAtto128,
		scale: 1e18,
		big:   1e9,
	})
}

func TestMilli32(t *testing.T) {
	runDecimalTests(t, decimalType[num.Milli32]{
		mk: func(whole, frac int64) num.Milli32 {
			return num.NewMilli32(int32(whole), int32(frac))
		},
		scale: 1e3,
		big:   1e6,
	})
}

func TestMilli64(t *testing.T) {
	runDecimalTests(t, decimalType[num.Milli64]{
		mk:    num.NewMilli64,
		scale: 1e3,
		big:   1e13,
	})
}
