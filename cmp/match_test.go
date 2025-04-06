package cmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMatchAny verifies that MatchAny correctly implements OR logic
//
//revive:disable-next-line:cognitive-complexity
func TestMatchAny(t *testing.T) {
	isEven := AsMatcher(func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := AsMatcher(func(n int) bool {
		return n%3 == 0
	})

	t.Run("with matchers", func(t *testing.T) {
		// Match if number is even OR divisible by 3
		matcher := MatchAny(isEven, isDivisibleBy3)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even only", 4, true},
			{"divisible by 3 only", 9, true},
			{"both even and divisible by 3", 6, true},
			{"neither even nor divisible by 3", 5, false},
			{"negative even", -4, true},
			{"negative divisible by 3", -9, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, matcher.Match(tt.value),
					"MatchAny(isEven, isDivisibleBy3).Match(%d) should return %v",
					tt.value, tt.expected)
			})
		}
	})

	t.Run("with nil matcher", func(t *testing.T) {
		// Match if number is even OR nil matcher
		matcher := MatchAny(isEven, nil)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even number", 4, true},
			{"odd number", 5, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, matcher.Match(tt.value),
					"MatchAny(isEven, nil).Match(%d) should return %v",
					tt.value, tt.expected)
			})
		}
	})

	t.Run("empty matcher list", func(t *testing.T) {
		// Empty OR should match nothing (return false)
		matcher := MatchAny[int]()
		assert.False(t, matcher.Match(42),
			"Empty MatchAny should match nothing")
	})

	// Test logical operations on MatchAny result
	t.Run("And operation on MatchAny result", func(t *testing.T) {
		isPositive := MatchFunc[int](func(n int) bool {
			return n > 0
		})

		// (even OR divisible by 3) AND positive
		matcher := MatchAny(isEven, isDivisibleBy3).And(isPositive)

		assert.True(t, matcher.Match(4), "Should match positive even number")
		assert.True(t, matcher.Match(9), "Should match positive number divisible by 3")
		assert.False(t, matcher.Match(-6), "Should not match negative number")
	})

	t.Run("Not operation on MatchAny result", func(t *testing.T) {
		// NOT(even OR divisible by 3) = odd AND not divisible by 3
		matcher := MatchAny(isEven, isDivisibleBy3).Not()

		assert.True(t, matcher.Match(5), "Should match odd number not divisible by 3")
		assert.False(t, matcher.Match(4), "Should not match even number")
		assert.False(t, matcher.Match(9), "Should not match number divisible by 3")
	})
}

// TestMatchAll verifies that MatchAll correctly implements AND logic
//
//revive:disable-next-line:cognitive-complexity
func TestMatchAll(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	t.Run("with matchers", func(t *testing.T) {
		// Match if number is even AND positive
		matcher := MatchAll(isEven, isPositive)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even and positive", 4, true},
			{"even but not positive", -2, false},
			{"positive but not even", 3, false},
			{"neither even nor positive", -3, false},
			{"zero", 0, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, matcher.Match(tt.value),
					"MatchAll(isEven, isPositive).Match(%d) should return %v",
					tt.value, tt.expected)
			})
		}
	})

	t.Run("with nil matcher", func(t *testing.T) {
		// Match if number is even AND nil matcher
		matcher := MatchAll(isEven, nil)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even number", 4, true},
			{"odd number", 5, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, matcher.Match(tt.value),
					"MatchAll(isEven, nil).Match(%d) should return %v",
					tt.value, tt.expected)
			})
		}
	})

	t.Run("empty matcher list", func(t *testing.T) {
		// Empty AND should match everything (return true)
		matcher := MatchAll[int]()
		assert.True(t, matcher.Match(42),
			"Empty MatchAll should match everything")
	})

	// Test logical operations on MatchAll result
	t.Run("Or operation on MatchAll result", func(t *testing.T) {
		isDivisibleBy3 := MatchFunc[int](func(n int) bool {
			return n%3 == 0
		})

		// (even AND positive) OR divisible by 3
		matcher := MatchAll(isEven, isPositive).Or(isDivisibleBy3)

		assert.True(t, matcher.Match(4), "Should match positive even number")
		assert.True(t, matcher.Match(9), "Should match number divisible by 3")
		assert.False(t, matcher.Match(-2), "Should not match negative even number")
	})

	t.Run("Not operation on MatchAll result", func(t *testing.T) {
		// NOT(even AND positive) = odd OR negative (or zero)
		matcher := MatchAll(isEven, isPositive).Not()

		assert.True(t, matcher.Match(3), "Should match odd positive")
		assert.True(t, matcher.Match(-2), "Should match negative even")
		assert.False(t, matcher.Match(4), "Should not match even positive")
	})
}

// TestAndsImplementation tests the ands implementation directly
func TestAndsImplementation(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isPositive := MatchFunc[int](func(n int) bool {
		return n > 0
	})

	t.Run("matching", func(t *testing.T) {
		andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even and positive", 4, true},
			{"even but not positive", -2, false},
			{"positive but not even", 3, false},
			{"neither", -3, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, andMatcher.Match(tt.value))
			})
		}
	})

	t.Run("with nil", func(t *testing.T) {
		andMatcher := ands[int]([]Matcher[int]{isEven, nil})
		assert.True(t, andMatcher.Match(4), "Should match even with nil")
		assert.False(t, andMatcher.Match(5), "Should not match odd")
	})

	t.Run("empty", func(t *testing.T) {
		andMatcher := ands[int]([]Matcher[int]{})
		assert.True(t, andMatcher.Match(42), "Empty AND should match everything")
	})

	t.Run("And operation", func(t *testing.T) {
		isDivisibleBy4 := MatchFunc[int](func(n int) bool {
			return n%4 == 0
		})

		// (even AND positive) AND divisible by 4
		andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})
		combinedMatcher := andMatcher.And(isDivisibleBy4)

		assert.True(t, combinedMatcher.Match(8), "Should match 8")
		assert.False(t, combinedMatcher.Match(6), "Should not match 6 (not divisible by 4)")
		assert.False(t, combinedMatcher.Match(-8), "Should not match -8 (not positive)")
	})

	t.Run("Or operation", func(t *testing.T) {
		isDivisibleBy3 := MatchFunc[int](func(n int) bool {
			return n%3 == 0
		})

		// (even AND positive) OR divisible by 3
		andMatcher := MatchAll(isEven, isPositive)
		combinedMatcher := andMatcher.Or(isDivisibleBy3)

		assert.True(t, combinedMatcher.Match(6), "Should match even and positive")
		assert.True(t, combinedMatcher.Match(9), "Should match divisible by 3")
		assert.True(t, combinedMatcher.Match(-3), "Should match negative divisible by 3")
	})

	t.Run("Not operation", func(t *testing.T) {
		andMatcher := ands[int]([]Matcher[int]{isEven, isPositive})
		notMatcher := andMatcher.Not()

		assert.False(t, notMatcher.Match(4), "Should not match even and positive")
		assert.True(t, notMatcher.Match(-2), "Should match even but not positive")
		assert.True(t, notMatcher.Match(3), "Should match positive but not even")
	})
}

// TestOrsImplementation tests the ors implementation directly
func TestOrsImplementation(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	t.Run("matching", func(t *testing.T) {
		orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even only", 4, true},
			{"divisible by 3 only", 9, true},
			{"both conditions", 6, true},
			{"neither condition", 5, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, orMatcher.Match(tt.value))
			})
		}
	})

	t.Run("with nil", func(t *testing.T) {
		orMatcher := ors[int]([]Matcher[int]{isEven, nil})
		assert.True(t, orMatcher.Match(4), "Should match even")
		assert.False(t, orMatcher.Match(5), "Should not match odd despite nil")
	})

	t.Run("empty", func(t *testing.T) {
		orMatcher := ors[int]([]Matcher[int]{})
		assert.False(t, orMatcher.Match(42), "Empty OR should match nothing")
	})

	t.Run("And operation", func(t *testing.T) {
		isPositive := MatchFunc[int](func(n int) bool {
			return n > 0
		})

		// (even OR divisible by 3) AND positive
		orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
		combinedMatcher := orMatcher.And(isPositive)

		assert.True(t, combinedMatcher.Match(4), "Should match positive even")
		assert.True(t, combinedMatcher.Match(9), "Should match positive divisible by 3")
		assert.False(t, combinedMatcher.Match(-6), "Should not match negative")
	})

	t.Run("Or operation", func(t *testing.T) {
		isNegative := MatchFunc[int](func(n int) bool {
			return n < 0
		})

		// (even OR divisible by 3) OR negative
		orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
		combinedMatcher := orMatcher.Or(isNegative)

		assert.True(t, combinedMatcher.Match(4), "Should match even")
		assert.True(t, combinedMatcher.Match(9), "Should match divisible by 3")
		assert.True(t, combinedMatcher.Match(-7), "Should match negative")
		assert.False(t, combinedMatcher.Match(5), "Should not match odd positive not divisible by 3")
	})

	t.Run("Not operation", func(t *testing.T) {
		orMatcher := ors[int]([]Matcher[int]{isEven, isDivisibleBy3})
		notMatcher := orMatcher.Not()

		assert.False(t, notMatcher.Match(4), "Should not match even")
		assert.False(t, notMatcher.Match(9), "Should not match divisible by 3")
		assert.True(t, notMatcher.Match(5), "Should match numbers that don't satisfy either condition")
	})
}

// TestQJoin tests the internal qJoin function
func TestQJoin(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	t.Run("with valid first matcher", func(t *testing.T) {
		result := qJoin(isEven, []Matcher[int]{isDivisibleBy3})
		assert.Len(t, result, 2, "Should have two matchers")

		// Check that the matchers are in correct order
		assert.True(t, result[0].Match(2) && !result[0].Match(3), "First should behave like isEven")
		assert.True(t, result[1].Match(9) && !result[1].Match(4), "Second should behave like isDivisibleBy3")
	})

	t.Run("with nil first matcher", func(t *testing.T) {
		result := qJoin(nil, []Matcher[int]{isEven, isDivisibleBy3})
		assert.Len(t, result, 2, "Should have two matchers")

		// Join with nil first should return others directly
		assert.True(t, result[0].Match(2) && !result[0].Match(3), "First should behave like isEven")
		assert.True(t, result[1].Match(3) && !result[1].Match(2), "Second should behave like isDivisibleBy3")
	})

	t.Run("with nils in others", func(t *testing.T) {
		result := qJoin(isEven, []Matcher[int]{nil, isDivisibleBy3, nil})
		assert.Len(t, result, 2, "Should have two matchers (nil removed)")

		// Join should clean nils from others
		assert.True(t, result[0].Match(2) && !result[0].Match(3), "First should behave like isEven")
		assert.True(t, result[1].Match(3) && !result[1].Match(2), "Second should behave like isDivisibleBy3")
	})
}

// TestQClean tests the internal qClean function
func TestQClean(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	isDivisibleBy3 := MatchFunc[int](func(n int) bool {
		return n%3 == 0
	})

	t.Run("with no nils", func(t *testing.T) {
		queries := []Matcher[int]{isEven, isDivisibleBy3}
		result := qClean(queries)
		assert.Len(t, result, 2, "Should keep both matchers")
	})

	t.Run("with nils", func(t *testing.T) {
		queries := []Matcher[int]{nil, isEven, nil, isDivisibleBy3, nil}
		result := qClean(queries)
		assert.Len(t, result, 2, "Should remove nils")
		assert.True(t, result[0].Match(2) && !result[0].Match(3), "First should behave like isEven")
		assert.True(t, result[1].Match(3) && !result[1].Match(2), "Second should behave like isDivisibleBy3")
	})

	t.Run("with all nils", func(t *testing.T) {
		queries := []Matcher[int]{nil, nil, nil}
		result := qClean(queries)
		assert.Len(t, result, 0, "Should return empty slice")
	})
}
