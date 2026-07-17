package num_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// maxWord is the largest uint64, the all-ones 64-bit word.
const maxWord = ^uint64(0)

// u builds a Uint128 from a single low word, for readable small values.
func u(lo uint64) num.Uint128 {
	return num.NewUint128(0, lo)
}

var (
	_ core.TestCase = uint128DivModTestCase{}
	_ core.TestCase = uint128MulTestCase{}
	_ core.TestCase = uint128MulDivModTestCase{}
	_ core.TestCase = uint128CmpTestCase{}
)

// uint128DivModTestCase exercises Div, Mod and DivMod together.
type uint128DivModTestCase struct {
	name  string
	a     num.Uint128
	b     num.Uint128
	wantQ num.Uint128
	wantR num.Uint128
}

func newUint128DivModTestCase(name string, a, b, wantQ,
	wantR num.Uint128) uint128DivModTestCase {
	return uint128DivModTestCase{
		name:  name,
		a:     a,
		b:     b,
		wantQ: wantQ,
		wantR: wantR,
	}
}

func (tc uint128DivModTestCase) Name() string { return tc.name }

func (tc uint128DivModTestCase) Test(t *testing.T) {
	t.Helper()

	q, r := tc.a.DivMod(tc.b)
	core.AssertEqual(t, tc.wantQ, q, "quotient")
	core.AssertEqual(t, tc.wantR, r, "remainder")
	core.AssertEqual(t, tc.wantQ, tc.a.Div(tc.b), "div")
	core.AssertEqual(t, tc.wantR, tc.a.Mod(tc.b), "mod")
	// invariant: a == q*b + r
	core.AssertEqual(t, tc.a, q.Mul(tc.b).Add(r), "identity")
}

func uint128DivModTestCases() []uint128DivModTestCase {
	return []uint128DivModTestCase{
		newUint128DivModTestCase("exact", u(100), u(10), u(10), u(0)),
		newUint128DivModTestCase("remainder", u(100), u(7), u(14), u(2)),
		newUint128DivModTestCase("divisor larger", u(3), u(5), u(0), u(3)),
		newUint128DivModTestCase("by one", u(100), u(1), u(100), u(0)),
		newUint128DivModTestCase("self", u(42), u(42), u(1), u(0)),
		newUint128DivModTestCase("high word", num.NewUint128(1, 0), u(2),
			num.NewUint128(0, 1<<63), u(0)),
		newUint128DivModTestCase("max by max", num.MaxUint128,
			num.MaxUint128, u(1), u(0)),
	}
}

func TestUint128DivMod(t *testing.T) {
	core.RunTestCases(t, uint128DivModTestCases())
}

func TestUint128DivByZeroPanics(t *testing.T) {
	core.AssertPanic(t, func() { u(1).DivMod(u(0)) }, num.ErrDivZero,
		"div-mod by zero")
	core.AssertPanic(t, func() { u(1).Div(u(0)) }, num.ErrDivZero, "div by zero")
	core.AssertPanic(t, func() { u(1).Mod(u(0)) }, num.ErrDivZero, "mod by zero")
}

// uint128MulDivModTestCase exercises the wide multiply-then-divide.
type uint128MulDivModTestCase struct {
	name  string
	a     num.Uint128
	b     num.Uint128
	d     num.Uint128
	wantQ num.Uint128
}

func newUint128MulDivModTestCase(name string, a, b, d,
	wantQ num.Uint128) uint128MulDivModTestCase {
	return uint128MulDivModTestCase{name: name, a: a, b: b, d: d, wantQ: wantQ}
}

func (tc uint128MulDivModTestCase) Name() string { return tc.name }

func (tc uint128MulDivModTestCase) Test(t *testing.T) {
	t.Helper()
	q, r := tc.a.MulDivMod(tc.b, tc.d)
	core.AssertEqual(t, tc.wantQ, q, "quotient")
	// the remainder is pinned by the identity a*b == q*d + r (mod 2^128).
	core.AssertEqual(t, tc.a.Mul(tc.b), q.Mul(tc.d).Add(r), "identity")
}

func uint128MulDivModTestCases() []uint128MulDivModTestCase {
	return []uint128MulDivModTestCase{
		newUint128MulDivModTestCase("product then divide", u(6), u(7), u(5),
			u(8)),
		newUint128MulDivModTestCase("exact", u(10), u(10), u(4), u(25)),
		newUint128MulDivModTestCase("divisor larger", u(2), u(3), u(10), u(0)),
		// product exceeds 128 bits but the quotient still fits.
		newUint128MulDivModTestCase("wide product", num.MaxUint128, u(2), u(2),
			num.MaxUint128),
		// product is exactly 2^128, so the quotient wraps to zero.
		newUint128MulDivModTestCase("quotient wraps", num.NewUint128(1, 0),
			num.NewUint128(1, 0), u(1), u(0)),
		// (2^65-1)^2 = 2^130 - 2^66 + 1 makes both word-1 cross-term
		// additions carry, so word 2 receives a carry of two; dividing by
		// 2^64 surfaces that word in the quotient's high half, guarding the
		// mul256 carry propagation.
		newUint128MulDivModTestCase("double carry into word 2",
			num.NewUint128(1, maxWord), num.NewUint128(1, maxWord),
			num.NewUint128(1, 0), num.NewUint128(3, maxWord-3)),
		// (2^128-1)*(2^128-2^64+1) overflows word 2 into word 3; dividing
		// the product back by the first factor recovers the second, so a
		// dropped word-3 carry in mul256 would corrupt the quotient.
		newUint128MulDivModTestCase("carry into word 3", num.MaxUint128,
			num.NewUint128(maxWord, 1), num.MaxUint128,
			num.NewUint128(maxWord, 1)),
	}
}

func TestUint128MulDivMod(t *testing.T) {
	core.RunTestCases(t, uint128MulDivModTestCases())
}

func TestUint128MulDivModByZeroPanics(t *testing.T) {
	core.AssertPanic(t, func() { u(1).MulDivMod(u(1), u(0)) }, num.ErrDivZero,
		"mul-div-mod by zero")
}

// uint128MulTestCase exercises the low-128-bit product.
type uint128MulTestCase struct {
	name string
	a    num.Uint128
	b    num.Uint128
	want num.Uint128
}

func newUint128MulTestCase(name string, a, b,
	want num.Uint128) uint128MulTestCase {
	return uint128MulTestCase{name: name, a: a, b: b, want: want}
}

func (tc uint128MulTestCase) Name() string { return tc.name }

func (tc uint128MulTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.a.Mul(tc.b), "product")
}

func uint128MulTestCases() []uint128MulTestCase {
	return []uint128MulTestCase{
		newUint128MulTestCase("small", u(6), u(7), u(42)),
		newUint128MulTestCase("by zero", u(9), u(0), u(0)),
		newUint128MulTestCase("crosses word", u(1<<32), u(1<<32),
			num.NewUint128(1, 0)),
		newUint128MulTestCase("wraps", num.MaxUint128, u(2),
			num.NewUint128(maxWord, maxWord-1)),
	}
}

func TestUint128Mul(t *testing.T) {
	core.RunTestCases(t, uint128MulTestCases())
}

// TestUint128AddSub pins the carry chain and the wrap-around policy.
func TestUint128AddSub(t *testing.T) {
	// the carry propagates into the high word, and the borrow back out.
	core.AssertEqual(t, num.NewUint128(1, 0),
		num.NewUint128(0, maxWord).Add(u(1)), "carry")
	core.AssertEqual(t, num.NewUint128(0, maxWord),
		num.NewUint128(1, 0).Sub(u(1)), "borrow")
	// arithmetic wraps at the 128-bit boundary in both directions.
	core.AssertEqual(t, num.ZeroUint128, num.MaxUint128.Add(u(1)), "add wraps")
	core.AssertEqual(t, num.MaxUint128, num.ZeroUint128.Sub(u(1)), "sub wraps")
}

// uint128CmpTestCase exercises ordering, Equal and IsZero.
type uint128CmpTestCase struct {
	name string
	a    num.Uint128
	b    num.Uint128
	want int
}

func newUint128CmpTestCase(name string, a, b num.Uint128,
	want int) uint128CmpTestCase {
	return uint128CmpTestCase{name: name, a: a, b: b, want: want}
}

func (tc uint128CmpTestCase) Name() string { return tc.name }

func (tc uint128CmpTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.a.Cmp(tc.b), "cmp")
	core.AssertEqual(t, tc.want == 0, tc.a.Equal(tc.b), "equal")
}

func uint128CmpTestCases() []uint128CmpTestCase {
	return []uint128CmpTestCase{
		newUint128CmpTestCase("equal low", u(5), u(5), 0),
		newUint128CmpTestCase("less low", u(4), u(5), -1),
		newUint128CmpTestCase("greater low", u(6), u(5), 1),
		newUint128CmpTestCase("high dominates", num.NewUint128(1, 0),
			num.NewUint128(0, maxWord), 1),
		newUint128CmpTestCase("max vs zero", num.MaxUint128,
			num.ZeroUint128, 1),
	}
}

func TestUint128Cmp(t *testing.T) {
	core.RunTestCases(t, uint128CmpTestCases())
}

func TestUint128IsZero(t *testing.T) {
	core.AssertTrue(t, num.ZeroUint128.IsZero(), "zero")
	core.AssertFalse(t, u(1).IsZero(), "non-zero")
	core.AssertFalse(t, num.NewUint128(1, 0).IsZero(), "high only")
}
