package cmp

import (
	"testing"

	"darvaza.org/core"
)

// TestMatchFunc verifies the MatchFunc implementation of the Matcher interface
func TestMatchFunc(t *testing.T) {
	t.Run("basic matching", runTestMatchFuncBasicMatching)
	t.Run("nil function", runTestMatchFuncNil)
}

// TestMatchFuncAnd verifies the And method of MatchFunc
func TestMatchFuncAnd(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// Combined matcher for even AND positive numbers
	evenAndPositive := isEven.And(isPositive)

	testCases := []matchTestCase[int]{
		newMatchTestCase("even and positive", 4, true, evenAndPositive),
		newMatchTestCase("even but not positive", -2, false, evenAndPositive),
		newMatchTestCase("positive but not even", 3, false, evenAndPositive),
		newMatchTestCase("neither even nor positive", -3, false, evenAndPositive),
		newMatchTestCase("zero", 0, false, evenAndPositive),
	}

	core.RunTestCases(t, testCases)

	// Test with nil matcher
	evenAndNil := isEven.And(nil)
	core.AssertTrue(t, evenAndNil.Match(4), "even with nil")
	core.AssertFalse(t, evenAndNil.Match(5), "odd with nil")

	// Test with multiple matchers including nil
	isBig := MatchFunc[int](func(n int) bool {
		return n > 100
	})
	complexMatcher := isEven.And(isPositive, isBig, nil)
	core.AssertTrue(t, complexMatcher.Match(102), "all conditions")
	core.AssertFalse(t, complexMatcher.Match(4), "not big")
}

// TestMatchFuncOr verifies the Or method of MatchFunc
func TestMatchFuncOr(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	// Combined matcher for numbers either even OR divisible by 3
	evenOrDivisibleBy3 := isEven.Or(isDivisibleBy3)

	testCases := []matchTestCase[int]{
		newMatchTestCase("even only", 4, true, evenOrDivisibleBy3),
		newMatchTestCase("divisible by 3 only", 9, true, evenOrDivisibleBy3),
		newMatchTestCase("both even and divisible by 3", 6, true, evenOrDivisibleBy3),
		newMatchTestCase("neither even nor divisible by 3", 5, false, evenOrDivisibleBy3),
		newMatchTestCase("negative even", -4, true, evenOrDivisibleBy3),
		newMatchTestCase("negative divisible by 3", -9, true, evenOrDivisibleBy3),
	}

	core.RunTestCases(t, testCases)

	// Test with nil matcher
	evenOrNil := isEven.Or(nil)
	core.AssertTrue(t, evenOrNil.Match(4), "even with nil")
	core.AssertFalse(t, evenOrNil.Match(5), "odd with nil")

	// Test with multiple matchers including nil
	isBig := MatchFunc[int](func(n int) bool {
		return n > 100
	})
	complexMatcher := isEven.Or(isDivisibleBy3, isBig, nil)
	core.AssertTrue(t, complexMatcher.Match(5000), "big")
	core.AssertTrue(t, complexMatcher.Match(4), "even")
	core.AssertFalse(t, complexMatcher.Match(5), "no condition")
}

// TestMatchFuncNot verifies the Not method of MatchFunc
func TestMatchFuncNot(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	// Negate even checker to get odd checker
	isOdd := isEven.Not()

	testCases := []matchTestCase[int]{
		newMatchTestCase("odd number", 5, true, isOdd),
		newMatchTestCase("even number", 4, false, isOdd),
		newMatchTestCase("zero", 0, false, isOdd),
		newMatchTestCase("negative odd", -3, true, isOdd),
		newMatchTestCase("negative even", -2, false, isOdd),
	}

	core.RunTestCases(t, testCases)
}

// TestAsMatcher verifies that AsMatcher converts MatchFunc to Matcher
func TestAsMatcher(t *testing.T) {
	t.Run("with valid function", runTestAsMatcherWithValidFunction)
	t.Run("with nil function", runTestAsMatcherWithNilFunction)
}

func runTestMatchFuncNil(t *testing.T) {
	t.Helper()
	// Nil MatchFunc should match everything (return true)
	var nilMatcher MatchFunc[string]
	core.AssertTrue(t, nilMatcher.Match("anything"), "nil matcher")
}

func runTestMatchFuncBasicMatching(t *testing.T) {
	t.Helper()
	// Simple matcher that checks if a number is even
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	testCases := []matchTestCase[int]{
		newMatchTestCase("even number", 4, true, isEven),
		newMatchTestCase("odd number", 5, false, isEven),
		newMatchTestCase("zero", 0, true, isEven),
		newMatchTestCase("negative even", -6, true, isEven),
		newMatchTestCase("negative odd", -3, false, isEven),
	}

	core.RunTestCases(t, testCases)
}

func runTestAsMatcherWithValidFunction(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	matcher := AsMatcher(isEven)
	core.AssertNotNil(t, matcher, "matcher")
	core.AssertTrue(t, matcher.Match(4), "even")
	core.AssertFalse(t, matcher.Match(5), "odd")
}

func runTestAsMatcherWithNilFunction(t *testing.T) {
	t.Helper()
	var nilFunc MatchFunc[int]
	matcher := AsMatcher(nilFunc)
	core.AssertNil(t, matcher, "nil function")
}

// TestM verifies that the M function correctly converts Matchers to functions
func TestM(t *testing.T) {
	t.Run("with nil matcher", runTestMWithNilMatcher)
	t.Run("with non-nil matcher", runTestMWithNonNilMatcher)
}

func runTestMWithNilMatcher(t *testing.T) {
	t.Helper()
	// When Matcher is nil, M returns a function that always returns true
	var nilMatcher Matcher[int]
	fn := M(nilMatcher)

	// Should return true for any value when matcher is nil
	core.AssertTrue(t, fn(0), "nil matcher for 0")
	core.AssertTrue(t, fn(42), "nil matcher for 42")
	core.AssertTrue(t, fn(-10), "nil matcher for -10")
}

func runTestMWithNonNilMatcher(t *testing.T) {
	t.Helper()
	// Create a matcher that checks for even numbers
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	// Convert to function
	fn := M(isEven)

	// Test the function behaviour
	core.AssertTrue(t, fn(0), "0 is even")
	core.AssertTrue(t, fn(2), "2 is even")
	core.AssertTrue(t, fn(-4), "-4 is even")
	core.AssertFalse(t, fn(1), "1 is odd")
	core.AssertFalse(t, fn(3), "3 is odd")
	core.AssertFalse(t, fn(-5), "-5 is odd")
}
