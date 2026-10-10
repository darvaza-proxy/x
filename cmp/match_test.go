package cmp_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/cmp"
)

// TestCase interface validations
var _ core.TestCase = matchTestCase[int]{}

// matchTestCase is a generic test case for match functions
type matchTestCase[T any] struct {
	matcher  cmp.Matcher[T]
	value    T
	name     string
	expected bool
}

func newMatchTestCase[T any](name string, value T, expected bool, matcher cmp.Matcher[T]) matchTestCase[T] {
	return matchTestCase[T]{
		name:     name,
		value:    value,
		expected: expected,
		matcher:  matcher,
	}
}

func (tc matchTestCase[T]) Name() string {
	return tc.name
}

func (tc matchTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.matcher.Match(tc.value)
	core.AssertEqual(t, tc.expected, result, "Match(%v)", tc.value)
}

// TestMatchAny verifies that MatchAny correctly implements OR logic
func TestMatchAny(t *testing.T) {
	t.Run("with matchers", runTestMatchAnyWithMatchers)
	t.Run("with nil matcher", runTestMatchAnyWithNil)
	t.Run("empty matcher list", runTestMatchAnyEmptyList)

	// Test logical operations on MatchAny result
	t.Run("And operation on MatchAny result", runTestMatchAnyAnd)
	t.Run("Not operation on MatchAny result", runTestMatchAnyNot)
}

func runTestMatchAnyAnd(t *testing.T) {
	t.Helper()
	isEven := cmp.AsMatcher(func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := cmp.AsMatcher(func(n int) bool {
		return n%3 == 0
	})
	isPositive := cmp.MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// (even OR divisible by 3) AND positive
	matcher := cmp.MatchAny(isEven, isDivisibleBy3).And(isPositive)

	core.AssertTrue(t, matcher.Match(4), "positive even")
	core.AssertTrue(t, matcher.Match(9), "positive divisible by 3")
	core.AssertFalse(t, matcher.Match(-6), "negative")
}

func runTestMatchAnyNot(t *testing.T) {
	t.Helper()
	isEven := cmp.AsMatcher(func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := cmp.AsMatcher(func(n int) bool {
		return n%3 == 0
	})

	// NOT(even OR divisible by 3) = odd AND not divisible by 3
	matcher := cmp.MatchAny(isEven, isDivisibleBy3).Not()

	core.AssertTrue(t, matcher.Match(5), "odd not divisible by 3")
	core.AssertFalse(t, matcher.Match(4), "even")
	core.AssertFalse(t, matcher.Match(9), "divisible by 3")
}

// TestMatchAll verifies that MatchAll correctly implements AND logic
func TestMatchAll(t *testing.T) {
	t.Run("with matchers", runTestMatchAllWithMatchers)
	t.Run("with nil matcher", runTestMatchAllWithNil)
	t.Run("empty matcher list", runTestMatchAllEmptyList)

	// Test logical operations on MatchAll result
	t.Run("Or operation on MatchAll result", runTestMatchAllOr)
	t.Run("Not operation on MatchAll result", runTestMatchAllNot)
}

func runTestMatchAllOr(t *testing.T) {
	t.Helper()
	isEven := cmp.MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := cmp.MatchFunc[int](func(n int) bool {
		return n > 0
	})
	isDivisibleBy3 := cmp.MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	// (even AND positive) OR divisible by 3
	matcher := cmp.MatchAll(isEven, isPositive).Or(isDivisibleBy3)

	core.AssertTrue(t, matcher.Match(4), "positive even")
	core.AssertTrue(t, matcher.Match(9), "divisible by 3")
	core.AssertFalse(t, matcher.Match(-2), "negative even")
}

func runTestMatchAllNot(t *testing.T) {
	t.Helper()
	isEven := cmp.MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := cmp.MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// NOT(even AND positive) = odd OR negative (or zero)
	matcher := cmp.MatchAll(isEven, isPositive).Not()

	core.AssertTrue(t, matcher.Match(3), "odd positive")
	core.AssertTrue(t, matcher.Match(-2), "negative even")
	core.AssertFalse(t, matcher.Match(4), "even positive")
}

func runTestMatchAnyWithMatchers(t *testing.T) {
	t.Helper()
	isEven := cmp.AsMatcher(func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := cmp.AsMatcher(func(n int) bool {
		return n%3 == 0
	})

	// Match if number is even OR divisible by 3
	matcher := cmp.MatchAny(isEven, isDivisibleBy3)

	tests := []matchTestCase[int]{
		newMatchTestCase("even only", 4, true, matcher),
		newMatchTestCase("divisible by 3 only", 9, true, matcher),
		newMatchTestCase("both even and divisible by 3", 6, true, matcher),
		newMatchTestCase("neither even nor divisible by 3", 5, false, matcher),
		newMatchTestCase("negative even", -4, true, matcher),
		newMatchTestCase("negative divisible by 3", -9, true, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestMatchAnyWithNil(t *testing.T) {
	t.Helper()
	isEven := cmp.AsMatcher(func(n int) bool {
		return n%2 == 0
	})

	// Match if number is even OR nil matcher
	matcher := cmp.MatchAny(isEven, nil)

	tests := []matchTestCase[int]{
		newMatchTestCase("even number", 4, true, matcher),
		newMatchTestCase("odd number", 5, false, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestMatchAllWithMatchers(t *testing.T) {
	t.Helper()
	isEven := cmp.MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := cmp.MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// Match if number is even AND positive
	matcher := cmp.MatchAll(isEven, isPositive)

	tests := []matchTestCase[int]{
		newMatchTestCase("even and positive", 4, true, matcher),
		newMatchTestCase("even but not positive", -2, false, matcher),
		newMatchTestCase("positive but not even", 3, false, matcher),
		newMatchTestCase("neither even nor positive", -3, false, matcher),
		newMatchTestCase("zero", 0, false, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestMatchAllWithNil(t *testing.T) {
	t.Helper()
	isEven := cmp.MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	// Match if number is even AND nil matcher
	matcher := cmp.MatchAll(isEven, nil)

	tests := []matchTestCase[int]{
		newMatchTestCase("even number", 4, true, matcher),
		newMatchTestCase("odd number", 5, false, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestMatchAnyEmptyList(t *testing.T) {
	t.Helper()
	// Empty OR should match nothing (return false)
	matcher := cmp.MatchAny[int]()
	core.AssertFalse(t, matcher.Match(42), "empty MatchAny")
}

func runTestMatchAllEmptyList(t *testing.T) {
	t.Helper()
	// Empty AND should match everything (return true)
	matcher := cmp.MatchAll[int]()
	core.AssertTrue(t, matcher.Match(42), "empty MatchAll")
}
