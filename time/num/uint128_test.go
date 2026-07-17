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
