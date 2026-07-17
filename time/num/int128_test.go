package num_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// i builds an Int128 from a signed 64-bit value, for readable rows.
func i(x int64) num.Int128 {
	return num.NewInt128(x)
}

var (
	_ core.TestCase = int128DivModTestCase{}
	_ core.TestCase = int128MulDivModTestCase{}
	_ core.TestCase = int128UnaryTestCase{}
	_ core.TestCase = int128CmpTestCase{}
)

// int128DivModTestCase exercises truncated signed division across the
// sign matrix.
type int128DivModTestCase struct {
	name  string
	a     num.Int128
	b     num.Int128
	wantQ num.Int128
	wantR num.Int128
}

func newInt128DivModTestCase(name string, a, b, wantQ,
	wantR num.Int128) int128DivModTestCase {
	return int128DivModTestCase{
		name:  name,
		a:     a,
		b:     b,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc int128DivModTestCase) Name() string { return tc.name }

func (tc int128DivModTestCase) Test(t *testing.T) {
	t.Helper()

	q, r := tc.a.DivMod(tc.b)
	core.AssertEqual(t, tc.wantQ, q, "quotient")
	core.AssertEqual(t, tc.wantR, r, "remainder")
	core.AssertEqual(t, tc.wantQ, tc.a.Div(tc.b), "div")
	core.AssertEqual(t, tc.wantR, tc.a.Mod(tc.b), "mod")
	// invariant: a == q*b + r
	core.AssertEqual(t, tc.a, q.Mul(tc.b).Add(r), "identity")
}

func int128DivModTestCases() []int128DivModTestCase {
	return []int128DivModTestCase{
		newInt128DivModTestCase("pos pos", i(7), i(2), i(3), i(1)),
		newInt128DivModTestCase("neg pos", i(-7), i(2), i(-3), i(-1)),
		newInt128DivModTestCase("pos neg", i(7), i(-2), i(-3), i(1)),
		newInt128DivModTestCase("neg neg", i(-7), i(-2), i(3), i(-1)),
		newInt128DivModTestCase("exact", i(6), i(3), i(2), i(0)),
		newInt128DivModTestCase("divisor larger", i(3), i(5), i(0), i(3)),
		newInt128DivModTestCase("neg divisor larger", i(-3), i(5), i(0),
			i(-3)),
		// MinInt128 / -1 overflows and wraps to MinInt128, matching Go.
		newInt128DivModTestCase("min over neg one", num.MinInt128, i(-1),
			num.MinInt128, i(0)),
	}
}

func TestInt128DivMod(t *testing.T) {
	core.RunTestCases(t, int128DivModTestCases())
}

func TestInt128DivByZeroPanics(t *testing.T) {
	core.AssertPanic(t, func() { i(1).DivMod(i(0)) }, num.ErrDivZero,
		"div-mod by zero")
	core.AssertPanic(t, func() { i(1).Div(i(0)) }, num.ErrDivZero, "div by zero")
	core.AssertPanic(t, func() { i(1).Mod(i(0)) }, num.ErrDivZero, "mod by zero")
}

// int128MulDivModTestCase exercises the signed wide multiply-then-divide
// across the sign matrix.
type int128MulDivModTestCase struct {
	name  string
	a     num.Int128
	b     num.Int128
	d     num.Int128
	wantQ num.Int128
}

func newInt128MulDivModTestCase(name string, a, b, d,
	wantQ num.Int128) int128MulDivModTestCase {
	return int128MulDivModTestCase{name: name, a: a, b: b, d: d, wantQ: wantQ}
}

func (tc int128MulDivModTestCase) Name() string { return tc.name }

func (tc int128MulDivModTestCase) Test(t *testing.T) {
	t.Helper()
	q, r := tc.a.MulDivMod(tc.b, tc.d)
	core.AssertEqual(t, tc.wantQ, q, "quotient")
	// the remainder is pinned by the identity a*b == q*d + r, which also
	// fixes its sign.
	core.AssertEqual(t, tc.a.Mul(tc.b), q.Mul(tc.d).Add(r), "identity")
}

func int128MulDivModTestCases() []int128MulDivModTestCase {
	return []int128MulDivModTestCase{
		newInt128MulDivModTestCase("pos", i(7), i(3), i(5), i(4)),
		newInt128MulDivModTestCase("neg product", i(-7), i(3), i(5), i(-4)),
		newInt128MulDivModTestCase("neg divisor", i(7), i(3), i(-5), i(-4)),
		newInt128MulDivModTestCase("neg product neg divisor", i(-7), i(3),
			i(-5), i(4)),
		newInt128MulDivModTestCase("both operands neg", i(-7), i(-3), i(5),
			i(4)),
		newInt128MulDivModTestCase("exact", i(6), i(2), i(3), i(4)),
		newInt128MulDivModTestCase("divisor larger", i(2), i(3), i(10), i(0)),
		// wide product: 10^12 * 10^12 / 10^6 = 10^18, exercising mul256.
		newInt128MulDivModTestCase("wide product", i(1e12), i(1e12), i(1e6),
			i(1e18)),
	}
}

func TestInt128MulDivMod(t *testing.T) {
	core.RunTestCases(t, int128MulDivModTestCases())
}

func TestInt128MulDivModByZeroPanics(t *testing.T) {
	core.AssertPanic(t, func() { i(1).MulDivMod(i(1), i(0)) }, num.ErrDivZero,
		"mul-div-mod by zero")
}

// int128UnaryTestCase exercises Neg, Abs and IsNegative.
type int128UnaryTestCase struct {
	name     string
	in       num.Int128
	wantNeg  num.Int128
	wantAbs  num.Int128
	negative bool
}

func newInt128UnaryTestCase(name string, in, wantNeg, wantAbs num.Int128,
	negative bool) int128UnaryTestCase {
	return int128UnaryTestCase{
		name:     name,
		in:       in,
		wantNeg:  wantNeg,
		wantAbs:  wantAbs,
		negative: negative,
	}
}

func (tc int128UnaryTestCase) Name() string { return tc.name }

func (tc int128UnaryTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.wantNeg, tc.in.Neg(), "neg")
	core.AssertEqual(t, tc.wantAbs, tc.in.Abs(), "abs")
	core.AssertEqual(t, tc.negative, tc.in.IsNegative(), "negative")
}

func int128UnaryTestCases() []int128UnaryTestCase {
	return []int128UnaryTestCase{
		newInt128UnaryTestCase("positive", i(5), i(-5), i(5), false),
		newInt128UnaryTestCase("negative", i(-5), i(5), i(5), true),
		newInt128UnaryTestCase("zero", i(0), i(0), i(0), false),
		// MinInt128 has no positive counterpart: Neg and Abs wrap to it.
		newInt128UnaryTestCase("min", num.MinInt128, num.MinInt128,
			num.MinInt128, true),
	}
}

func TestInt128Unary(t *testing.T) {
	core.RunTestCases(t, int128UnaryTestCases())
}

// int128CmpTestCase exercises signed ordering and Equal.
type int128CmpTestCase struct {
	name string
	a    num.Int128
	b    num.Int128
	want int
}

func newInt128CmpTestCase(name string, a, b num.Int128,
	want int) int128CmpTestCase {
	return int128CmpTestCase{name: name, a: a, b: b, want: want}
}

func (tc int128CmpTestCase) Name() string { return tc.name }

func (tc int128CmpTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.a.Cmp(tc.b), "cmp")
	core.AssertEqual(t, tc.want == 0, tc.a.Equal(tc.b), "equal")
}

func int128CmpTestCases() []int128CmpTestCase {
	return []int128CmpTestCase{
		newInt128CmpTestCase("equal", i(5), i(5), 0),
		newInt128CmpTestCase("less", i(4), i(5), -1),
		newInt128CmpTestCase("greater", i(6), i(5), 1),
		newInt128CmpTestCase("neg less than pos", i(-1), i(1), -1),
		newInt128CmpTestCase("min less than max", num.MinInt128,
			num.MaxInt128, -1),
		// high word dominates the other way: exercises the sv > sw arm.
		newInt128CmpTestCase("max greater than min", num.MaxInt128,
			num.MinInt128, 1),
	}
}

func TestInt128Cmp(t *testing.T) {
	core.RunTestCases(t, int128CmpTestCases())
}

func TestInt128Basics(t *testing.T) {
	core.AssertTrue(t, i(0).IsZero(), "zero")
	core.AssertFalse(t, i(1).IsZero(), "non-zero")
	core.AssertFalse(t, num.MinInt128.IsZero(), "min non-zero")
	// 5 - 3 == 2, and 3 - 5 == -2 across zero.
	core.AssertEqual(t, i(2), i(5).Sub(i(3)), "sub positive")
	core.AssertEqual(t, i(-2), i(3).Sub(i(5)), "sub crossing zero")
}
