package cmp

import (
	"testing"

	"darvaza.org/core"
)

// TestMatchEq verifies that the MatchEq function correctly creates matchers
// that check for equality with a given value.
func TestMatchEq(t *testing.T) {
	t.Run("with integers", runTestMatchEqWithIntegers)
	t.Run("with strings", runTestMatchEqWithStrings)
}

func runTestMatchEqWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	matchFive := MatchEq(5)
	matchZero := MatchEq(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("equal to 5", 5, true, matchFive),
		newMatchTestCase("not equal to 5", 10, false, matchFive),
		newMatchTestCase("equal to 0", 0, true, matchZero),
		newMatchTestCase("not equal to 0", 1, false, matchZero),
		newMatchTestCase("negative value", -5, false, matchFive),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchEqWithStrings(t *testing.T) {
	t.Helper()
	// Create matchers
	matchHello := MatchEq("hello")
	matchEmpty := MatchEq("")

	testCases := []matchTestCase[string]{
		newMatchTestCase("equal strings", "hello", true, matchHello),
		newMatchTestCase("different strings", "world", false, matchHello),
		newMatchTestCase("empty string match", "", true, matchEmpty),
		newMatchTestCase("non-empty vs empty", "hello", false, matchEmpty),
		newMatchTestCase("case sensitivity", "Hello", false, matchHello),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchEqFnWithIntegers(t *testing.T) {
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

	// Create matchers
	matchFive := MatchEqFn(5, cmp)
	matchZero := MatchEqFn(0, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("equal to 5", 5, true, matchFive),
		newMatchTestCase("not equal to 5", 10, false, matchFive),
		newMatchTestCase("equal to 0", 0, true, matchZero),
		newMatchTestCase("not equal to 0", 1, false, matchZero),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchEqFnWithCustomStruct(t *testing.T) {
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

	matchPi := MatchEqFn(s1, cmp)

	core.AssertTrue(t, matchPi.Match(s3), "equal scores")
	core.AssertFalse(t, matchPi.Match(s2), "different scores")
}

// TestMatchEqFn verifies that MatchEqFn correctly creates matchers that
// check for equality using a custom comparison function.
func TestMatchEqFn(t *testing.T) {
	t.Run("with integers", runTestMatchEqFnWithIntegers)
	t.Run("with custom struct", runTestMatchEqFnWithCustomStruct)
}

// TestMatchEqFnPanic verifies that MatchEqFn panics when given a nil comparison function.
func TestMatchEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchEqFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchEqFn2 verifies that MatchEqFn2 correctly creates matchers that check
// for equality using a custom equality condition function.
func TestMatchEqFn2(t *testing.T) {
	t.Run("with integers", runTestMatchEqFn2WithIntegers)

	t.Run("with custom struct", runTestMatchEqFn2WithCustomStruct)
}

// TestMatchEqFn2Panic verifies that MatchEqFn2 panics when given a nil condition function.
func TestMatchEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchEqFn2(1, nil)
	}, expectedNilCondFuncErr, "nil condition")
}

// TestMatchNotEq verifies that the MatchNotEq function correctly creates
// matchers that check for inequality with a given value.
func TestMatchNotEq(t *testing.T) {
	t.Run("with integers", runTestMatchNotEqWithIntegers)
}

func runTestMatchNotEqWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	notFive := MatchNotEq(5)
	notZero := MatchNotEq(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("equal to 5", 5, false, notFive),
		newMatchTestCase("not equal to 5", 10, true, notFive),
		newMatchTestCase("equal to 0", 0, false, notZero),
		newMatchTestCase("not equal to 0", 1, true, notZero),
		newMatchTestCase("negative value", -5, true, notFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchNotEqFn verifies that MatchNotEqFn correctly creates matchers that
// check if a value is not equal to a given value using a custom comparison function.
func TestMatchNotEqFn(t *testing.T) {
	t.Run("with integers", runTestMatchNotEqFnWithIntegers)
	t.Run("panic on nil", runTestMatchNotEqFnPanic)
}

func runTestMatchNotEqFnWithIntegers(t *testing.T) {
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

	// Create matchers
	notFive := MatchNotEqFn(5, cmp)
	notZero := MatchNotEqFn(0, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("equal to 5", 5, false, notFive),
		newMatchTestCase("not equal to 5", 10, true, notFive),
		newMatchTestCase("equal to 0", 0, false, notZero),
		newMatchTestCase("not equal to 0", 1, true, notZero),
		newMatchTestCase("negative value", -5, true, notFive),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchNotEqFnPanic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		MatchNotEqFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchNotEqFn2 verifies that MatchNotEqFn2 correctly creates matchers that
// check if a value is not equal to a given value using a custom equality function.
func TestMatchNotEqFn2(t *testing.T) {
	t.Run("with integers", runTestMatchNotEqFn2WithIntegers)
	t.Run("panic on nil", runTestMatchNotEqFn2Panic)
}

func runTestMatchNotEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Custom equality function
	eq := func(a, b int) bool {
		return a == b
	}

	// Create matchers
	notFive := MatchNotEqFn2(5, eq)
	notZero := MatchNotEqFn2(0, eq)

	testCases := []matchTestCase[int]{
		newMatchTestCase("equal to 5", 5, false, notFive),
		newMatchTestCase("not equal to 5", 10, true, notFive),
		newMatchTestCase("equal to 0", 0, false, notZero),
		newMatchTestCase("not equal to 0", 1, true, notZero),
		newMatchTestCase("negative value", -5, true, notFive),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchNotEqFn2Panic(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		MatchNotEqFn2(1, nil)
	}, expectedNilCondFuncErr, "nil condition")
}

// TestMatchGt verifies that the MatchGt function correctly creates matchers
// that check if a value is greater than a given value.
func TestMatchGt(t *testing.T) {
	t.Run("with integers", runTestMatchGtWithIntegers)
}

func runTestMatchGtWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	greaterThanFive := MatchGt(5)
	greaterThanZero := MatchGt(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("greater than 5", 10, true, greaterThanFive),
		newMatchTestCase("equal to 5", 5, false, greaterThanFive),
		newMatchTestCase("less than 5", 3, false, greaterThanFive),
		newMatchTestCase("greater than 0", 1, true, greaterThanZero),
		newMatchTestCase("equal to 0", 0, false, greaterThanZero),
		newMatchTestCase("less than 0", -1, false, greaterThanZero),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchGtFn verifies that MatchGtFn correctly creates matchers that check
// if a value is greater than a given value using a custom comparison function.
func TestMatchGtFn(t *testing.T) {
	t.Run("with integers", runTestMatchGtFnWithIntegers)
}

func runTestMatchGtFnWithIntegers(t *testing.T) {
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

	// Create matcher
	greaterThanFive := MatchGtFn(5, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("greater than 5", 10, true, greaterThanFive),
		newMatchTestCase("equal to 5", 5, false, greaterThanFive),
		newMatchTestCase("less than 5", 3, false, greaterThanFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchGtFnPanic verifies that MatchGtFn panics when given a nil comparison function.
func TestMatchGtFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchGtFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchLt verifies that the MatchLt function correctly creates matchers
// that check if a value is less than a given value.
func TestMatchLt(t *testing.T) {
	t.Run("with integers", runTestMatchLtWithIntegers)
}

func runTestMatchLtWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	lessThanFive := MatchLt(5)
	lessThanZero := MatchLt(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, lessThanFive),
		newMatchTestCase("equal to 5", 5, false, lessThanFive),
		newMatchTestCase("greater than 5", 7, false, lessThanFive),
		newMatchTestCase("less than 0", -1, true, lessThanZero),
		newMatchTestCase("equal to 0", 0, false, lessThanZero),
		newMatchTestCase("greater than 0", 1, false, lessThanZero),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtFn verifies that MatchLtFn correctly creates matchers that check
// if a value is less than a given value using a custom comparison function.
func TestMatchLtFn(t *testing.T) {
	t.Run("with integers", runTestMatchLtFnWithIntegers)
}

func runTestMatchLtFnWithIntegers(t *testing.T) {
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

	// Create matcher
	lessThanFive := MatchLtFn(5, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, lessThanFive),
		newMatchTestCase("equal to 5", 5, false, lessThanFive),
		newMatchTestCase("greater than 5", 7, false, lessThanFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtFnPanic verifies that MatchLtFn panics when given a nil comparison function.
func TestMatchLtFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchLtFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchLtFn2 verifies that MatchLtFn2 correctly creates matchers that check
// if a value is less than a given value using a custom condition function.
func TestMatchLtFn2(t *testing.T) {
	t.Run("with integers", runTestMatchLtFn2WithIntegers)
}

func runTestMatchLtFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Less than condition function
	isLess := func(a, b int) bool {
		return a < b
	}

	// Create matcher
	lessThanFive := MatchLtFn2(5, isLess)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, lessThanFive),
		newMatchTestCase("equal to 5", 5, false, lessThanFive),
		newMatchTestCase("greater than 5", 7, false, lessThanFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtFn2Panic verifies that MatchLtFn2 panics when given a nil condition function.
func TestMatchLtFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchLtFn2(1, nil)
	}, expectedNilCondFuncErr, "nil condition")
}

// TestMatchGtEq verifies that the MatchGtEq function correctly creates
// matchers that check if a value is greater than or equal to a given value.
func TestMatchGtEq(t *testing.T) {
	t.Run("with integers", runTestMatchGtEqWithIntegers)
}

func runTestMatchGtEqWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	gtEqFive := MatchGtEq(5)
	gtEqZero := MatchGtEq(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("greater than 5", 10, true, gtEqFive),
		newMatchTestCase("equal to 5", 5, true, gtEqFive),
		newMatchTestCase("less than 5", 3, false, gtEqFive),
		newMatchTestCase("greater than 0", 1, true, gtEqZero),
		newMatchTestCase("equal to 0", 0, true, gtEqZero),
		newMatchTestCase("less than 0", -1, false, gtEqZero),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchGtEqFn verifies that MatchGtEqFn correctly creates matchers that
// check if a value is greater than or equal to a given value using a custom
// comparison function.
func TestMatchGtEqFn(t *testing.T) {
	t.Run("with integers", runTestMatchGtEqFnWithIntegers)
}

func runTestMatchGtEqFnWithIntegers(t *testing.T) {
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

	// Create matcher
	gtEqFive := MatchGtEqFn(5, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("greater than 5", 10, true, gtEqFive),
		newMatchTestCase("equal to 5", 5, true, gtEqFive),
		newMatchTestCase("less than 5", 3, false, gtEqFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchGtEqFnPanic verifies that MatchGtEqFn panics when given a nil
// comparison function.
func TestMatchGtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchGtEqFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchGtEqFn2 verifies that MatchGtEqFn2 correctly creates matchers
// that check if a value is greater than or equal to a given value using a
// custom condition function.
func TestMatchGtEqFn2(t *testing.T) {
	t.Run("with integers", runTestMatchGtEqFn2WithIntegers)
}

func runTestMatchGtEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Less than condition function
	isLess := func(a, b int) bool {
		return a < b
	}

	// Create matcher
	gtEqFive := MatchGtEqFn2(5, isLess)

	testCases := []matchTestCase[int]{
		newMatchTestCase("greater than 5", 10, true, gtEqFive),
		newMatchTestCase("equal to 5", 5, true, gtEqFive),
		newMatchTestCase("less than 5", 3, false, gtEqFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchGtEqFn2Panic verifies that MatchGtEqFn2 panics when given a nil
// condition function.
func TestMatchGtEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchGtEqFn2(1, nil)
	}, expectedNilCondFuncErr, "nil condition")
}

// TestMatchLtEq verifies that the MatchLtEq function correctly creates
// matchers that check if a value is less than or equal to a given value.
func TestMatchLtEq(t *testing.T) {
	t.Run("with integers", runTestMatchLtEqWithIntegers)
}

func runTestMatchLtEqWithIntegers(t *testing.T) {
	t.Helper()
	// Create matchers
	ltEqFive := MatchLtEq(5)
	ltEqZero := MatchLtEq(0)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, ltEqFive),
		newMatchTestCase("equal to 5", 5, true, ltEqFive),
		newMatchTestCase("greater than 5", 7, false, ltEqFive),
		newMatchTestCase("less than 0", -1, true, ltEqZero),
		newMatchTestCase("equal to 0", 0, true, ltEqZero),
		newMatchTestCase("greater than 0", 1, false, ltEqZero),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtEqFn verifies that MatchLtEqFn correctly creates matchers that
// check if a value is less than or equal to a given value using a custom
// comparison function.
func TestMatchLtEqFn(t *testing.T) {
	t.Run("with integers", runTestMatchLtEqFnWithIntegers)
}

func runTestMatchLtEqFnWithIntegers(t *testing.T) {
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

	// Create matcher
	ltEqFive := MatchLtEqFn(5, cmp)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, ltEqFive),
		newMatchTestCase("equal to 5", 5, true, ltEqFive),
		newMatchTestCase("greater than 5", 7, false, ltEqFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtEqFnPanic verifies that MatchLtEqFn panics when given a nil
// comparison function.
func TestMatchLtEqFnPanic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchLtEqFn(1, nil)
	}, expectedNilCompFuncErr, "nil comparison")
}

// TestMatchLtEqFn2 verifies that MatchLtEqFn2 correctly creates matchers
// that check if a value is less than or equal to a given value using a
// custom condition function.
func TestMatchLtEqFn2(t *testing.T) {
	t.Run("with integers", runTestMatchLtEqFn2WithIntegers)
}

func runTestMatchLtEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Less than condition function
	isLess := func(a, b int) bool {
		return a < b
	}

	// Create matcher
	ltEqFive := MatchLtEqFn2(5, isLess)

	testCases := []matchTestCase[int]{
		newMatchTestCase("less than 5", 3, true, ltEqFive),
		newMatchTestCase("equal to 5", 5, true, ltEqFive),
		newMatchTestCase("greater than 5", 7, false, ltEqFive),
	}

	core.RunTestCases(t, testCases)
}

// TestMatchLtEqFn2Panic verifies that MatchLtEqFn2 panics when given a nil
// condition function.
func TestMatchLtEqFn2Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		MatchLtEqFn2(1, nil)
	}, expectedNilCondFuncErr, "nil condition")
}

// TestCustomTypeWithMatchers tests using matchers with custom types
func TestCustomTypeWithMatchers(t *testing.T) {
	// Test various matchers with custom types
	t.Run("equality matchers", runTestCustomTypeEqualityMatchers)

	t.Run("comparison matchers", runTestCustomTypeComparisonMatchers)
}

func runTestMatchEqFn2WithIntegers(t *testing.T) {
	t.Helper()
	// Equality function that considers numbers equal if they have the same parity
	sameParity := func(a, b int) bool {
		return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
	}

	// Create matchers for even and odd numbers
	matchEven := MatchEqFn2(4, sameParity) // will match any even number
	matchOdd := MatchEqFn2(5, sameParity)  // will match any odd number

	testCases := []matchTestCase[int]{
		newMatchTestCase("even matches even", 6, true, matchEven),
		newMatchTestCase("even doesn't match odd", 7, false, matchEven),
		newMatchTestCase("odd matches odd", 9, true, matchOdd),
		newMatchTestCase("odd doesn't match even", 8, false, matchOdd),
		newMatchTestCase("zero is even", 0, true, matchEven),
		newMatchTestCase("negative even number", -4, true, matchEven),
		newMatchTestCase("negative odd number", -3, true, matchOdd),
	}

	core.RunTestCases(t, testCases)
}

func runTestMatchEqFn2WithCustomStruct(t *testing.T) {
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

	matchAge30 := MatchEqFn2(p1, sameAge)

	core.AssertTrue(t, matchAge30.Match(p2), "same age")
	core.AssertFalse(t, matchAge30.Match(p3), "different age")
}

func runTestCustomTypeEqualityMatchers(t *testing.T) {
	t.Helper()
	type customType struct {
		value int
	}

	cmp := func(a, b customType) int {
		switch {
		case a.value < b.value:
			return -1
		case a.value > b.value:
			return 1
		default:
			return 0
		}
	}

	// Direct equality check
	isEqual := func(a, b customType) bool {
		return a.value == b.value
	}

	// Sample values
	a := customType{value: 5}
	b := customType{value: 10}
	c := customType{value: 5}

	equalToA := MatchEqFn(a, cmp)
	equalToADirect := MatchEqFn2(a, isEqual)

	core.AssertTrue(t, equalToA.Match(c), "same value")
	core.AssertFalse(t, equalToA.Match(b), "different value")
	core.AssertTrue(t, equalToADirect.Match(c), "direct same value")
}

func runTestCustomTypeComparisonMatchers(t *testing.T) {
	t.Helper()
	type customType struct {
		value int
	}

	cmp := func(a, b customType) int {
		switch {
		case a.value < b.value:
			return -1
		case a.value > b.value:
			return 1
		default:
			return 0
		}
	}

	// Less than check
	isLess := func(a, b customType) bool {
		return a.value < b.value
	}

	// Sample values
	a := customType{value: 5}
	b := customType{value: 10}
	d := customType{value: 3}

	greaterThanA := MatchGtFn(a, cmp)
	lessThanB := MatchLtFn(b, cmp)
	lessThanOrEqualToB := MatchLtEqFn2(b, isLess)

	core.AssertTrue(t, greaterThanA.Match(b), "b > a")
	core.AssertFalse(t, greaterThanA.Match(d), "d not > a")
	core.AssertTrue(t, lessThanB.Match(a), "a < b")
	core.AssertTrue(t, lessThanOrEqualToB.Match(a), "a <= b")
	core.AssertTrue(t, lessThanOrEqualToB.Match(b), "b <= b")
}
