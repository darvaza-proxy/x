package num_test

import (
	"fmt"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = pow10Case{}
	_ core.TestCase = pow10ValueCase{}
)

// pow10Case checks Pow10 at one exponent against the decimal text of
// the power.
type pow10Case struct {
	name string
	want string
	n    int
}

func newPow10Case(n int, want string) pow10Case {
	return pow10Case{
		name: fmt.Sprintf("10^%d", n),
		want: want,
		n:    n,
	}
}

func (tc pow10Case) Name() string { return tc.name }

func (tc pow10Case) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, num.Pow10(tc.n).String(), "Pow10(%d)", tc.n)
}

// pow10Cases returns a row for every power of ten a Uint128 holds,
// each expecting a one followed by n zeros.
func pow10Cases() []pow10Case {
	cases := make([]pow10Case, 0, 39)
	for n := range 39 {
		cases = append(cases, newPow10Case(n, "1"+strings.Repeat("0", n)))
	}
	return cases
}

// pow10ValueCase checks Pow10 at one exponent: the power as a Uint128,
// or for an exponent outside the table, a panic with wantErr, its
// stack starting at callPow10.
type pow10ValueCase struct {
	wantErr error
	name    string
	want    num.Uint128
	n       int
}

func newPow10ValueCase(n int, want num.Uint128) pow10ValueCase {
	return pow10ValueCase{
		name:    fmt.Sprintf("10^%d", n),
		wantErr: nil,
		want:    want,
		n:       n,
	}
}

func newPow10PanicCase(name string, n int, wantErr error) pow10ValueCase {
	return pow10ValueCase{
		name:    name,
		wantErr: wantErr,
		want:    num.ZeroUint128,
		n:       n,
	}
}

func (tc pow10ValueCase) Name() string { return tc.name }

func (tc pow10ValueCase) Test(t *testing.T) {
	t.Helper()
	if tc.wantErr != nil {
		tc.testPanic(t)
		return
	}
	core.AssertEqual(t, tc.want, num.Pow10(tc.n), "Pow10(%d)", tc.n)
}

func (tc pow10ValueCase) testPanic(t *testing.T) {
	t.Helper()
	err := core.Catch(func() error {
		callPow10(tc.n)
		return nil
	})
	core.AssertErrorIs(t, err, tc.wantErr, "Pow10(%d)", tc.n)
	core.AssertErrorIs(t, err, core.ErrInvalid, "Pow10(%d) invalid", tc.n)
	core.AssertTopFrame(t, err, "callPow10", "Pow10(%d) stack", tc.n)
}

// pow10ValueCases returns a row for every power of ten a Uint128
// holds, each expecting the power multiplied up by ten at each step
// from one, and a row on either side of the table expecting a panic.
func pow10ValueCases() []pow10ValueCase {
	ten := num.AsUint128(10)
	want := num.AsUint128(1)
	cases := make([]pow10ValueCase, 0, 41)
	for n := range 39 {
		cases = append(cases, newPow10ValueCase(n, want))
		want = want.Mul(ten)
	}
	return append(cases,
		newPow10PanicCase("below range", -1, num.ErrPow10Range),
		newPow10PanicCase("above range", 39, num.ErrPow10Range),
	)
}

// TestPow10 checks every power of ten Pow10 holds, against its decimal
// text and against the power multiplied up by ten, and that an
// exponent outside 0 to 38 panics with ErrPow10Range from the caller.
func TestPow10(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		core.RunTestCases(t, pow10Cases())
	})
	t.Run("value", func(t *testing.T) {
		core.RunTestCases(t, pow10ValueCases())
	})
}

// callPow10 calls Pow10, a named caller for the panic stack to start
// at.
func callPow10(n int) {
	num.Pow10(n)
}
