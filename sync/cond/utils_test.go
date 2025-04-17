package cond

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMakeAnyMatch verifies that makeAnyMatch correctly combines predicates
// according to the specification.
func TestMakeAnyMatch(t *testing.T) {
	t.Parallel()

	// Define test cases with expected behaviour
	testCases := []struct {
		name     string
		funcs    []func(int) bool
		inputs   []int
		expected []bool
	}{
		{
			name:     "empty funcs slice returns always true",
			funcs:    []func(int) bool{},
			inputs:   []int{0, 1, -1, 42},
			expected: []bool{true, true, true, true},
		},
		{
			name:     "nil functions are the same as no functions",
			funcs:    []func(int) bool{nil, nil},
			inputs:   []int{0, 1, -1},
			expected: []bool{true, true, true},
		},
		{
			name: "single function is returned as is",
			funcs: []func(int) bool{
				func(n int) bool { return n > 0 },
			},
			inputs:   []int{-1, 0, 1},
			expected: []bool{false, false, true},
		},
		{
			name: "combined functions return true if any match",
			funcs: []func(int) bool{
				func(n int) bool { return n < 0 },
				func(n int) bool { return n > 10 },
			},
			inputs:   []int{-5, 0, 5, 15},
			expected: []bool{true, false, false, true},
		},
		{
			name: "short-circuit evaluation",
			funcs: []func(int) bool{
				func(n int) bool { return n < 0 },
				func(_ int) bool {
					t.Error("This should not be called when first func returns true")
					return true
				},
			},
			inputs:   []int{-1},
			expected: []bool{true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runMakeAnyMatchTest(t, tc.funcs, tc.inputs, tc.expected)
		})
	}
}

// TestMakeAnyMatchWithStrings verifies that makeAnyMatch works with
// different types like strings.
func TestMakeAnyMatchWithStrings(t *testing.T) {
	t.Parallel()

	// Define predicates for strings
	funcs := []func(string) bool{
		func(s string) bool { return len(s) > 5 },
		func(s string) bool { return s == "match" },
	}

	testCases := []struct {
		input    string
		expected bool
	}{
		{"short", false},
		{"veryLongString", true},
		{"match", true},
	}

	// Create the combined function
	resultFunc := makeAnyMatch(funcs)

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			actual := resultFunc(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

// runMakeAnyMatchTest executes a single test case for the makeAnyMatch function.
func runMakeAnyMatchTest(
	t *testing.T,
	funcs []func(int) bool,
	inputs []int,
	expected []bool,
) {
	t.Helper()

	// Create the combined function
	resultFunc := makeAnyMatch(funcs)
	assert.NotNil(t, resultFunc, "makeAnyMatch should never return nil")

	// Test each input
	for i, input := range inputs {
		actual := resultFunc(input)
		assert.Equal(t, expected[i], actual, "Failed for input %v", input)
	}
}

// BenchmarkMakeAnyMatch measures the performance of makeAnyMatch with
// varying numbers of predicates.
func BenchmarkMakeAnyMatch(b *testing.B) {
	// Prepare a set of predicates
	predicates := []func(int) bool{
		func(n int) bool { return n < 0 },
		func(n int) bool { return n > 100 },
		func(n int) bool { return n%2 == 0 },
		func(n int) bool { return n%3 == 0 },
		func(n int) bool { return n%5 == 0 },
	}

	benchCases := []struct {
		name string
		n    int // Number of predicates to use
	}{
		{"Single", 1},
		{"Two", 2},
		{"Five", 5},
	}

	for _, bc := range benchCases {
		funcs := predicates[:bc.n]
		resultFunc := makeAnyMatch(funcs)

		b.Run(bc.name, func(b *testing.B) {
			for i := range b.N {
				// Use i as the input to get varying results
				_ = resultFunc(i)
			}
		})
	}
}
