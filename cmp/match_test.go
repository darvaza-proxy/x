package cmp

import (
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var _ core.TestCase = matchTestCase[int]{}

// matchTestCase is a generic test case for match functions
type matchTestCase[T any] struct {
	name     string
	value    T
	expected bool
	matcher  Matcher[T]
}

func newMatchTestCase[T any](name string, value T, expected bool, matcher Matcher[T]) matchTestCase[T] {
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
	isEven := AsMatcher(func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := AsMatcher(func(n int) bool {
		return n%3 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// (even OR divisible by 3) AND positive
	matcher := MatchAny(isEven, isDivisibleBy3).And(isPositive)

	core.AssertTrue(t, matcher.Match(4), "positive even")
	core.AssertTrue(t, matcher.Match(9), "positive divisible by 3")
	core.AssertFalse(t, matcher.Match(-6), "negative")
}

func runTestMatchAnyNot(t *testing.T) {
	t.Helper()
	isEven := AsMatcher(func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := AsMatcher(func(n int) bool {
		return n%3 == 0
	})

	// NOT(even OR divisible by 3) = odd AND not divisible by 3
	matcher := MatchAny(isEven, isDivisibleBy3).Not()

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
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})
	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	// (even AND positive) OR divisible by 3
	matcher := MatchAll(isEven, isPositive).Or(isDivisibleBy3)

	core.AssertTrue(t, matcher.Match(4), "positive even")
	core.AssertTrue(t, matcher.Match(9), "divisible by 3")
	core.AssertFalse(t, matcher.Match(-2), "negative even")
}

func runTestMatchAllNot(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// NOT(even AND positive) = odd OR negative (or zero)
	matcher := MatchAll(isEven, isPositive).Not()

	core.AssertTrue(t, matcher.Match(3), "odd positive")
	core.AssertTrue(t, matcher.Match(-2), "negative even")
	core.AssertFalse(t, matcher.Match(4), "even positive")
}

// TestAndsImplementation tests the ands implementation directly
func TestAndsImplementation(t *testing.T) {
	t.Run("matching", runTestAndsMatching)
	t.Run("with nil", runTestAndsWithNil)
	t.Run("empty", runTestAndsEmpty)
	t.Run("And operation", runTestAndsAnd)
	t.Run("Or operation", runTestAndsOr)
	t.Run("Not operation", runTestAndsNot)
}

func runTestAndsAnd(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})
	isDivisibleBy4 := MatchFunc[int](func(n int) bool {
		return n%4 == 0
	})

	// (even AND positive) AND divisible by 4
	andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})
	combinedMatcher := andMatcher.And(isDivisibleBy4)

	core.AssertTrue(t, combinedMatcher.Match(8), "match 8")
	core.AssertFalse(t, combinedMatcher.Match(6), "not divisible by 4")
	core.AssertFalse(t, combinedMatcher.Match(-8), "not positive")
}

func runTestAndsOr(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})
	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	// (even AND positive) OR divisible by 3
	andMatcher := MatchAll(isEven, isPositive)
	combinedMatcher := andMatcher.Or(isDivisibleBy3)

	core.AssertTrue(t, combinedMatcher.Match(6), "even and positive")
	core.AssertTrue(t, combinedMatcher.Match(9), "divisible by 3")
	core.AssertTrue(t, combinedMatcher.Match(-3), "negative divisible by 3")
}

func runTestAndsNot(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})
	notMatcher := andMatcher.Not()

	core.AssertFalse(t, notMatcher.Match(4), "not even and positive")
	core.AssertTrue(t, notMatcher.Match(-2), "even but not positive")
	core.AssertTrue(t, notMatcher.Match(3), "positive but not even")
}

// TestOrsImplementation tests the ors implementation directly
func TestOrsImplementation(t *testing.T) {
	t.Run("matching", runTestOrsMatching)
	t.Run("with nil", runTestOrsWithNil)
	t.Run("empty", runTestOrsEmpty)
	t.Run("And operation", runTestOrsAnd)
	t.Run("Or operation", runTestOrsOr)
	t.Run("Not operation", runTestOrsNot)
}

func runTestOrsAnd(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})
	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// (even OR divisible by 3) AND positive
	orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
	combinedMatcher := orMatcher.And(isPositive)

	core.AssertTrue(t, combinedMatcher.Match(4), "positive even")
	core.AssertTrue(t, combinedMatcher.Match(9), "positive divisible by 3")
	core.AssertFalse(t, combinedMatcher.Match(-6), "negative")
}

func runTestOrsOr(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})
	isNegative := MatchFunc[int](func(n int) bool {
		return n < 0
	})

	// (even OR divisible by 3) OR negative
	orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
	combinedMatcher := orMatcher.Or(isNegative)

	core.AssertTrue(t, combinedMatcher.Match(4), "even")
	core.AssertTrue(t, combinedMatcher.Match(9), "divisible by 3")
	core.AssertTrue(t, combinedMatcher.Match(-7), "negative")
	core.AssertFalse(t, combinedMatcher.Match(5), "odd positive not divisible by 3")
}

func runTestOrsNot(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})
	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
	notMatcher := orMatcher.Not()

	core.AssertFalse(t, notMatcher.Match(4), "even")
	core.AssertFalse(t, notMatcher.Match(9), "divisible by 3")
	core.AssertTrue(t, notMatcher.Match(5), "neither condition")
}

// TestQJoin tests the internal qJoin function
func TestQJoin(t *testing.T) {
	t.Run("with valid first matcher", runTestQJoinValidFirst)
	t.Run("with nil first matcher", runTestQJoinNilFirst)
	t.Run("with nils in others", runTestQJoinNilsInOthers)
}

// TestQClean tests the internal qClean function
func TestQClean(t *testing.T) {
	t.Run("with no nils", runTestQCleanNoNils)
	t.Run("with nils", runTestQCleanWithNils)
	t.Run("with all nils", runTestQCleanAllNils)
}

func runTestMatchAnyWithMatchers(t *testing.T) {
	t.Helper()
	isEven := AsMatcher(func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := AsMatcher(func(n int) bool {
		return n%3 == 0
	})

	// Match if number is even OR divisible by 3
	matcher := MatchAny(isEven, isDivisibleBy3)

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
	isEven := AsMatcher(func(n int) bool {
		return n%2 == 0
	})

	// Match if number is even OR nil matcher
	matcher := MatchAny(isEven, nil)

	tests := []matchTestCase[int]{
		newMatchTestCase("even number", 4, true, matcher),
		newMatchTestCase("odd number", 5, false, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestMatchAllWithMatchers(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	// Match if number is even AND positive
	matcher := MatchAll(isEven, isPositive)

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
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	// Match if number is even AND nil matcher
	matcher := MatchAll(isEven, nil)

	tests := []matchTestCase[int]{
		newMatchTestCase("even number", 4, true, matcher),
		newMatchTestCase("odd number", 5, false, matcher),
	}

	core.RunTestCases(t, tests)
}

func runTestAndsMatching(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})

	tests := []matchTestCase[int]{
		newMatchTestCase("even and positive", 4, true, andMatcher),
		newMatchTestCase("even but not positive", -2, false, andMatcher),
		newMatchTestCase("positive but not even", 3, false, andMatcher),
		newMatchTestCase("neither", -3, false, andMatcher),
	}

	core.RunTestCases(t, tests)
}

func runTestOrsMatching(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})

	tests := []matchTestCase[int]{
		newMatchTestCase("even only", 4, true, orMatcher),
		newMatchTestCase("divisible by 3 only", 9, true, orMatcher),
		newMatchTestCase("both conditions", 6, true, orMatcher),
		newMatchTestCase("neither condition", 5, false, orMatcher),
	}

	core.RunTestCases(t, tests)
}

func runTestQJoinValidFirst(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	result := qJoin(isEven, []Matcher[int]{isDivisibleBy3})
	core.AssertEqual(t, 2, len(result), "length")

	// Check that the matchers are in correct order
	core.AssertTrue(t, result[0].Match(2) && !result[0].Match(3), "first isEven")
	core.AssertTrue(t, result[1].Match(9) && !result[1].Match(4), "second isDivisibleBy3")
}

func runTestQJoinNilFirst(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	result := qJoin(nil, []Matcher[int]{isEven, isDivisibleBy3})
	core.AssertEqual(t, 2, len(result), "length")

	// Join with nil first should return others directly
	core.AssertTrue(t, result[0].Match(2) && !result[0].Match(3), "first isEven")
	core.AssertTrue(t, result[1].Match(3) && !result[1].Match(2), "second isDivisibleBy3")
}

func runTestQJoinNilsInOthers(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	result := qJoin(isEven, []Matcher[int]{nil, isDivisibleBy3, nil})
	core.AssertEqual(t, 2, len(result), "length")

	// Join should clean nils from others
	core.AssertTrue(t, result[0].Match(2) && !result[0].Match(3), "first isEven")
	core.AssertTrue(t, result[1].Match(3) && !result[1].Match(2), "second isDivisibleBy3")
}

func runTestQCleanWithNils(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	queries := []Matcher[int]{nil, isEven, nil, isDivisibleBy3, nil}
	result := qClean(queries)
	core.AssertEqual(t, 2, len(result), "length")
	core.AssertTrue(t, result[0].Match(2) && !result[0].Match(3), "first isEven")
	core.AssertTrue(t, result[1].Match(3) && !result[1].Match(2), "second isDivisibleBy3")
}

func runTestAndsWithNil(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	andMatcher := ands[int]([]Matcher[int]{isEven, nil})
	core.AssertTrue(t, andMatcher.Match(4), "even with nil")
	core.AssertFalse(t, andMatcher.Match(5), "odd")
}

func runTestOrsWithNil(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	orMatcher := ors[int]([]Matcher[int]{isEven, nil})
	core.AssertTrue(t, orMatcher.Match(4), "even")
	core.AssertFalse(t, orMatcher.Match(5), "odd despite nil")
}

func runTestMatchAnyEmptyList(t *testing.T) {
	t.Helper()
	// Empty OR should match nothing (return false)
	matcher := MatchAny[int]()
	core.AssertFalse(t, matcher.Match(42), "empty MatchAny")
}

func runTestMatchAllEmptyList(t *testing.T) {
	t.Helper()
	// Empty AND should match everything (return true)
	matcher := MatchAll[int]()
	core.AssertTrue(t, matcher.Match(42), "empty MatchAll")
}

func runTestAndsEmpty(t *testing.T) {
	t.Helper()
	andMatcher := ands[int]([]Matcher[int]{})
	core.AssertTrue(t, andMatcher.Match(42), "empty AND")
}

func runTestOrsEmpty(t *testing.T) {
	t.Helper()
	orMatcher := ors[int]([]Matcher[int]{})
	core.AssertFalse(t, orMatcher.Match(42), "empty OR")
}

func runTestQCleanNoNils(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	queries := []Matcher[int]{isEven, isDivisibleBy3}
	result := qClean(queries)
	core.AssertEqual(t, 2, len(result), "length")
}

func runTestQCleanAllNils(t *testing.T) {
	t.Helper()
	queries := []Matcher[int]{nil, nil, nil}
	result := qClean(queries)
	core.AssertEqual(t, 0, len(result), "length")
}
