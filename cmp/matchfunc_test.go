package cmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMatchFunc verifies the MatchFunc implementation of the Matcher interface
func TestMatchFunc(t *testing.T) {
	t.Run("basic matching", func(t *testing.T) {
		// Simple matcher that checks if a number is even
		isEven := MatchFunc[int](func(n int) bool {
			return n%2 == 0
		})

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"even number", 4, true},
			{"odd number", 5, false},
			{"zero", 0, true},
			{"negative even", -6, true},
			{"negative odd", -3, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, isEven.Match(tt.value),
					"isEven.Match(%d) should return %v", tt.value, tt.expected)
			})
		}
	})

	t.Run("nil function", func(t *testing.T) {
		// Nil MatchFunc should match everything (return true)
		var nilMatcher MatchFunc[string]
		assert.True(t, nilMatcher.Match("anything"),
			"Nil MatchFunc should match everything")
	})
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
			assert.Equal(t, tt.expected, evenAndPositive.Match(tt.value),
				"Combined AND matcher for %d should return %v",
				tt.value, tt.expected)
		})
	}

	// Test with nil matcher
	evenAndNil := isEven.And(nil)
	assert.True(t, evenAndNil.Match(4), "Even number should match with nil in AND")
	assert.False(t, evenAndNil.Match(5), "Odd number should not match with nil in AND")

	// Test with multiple matchers including nil
	isBig := MatchFunc[int](func(n int) bool {
		return n > 100
	})
	complexMatcher := isEven.And(isPositive, isBig, nil)
	assert.True(t, complexMatcher.Match(102), "Should match all conditions")
	assert.False(t, complexMatcher.Match(4), "Should not match (not big)")
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
			assert.Equal(t, tt.expected, evenOrDivisibleBy3.Match(tt.value),
				"Combined OR matcher for %d should return %v",
				tt.value, tt.expected)
		})
	}

	// Test with nil matcher
	evenOrNil := isEven.Or(nil)
	assert.True(t, evenOrNil.Match(4), "Even number should match with nil in OR")
	assert.False(t, evenOrNil.Match(5), "Odd number should not match with nil in OR")

	// Test with multiple matchers including nil
	isBig := MatchFunc[int](func(n int) bool {
		return n > 100
	})
	complexMatcher := isEven.Or(isDivisibleBy3, isBig, nil)
	assert.True(t, complexMatcher.Match(5000), "Should match (big)")
	assert.True(t, complexMatcher.Match(4), "Should match (even)")
	assert.False(t, complexMatcher.Match(5), "Should not match any condition")
}

// TestMatchFuncNot verifies the Not method of MatchFunc
func TestMatchFuncNot(t *testing.T) {
	isEven := MatchFunc[int](func(n int) bool {
		return n%2 == 0
	})

	// Negate even checker to get odd checker
	isOdd := isEven.Not()

	tests := []struct {
		name     string
		value    int
		expected bool
	}{
		{"odd number", 5, true},
		{"even number", 4, false},
		{"zero", 0, false},
		{"negative odd", -3, true},
		{"negative even", -2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isOdd.Match(tt.value),
				"Negated matcher for %d should return %v",
				tt.value, tt.expected)
		})
	}
}

// TestAsMatcher verifies that AsMatcher converts MatchFunc to Matcher
func TestAsMatcher(t *testing.T) {
	t.Run("with valid function", func(t *testing.T) {
		isEven := MatchFunc[int](func(n int) bool {
			return n%2 == 0
		})

		matcher := AsMatcher(isEven)
		assert.NotNil(t, matcher, "AsMatcher should return non-nil for valid function")
		assert.True(t, matcher.Match(4), "Matcher should work with even number")
		assert.False(t, matcher.Match(5), "Matcher should reject odd number")
	})

	t.Run("with nil function", func(t *testing.T) {
		var nilFunc MatchFunc[int]
		matcher := AsMatcher(nilFunc)
		assert.Nil(t, matcher, "AsMatcher should return nil for nil function")
	})
}
