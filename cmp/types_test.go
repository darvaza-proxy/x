package cmp_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/cmp"
)

// TestCase interface validations for types_test
var _ core.TestCase = reverseTestCase[int]{}
var _ core.TestCase = reverseChainedTestCase{}
var _ core.TestCase = asLessTestCase[string]{}
var _ core.TestCase = asLessTestCase[int]{}
var _ core.TestCase = asEqualTestCase[float64]{}
var _ core.TestCase = asCmpTestCase[int]{}
var _ core.TestCase = asCmpTestCase[string]{}
var _ core.TestCase = asCmpAsLessCycleTestCase[int]{}

// reverseTestCase tests the Reverse function
type reverseTestCase[T any] struct {
	a, b     T
	cmp      cmp.CompFunc[T]
	name     string
	expected int
}

func newReverseTestCase[T any](name string, a, b T, expected int, fn cmp.CompFunc[T]) reverseTestCase[T] {
	return reverseTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cmp:      fn,
	}
}

func (tc reverseTestCase[T]) Name() string {
	return tc.name
}

func (tc reverseTestCase[T]) Test(t *testing.T) {
	t.Helper()
	reversedCmp := cmp.Reverse(tc.cmp)
	result := reversedCmp(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, "Reverse")
}

// runTestReverseWithIntegers tests Reverse with integer comparison
func runTestReverseWithIntegers(t *testing.T) {
	t.Helper()
	fn := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tests := []reverseTestCase[int]{
		newReverseTestCase("less than", 5, 10, 1, fn),
		newReverseTestCase("greater than", 10, 5, -1, fn),
		newReverseTestCase("equal", 5, 5, 0, fn),
		newReverseTestCase("with zero", 0, 1, 1, fn),
		newReverseTestCase("negative numbers", -5, -3, 1, fn),
		newReverseTestCase("mixed signs", -1, 1, 1, fn),
	}

	core.RunTestCases(t, tests)
}

// TestReverse ensures the Reverse function properly inverts comparison results
// for various data types including primitives and custom structures.
func TestReverse(t *testing.T) {
	t.Run("with integers", runTestReverseWithIntegers)
	t.Run("with custom struct", runTestReverseCustomStruct)
}

// runTestReverseCustomStruct tests Reverse with a custom struct type
func runTestReverseCustomStruct(t *testing.T) {
	t.Helper()
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

	reversedScoreCmp := cmp.Reverse(scoreCmp)
	s1 := score{3.14}
	s2 := score{2.71}

	core.AssertEqual(t, -1, reversedScoreCmp(s1, s2), "reversed higher to lower")
	core.AssertEqual(t, 1, reversedScoreCmp(s2, s1), "reversed lower to higher")
	core.AssertEqual(t, 0, reversedScoreCmp(s1, s1), "reversed equal")
}

// reverseChainedTestCase tests double reverse behaviour
type reverseChainedTestCase struct {
	name     string
	a, b     int
	expected int
}

func newReverseChainedTestCase(name string, a, b, expected int) reverseChainedTestCase {
	return reverseChainedTestCase{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
	}
}

func (tc reverseChainedTestCase) Name() string {
	return tc.name
}

func (tc reverseChainedTestCase) Test(t *testing.T) {
	t.Helper()
	fn := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	doubleReversed := cmp.Reverse(cmp.Reverse(fn))
	result := doubleReversed(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, "double reversed")
	core.AssertEqual(t, tc.expected, fn(tc.a, tc.b), "original")
}

// TestReverseChained confirms that applying Reverse twice returns
// a function equivalent to the original comparator.
func TestReverseChained(t *testing.T) {
	tests := []reverseChainedTestCase{
		newReverseChainedTestCase("less than", 5, 10, -1),
		newReverseChainedTestCase("greater than", 10, 5, 1),
		newReverseChainedTestCase("equal", 5, 5, 0),
	}

	core.RunTestCases(t, tests)
}

// TestReverseNil confirms that Reverse panics when given a nil function.
func TestReverseNil(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.Reverse[int](nil)
	}, expectedNilCompFuncErr, "Reverse(nil)")
}

// asLessTestCase tests the AsLess function
type asLessTestCase[T any] struct {
	a, b     T
	cmp      cmp.CompFunc[T]
	name     string
	expected bool
}

func newAsLessTestCase[T any](name string, a, b T, expected bool, fn cmp.CompFunc[T]) asLessTestCase[T] {
	return asLessTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cmp:      fn,
	}
}

func (tc asLessTestCase[T]) Name() string {
	return tc.name
}

func (tc asLessTestCase[T]) Test(t *testing.T) {
	t.Helper()
	less := cmp.AsLess(tc.cmp)
	result := less(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, "AsLess")
}

// runTestAsLessWithStrings tests AsLess with string comparison
func runTestAsLessWithStrings(t *testing.T) {
	t.Helper()
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

	tests := []asLessTestCase[string]{
		newAsLessTestCase("less than", "apple", "banana", true, strCmp),
		newAsLessTestCase("greater than", "banana", "apple", false, strCmp),
		newAsLessTestCase("equal", "cherry", "cherry", false, strCmp),
		newAsLessTestCase("empty string", "", "something", true, strCmp),
	}

	core.RunTestCases(t, tests)
}

// runTestAsLessWithIntegers tests AsLess with integer comparison
func runTestAsLessWithIntegers(t *testing.T) {
	t.Helper()
	fn := func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tests := []asLessTestCase[int]{
		newAsLessTestCase("less than", 3, 7, true, fn),
		newAsLessTestCase("equal", 5, 5, false, fn),
		newAsLessTestCase("greater than", 8, 4, false, fn),
		newAsLessTestCase("with zero", 0, 1, true, fn),
		newAsLessTestCase("negative numbers", -3, -2, true, fn),
		newAsLessTestCase("mixed signs", -1, 1, true, fn),
		newAsLessTestCase("large numbers", 1000000, 1000001, true, fn),
		newAsLessTestCase("equal large numbers", 1000000, 1000000, false, fn),
	}

	core.RunTestCases(t, tests)
}

// TestAsLess ensures that AsLess correctly converts comparison functions
// to "less than" predicate functions for various data types.
func TestAsLess(t *testing.T) {
	t.Run("with strings", runTestAsLessWithStrings)
	t.Run("with integers", runTestAsLessWithIntegers)
	t.Run("with custom struct", runTestAsLessCustomStruct)
}

// runTestAsLessCustomStruct tests AsLess with a custom struct type
func runTestAsLessCustomStruct(t *testing.T) {
	t.Helper()
	type version struct {
		major, minor, patch int
	}

	fn := func(a, b version) int {
		if a.major != b.major {
			return a.major - b.major
		}
		if a.minor != b.minor {
			return a.minor - b.minor
		}
		return a.patch - b.patch
	}

	lessFn := cmp.AsLess(fn)

	// Test cases
	v1 := version{1, 0, 0}
	v2 := version{2, 0, 0}
	v3 := version{1, 2, 0}
	v4 := version{1, 3, 0}
	v5 := version{1, 2, 3}
	v6 := version{1, 2, 4}
	v7 := version{1, 9, 9}

	core.AssertTrue(t, lessFn(v1, v2), "lower major version")
	core.AssertTrue(t, lessFn(v3, v4), "same major different minor")
	core.AssertTrue(t, lessFn(v5, v6), "same major and minor different patch")
	core.AssertFalse(t, lessFn(v5, v5), "identical versions")
	core.AssertFalse(t, lessFn(v2, v7), "higher version")
}

// TestAsLessNil confirms that AsLess panics when given a nil function.
func TestAsLessNil(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.AsLess[string](nil)
	}, expectedNilCompFuncErr, "AsLess(nil)")
}

// asEqualTestCase tests the AsEqual function
type asEqualTestCase[T any] struct {
	a, b     T
	cmp      cmp.CompFunc[T]
	name     string
	expected bool
}

func newAsEqualTestCase[T any](name string, a, b T, expected bool, fn cmp.CompFunc[T]) asEqualTestCase[T] {
	return asEqualTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		cmp:      fn,
	}
}

func (tc asEqualTestCase[T]) Name() string {
	return tc.name
}

func (tc asEqualTestCase[T]) Test(t *testing.T) {
	t.Helper()
	equal := cmp.AsEqual(tc.cmp)
	result := equal(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, "AsEqual")
}

// runTestAsEqualWithFloats tests AsEqual with float comparison
func runTestAsEqualWithFloats(t *testing.T) {
	t.Helper()
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

	tests := []asEqualTestCase[float64]{
		newAsEqualTestCase("equal positive", 1.0, 1.0, true, floatCmp),
		newAsEqualTestCase("less than", 1.5, 2.0, false, floatCmp),
		newAsEqualTestCase("greater than", 3.0, 2.5, false, floatCmp),
		newAsEqualTestCase("equal zero", 0.0, 0.0, true, floatCmp),
		newAsEqualTestCase("equal negative", -1.0, -1.0, true, floatCmp),
		newAsEqualTestCase("different signs", -1.0, 1.0, false, floatCmp),
	}

	core.RunTestCases(t, tests)
}

// TestAsEqual ensures that AsEqual correctly converts comparison functions
// to equality predicate functions.
func TestAsEqual(t *testing.T) {
	t.Run("with floats", runTestAsEqualWithFloats)
}

// TestAsEqualNil confirms that AsEqual panics when given a nil function.
func TestAsEqualNil(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.AsEqual[float64](nil)
	}, expectedNilCompFuncErr, "AsEqual(nil)")
}

// asCmpTestCase tests the AsCmp function
type asCmpTestCase[T any] struct {
	a, b     T
	less     cmp.CondFunc[T]
	name     string
	expected int
}

func newAsCmpTestCase[T any](name string, a, b T, expected int, less cmp.CondFunc[T]) asCmpTestCase[T] {
	return asCmpTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		less:     less,
	}
}

func (tc asCmpTestCase[T]) Name() string {
	return tc.name
}

func (tc asCmpTestCase[T]) Test(t *testing.T) {
	t.Helper()
	fn := cmp.AsCmp(tc.less)
	result := fn(tc.a, tc.b)
	core.AssertEqual(t, tc.expected, result, "AsCmp")
}

// runTestAsCmpWithIntegers tests AsCmp with integer less function
func runTestAsCmpWithIntegers(t *testing.T) {
	t.Helper()
	// Create a less-than condition function
	less := func(a, b int) bool {
		return a < b
	}

	tests := []asCmpTestCase[int]{
		newAsCmpTestCase("less than", 5, 10, -1, less),
		newAsCmpTestCase("greater than", 10, 5, 1, less),
		newAsCmpTestCase("equal", 5, 5, 0, less),
		newAsCmpTestCase("with zero", 0, 1, -1, less),
		newAsCmpTestCase("negative numbers", -5, -3, -1, less),
		newAsCmpTestCase("mixed signs", -1, 1, -1, less),
	}

	core.RunTestCases(t, tests)
}

// runTestAsCmpWithStrings tests AsCmp with string less function
func runTestAsCmpWithStrings(t *testing.T) {
	t.Helper()
	// Create a less-than condition function for strings
	less := func(a, b string) bool {
		return a < b
	}

	tests := []asCmpTestCase[string]{
		newAsCmpTestCase("less than", "apple", "banana", -1, less),
		newAsCmpTestCase("greater than", "zebra", "apple", 1, less),
		newAsCmpTestCase("equal", "cherry", "cherry", 0, less),
		newAsCmpTestCase("empty string", "", "something", -1, less),
	}

	core.RunTestCases(t, tests)
}

// TestAsCmp ensures that AsCmp correctly converts a less-than condition function
// to a comparison function for various data types.
func TestAsCmp(t *testing.T) {
	t.Run("with integers", runTestAsCmpWithIntegers)
	t.Run("with strings", runTestAsCmpWithStrings)
	t.Run("with custom struct", runTestAsCmpCustomStruct)
}

// runTestAsCmpCustomStruct tests AsCmp with a custom struct type
func runTestAsCmpCustomStruct(t *testing.T) {
	t.Helper()
	type version struct {
		major, minor, patch int
	}

	// A less condition for semantic versioning
	less := func(a, b version) bool {
		if a.major != b.major {
			return a.major < b.major
		}
		if a.minor != b.minor {
			return a.minor < b.minor
		}
		return a.patch < b.patch
	}

	fn := cmp.AsCmp(less)

	// Test cases
	v1 := version{1, 0, 0}
	v2 := version{2, 0, 0}
	v3 := version{1, 2, 0}
	v4 := version{1, 3, 0}
	v5 := version{1, 2, 3}
	v6 := version{1, 2, 4}

	core.AssertEqual(t, -1, fn(v1, v2), "lower major version")
	core.AssertEqual(t, 1, fn(v2, v1), "higher major version")
	core.AssertEqual(t, -1, fn(v3, v4), "same major different minor")
	core.AssertEqual(t, -1, fn(v5, v6), "same major and minor different patch")
	core.AssertEqual(t, 0, fn(v5, v5), "identical versions")
}

// TestAsCmpNil confirms that AsCmp panics when given a nil function.
func TestAsCmpNil(t *testing.T) {
	core.AssertPanic(t, func() {
		cmp.AsCmp[string](nil)
	}, expectedNilCondFuncErr, "AsCmp(nil)")
}

// asCmpAsLessCycleTestCase tests the AsCmp and AsLess function composition
type asCmpAsLessCycleTestCase[T core.Ordered] struct {
	a, b     T
	name     string
	expected bool
}

func newAsCmpAsLessCycleTestCase[T core.Ordered](name string, a, b T, expected bool) asCmpAsLessCycleTestCase[T] {
	return asCmpAsLessCycleTestCase[T]{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
	}
}

func (tc asCmpAsLessCycleTestCase[T]) Name() string {
	return tc.name
}

func (tc asCmpAsLessCycleTestCase[T]) Test(t *testing.T) {
	t.Helper()
	// Original less function
	less := func(a, b T) bool {
		return a < b
	}

	// Convert to comparison function and back to less function
	backToLess := cmp.AsLess(cmp.AsCmp(less))

	// Function composition should preserve original behaviour
	core.AssertEqual(t, tc.expected, backToLess(tc.a, tc.b), "composed")
	core.AssertEqual(t, tc.expected, less(tc.a, tc.b), "original")
}

// TestAsCmpAsLessCycle confirms that combining AsCmp and AsLess returns
// a function equivalent to the original condition function.
func TestAsCmpAsLessCycle(t *testing.T) {
	tests := []core.TestCase{
		newAsCmpAsLessCycleTestCase("less than", 3, 8, true),
		newAsCmpAsLessCycleTestCase("greater than", 10, 5, false),
		newAsCmpAsLessCycleTestCase("equal", 7, 7, false),
		newAsCmpAsLessCycleTestCase("negative numbers", -5, -2, true),
	}

	core.RunTestCases(t, tests)
}
