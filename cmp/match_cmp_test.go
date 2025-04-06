package cmp

//revive:disable:cognitive-complexity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMatchEq verifies the MatchEq function correctly creates matchers
// that check for equality with a given value.
func TestMatchEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		matchFive := MatchEq(5)
		matchZero := MatchEq(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"equal to 5", 5, matchFive, true},
			{"not equal to 5", 10, matchFive, false},
			{"equal to 0", 0, matchZero, true},
			{"not equal to 0", 1, matchZero, false},
			{"negative value", -5, matchFive, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchEq(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})

	t.Run("with strings", func(t *testing.T) {
		// Create matchers
		matchHello := MatchEq("hello")
		matchEmpty := MatchEq("")

		tests := []struct {
			name     string
			value    string
			matcher  Matcher[string]
			expected bool
		}{
			{"equal strings", "hello", matchHello, true},
			{"different strings", "world", matchHello, false},
			{"empty string match", "", matchEmpty, true},
			{"non-empty vs empty", "hello", matchEmpty, false},
			{"case sensitivity", "Hello", matchHello, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchEq(%q)(%q) should return %v", "hello", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchEqFn verifies that MatchEqFn correctly creates matchers that check
// for equality using a custom comparison function.
func TestMatchEqFn(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
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

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"equal to 5", 5, matchFive, true},
			{"not equal to 5", 10, matchFive, false},
			{"equal to 0", 0, matchZero, true},
			{"not equal to 0", 1, matchZero, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchEqFn(5, cmp)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})

	t.Run("with custom struct", func(t *testing.T) {
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

		assert.True(t, matchPi.Match(s3), "Equal scores should match")
		assert.False(t, matchPi.Match(s2), "Different scores should not match")
	})
}

// TestMatchEqFnPanic verifies that MatchEqFn panics when given a nil comparison function.
func TestMatchEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		MatchEqFn(1, nil)
	}, "MatchEqFn with nil comparison function should panic")
}

// TestMatchEqFn2 verifies that MatchEqFn2 correctly creates matchers that check
// for equality using a custom equality condition function.
func TestMatchEqFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Equality function that considers numbers equal if they have the same parity
		sameParity := func(a, b int) bool {
			return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
		}

		// Create matchers for even and odd numbers
		matchEven := MatchEqFn2(4, sameParity) // will match any even number
		matchOdd := MatchEqFn2(5, sameParity)  // will match any odd number

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"even matches even", 6, matchEven, true},
			{"even doesn't match odd", 7, matchEven, false},
			{"odd matches odd", 9, matchOdd, true},
			{"odd doesn't match even", 8, matchOdd, false},
			{"zero is even", 0, matchEven, true},
			{"negative even number", -4, matchEven, true},
			{"negative odd number", -3, matchOdd, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchEqFn2 with sameParity matcher should return %v for %d", tt.expected, tt.value)
			})
		}
	})

	t.Run("with custom struct", func(t *testing.T) {
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

		assert.True(t, matchAge30.Match(p2), "People with same age should match")
		assert.False(t, matchAge30.Match(p3), "People with different age should not match")
	})
}

// TestMatchEqFn2Panic verifies that MatchEqFn2 panics when given a nil condition function.
func TestMatchEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		MatchEqFn2(1, nil)
	}, "MatchEqFn2 with nil condition function should panic")
}

// TestMatchNotEq verifies the MatchNotEq function correctly creates matchers
// that check for inequality with a given value.
func TestMatchNotEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		notFive := MatchNotEq(5)
		notZero := MatchNotEq(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"equal to 5", 5, notFive, false},
			{"not equal to 5", 10, notFive, true},
			{"equal to 0", 0, notZero, false},
			{"not equal to 0", 1, notZero, true},
			{"negative value", -5, notFive, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchNotEq(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGt verifies the MatchGt function correctly creates matchers
// that check if a value is greater than a given value.
func TestMatchGt(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		greaterThanFive := MatchGt(5)
		greaterThanZero := MatchGt(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"greater than 5", 10, greaterThanFive, true},
			{"equal to 5", 5, greaterThanFive, false},
			{"less than 5", 3, greaterThanFive, false},
			{"greater than 0", 1, greaterThanZero, true},
			{"equal to 0", 0, greaterThanZero, false},
			{"less than 0", -1, greaterThanZero, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchGt(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGtFn verifies that MatchGtFn correctly creates matchers that check
// if a value is greater than a given value using a custom comparison function.
func TestMatchGtFn(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
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

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"greater than 5", 10, true},
			{"equal to 5", 5, false},
			{"less than 5", 3, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, greaterThanFive.Match(tt.value),
					"MatchGtFn(5, cmp)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGtFnPanic verifies that MatchGtFn panics when given a nil comparison function.
func TestMatchGtFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		MatchGtFn(1, nil)
	}, "MatchGtFn with nil comparison function should panic")
}

// TestMatchLt verifies the MatchLt function correctly creates matchers
// that check if a value is less than a given value.
func TestMatchLt(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		lessThanFive := MatchLt(5)
		lessThanZero := MatchLt(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"less than 5", 3, lessThanFive, true},
			{"equal to 5", 5, lessThanFive, false},
			{"greater than 5", 7, lessThanFive, false},
			{"less than 0", -1, lessThanZero, true},
			{"equal to 0", 0, lessThanZero, false},
			{"greater than 0", 1, lessThanZero, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchLt(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtFn verifies that MatchLtFn correctly creates matchers that check
// if a value is less than a given value using a custom comparison function.
func TestMatchLtFn(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
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

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"less than 5", 3, true},
			{"equal to 5", 5, false},
			{"greater than 5", 7, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, lessThanFive.Match(tt.value),
					"MatchLtFn(5, cmp)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtFnPanic verifies that MatchLtFn panics when given a nil comparison function.
func TestMatchLtFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		MatchLtFn(1, nil)
	}, "MatchLtFn with nil comparison function should panic")
}

// TestMatchLtFn2 verifies that MatchLtFn2 correctly creates matchers that check
// if a value is less than a given value using a custom condition function.
func TestMatchLtFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Less than condition function
		isLess := func(a, b int) bool {
			return a < b
		}

		// Create matcher
		lessThanFive := MatchLtFn2(5, isLess)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"less than 5", 3, true},
			{"equal to 5", 5, false},
			{"greater than 5", 7, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, lessThanFive.Match(tt.value),
					"MatchLtFn2(5, isLess)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtFn2Panic verifies that MatchLtFn2 panics when given a nil condition function.
func TestMatchLtFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		MatchLtFn2(1, nil)
	}, "MatchLtFn2 with nil condition function should panic")
}

// TestMatchGtEq verifies the MatchGtEq function correctly creates matchers
// that check if a value is greater than or equal to a given value.
func TestMatchGtEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		gtEqFive := MatchGtEq(5)
		gtEqZero := MatchGtEq(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"greater than 5", 10, gtEqFive, true},
			{"equal to 5", 5, gtEqFive, true},
			{"less than 5", 3, gtEqFive, false},
			{"greater than 0", 1, gtEqZero, true},
			{"equal to 0", 0, gtEqZero, true},
			{"less than 0", -1, gtEqZero, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchGtEq(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGtEqFn verifies that MatchGtEqFn correctly creates matchers that check
// if a value is greater than or equal to a given value using a custom comparison function.
func TestMatchGtEqFn(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
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

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"greater than 5", 10, true},
			{"equal to 5", 5, true},
			{"less than 5", 3, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, gtEqFive.Match(tt.value),
					"MatchGtEqFn(5, cmp)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGtEqFnPanic verifies that MatchGtEqFn panics when given a nil comparison function.
func TestMatchGtEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		MatchGtEqFn(1, nil)
	}, "MatchGtEqFn with nil comparison function should panic")
}

// TestMatchGtEqFn2 verifies that MatchGtEqFn2 correctly creates matchers that check
// if a value is greater than or equal to a given value using a custom condition function.
func TestMatchGtEqFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Less than condition function
		isLess := func(a, b int) bool {
			return a < b
		}

		// Create matcher
		gtEqFive := MatchGtEqFn2(5, isLess)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"greater than 5", 10, true},
			{"equal to 5", 5, true},
			{"less than 5", 3, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, gtEqFive.Match(tt.value),
					"MatchGtEqFn2(5, isLess)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchGtEqFn2Panic verifies that MatchGtEqFn2 panics when given a nil condition function.
func TestMatchGtEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		MatchGtEqFn2(1, nil)
	}, "MatchGtEqFn2 with nil condition function should panic")
}

// TestMatchLtEq verifies the MatchLtEq function correctly creates matchers
// that check if a value is less than or equal to a given value.
func TestMatchLtEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Create matchers
		ltEqFive := MatchLtEq(5)
		ltEqZero := MatchLtEq(0)

		tests := []struct {
			name     string
			value    int
			matcher  Matcher[int]
			expected bool
		}{
			{"less than 5", 3, ltEqFive, true},
			{"equal to 5", 5, ltEqFive, true},
			{"greater than 5", 7, ltEqFive, false},
			{"less than 0", -1, ltEqZero, true},
			{"equal to 0", 0, ltEqZero, true},
			{"greater than 0", 1, ltEqZero, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.matcher.Match(tt.value),
					"MatchLtEq(%d)(%d) should return %v", 5, tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtEqFn verifies that MatchLtEqFn correctly creates matchers that check
// if a value is less than or equal to a given value using a custom comparison function.
func TestMatchLtEqFn(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
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

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"less than 5", 3, true},
			{"equal to 5", 5, true},
			{"greater than 5", 7, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, ltEqFive.Match(tt.value),
					"MatchLtEqFn(5, cmp)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtEqFnPanic verifies that MatchLtEqFn panics when given a nil comparison function.
func TestMatchLtEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		MatchLtEqFn(1, nil)
	}, "MatchLtEqFn with nil comparison function should panic")
}

// TestMatchLtEqFn2 verifies that MatchLtEqFn2 correctly creates matchers that check
// if a value is less than or equal to a given value using a custom condition function.
func TestMatchLtEqFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Less than condition function
		isLess := func(a, b int) bool {
			return a < b
		}

		// Create matcher
		ltEqFive := MatchLtEqFn2(5, isLess)

		tests := []struct {
			name     string
			value    int
			expected bool
		}{
			{"less than 5", 3, true},
			{"equal to 5", 5, true},
			{"greater than 5", 7, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, ltEqFive.Match(tt.value),
					"MatchLtEqFn2(5, isLess)(%d) should return %v", tt.value, tt.expected)
			})
		}
	})
}

// TestMatchLtEqFn2Panic verifies that MatchLtEqFn2 panics when given a nil condition function.
func TestMatchLtEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		MatchLtEqFn2(1, nil)
	}, "MatchLtEqFn2 with nil condition function should panic")
}

// TestCustomTypeWithMatchers tests using matchers with custom types
func TestCustomTypeWithMatchers(t *testing.T) {
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

	// Less than check
	isLess := func(a, b customType) bool {
		return a.value < b.value
	}

	// Sample values
	a := customType{value: 5}
	b := customType{value: 10}
	c := customType{value: 5}
	d := customType{value: 3}

	// Test various matchers with custom types
	t.Run("equality matchers", func(t *testing.T) {
		equalToA := MatchEqFn(a, cmp)
		equalToADirect := MatchEqFn2(a, isEqual)

		assert.True(t, equalToA.Match(c), "equalToA should match c (same value)")
		assert.False(t, equalToA.Match(b), "equalToA should not match b (different value)")
		assert.True(t, equalToADirect.Match(c), "equalToADirect should match c (same value)")
	})

	t.Run("comparison matchers", func(t *testing.T) {
		greaterThanA := MatchGtFn(a, cmp)
		lessThanB := MatchLtFn(b, cmp)
		lessThanOrEqualToB := MatchLtEqFn2(b, isLess)

		assert.True(t, greaterThanA.Match(b), "b should be greater than a")
		assert.False(t, greaterThanA.Match(d), "d should not be greater than a")
		assert.True(t, lessThanB.Match(a), "a should be less than b")
		assert.True(t, lessThanOrEqualToB.Match(a), "a should be less than or equal to b")
		assert.True(t, lessThanOrEqualToB.Match(b), "b should be less than or equal to b (equal)")
	})
}
