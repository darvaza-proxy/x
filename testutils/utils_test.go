package testutils_test

import (
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// Interface validation for TestCase types.
var _ core.TestCase = newArrayTestCaseDataTestCase{}
var _ core.TestCase = arrayTestCaseDataIndexesTestCase{}
var _ core.TestCase = arrayTestCaseDataNamesTestCase{}
var _ core.TestCase = arrayTestCaseDataMakeTestCase{}
var _ core.TestCase = runTestTestCase{}
var _ core.TestCase = uniqueSliceOrderedFnTestCase{}
var _ core.TestCase = dummyTestCaseTestCase{}

// ============================================================================
// ARRAY TEST CASE DATA TESTS
// ============================================================================

// newArrayTestCaseDataTestCase tests the NewArrayTestCaseData factory
type newArrayTestCaseDataTestCase struct {
	name             string
	inputName        string
	inputFailIndexes []int
	inputTotalCount  int
	expectedName     string
	expectedIndexes  []int
	expectedTotal    int
}

func (tc newArrayTestCaseDataTestCase) Name() string { return tc.name }

func (tc newArrayTestCaseDataTestCase) Test(t *testing.T) {
	t.Helper()

	data := testutils.NewArrayTestCaseData(tc.inputName, tc.inputFailIndexes, tc.inputTotalCount)

	core.AssertEqual(t, tc.expectedName, data.Name(), "name")
	core.AssertEqual(t, tc.expectedTotal, data.TotalCount(), "total count")
	core.AssertSliceEqual(t, tc.expectedIndexes, data.FailIndexes(), "fail indexes")
}

func newNewArrayTestCaseDataTestCase(name string, inputFailIndexes []int, inputTotalCount int,
	expectedIndexes []int, expectedTotal int) newArrayTestCaseDataTestCase {
	return newArrayTestCaseDataTestCase{
		name:             name,
		inputName:        name,
		inputFailIndexes: inputFailIndexes,
		inputTotalCount:  inputTotalCount,
		expectedName:     name,
		expectedIndexes:  expectedIndexes,
		expectedTotal:    expectedTotal,
	}
}

func makeTestNewArrayTestCaseDataTestCases() []core.TestCase {
	return []core.TestCase{
		newNewArrayTestCaseDataTestCase("valid input", []int{1, 3}, 5,
			[]int{1, 3}, 5),
		newNewArrayTestCaseDataTestCase("duplicate indexes", []int{1, 3, 1, 3}, 5,
			[]int{1, 3}, 5),
		newNewArrayTestCaseDataTestCase("out of range indexes", []int{-1, 1, 5, 10}, 3,
			[]int{1}, 3),
		newNewArrayTestCaseDataTestCase("empty fail indexes", []int{}, 3,
			[]int{}, 3),
		newNewArrayTestCaseDataTestCase("auto calculate from max", []int{2, 5, 1}, 0,
			[]int{1, 2, 5}, 6),
		newNewArrayTestCaseDataTestCase("empty indexes with zero total", []int{}, 0,
			[]int{}, 0),
	}
}

// TestNewArrayTestCaseData tests the NewArrayTestCaseData constructor
func TestNewArrayTestCaseData(t *testing.T) {
	core.RunTestCases(t, makeTestNewArrayTestCaseDataTestCases())
}

// arrayTestCaseDataIndexesTestCase tests both FailIndexes and PassIndexes methods
type arrayTestCaseDataIndexesTestCase struct {
	data           testutils.ArrayTestCaseData
	expectedFails  []int
	expectedPasses []int
	name           string
}

func (tc arrayTestCaseDataIndexesTestCase) Name() string { return tc.name }

func (tc arrayTestCaseDataIndexesTestCase) Test(t *testing.T) {
	t.Helper()

	failIndexes := tc.data.FailIndexes()
	passIndexes := tc.data.PassIndexes()

	core.AssertSliceEqual(t, tc.expectedFails, failIndexes, "fail indexes")
	core.AssertSliceEqual(t, tc.expectedPasses, passIndexes, "pass indexes")
}

func newArrayTestCaseDataIndexesTestCase(name string, data testutils.ArrayTestCaseData,
	expectedFails, expectedPasses []int) arrayTestCaseDataIndexesTestCase {
	return arrayTestCaseDataIndexesTestCase{
		data:           data,
		expectedFails:  expectedFails,
		expectedPasses: expectedPasses,
		name:           name,
	}
}

// arrayTestCaseDataNamesTestCase tests both FailNames and PassNames methods
type arrayTestCaseDataNamesTestCase struct {
	data           testutils.ArrayTestCaseData
	expectedFails  []string
	expectedPasses []string
	name           string
}

func (tc arrayTestCaseDataNamesTestCase) Name() string { return tc.name }

func (tc arrayTestCaseDataNamesTestCase) Test(t *testing.T) {
	t.Helper()

	failNames := tc.data.FailNames()
	passNames := tc.data.PassNames()

	core.AssertSliceEqual(t, tc.expectedFails, failNames, "fail names")
	core.AssertSliceEqual(t, tc.expectedPasses, passNames, "pass names")
}

func newArrayTestCaseDataNamesTestCase(name string, data testutils.ArrayTestCaseData,
	expectedFails, expectedPasses []string) arrayTestCaseDataNamesTestCase {
	return arrayTestCaseDataNamesTestCase{
		data:           data,
		expectedFails:  expectedFails,
		expectedPasses: expectedPasses,
		name:           name,
	}
}

// makeArrayTestCaseDataIndexesTestCases creates test cases for testing indexes methods
func makeTestArrayTestCaseDataIndexesTestCases() []core.TestCase {
	return []core.TestCase{
		newArrayTestCaseDataIndexesTestCase("mixed pass fail",
			testutils.NewArrayTestCaseData("mixed", []int{1, 3}, 5),
			[]int{1, 3}, []int{0, 2, 4}),
		newArrayTestCaseDataIndexesTestCase("all pass",
			testutils.NewArrayTestCaseData("all pass", []int{}, 3),
			[]int{}, []int{0, 1, 2}),
		newArrayTestCaseDataIndexesTestCase("all fail",
			testutils.NewArrayTestCaseData("all fail", []int{0, 1, 2}, 3),
			[]int{0, 1, 2}, []int{}),
		newArrayTestCaseDataIndexesTestCase("single fail",
			testutils.NewArrayTestCaseData("single fail", []int{1}, 4),
			[]int{1}, []int{0, 2, 3}),
	}
}

// makeArrayTestCaseDataNamesTestCases creates test cases for testing names methods
func makeTestArrayTestCaseDataNamesTestCases() []core.TestCase {
	return []core.TestCase{
		newArrayTestCaseDataNamesTestCase("mixed pass fail",
			testutils.NewArrayTestCaseData("mixed", []int{1, 3}, 5),
			[]string{"test1", "test3"}, []string{"test0", "test2", "test4"}),
		newArrayTestCaseDataNamesTestCase("all pass",
			testutils.NewArrayTestCaseData("all pass", []int{}, 3),
			[]string{}, []string{"test0", "test1", "test2"}),
		newArrayTestCaseDataNamesTestCase("all fail",
			testutils.NewArrayTestCaseData("all fail", []int{0, 1, 2}, 3),
			[]string{"test0", "test1", "test2"}, []string{}),
		newArrayTestCaseDataNamesTestCase("single fail",
			testutils.NewArrayTestCaseData("single fail", []int{1}, 4),
			[]string{"test1"}, []string{"test0", "test2", "test3"}),
	}
}

// TestArrayTestCaseDataIndexes tests the FailIndexes and PassIndexes methods
func TestArrayTestCaseDataIndexes(t *testing.T) {
	core.RunTestCases(t, makeTestArrayTestCaseDataIndexesTestCases())
}

// TestArrayTestCaseDataNames tests the FailNames and PassNames methods
func TestArrayTestCaseDataNames(t *testing.T) {
	core.RunTestCases(t, makeTestArrayTestCaseDataNamesTestCases())
}

// arrayTestCaseDataMakeTestCase tests the Make method
type arrayTestCaseDataMakeTestCase struct {
	name        string
	failIndexes []int
	totalCount  int
	expectPass  []bool // true if test should pass, false if should fail
}

func (tc arrayTestCaseDataMakeTestCase) Name() string { return tc.name }

func (tc arrayTestCaseDataMakeTestCase) Test(t *testing.T) {
	t.Helper()

	data := testutils.NewArrayTestCaseData("test", tc.failIndexes, tc.totalCount)
	testCases := data.Make()

	core.AssertEqual(t, tc.totalCount, len(testCases), "test case count")

	// Verify each test case has the correct pass/fail configuration
	for i, testCase := range testCases {
		stc, ok := core.AssertTypeIs[*testutils.DummyTestCase](t, testCase, "test%d type", i)
		if !ok {
			continue
		}

		expectedName := fmt.Sprintf("test%d", i)
		core.AssertEqual(t, expectedName, stc.TestName, "test%d name", i)
		core.AssertEqual(t, tc.expectPass[i], stc.ShouldPass, "test%d should pass", i)
	}
}

func newArrayTestCaseDataMakeTestCase(name string, failIndexes []int, totalCount int,
	expectPass []bool) arrayTestCaseDataMakeTestCase {
	return arrayTestCaseDataMakeTestCase{
		name:        name,
		failIndexes: failIndexes,
		totalCount:  totalCount,
		expectPass:  expectPass,
	}
}

func makeTestArrayTestCaseDataMakeTestCases() []core.TestCase {
	return []core.TestCase{
		newArrayTestCaseDataMakeTestCase("mixed pass fail", []int{1}, 3, []bool{true, false, true}),
		newArrayTestCaseDataMakeTestCase("all pass", []int{}, 2, []bool{true, true}),
		newArrayTestCaseDataMakeTestCase("all fail", []int{0, 1}, 2, []bool{false, false}),
		newArrayTestCaseDataMakeTestCase("single item pass", []int{}, 1, []bool{true}),
		newArrayTestCaseDataMakeTestCase("single item fail", []int{0}, 1, []bool{false}),
	}
}

// TestArrayTestCaseDataMake tests the Make method
func TestArrayTestCaseDataMake(t *testing.T) {
	core.RunTestCases(t, makeTestArrayTestCaseDataMakeTestCases())
}

// ============================================================================
// DUMMY TEST CASE TESTS
// ============================================================================

// dummyTestCaseTestCase tests DummyTestCase behaviour
type dummyTestCaseTestCase struct {
	name       string
	shouldPass bool
	testName   string
}

func (tc dummyTestCaseTestCase) Name() string { return tc.name }

func (tc dummyTestCaseTestCase) doTest() bool {
	// Create DummyTestCase and test it using the RunTest utility
	stc := testutils.NewDummyTestCase(tc.testName, tc.shouldPass)
	return testutils.RunTest(tc.testName, stc.Test)
}

func (tc dummyTestCaseTestCase) Test(t *testing.T) {
	t.Helper()

	// Create DummyTestCase
	stc := testutils.NewDummyTestCase(tc.testName, tc.shouldPass)

	// Test Name method
	core.AssertEqual(t, tc.testName, stc.Name(), "test case name")

	passed := tc.doTest()
	if tc.shouldPass {
		core.AssertTrue(t, passed, "test should pass when shouldPass=true")
	} else {
		core.AssertFalse(t, passed, "test should fail when shouldPass=false")
	}
}

func newDummyTestCaseTestCase(name, testName string, shouldPass bool) dummyTestCaseTestCase {
	return dummyTestCaseTestCase{
		name:       name,
		shouldPass: shouldPass,
		testName:   testName,
	}
}

func makeTestDummyTestCaseTestCases() []core.TestCase {
	return []core.TestCase{
		newDummyTestCaseTestCase("passing test case", "test_pass", true),
		newDummyTestCaseTestCase("failing test case", "test_fail", false),
		newDummyTestCaseTestCase("empty name passing", "", true),
		newDummyTestCaseTestCase("empty name failing", "", false),
	}
}

// TestDummyTestCase tests DummyTestCase behaviour
func TestDummyTestCase(t *testing.T) {
	core.RunTestCases(t, makeTestDummyTestCaseTestCases())
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

// Test the convenience helper functions
func TestNewPassingTestCase(t *testing.T) {
	tc := testutils.NewPassingTestCase("test %d", 42)

	core.AssertEqual(t, "test 42", tc.Name(), "formatted name")
	core.AssertTrue(t, tc.ShouldPass, "should pass")
}

func TestNewFailingTestCase(t *testing.T) {
	tc := testutils.NewFailingTestCase("error %s", "message")

	core.AssertEqual(t, "error message", tc.Name(), "formatted name")
	core.AssertFalse(t, tc.ShouldPass, "should fail")
}

// runTestTestCase tests the RunTest utility function
type runTestTestCase struct {
	name     string
	testFn   func(*testing.T)
	expected bool
}

func (tc runTestTestCase) Name() string { return tc.name }

func (tc runTestTestCase) Test(t *testing.T) {
	t.Helper()

	result := testutils.RunTest(tc.name, tc.testFn)
	core.AssertEqual(t, tc.expected, result, "RunTest result")
}

func newRunTestTestCase(name string, testFn func(*testing.T), expected bool) runTestTestCase {
	return runTestTestCase{
		name:     name,
		testFn:   testFn,
		expected: expected,
	}
}

// makeTestRunTestTestCases creates test cases for TestRunTest
func makeTestRunTestTestCases() []core.TestCase {
	return []core.TestCase{
		newRunTestTestCase("passing test", func(t *testing.T) { t.Log("pass") }, true),
		newRunTestTestCase("failing test", func(t *testing.T) { t.Error("fail") }, false),
		newRunTestTestCase("fatal test", func(t *testing.T) { t.Fatal("fatal") }, false),
	}
}

// TestRunTest tests the RunTest utility function
func TestRunTest(t *testing.T) {
	core.RunTestCases(t, makeTestRunTestTestCases())
}

// uniqueSliceOrderedFnTestCase tests the UniqueSliceOrderedFn utility
type uniqueSliceOrderedFnTestCase struct {
	name     string
	input    []int
	cond     func(int) bool
	expected []int
}

func (tc uniqueSliceOrderedFnTestCase) Name() string { return tc.name }

func (tc uniqueSliceOrderedFnTestCase) Test(t *testing.T) {
	t.Helper()

	result := testutils.UniqueSliceOrderedFn(tc.input, tc.cond)
	core.AssertSliceEqual(t, tc.expected, result, "unique ordered result")
}

func newUniqueSliceOrderedFnTestCase(name string, input []int, cond func(int) bool,
	expected []int) uniqueSliceOrderedFnTestCase {
	return uniqueSliceOrderedFnTestCase{
		name:     name,
		input:    input,
		cond:     cond,
		expected: expected,
	}
}

// makeTestUniqueSliceOrderedFnTestCases creates test cases for TestUniqueSliceOrderedFn
func makeTestUniqueSliceOrderedFnTestCases() []core.TestCase {
	return []core.TestCase{
		newUniqueSliceOrderedFnTestCase("with condition", []int{3, -1, 2, 3, -2, 1, 2},
			func(v int) bool { return v >= 0 }, []int{1, 2, 3}),
		newUniqueSliceOrderedFnTestCase("without condition", []int{3, 1, 2, 3, 1, 2},
			nil, []int{1, 2, 3}),
		newUniqueSliceOrderedFnTestCase("empty input", []int{}, nil, []int{}),
	}
}

// TestUniqueSliceOrderedFn tests the UniqueSliceOrderedFn utility
func TestUniqueSliceOrderedFn(t *testing.T) {
	core.RunTestCases(t, makeTestUniqueSliceOrderedFnTestCases())
}
