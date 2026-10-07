package num

import (
	"fmt"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = pow10ExpCase{}

// pow10ExpCase checks pow10Exp on one power of ten against its
// exponent.
type pow10ExpCase struct {
	name string
	p    Uint128
	want int
}

func newPow10ExpCase(p Uint128, want int) pow10ExpCase {
	return pow10ExpCase{
		name: fmt.Sprintf("10^%d", want),
		p:    p,
		want: want,
	}
}

func (tc pow10ExpCase) Name() string { return tc.name }

func (tc pow10ExpCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, pow10Exp(tc.p), "pow10Exp")
}

// pow10ExpCases returns a row for every power of ten in pow10Table,
// each expecting its exponent.
func pow10ExpCases() []pow10ExpCase {
	cases := make([]pow10ExpCase, 0, len(pow10Table))
	for n := range len(pow10Table) {
		cases = append(cases, newPow10ExpCase(pow10Table[n], n))
	}
	return cases
}

// TestPow10Exp checks pow10Exp against the exponent of every power of
// ten in pow10Table.
func TestPow10Exp(t *testing.T) {
	core.RunTestCases(t, pow10ExpCases())
}

// TestPow10Bound checks that each entry of pow10Bound is the largest
// magnitude whose product by its power of ten stays within MaxInt128:
// the product fits, and the room left is less than one more step.
func TestPow10Bound(t *testing.T) {
	limit := MaxInt128.bits()
	for k, p := range pow10Table {
		prod := pow10Bound[k].Mul(p)
		core.AssertTrue(t, prod.Cmp(limit) <= 0, "10^%d fits", k)
		core.AssertTrue(t, limit.Sub(prod).Cmp(p) < 0, "10^%d largest", k)
	}
}
