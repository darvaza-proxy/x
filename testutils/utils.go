package testutils

import (
	"fmt"
	"testing"

	"darvaza.org/core"
)

// Interface validation for TestCase types.
var _ core.TestCase = DummyTestCase{}

// DummyTestCase is a test case implementation that can be configured to pass or fail.
// This is useful for meta-testing and testing custom TestCase implementations.
type DummyTestCase struct {
	TestName   string
	ShouldPass bool
}

// Name returns the test case name.
func (tc DummyTestCase) Name() string { return tc.TestName }

// Test executes the test case, passing or failing based on ShouldPass.
func (tc DummyTestCase) Test(t *testing.T) {
	t.Helper()
	if tc.ShouldPass {
		t.Logf("This test passes")
	} else {
		t.Errorf("This test fails")
	}
}

// NewDummyTestCase creates a test case that passes or fails based on shouldPass.
func NewDummyTestCase(name string, shouldPass bool) *DummyTestCase {
	return &DummyTestCase{
		TestName:   name,
		ShouldPass: shouldPass,
	}
}

// NewPassingTestCase creates a test case that passes, with support for formatted names.
func NewPassingTestCase(format string, args ...any) *DummyTestCase {
	name := fmt.Sprintf(format, args...)
	return NewDummyTestCase(name, true)
}

// NewFailingTestCase creates a test case that fails, with support for formatted names.
func NewFailingTestCase(format string, args ...any) *DummyTestCase {
	name := fmt.Sprintf(format, args...)
	return NewDummyTestCase(name, false)
}

// RunTest executes a single test function and returns whether it passed.
// This is useful for testing TestCase implementations or other test functions.
func RunTest(name string, fn func(*testing.T)) bool {
	internalTest := testing.InternalTest{
		Name: name,
		F:    fn,
	}

	return testing.RunTests(matchStringDummy, []testing.InternalTest{internalTest})
}

// matchStringDummy is a match function for testing.RunTests that
// always matches regardless of the input and used solely by [RunTest].
func matchStringDummy(_, _ string) (bool, error) {
	return true, nil
}

// ArrayTestCaseData defines a test scenario with specific pass/fail expectations.
// This is useful for generating arrays of test cases with controlled behaviour.
type ArrayTestCaseData struct {
	name        string
	failIndexes []int // indices of test cases that should fail (rest will pass)
	totalCount  int   // Total number of test cases in array
}

// Name returns the test scenario name.
func (tc ArrayTestCaseData) Name() string {
	return tc.name
}

// FailIndexes returns the indices of tests that should fail.
func (tc ArrayTestCaseData) FailIndexes() []int {
	return core.SliceCopy(tc.failIndexes)
}

// TotalCount returns the total number of test cases.
func (tc ArrayTestCaseData) TotalCount() int {
	return tc.totalCount
}

// PassIndexes returns the indices of tests that should pass (not in fail list).
// Since `failIndexes` is sorted, this uses an efficient merge-like algorithm.
func (tc ArrayTestCaseData) PassIndexes() []int {
	passIndexes := make([]int, 0, tc.totalCount-len(tc.failIndexes))

	failIdx := 0
	for i := range tc.totalCount {
		// Skip if this index is in the sorted fail list
		if failIdx < len(tc.failIndexes) && tc.failIndexes[failIdx] == i {
			failIdx++
		} else {
			passIndexes = append(passIndexes, i)
		}
	}
	return passIndexes
}

// FailNames returns the names of tests that should fail.
func (tc ArrayTestCaseData) FailNames() []string {
	failNames := make([]string, len(tc.failIndexes))
	for i, idx := range tc.failIndexes {
		failNames[i] = fmt.Sprintf("test%d", idx)
	}
	return failNames
}

// PassNames returns the names of tests that should pass.
func (tc ArrayTestCaseData) PassNames() []string {
	passIndexes := tc.PassIndexes()
	passNames := make([]string, len(passIndexes))
	for i, idx := range passIndexes {
		passNames[i] = fmt.Sprintf("test%d", idx)
	}
	return passNames
}

// Make creates test cases based on the configuration.
func (tc ArrayTestCaseData) Make() []core.TestCase {
	testCases := make([]*DummyTestCase, tc.totalCount)

	// Mark all as passing by default
	for i := range tc.totalCount {
		testName := fmt.Sprintf("test%d", i)
		testCases[i] = NewDummyTestCase(testName, true)
	}

	// Override failing indexes
	for _, failIndex := range tc.failIndexes {
		testCases[failIndex].ShouldPass = false
	}

	return core.SliceAs[*DummyTestCase, core.TestCase](testCases)
}

// UniqueSliceOrderedFn returns a sorted slice with unique elements, optionally filtered by cond.
func UniqueSliceOrderedFn[T core.Ordered](input []T, cond func(v T) bool) []T {
	var s []T

	// Copy input.
	if cond != nil {
		s = core.SliceCopyFn(input, func(_ []T, v T) (T, bool) {
			return v, cond(v)
		})
	} else {
		s = core.SliceCopy(input)
	}

	// Sort.
	core.SliceSortOrdered(s)

	// Unique.
	return core.SliceReplaceFn(s, func(prev []T, v T) (T, bool) {
		if l := len(prev); l > 0 {
			if prev[l-1] == v {
				// Skip duplicate.
				return v, false
			}
		}
		return v, true
	})
}

// sanitiseIndexes returns sanitised fail indices and total count.
// If totalCount is 0, it calculates the total from the maximum index + 1.
func sanitiseIndexes(failIndexes []int, totalCount int) ([]int, int) {
	// Compose filter that considers totalCount when positive, and
	// always excludes negatives.
	filter := func(v int) bool {
		switch {
		case v < 0:
			return false
		case totalCount > 0 && v >= totalCount:
			return false
		default:
			return true
		}
	}

	// Get sorted, unique, valid indices.
	validIndexes := UniqueSliceOrderedFn(failIndexes, filter)

	// Handle totalCount == 0 case.
	if totalCount == 0 {
		if len(validIndexes) == 0 {
			return []int{}, 0
		}
		totalCount = validIndexes[len(validIndexes)-1] + 1
	}

	return validIndexes, totalCount
}

// NewArrayTestCaseData creates a test case for array testing.
// It filters out invalid indices and removes duplicates.
// If totalCount is 0, it calculates the total from the maximum index + 1.
func NewArrayTestCaseData(name string, failIndexes []int, totalCount int) ArrayTestCaseData {
	sanitisedIndexes, sanitisedTotal := sanitiseIndexes(failIndexes, totalCount)

	return ArrayTestCaseData{
		name:        name,
		failIndexes: sanitisedIndexes,
		totalCount:  sanitisedTotal,
	}
}
