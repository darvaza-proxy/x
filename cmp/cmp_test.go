package cmp_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/cmp"
)

// TestCase interface validations
var _ core.TestCase = cmpTestCase[int]{}
var _ core.TestCase = cmpTestCase[string]{}

// cmpTestCase is a generic test case for comparison functions
type cmpTestCase[T any] struct {
	a, b     T
	fn       func(a, b T) bool
	name     string
	fmt      string // format string for error messages
	expected bool
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
		newCmpTestCase("equal values", 5, 5, true, cmp.Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("different values", 5, 10, false, cmp.Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("negative values equal", -5, -5, true, cmp.Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("negative values different", -5, -10, false, cmp.Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("zero and non-zero", 0, 5, false, cmp.Eq[int], "Eq(%d, %d)"),
		newCmpTestCase("zero and zero", 0, 0, true, cmp.Eq[int], "Eq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestEqWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("equal strings", "hello", "hello", true, cmp.Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("different strings", "hello", "world", false, cmp.Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("empty strings", "", "", true, cmp.Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("empty and non-empty", "", "hello", false, cmp.Eq[string], "Eq(%q, %q)"),
		newCmpTestCase("case sensitivity", "Hello", "hello", false, cmp.Eq[string], "Eq(%q, %q)"),
	}

	core.RunTestCases(t, tests)
}

// checkFn is a comparison function taking a custom comparator
type checkFn[T any] func(a, b T, fn cmp.CompFunc[T]) bool

// cmpFnTestCase is a generic test case for comparison functions with custom comparator
type cmpFnTestCase[T any] struct {
	a, b     T
	check    checkFn[T]
	cmp      cmp.CompFunc[T]
	name     string
	fmt      string
	expected bool
}

var _ core.TestCase = cmpFnTestCase[int]{}

//revive:disable-next-line:argument-limit
func newCmpFnTestCase[T any](name string, a, b T, expected bool,
	fn cmp.CompFunc[T], check checkFn[T], fmt string) cmpFnTestCase[T] {
	return cmpFnTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cmp:      fn,
		check:    check,
		fmt:      fmt,
	}
}

func (tc cmpFnTestCase[T]) Name() string {
	return tc.name
}

func (tc cmpFnTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.check(tc.a, tc.b, tc.cmp)
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
	fn := func(a, b int) int {
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
		newCmpFnTestCase("equal values", 5, 5, true, fn, cmp.EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("different values", 5, 10, false, fn, cmp.EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("negative values equal", -5, -5, true, fn, cmp.EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("negative values different", -5, -10, false, fn, cmp.EqFn[int], "EqFn(%d, %d)"),
		newCmpFnTestCase("zero and non-zero", 0, 5, false, fn, cmp.EqFn[int], "EqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestEqFnCustomStruct(t *testing.T) {
	t.Helper()
	type score struct {
		value float64
	}

	fn := func(a, b score) int {
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

	core.AssertTrue(t, cmp.EqFn(s1, s3, fn), "equal scores")
	core.AssertFalse(t, cmp.EqFn(s1, s2, fn), "different scores")
}

func runTestEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Equality function that considers numbers equal if they have the same parity
	sameParity := func(a, b int) bool {
		return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
	}

	tests := []cmpFn2TestCase[int]{
		newCmpFn2TestCase("both even", 4, 6, true, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("both odd", 5, 7, true, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("even and odd", 4, 7, false, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("odd and even", 5, 8, false, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("zero and even", 0, 2, true, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
		newCmpFn2TestCase("negative numbers", -3, -5, true, sameParity, cmp.EqFn2[int], "EqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestNotEqFnWithIntegers(t *testing.T) {
	t.Helper()
	fn := func(a, b int) int {
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
		newCmpFnTestCase("equal values", 5, 5, false, fn, cmp.NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("different values", 5, 10, true, fn, cmp.NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("negative values equal", -5, -5, false, fn, cmp.NotEqFn[int], "NotEqFn(%d, %d)"),
		newCmpFnTestCase("negative values different", -5, -10, true, fn, cmp.NotEqFn[int], "NotEqFn(%d, %d)"),
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
		newCmpFn2TestCase("both even", 4, 6, false, sameParity, cmp.NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("both odd", 5, 7, false, sameParity, cmp.NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("even and odd", 4, 7, true, sameParity, cmp.NotEqFn2[int], "NotEqFn2(%d, %d)"),
		newCmpFn2TestCase("odd and even", 5, 8, true, sameParity, cmp.NotEqFn2[int], "NotEqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtFnWithIntegers(t *testing.T) {
	t.Helper()
	fn := func(a, b int) int {
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
		newCmpFnTestCase("less", 5, 10, true, fn, cmp.LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("greater", 10, 5, false, fn, cmp.LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, false, fn, cmp.LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("negative numbers", -5, -3, true, fn, cmp.LtFn[int], "LtFn(%d, %d)"),
		newCmpFnTestCase("mixed signs", -1, 1, true, fn, cmp.LtFn[int], "LtFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestEqFnPanic verifies that EqFn panics when given a nil comparison
// function.
func TestEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.EqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "EqFn nil cmp")
}

// cmpFn2TestCase is a generic test case for comparison functions with condition function
type cmpFn2TestCase[T any] struct {
	a, b     T
	fn       func(a, b T, cond cmp.CondFunc[T]) bool
	cond     cmp.CondFunc[T]
	name     string
	fmt      string
	expected bool
}

var _ core.TestCase = cmpFn2TestCase[int]{}

//revive:disable-next-line:argument-limit
func newCmpFn2TestCase[T any](name string, a, b T, expected bool,
	cond cmp.CondFunc[T], fn func(a, b T, cond cmp.CondFunc[T]) bool, fmt string) cmpFn2TestCase[T] {
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

	p1 := person{nameAlice, 30}
	p2 := person{nameBob, 30}
	p3 := person{nameCharlie, 25}

	core.AssertTrue(t, cmp.EqFn2(p1, p2, sameAge), "same age")
	core.AssertFalse(t, cmp.EqFn2(p1, p3, sameAge), "different age")
}

// TestEqFn2Panic verifies that EqFn2 panics when given a nil condition
// function.
func TestEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.EqFn2(1, 2, nil)
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
		newCmpTestCase("equal values", 5, 5, false, cmp.NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("different values", 5, 10, true, cmp.NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("zero and non-zero", 0, 5, true, cmp.NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("zero and zero", 0, 0, false, cmp.NotEq[int], "NotEq(%d, %d)"),
		newCmpTestCase("negative values", -5, -10, true, cmp.NotEq[int], "NotEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestNotEqWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("equal strings", "hello", "hello", false, cmp.NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("different strings", "hello", "world", true, cmp.NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("empty strings", "", "", false, cmp.NotEq[string], "NotEq(%q, %q)"),
		newCmpTestCase("empty and non-empty", "", "hello", true, cmp.NotEq[string], "NotEq(%q, %q)"),
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
		cmp.NotEqFn(1, 2, nil)
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
		cmp.NotEqFn2(1, 2, nil)
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
		newCmpTestCase("less", 5, 10, true, cmp.Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("greater", 10, 5, false, cmp.Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("equal", 5, 5, false, cmp.Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, true, cmp.Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("negative values", -10, -5, true, cmp.Lt[int], "Lt(%d, %d)"),
		newCmpTestCase("with zero", 0, 5, true, cmp.Lt[int], "Lt(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtWithStrings(t *testing.T) {
	t.Helper()
	tests := []cmpTestCase[string]{
		newCmpTestCase("lexicographically less", "apple", "banana", true, cmp.Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("lexicographically greater", "zebra", "apple", false, cmp.Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("equal strings", "apple", "apple", false, cmp.Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("empty string", "", "a", true, cmp.Lt[string], "Lt(%q, %q)"),
		newCmpTestCase("case sensitivity", "Z", "a", true, cmp.Lt[string], "Lt(%q, %q)"), // ASCII 'Z' comes before 'a'
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
		cmp.LtFn(1, 2, nil)
	}, expectedNilCompFuncErr, "LtFn nil cmp")
}

// TestGt verifies the Gt function correctly determines "greater than"
// relationships for ordered types.
func TestGt(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("greater", 10, 5, true, cmp.Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("less", 5, 10, false, cmp.Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("equal", 5, 5, false, cmp.Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, false, cmp.Gt[int], "Gt(%d, %d)"),
		newCmpTestCase("negative values", -5, -10, true, cmp.Gt[int], "Gt(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtFn verifies that GtFn correctly determines "greater than"
// relationships using a custom comparison function.
func TestGtFn(t *testing.T) {
	fn := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("greater", 10, 5, true, fn, cmp.GtFn[int], "GtFn(%d, %d)"),
		newCmpFnTestCase("less", 5, 10, false, fn, cmp.GtFn[int], "GtFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, false, fn, cmp.GtFn[int], "GtFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtFnPanic verifies that GtFn panics when given a nil comparison
// function.
func TestGtFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.GtFn(1, 2, nil)
	}, expectedNilCompFuncErr, "GtFn nil cmp")
}

// TestGtEq verifies the GtEq function correctly determines "greater than
// or equal to" relationships for ordered types.
func TestGtEq(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("greater", 10, 5, true, cmp.GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("less", 5, 10, false, cmp.GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("equal", 5, 5, true, cmp.GtEq[int], "GtEq(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, false, cmp.GtEq[int], "GtEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtEqFn verifies that GtEqFn correctly determines "greater than or
// equal to" relationships using a custom comparison function.
func TestGtEqFn(t *testing.T) {
	fn := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("greater", 10, 5, true, fn, cmp.GtEqFn[int], "GtEqFn(%d, %d)"),
		newCmpFnTestCase("less", 5, 10, false, fn, cmp.GtEqFn[int], "GtEqFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, true, fn, cmp.GtEqFn[int], "GtEqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestGtEqFnPanic verifies that GtEqFn panics when given a nil comparison
// function.
func TestGtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.GtEqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "GtEqFn nil cmp")
}

// TestLtEq verifies the LtEq function correctly determines "less than or
// equal to" relationships for ordered types.
func TestLtEq(t *testing.T) {
	tests := []cmpTestCase[int]{
		newCmpTestCase("less", 5, 10, true, cmp.LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("greater", 10, 5, false, cmp.LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("equal", 5, 5, true, cmp.LtEq[int], "LtEq(%d, %d)"),
		newCmpTestCase("negative and positive", -5, 5, true, cmp.LtEq[int], "LtEq(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestLtEqFn verifies that LtEqFn correctly determines "less than or
// equal to" relationships using a custom comparison function.
func TestLtEqFn(t *testing.T) {
	fn := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tests := []cmpFnTestCase[int]{
		newCmpFnTestCase("less", 5, 10, true, fn, cmp.LtEqFn[int], "LtEqFn(%d, %d)"),
		newCmpFnTestCase("greater", 10, 5, false, fn, cmp.LtEqFn[int], "LtEqFn(%d, %d)"),
		newCmpFnTestCase("equal", 5, 5, true, fn, cmp.LtEqFn[int], "LtEqFn(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

// TestLtEqFnPanic verifies that LtEqFn panics when given a nil comparison
// function.
func TestLtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.LtEqFn(1, 2, nil)
	}, expectedNilCompFuncErr, "LtEqFn nil cmp")
}

// TestCustomTypeComparison verifies that comparison operations work correctly
// with custom types using appropriate comparison functions.
type customType struct {
	value int
}

func TestCustomTypeComparison(t *testing.T) {
	fn := func(a, b customType) int {
		return a.value - b.value
	}

	a := customType{value: 5}
	b := customType{value: 10}
	c := customType{value: 5}

	core.AssertTrue(t, cmp.EqFn(a, c, fn), "EqFn(a, c)")
	core.AssertFalse(t, cmp.EqFn(a, b, fn), "EqFn(a, b)")
	core.AssertTrue(t, cmp.NotEqFn(a, b, fn), "NotEqFn(a, b)")
	core.AssertTrue(t, cmp.LtFn(a, b, fn), "LtFn(a, b)")
	core.AssertTrue(t, cmp.LtEqFn(a, b, fn), "LtEqFn(a, b)")
	core.AssertTrue(t, cmp.LtEqFn(a, c, fn), "LtEqFn(a, c)")
	core.AssertFalse(t, cmp.GtFn(a, b, fn), "GtFn(a, b)")
	core.AssertTrue(t, cmp.GtFn(b, a, fn), "GtFn(b, a)")
	core.AssertTrue(t, cmp.GtEqFn(a, c, fn), "GtEqFn(a, c)")
}

// ltEqFn2TestCase is a test case for LtEqFn2 function
type ltEqFn2TestCase[T any] struct {
	a, b     T
	less     cmp.CondFunc[T]
	name     string
	fmt      string
	expected bool
}

var _ core.TestCase = ltEqFn2TestCase[int]{}

//revive:disable-next-line:argument-limit
func newLtEqFn2TestCase[T any](name string, a, b T, expected bool,
	less cmp.CondFunc[T], fmt string) ltEqFn2TestCase[T] {
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
	result := cmp.LtEqFn2(tc.a, tc.b, tc.less)
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
		cmp.LtEqFn2(1, 2, nil)
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
		newCmpFn2TestCase("a less than b", 3, 5, true, less, cmp.LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("a equal to b", 5, 5, false, less, cmp.LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("a greater than b", 7, 5, false, less, cmp.LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("negative values", -10, -5, true, less, cmp.LtFn2[int], "LtFn2(%d, %d)"),
		newCmpFn2TestCase("zero comparison", 0, 1, true, less, cmp.LtFn2[int], "LtFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestLtFn2Panic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		cmp.LtFn2(1, 2, nil)
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
		newCmpFn2TestCase("a greater than b", 7, 5, true, less, cmp.GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("a equal to b", 5, 5, true, less, cmp.GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("a less than b", 3, 5, false, less, cmp.GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("negative values", -5, -10, true, less, cmp.GtEqFn2[int], "GtEqFn2(%d, %d)"),
		newCmpFn2TestCase("zero comparison", 1, 0, true, less, cmp.GtEqFn2[int], "GtEqFn2(%d, %d)"),
	}

	core.RunTestCases(t, tests)
}

func runTestGtEqFn2Panic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		cmp.GtEqFn2(1, 2, nil)
	}, expectedNilCondFuncErr, "GtEqFn2 nil less")
}
