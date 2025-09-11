package cmp

import (
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var _ core.TestCase = cmpTestCase[int]{}
var _ core.TestCase = cmpTestCase[string]{}

// cmpTestCase is a generic test case for comparison functions
type cmpTestCase[T any] struct {
	name     string
	a, b     T
	expected bool
	fn       func(a, b T) bool
	fmt      string // format string for error messages
}

//revive:disable-next-line:argument-limit
func newCmpTestCase[T any](name string, a, b T, expected bool, fn func(a, b T) bool, fmt string) cmpTestCase[T] {
	return cmpTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		fn:       fn,
		fmt:      fmt,
	}
}

func (tc cmpTestCase[T]) Name() string {
	return tc.name
}

func (tc cmpTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.fn(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, tc.fmt, tc.a, tc.b)
}

// TestEq verifies that the Eq function correctly determines equality
// for various comparable types including integers and strings.
func TestEq(t *testing.T) {
	t.Run("with integers", runTestEqWithIntegers)
	t.Run("with strings", runTestEqWithStrings)
}

func runTestEqWithIntegers(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[int]{
		newCmpTestCase("equal values", 5, 5, true, Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("different values", 5, 10, false, Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("negative values equal", -5, -5, true, Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("negative values different", -5, -10, false, Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("zero and non-zero", 0, 5, false, Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("zero and zero", 0, 0, true, Eq[int], "Eq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestEqWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("equal strings", "hello", "hello", true, Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("different strings", "hello", "world", false, Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("empty strings", "", "", true, Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("empty and non-empty", "", "hello", false, Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("case sensitivity", "Hello", "hello", false, Eq[string], "Eq(%q, %q)"),
	}

	core.RunTestCases(t, tests)
}

// cmpFnTestCase is a generic test case for comparison functions with custom comparator
type cmpFnTestCase[T any] struct {
	name     string
	a, b     T
	expected bool
	cmp      CompFunc[T]
	fn       func(a, b T, cmp CompFunc[T]) bool
	fmt      string
}

var _ core.TestCase = cmpFnTestCase[int]{}

//revive:disable-next-line:argument-limit
func newCmpFnTestCase[T any](name string, a, b T, expected bool,
	cmp CompFunc[T], fn func(a, b T, cmp CompFunc[T]) bool, fmt string) cmpFnTestCase[T] {
	return cmpFnTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cmp:      cmp,
		fn:       fn,
		fmt:      fmt,
	}
}

func (tc cmpFnTestCase[T]) Name() string {
	return tc.name
}

func (tc cmpFnTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.fn(tc.a, tc.b, tc.cmp)
	core.AssertEqual(t, tc.expected, result, tc.fmt, tc.a, tc.b)
}

// TestEqFn verifies that EqFn correctly determines equality using
// a custom comparison function for different data types.
func TestEqFn(t *testing.T) {
	t.Run("with integers", runTestEqFnWithIntegers)
	t.Run("with custom struct", runTestEqFnCustomStruct)
}

func runTestEqFnWithIntegers(t *testing.T) {
	t.Helper()
	cmp := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("equal values", 5, 5, true, cmp, EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("different values", 5, 10, false, cmp, EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("negative values equal", -5, -5, true, cmp, EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("negative values different", -5, -10, false, cmp, EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("zero and non-zero", 0, 5, false, cmp, EqFn[int], "EqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestEqFnCustomStruct(t *testing.T) {
	t.Helper()
	type score struct {
		value float64
	}

	cmp := func(a, b score) int {
		switch {
		case a.value < b.value:
			return -1
		case a.value > b.value:
			return 1
		default:
			return 0
		}
	}

	s1 := score{3.14}
	s2 := score{2.71}
	s3 := score{3.14}

	core.AssertTrue(t, EqFn(s1, s3, cmp), "equal scores")
	core.AssertFalse(t, EqFn(s1, s2, cmp), "different scores")
}

func runTestEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Equality function that considers numbers equal if they have the same parity
	sameParity := func(a, b int) bool {
		return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
	}

	tests := []cmpFn2TestCase[int]{
		newCmpFn2TestCase("both even", 4, 6, true, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("both odd", 5, 7, true, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("even and odd", 4, 7, false, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("odd and even", 5, 8, false, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("zero and even", 0, 2, true, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("negative numbers", -3, -5, true, sameParity, EqFn2[int], "EqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestNotEqFnWithIntegers(t *testing.T) {
	t.Helper()
	cmp := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("equal values", 5, 5, false, cmp, NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("different values", 5, 10, true, cmp, NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("negative values equal", -5, -5, false, cmp, NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("negative values different", -5, -10, true, cmp, NotEqFn[int], "NotEqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestNotEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Equality function that considers numbers equal if they have the same parity
	sameParity := func(a, b int) bool {
		return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
	}

	tests := []cmpFn2TestCase[int]{
		newCmpFn2TestCase("both even", 4, 6, false, sameParity, NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("both odd", 5, 7, false, sameParity, NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("even and odd", 4, 7, true, sameParity, NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("odd and even", 5, 8, true, sameParity, NotEqFn2[int], "NotEqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtFnWithIntegers(t *testing.T) {
	t.Helper()
	cmp := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("less", 5, 10, true, cmp, LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("greater", 10, 5, false, cmp, LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, false, cmp, LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("negative numbers", -5, -3, true, cmp, LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("mixed signs", -1, 1, true, cmp, LtFn[int], "LtFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestEqFnPanic verifies that EqFn panics when given a nil comparison
// function.
func TestEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		EqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "EqFn nil cmp")
}

// cmpFn2TestCase is a generic test case for comparison functions with condition function
type cmpFn2TestCase[T any] struct {
	name     string
	a, b     T
	expected bool
	cond     CondFunc[T]
	fn       func(a, b T, cond CondFunc[T]) bool
	fmt      string
}

var _ core.TestCase = cmpFn2TestCase[int]{}

//revive:disable-next-line:argument-limit
func newCmpFn2TestCase[T any](name string, a, b T, expected bool,
	cond CondFunc[T], fn func(a, b T, cond CondFunc[T]) bool, fmt string) cmpFn2TestCase[T] {
	return cmpFn2TestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cond:     cond,
		fn:       fn,
		fmt:      fmt,
	}
}

func (tc cmpFn2TestCase[T]) Name() string {
	return tc.name
}

func (tc cmpFn2TestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.fn(tc.a, tc.b, tc.cond)
	core.AssertEqual(t, tc.expected, result, tc.fmt, tc.a, tc.b)
}

// TestEqFn2 verifies that EqFn2 correctly determines equality using
// a custom equality condition function.
func TestEqFn2(t *testing.T) {
	t.Run("with integers", runTestEqFn2WithIntegers)
	t.Run("with custom struct", runTestEqFn2CustomStruct)
}

func runTestEqFn2CustomStruct(t *testing.T) {
	t.Helper()
	type person struct {
		name string
		age  int
	}

	// Equality function that considers people equal if they have the same age
	sameAge := func(a, b person) bool {
		return a.age == b.age
	}

	p1 := person{"Alice", 30}
	p2 := person{"Bob", 30}
	p3 := person{"Charlie", 25}

	core.AssertTrue(t, EqFn2(p1, p2, sameAge), "same age")
	core.AssertFalse(t, EqFn2(p1, p3, sameAge), "different age")
}

// TestEqFn2Panic verifies that EqFn2 panics when given a nil condition
// function.
func TestEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		EqFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "EqFn2 nil cond")
}

// TestNotEq verifies the NotEq function correctly determines inequality
// for various comparable types.
func TestNotEq(t *testing.T) {
	t.Run("with integers", runTestNotEqWithIntegers)
	t.Run("with strings", runTestNotEqWithStrings)
}

func runTestNotEqWithIntegers(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[int]{
		newCmpTestCase("equal values", 5, 5, false, NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("different values", 5, 10, true, NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("zero and non-zero", 0, 5, true, NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("zero and zero", 0, 0, false, NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("negative values", -5, -10, true, NotEq[int], "NotEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestNotEqWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("equal strings", "hello", "hello", false, NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("different strings", "hello", "world", true, NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("empty strings", "", "", false, NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("empty and non-empty", "", "hello", true, NotEq[string], "NotEq(%q, %q)"),
	}

	core.RunTestCases(t, tests)
}

// TestNotEqFn verifies that NotEqFn correctly determines inequality
// using a custom comparison function.
func TestNotEqFn(t *testing.T) {
	t.Run("with integers", runTestNotEqFnWithIntegers)
}

// TestNotEqFnPanic verifies that NotEqFn panics when given a nil
// comparison function.
func TestNotEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		NotEqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "NotEqFn nil cmp")
}

// TestNotEqFn2 verifies that NotEqFn2 correctly negates the result
// of a custom equality condition function.
func TestNotEqFn2(t *testing.T) {
	t.Run("with integers", runTestNotEqFn2WithIntegers)
}

// TestNotEqFn2Panic verifies that NotEqFn2 panics when given a nil
// condition function.
func TestNotEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		NotEqFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "NotEqFn2 nil cond")
}

// TestLt verifies the Lt function correctly determines "less than"
// relationships for ordered types.
func TestLt(t *testing.T) {
	t.Run("with integers", runTestLtWithIntegers)
	t.Run("with strings", runTestLtWithStrings)
}

func runTestLtWithIntegers(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[int]{
		newCmpTestCase("less", 5, 10, true, Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("greater", 10, 5, false, Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("equal", 5, 5, false, Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, true, Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("negative values", -10, -5, true, Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("with zero", 0, 5, true, Lt[int], "Lt(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("lexicographically less", "apple", "banana", true, Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("lexicographically greater", "zebra", "apple", false, Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("equal strings", "apple", "apple", false, Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("empty string", "", "a", true, Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("case sensitivity", "Z", "a", true, Lt[string], "Lt(%q, %q)"), // ASCII 'Z' comes before 'a'
	}

	core.RunTestCases(t, tests)
}

// TestLtFn verifies that LtFn correctly determines "less than"
// relationships using a custom comparison function.
func TestLtFn(t *testing.T) {
	t.Run("with integers", runTestLtFnWithIntegers)
}

// TestLtFnPanic verifies that LtFn panics when given a nil comparison
// function.
func TestLtFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		LtFn(1, 2, nil)
	}, expectedNilCompFuncErr, "LtFn nil cmp")
}

// TestGt verifies the Gt function correctly determines "greater than"
// relationships for ordered types.
func TestGt(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("greater", 10, 5, true, Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("less", 5, 10, false, Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("equal", 5, 5, false, Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, false, Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("negative values", -5, -10, true, Gt[int], "Gt(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtFn verifies that GtFn correctly determines "greater than"
// relationships using a custom comparison function.
func TestGtFn(t *testing.T) {
	cmp := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("greater", 10, 5, true, cmp, GtFn[int], "GtFn(%d, %d)"),
		newCmpFnTestCase("less", 5, 10, false, cmp, GtFn[int], "GtFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, false, cmp, GtFn[int], "GtFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtFnPanic verifies that GtFn panics when given a nil comparison
// function.
func TestGtFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		GtFn(1, 2, nil)
	}, expectedNilCompFuncErr, "GtFn nil cmp")
}

// TestGtEq verifies the GtEq function correctly determines "greater than
// or equal to" relationships for ordered types.
func TestGtEq(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("greater", 10, 5, true, GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("less", 5, 10, false, GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("equal", 5, 5, true, GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, false, GtEq[int], "GtEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtEqFn verifies that GtEqFn correctly determines "greater than or
// equal to" relationships using a custom comparison function.
func TestGtEqFn(t *testing.T) {
	cmp := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("greater", 10, 5, true, cmp, GtEqFn[int], "GtEqFn(%d, %d)"),
		newCmpFnTestCase("less", 5, 10, false, cmp, GtEqFn[int], "GtEqFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, true, cmp, GtEqFn[int], "GtEqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtEqFnPanic verifies that GtEqFn panics when given a nil comparison
// function.
func TestGtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		GtEqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "GtEqFn nil cmp")
}

// TestLtEq verifies the LtEq function correctly determines "less than or
// equal to" relationships for ordered types.
func TestLtEq(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("less", 5, 10, true, LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("greater", 10, 5, false, LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("equal", 5, 5, true, LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, true, LtEq[int], "LtEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestLtEqFn verifies that LtEqFn correctly determines "less than or
// equal to" relationships using a custom comparison function.
func TestLtEqFn(t *testing.T) {
	cmp := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("less", 5, 10, true, cmp, LtEqFn[int], "LtEqFn(%d, %d)"),
		newCmpFnTestCase("greater", 10, 5, false, cmp, LtEqFn[int], "LtEqFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, true, cmp, LtEqFn[int], "LtEqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestLtEqFnPanic verifies that LtEqFn panics when given a nil comparison
// function.
func TestLtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		LtEqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "LtEqFn nil cmp")
}

// TestCustomTypeComparison verifies that comparison operations work correctly
// with custom types using appropriate comparison functions.
type customType struct {
	value int
}

func TestCustomTypeComparison(t *testing.T) {
	cmp := func(a, b customType) int {
		return a.value - b.value
	}

	a := customType{value: 5}
	b := customType{value: 10}
	c := customType{value: 5}

	core.AssertTrue(t, EqFn(a, c, cmp), "EqFn(a, c)")
	core.AssertFalse(t, EqFn(a, b, cmp), "EqFn(a, b)")
	core.AssertTrue(t, NotEqFn(a, b, cmp), "NotEqFn(a, b)")
	core.AssertTrue(t, LtFn(a, b, cmp), "LtFn(a, b)")
	core.AssertTrue(t, LtEqFn(a, b, cmp), "LtEqFn(a, b)")
	core.AssertTrue(t, LtEqFn(a, c, cmp), "LtEqFn(a, c)")
	core.AssertFalse(t, GtFn(a, b, cmp), "GtFn(a, b)")
	core.AssertTrue(t, GtFn(b, a, cmp), "GtFn(b, a)")
	core.AssertTrue(t, GtEqFn(a, c, cmp), "GtEqFn(a, c)")
}

// ltEqFn2TestCase is a test case for LtEqFn2 function
type ltEqFn2TestCase[T any] struct {
	name     string
	a, b     T
	expected bool
	less     CondFunc[T]
	fmt      string
}

var _ core.TestCase = ltEqFn2TestCase[int]{}

//revive:disable-next-line:argument-limit
func newLtEqFn2TestCase[T any](name string, a, b T, expected bool, less CondFunc[T], fmt string) ltEqFn2TestCase[T] {
	return ltEqFn2TestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		less:     less,
		fmt:      fmt,
	}
}

func (tc ltEqFn2TestCase[T]) Name() string {
	return tc.name
}

func (tc ltEqFn2TestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := LtEqFn2(tc.a, tc.b, tc.less)
	core.AssertEqual(t, tc.expected, result, tc.fmt, tc.a, tc.b)
}

// TestLtEqFn2 verifies that LtEqFn2 correctly determines "less than or
// equal to" relationships using direct less-than comparison functions.
func TestLtEqFn2(t *testing.T) {
	less := func(a, b int) bool {
		return a < b
	}

	tests := []ltEqFn2TestCase[int]{
		newLtEqFn2TestCase("less with positive numbers", 3, 5, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("equal with positive numbers", 5, 5, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("greater with positive numbers", 7, 5, false, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("less with negative numbers", -7, -5, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("equal with negative numbers", -5, -5, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("greater with negative numbers", -3, -5, false, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("less with mixed signs", -5, 3, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("comparing with zero", 0, 1, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("zero equality", 0, 0, true, less, "LtEqFn2(%d, %d)"),
		newLtEqFn2TestCase("large number comparison", 1000000, 1000001, true, less, "LtEqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)

	// Test with custom type
	t.Run("with temperature", runTestLtEqFn2Temperature)
}

func runTestLtEqFn2Temperature(t *testing.T) {
	t.Helper()
	type temperature struct {
		celsius float64
	}
	tempLess := func(a, b temperature) bool {
		return a.celsius < b.celsius
	}

	tempTests := []ltEqFn2TestCase[temperature]{
		newLtEqFn2TestCase("freezing point comparison", temperature{0}, temperature{0},
			true, tempLess, "LtEqFn2(%.1f, %.1f)"),
		newLtEqFn2TestCase("below freezing", temperature{-5.5}, temperature{-2.2},
			true, tempLess, "LtEqFn2(%.1f, %.1f)"),
		newLtEqFn2TestCase("above freezing", temperature{25.5}, temperature{30.2},
			true, tempLess, "LtEqFn2(%.1f, %.1f)"),
		newLtEqFn2TestCase("high temperature", temperature{100}, temperature{90},
			false, tempLess, "LtEqFn2(%.1f, %.1f)"),
	}

	core.RunTestCases(t, tempTests)
}

// TestLtEqFn2Panic verifies that LtEqFn2 panics when given a nil
// condition function.
func TestLtEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		LtEqFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "LtEqFn2 nil less")
}

// TestLtFn2 verifies that LtFn2 correctly determines "less than"
// relationships using a custom less-than condition function.
func TestLtFn2(t *testing.T) {
	t.Run("with integers", runTestLtFn2WithIntegers)
	t.Run("panic on nil", runTestLtFn2Panic)
}

func runTestLtFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Custom less-than function
	less := func(a, b int) bool {
		return a < b
	}

	tests := []cmpFn2TestCase[int]{
		newCmpFn2TestCase("a less than b", 3, 5, true, less, LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("a equal to b", 5, 5, false, less, LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("a greater than b", 7, 5, false, less, LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("negative values", -10, -5, true, less, LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("zero comparison", 0, 1, true, less, LtFn2[int], "LtFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtFn2Panic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		LtFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "LtFn2 nil less")
}

// TestGtEqFn2 verifies that GtEqFn2 correctly determines "greater than or
// equal" relationships using a custom less-than condition function.
func TestGtEqFn2(t *testing.T) {
	t.Run("with integers", runTestGtEqFn2WithIntegers)
	t.Run("panic on nil", runTestGtEqFn2Panic)
}

func runTestGtEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Custom less-than function
	less := func(a, b int) bool {
		return a < b
	}

	tests := []cmpFn2TestCase[int]{
		newCmpFn2TestCase("a greater than b", 7, 5, true, less, GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("a equal to b", 5, 5, true, less, GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("a less than b", 3, 5, false, less, GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("negative values", -5, -10, true, less, GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("zero comparison", 1, 0, true, less, GtEqFn2[int], "GtEqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestGtEqFn2Panic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		GtEqFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "GtEqFn2 nil less")
}
