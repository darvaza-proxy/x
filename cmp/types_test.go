package cmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReverse verifies the Reverse function correctly inverts the comparison results
// for various data types including primitives and custom structs.
//
//revive:disable-next-line:cognitive-complexity
func TestReverse(t *testing.T) {
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
			expected int
		}{
			{"less than", 5, 10, 1},
			{"greater than", 10, 5, -1},
			{"equal", 5, 5, 0},
			{"with zero", 0, 1, 1},
			{"negative numbers", -5, -3, 1},
			{"mixed signs", -1, 1, 1},
		}

		reversedCmp := Reverse(cmp)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, reversedCmp(tt.a, tt.b),
					"Reverse should invert comparison result for %d and %d", tt.a, tt.b)
			})
		}
	})

	t.Run("with custom struct", func(t *testing.T) {
		type score struct {
			value float64
		}

		scoreCmp := func(a, b score) int {
			switch {
			case a.value < b.value:
				return -1
			case a.value > b.value:
				return 1
			default:
				return 0
			}
		}

		reversedScoreCmp := Reverse(scoreCmp)
		s1 := score{3.14}
		s2 := score{2.71}

		assert.Equal(t, -1, reversedScoreCmp(s1, s2), "Reversed comparison of higher to lower score")
		assert.Equal(t, 1, reversedScoreCmp(s2, s1), "Reversed comparison of lower to higher score")
		assert.Equal(t, 0, reversedScoreCmp(s1, s1), "Reversed comparison of equal scores")
	})
}

// TestReverseChained verifies that applying Reverse twice returns
// a function equivalent to the original comparator.
func TestReverseChained(t *testing.T) {
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

	// Test double reverse returns to original ordering
	doubleReversed := Reverse(Reverse(cmp))

	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"less than", 5, 10, -1},
		{"greater than", 10, 5, 1},
		{"equal", 5, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Double reversed should equal original
			assert.Equal(t, tt.expected, doubleReversed(tt.a, tt.b),
				"Double reversed comparison should match original for %d and %d", tt.a, tt.b)
			assert.Equal(t, tt.expected, cmp(tt.a, tt.b),
				"Original comparison for %d and %d", tt.a, tt.b)
		})
	}
}

// TestReverseNil verifies that Reverse panics when given a nil function.
func TestReverseNil(t *testing.T) {
	assert.Panics(t, func() {
		Reverse[int](nil)
	}, "Reverse(nil) should panic with appropriate error message")
}

// TestAsLess verifies that AsLess correctly converts comparison functions
// to "less than" predicate functions for various data types.
//
//revive:disable-next-line:cognitive-complexity
//revive:disable-next-line:cyclomatic
func TestAsLess(t *testing.T) {
	t.Run("with strings", func(t *testing.T) {
		strCmp := func(a, b string) int {
			switch {
			case a < b:
				return -1
			case a > b:
				return 1
			default:
				return 0
			}
		}

		less := AsLess(strCmp)

		tests := []struct {
			name     string
			a, b     string
			expected bool
		}{
			{"less than", "apple", "banana", true},
			{"greater than", "banana", "apple", false},
			{"equal", "cherry", "cherry", false},
			{"empty string", "", "something", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := less(tt.a, tt.b)
				assert.Equal(t, tt.expected, result,
					"AsLess(cmp)(%q, %q) should return %v", tt.a, tt.b, tt.expected)
			})
		}
	})

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
			{"less than", 3, 7, true},
			{"equal", 5, 5, false},
			{"greater than", 8, 4, false},
			{"with zero", 0, 1, true},
			{"negative numbers", -3, -2, true},
			{"mixed signs", -1, 1, true},
			{"large numbers", 1000000, 1000001, true},
			{"equal large numbers", 1000000, 1000000, false},
		}

		lessFn := AsLess(cmp)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, lessFn(tt.a, tt.b),
					"AsLess should correctly determine if %d < %d", tt.a, tt.b)
			})
		}
	})

	t.Run("with custom struct", func(t *testing.T) {
		type version struct {
			major, minor, patch int
		}

		cmp := func(a, b version) int {
			if a.major != b.major {
				return a.major - b.major
			}
			if a.minor != b.minor {
				return a.minor - b.minor
			}
			return a.patch - b.patch
		}

		tests := []struct {
			name     string
			a, b     version
			expected bool
		}{
			{"lower major version", version{1, 0, 0}, version{2, 0, 0}, true},
			{"same major different minor", version{1, 2, 0}, version{1, 3, 0}, true},
			{"same major and minor different patch", version{1, 2, 3}, version{1, 2, 4}, true},
			{"identical versions", version{1, 2, 3}, version{1, 2, 3}, false},
			{"higher version", version{2, 0, 0}, version{1, 9, 9}, false},
		}

		lessFn := AsLess(cmp)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, lessFn(tt.a, tt.b),
					"AsLess should correctly compare version %v with %v", tt.a, tt.b)
			})
		}
	})
}

// TestAsLessNil verifies that AsLess panics when given a nil function.
func TestAsLessNil(t *testing.T) {
	assert.Panics(t, func() {
		AsLess[string](nil)
	}, "AsLess(nil) should panic with appropriate error message")
}

// TestAsEqual verifies that AsEqual correctly converts comparison functions
// to equality predicate functions.
//
//revive:disable-next-line:cognitive-complexity
func TestAsEqual(t *testing.T) {
	t.Run("with floats", func(t *testing.T) {
		floatCmp := func(a, b float64) int {
			switch {
			case a < b:
				return -1
			case a > b:
				return 1
			default:
				return 0
			}
		}

		equal := AsEqual(floatCmp)

		tests := []struct {
			name     string
			a, b     float64
			expected bool
		}{
			{"equal positive", 1.0, 1.0, true},
			{"less than", 1.5, 2.0, false},
			{"greater than", 3.0, 2.5, false},
			{"equal zero", 0.0, 0.0, true},
			{"equal negative", -1.0, -1.0, true},
			{"different signs", -1.0, 1.0, false},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				result := equal(tc.a, tc.b)
				assert.Equal(t, tc.expected, result,
					"AsEqual(cmp)(%.1f, %.1f) should return %v", tc.a, tc.b, tc.expected)
			})
		}
	})
}

// TestAsEqualNil verifies that AsEqual panics when given a nil function.
func TestAsEqualNil(t *testing.T) {
	assert.Panics(t, func() {
		AsEqual[float64](nil)
	}, "AsEqual(nil) should panic with appropriate error message")
}
