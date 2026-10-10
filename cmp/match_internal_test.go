package cmp

import (
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var _ core.TestCase = matchImplTestCase[int]{}

// matchImplTestCase is a generic test case for match functions,
// repeating matchTestCase for the white-box tests
type matchImplTestCase[T any] struct {
	matcher  Matcher[T]
	value    T
	name     string
	expected bool
}

func newMatchImplTestCase[T any](name string, value T, expected bool, matcher Matcher[T]) matchImplTestCase[T] {
	return matchImplTestCase[T]{
		name:     name,
		value:    value,
		expected: expected,
		matcher:  matcher,
	}
}

func (tc matchImplTestCase[T]) Name() string {
	return tc.name
}

func (tc matchImplTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.matcher.Match(tc.value)
	core.AssertEqual(t, tc.expected, result, "Match(%v)", tc.value)
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

func runTestAndsMatching(t *testing.T) {
	t.Helper()
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})

	tests := []matchImplTestCase[int]{
		newMatchImplTestCase("even and positive", 4, true, andMatcher),
		newMatchImplTestCase("even but not positive", -2, false, andMatcher),
		newMatchImplTestCase("positive but not even", 3, false, andMatcher),
		newMatchImplTestCase("neither", -3, false, andMatcher),
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

	tests := []matchImplTestCase[int]{
		newMatchImplTestCase("even only", 4, true, orMatcher),
		newMatchImplTestCase("divisible by 3 only", 9, true, orMatcher),
		newMatchImplTestCase("both conditions", 6, true, orMatcher),
		newMatchImplTestCase("neither condition", 5, false, orMatcher),
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
