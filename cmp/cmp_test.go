package cmp

//revive:disable:cognitive-complexity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEq verifies that the Eq function correctly determines equality
// for various comparable types including integers and strings.
func TestEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"equal values", 5, 5, true},
			{"different values", 5, 10, false},
			{"negative values equal", -5, -5, true},
			{"negative values different", -5, -10, false},
			{"zero and non-zero", 0, 5, false},
			{"zero and zero", 0, 0, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, Eq(tt.a, tt.b),
					"Eq(%d, %d) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})

	t.Run("with strings", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     string
			expected bool
		}{
			{"equal strings", "hello", "hello", true},
			{"different strings", "hello", "world", false},
			{"empty strings", "", "", true},
			{"empty and non-empty", "", "hello", false},
			{"case sensitivity", "Hello", "hello", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, Eq(tt.a, tt.b),
					"Eq(%q, %q) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestEqFn verifies that EqFn correctly determines equality using
// a custom comparison function for different data types.
func TestEqFn(t *testing.T) {
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

		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"equal values", 5, 5, true},
			{"different values", 5, 10, false},
			{"negative values equal", -5, -5, true},
			{"negative values different", -5, -10, false},
			{"zero and non-zero", 0, 5, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, EqFn(tt.a, tt.b, cmp),
					"EqFn(%d, %d, cmp) should return %v", tt.a, tt.b, tt.expected)
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

		assert.True(t, EqFn(s1, s3, cmp), "Equal scores should return true")
		assert.False(t, EqFn(s1, s2, cmp), "Different scores should return false")
	})
}

// TestEqFnPanic verifies that EqFn panics when given a nil comparison
// function.
func TestEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		EqFn(1, 2, nil)
	}, "EqFn with nil comparison function should panic")
}

// TestEqFn2 verifies that EqFn2 correctly determines equality using
// a custom equality condition function.
func TestEqFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Equality function that considers numbers equal if they have the same parity
		sameParity := func(a, b int) bool {
			return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
		}

		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"both even", 4, 6, true},
			{"both odd", 5, 7, true},
			{"even and odd", 4, 7, false},
			{"odd and even", 5, 8, false},
			{"zero and even", 0, 2, true},
			{"negative numbers", -3, -5, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, EqFn2(tt.a, tt.b, sameParity),
					"EqFn2(%d, %d, sameParity) should return %v", tt.a, tt.b, tt.expected)
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

		assert.True(t, EqFn2(p1, p2, sameAge),
			"People with same age should be equal")
		assert.False(t, EqFn2(p1, p3, sameAge),
			"People with different age should not be equal")
	})
}

// TestEqFn2Panic verifies that EqFn2 panics when given a nil condition
// function.
func TestEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		EqFn2(1, 2, nil)
	}, "EqFn2 with nil condition function should panic")
}

// TestNotEq verifies the NotEq function correctly determines inequality
// for various comparable types.
func TestNotEq(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"equal values", 5, 5, false},
			{"different values", 5, 10, true},
			{"zero and non-zero", 0, 5, true},
			{"zero and zero", 0, 0, false},
			{"negative values", -5, -10, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, NotEq(tt.a, tt.b),
					"NotEq(%d, %d) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})

	t.Run("with strings", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     string
			expected bool
		}{
			{"equal strings", "hello", "hello", false},
			{"different strings", "hello", "world", true},
			{"empty strings", "", "", false},
			{"empty and non-empty", "", "hello", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, NotEq(tt.a, tt.b),
					"NotEq(%q, %q) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestNotEqFn verifies that NotEqFn correctly determines inequality
// using a custom comparison function.
func TestNotEqFn(t *testing.T) {
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

		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"equal values", 5, 5, false},
			{"different values", 5, 10, true},
			{"negative values equal", -5, -5, false},
			{"negative values different", -5, -10, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, NotEqFn(tt.a, tt.b, cmp),
					"NotEqFn(%d, %d, cmp) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestNotEqFnPanic verifies that NotEqFn panics when given a nil
// comparison function.
func TestNotEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		NotEqFn(1, 2, nil)
	}, "NotEqFn with nil comparison function should panic")
}

// TestNotEqFn2 verifies that NotEqFn2 correctly negates the result
// of a custom equality condition function.
func TestNotEqFn2(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		// Equality function that considers numbers equal if they have the same parity
		sameParity := func(a, b int) bool {
			return (a%2 == 0 && b%2 == 0) || (a%2 != 0 && b%2 != 0)
		}

		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"both even", 4, 6, false},
			{"both odd", 5, 7, false},
			{"even and odd", 4, 7, true},
			{"odd and even", 5, 8, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, NotEqFn2(tt.a, tt.b, sameParity),
					"NotEqFn2(%d, %d, sameParity) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestNotEqFn2Panic verifies that NotEqFn2 panics when given a nil
// condition function.
func TestNotEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		NotEqFn2(1, 2, nil)
	}, "NotEqFn2 with nil condition function should panic")
}

// TestLt verifies the Lt function correctly determines "less than"
// relationships for ordered types.
func TestLt(t *testing.T) {
	t.Run("with integers", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"less", 5, 10, true},
			{"greater", 10, 5, false},
			{"equal", 5, 5, false},
			{"negative and positive", -5, 5, true},
			{"negative values", -10, -5, true},
			{"with zero", 0, 5, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, Lt(tt.a, tt.b),
					"Lt(%d, %d) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})

	t.Run("with strings", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     string
			expected bool
		}{
			{"lexicographically less", "apple", "banana", true},
			{"lexicographically greater", "zebra", "apple", false},
			{"equal strings", "apple", "apple", false},
			{"empty string", "", "a", true},
			{"case sensitivity", "Z", "a", true}, // ASCII 'Z' comes before 'a'
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, Lt(tt.a, tt.b),
					"Lt(%q, %q) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestLtFn verifies that LtFn correctly determines "less than"
// relationships using a custom comparison function.
func TestLtFn(t *testing.T) {
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

		tests := []struct {
			name     string
			a, b     int
			expected bool
		}{
			{"less", 5, 10, true},
			{"greater", 10, 5, false},
			{"equal", 5, 5, false},
			{"negative numbers", -5, -3, true},
			{"mixed signs", -1, 1, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, LtFn(tt.a, tt.b, cmp),
					"LtFn(%d, %d, cmp) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})
}

// TestLtFnPanic verifies that LtFn panics when given a nil comparison
// function.
func TestLtFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		LtFn(1, 2, nil)
	})
}

// TestGt verifies the Gt function correctly determines "greater than"
// relationships for ordered types.
func TestGt(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"greater", 10, 5, true},
		{"less", 5, 10, false},
		{"equal", 5, 5, false},
		{"negative and positive", -5, 5, false},
		{"negative values", -5, -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Gt(tt.a, tt.b))
		})
	}
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

	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"greater", 10, 5, true},
		{"less", 5, 10, false},
		{"equal", 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GtFn(tt.a, tt.b, cmp))
		})
	}
}

// TestGtFnPanic verifies that GtFn panics when given a nil comparison
// function.
func TestGtFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		GtFn(1, 2, nil)
	})
}

// TestGtEq verifies the GtEq function correctly determines "greater than
// or equal to" relationships for ordered types.
func TestGtEq(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"greater", 10, 5, true},
		{"less", 5, 10, false},
		{"equal", 5, 5, true},
		{"negative and positive", -5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GtEq(tt.a, tt.b))
		})
	}
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

	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"greater", 10, 5, true},
		{"less", 5, 10, false},
		{"equal", 5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GtEqFn(tt.a, tt.b, cmp))
		})
	}
}

// TestGtEqFnPanic verifies that GtEqFn panics when given a nil comparison
// function.
func TestGtEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		GtEqFn(1, 2, nil)
	})
}

// TestLtEq verifies the LtEq function correctly determines "less than or
// equal to" relationships for ordered types.
func TestLtEq(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"less", 5, 10, true},
		{"greater", 10, 5, false},
		{"equal", 5, 5, true},
		{"negative and positive", -5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LtEq(tt.a, tt.b))
		})
	}
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

	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"less", 5, 10, true},
		{"greater", 10, 5, false},
		{"equal", 5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LtEqFn(tt.a, tt.b, cmp))
		})
	}
}

// TestLtEqFnPanic verifies that LtEqFn panics when given a nil comparison
// function.
func TestLtEqFnPanic(t *testing.T) {
	assert.Panics(t, func() {
		LtEqFn(1, 2, nil)
	})
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

	assert.True(t, EqFn(a, c, cmp))
	assert.False(t, EqFn(a, b, cmp))
	assert.True(t, NotEqFn(a, b, cmp))
	assert.True(t, LtFn(a, b, cmp))
	assert.True(t, LtEqFn(a, b, cmp))
	assert.True(t, LtEqFn(a, c, cmp))
	assert.False(t, GtFn(a, b, cmp))
	assert.True(t, GtFn(b, a, cmp))
	assert.True(t, GtEqFn(a, c, cmp))
}

// TestLtEqFn2 verifies that LtEqFn2 correctly determines "less than or
// equal to" relationships using direct less-than comparison functions.
func TestLtEqFn2(t *testing.T) {
	less := func(a, b int) bool {
		return a < b
	}

	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		{"less with positive numbers", 3, 5, true},
		{"equal with positive numbers", 5, 5, true},
		{"greater with positive numbers", 7, 5, false},
		{"less with negative numbers", -7, -5, true},
		{"equal with negative numbers", -5, -5, true},
		{"greater with negative numbers", -3, -5, false},
		{"less with mixed signs", -5, 3, true},
		{"comparing with zero", 0, 1, true},
		{"zero equality", 0, 0, true},
		{"large number comparison", 1000000, 1000001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LtEqFn2(tt.a, tt.b, less))
		})
	}

	// Test with custom type
	type temperature struct {
		celsius float64
	}
	tempLess := func(a, b temperature) bool {
		return a.celsius < b.celsius
	}

	tempTests := []struct {
		name     string
		a, b     temperature
		expected bool
	}{
		{"freezing point comparison", temperature{0}, temperature{0}, true},
		{"below freezing", temperature{-5.5}, temperature{-2.2}, true},
		{"above freezing", temperature{25.5}, temperature{30.2}, true},
		{"high temperature", temperature{100}, temperature{90}, false},
	}

	for _, tt := range tempTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LtEqFn2(tt.a, tt.b, tempLess))
		})
	}
}

// TestLtEqFn2Panic verifies that LtEqFn2 panics when given a nil
// condition function.
func TestLtEqFn2Panic(t *testing.T) {
	assert.Panics(t, func() {
		LtEqFn2(1, 2, nil)
	})
}
