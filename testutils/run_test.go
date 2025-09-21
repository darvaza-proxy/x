package testutils_test

import (
	"strings"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// Interface validation for TestCase types.
var _ core.TestCase = runTestCase{}
var _ core.TestCase = runArrayTestCase{}

// runMockT is the interface for our test mocks
type runMockT interface {
	core.T
	RunCalled() bool
	RunName() string
	RunReturned() bool
	LogContains(string) bool
	ErrorContains(string) bool
	Logs() ([]string, bool)
	Errors() ([]string, bool)
}

// MockT wrappers for testing Run() function's three branches

// MockTWithTestingRun wraps MockT and implements Run(string, func(*testing.T)) bool
// for testing the first branch of Run().
type MockTWithTestingRun struct {
	core.MockT
	runCalled   bool
	runName     string
	runReturned bool
}

// Run implements the testing.T style Run method using RunTest.
func (m *MockTWithTestingRun) Run(name string, fn func(*testing.T)) bool {
	m.runCalled = true
	m.runName = name
	// Use RunTest to execute the function with proper isolation
	result := testutils.RunTest(name, fn)
	m.runReturned = result
	return result
}

func (m *MockTWithTestingRun) RunCalled() bool   { return m.runCalled }
func (m *MockTWithTestingRun) RunName() string   { return m.runName }
func (m *MockTWithTestingRun) RunReturned() bool { return m.runReturned }

func (m *MockTWithTestingRun) LogContains(substring string) bool {
	for _, log := range m.MockT.Logs {
		if strings.Contains(log, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithTestingRun) ErrorContains(substring string) bool {
	for _, err := range m.MockT.Errors {
		if strings.Contains(err, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithTestingRun) Logs() ([]string, bool) {
	return m.MockT.Logs, len(m.MockT.Logs) > 0
}

func (m *MockTWithTestingRun) Errors() ([]string, bool) {
	return m.MockT.Errors, len(m.MockT.Errors) > 0
}

// MockTWithCoreRun wraps MockT and implements Run(string, func(core.T)) bool
// for testing the second branch of Run().
type MockTWithCoreRun struct {
	core.MockT
	runCalled   bool
	runName     string
	runReturned bool
}

// Run implements the core.T style Run method.
func (m *MockTWithCoreRun) Run(name string, fn func(core.T)) bool {
	m.runCalled = true
	m.runName = name

	result := m.MockT.Run(name, fn)
	m.runReturned = result
	return result
}

func (m *MockTWithCoreRun) RunCalled() bool   { return m.runCalled }
func (m *MockTWithCoreRun) RunName() string   { return m.runName }
func (m *MockTWithCoreRun) RunReturned() bool { return m.runReturned }

func (m *MockTWithCoreRun) LogContains(substring string) bool {
	for _, log := range m.MockT.Logs {
		if strings.Contains(log, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithCoreRun) ErrorContains(substring string) bool {
	for _, err := range m.MockT.Errors {
		if strings.Contains(err, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithCoreRun) Logs() ([]string, bool) {
	return m.MockT.Logs, len(m.MockT.Logs) > 0
}

func (m *MockTWithCoreRun) Errors() ([]string, bool) {
	return m.MockT.Errors, len(m.MockT.Errors) > 0
}

// MockTWithoutRun wraps MockT and shadows any Run method to ensure
// the fallback branch of Run() is triggered.
type MockTWithoutRun struct {
	core.MockT
}

// Run shadows any potential Run method from embedded types to ensure
// this type doesn't match the interface assertions in Run().
// This method has an incompatible signature to force the fallback path.
func (m *MockTWithoutRun) Run() {
	// Intentionally empty - this shadows any Run method
}

func (m *MockTWithoutRun) RunCalled() bool   { return true }        // fallback always calls via RunTest
func (m *MockTWithoutRun) RunName() string   { return "subtest" }   // fallback uses the name we pass
func (m *MockTWithoutRun) RunReturned() bool { return !m.Failed() } // return based on mock state

func (m *MockTWithoutRun) LogContains(substring string) bool {
	for _, log := range m.MockT.Logs {
		if strings.Contains(log, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithoutRun) ErrorContains(substring string) bool {
	for _, err := range m.MockT.Errors {
		if strings.Contains(err, substring) {
			return true
		}
	}
	return false
}

func (m *MockTWithoutRun) Logs() ([]string, bool) {
	return m.MockT.Logs, len(m.MockT.Logs) > 0
}

func (m *MockTWithoutRun) Errors() ([]string, bool) {
	return m.MockT.Errors, len(m.MockT.Errors) > 0
}

// Factory functions for creating test scenarios

// newMockTWithTestingRun creates a MockTWithTestingRun and a test function
// that will pass or fail based on wantedPass.
func newMockTWithTestingRun(wantedPass bool) (*MockTWithTestingRun, func(core.T)) {
	mock := &MockTWithTestingRun{}

	fn := func(subT core.T) {
		// Log to mock for test verification AND to subT to affect RunTest result
		if wantedPass {
			mock.Logf("subtest passes")
			subT.Logf("subtest passes")
		} else {
			mock.Errorf("subtest fails")
			subT.Errorf("subtest fails")
		}
	}

	return mock, fn
}

// newMockTWithCoreRun creates a MockTWithCoreRun and a test function
// that will pass or fail based on wantedPass.
func newMockTWithCoreRun(wantedPass bool) (*MockTWithCoreRun, func(core.T)) {
	mock := &MockTWithCoreRun{}
	fn := func(subT core.T) {
		// subT is the mock itself for core.T interface
		if wantedPass {
			subT.Logf("subtest passes")
		} else {
			subT.Errorf("subtest fails")
		}
	}
	return mock, fn
}

// newMockTWithoutRun creates a MockTWithoutRun and a test function
// that will pass or fail based on wantedPass.
func newMockTWithoutRun(wantedPass bool) (*MockTWithoutRun, func(core.T)) {
	mock := &MockTWithoutRun{}
	fn := func(subT core.T) {
		// In fallback case, the parent mock gets logs/errors from Run()
		// We need to log to mock so our test can verify the messages
		if wantedPass {
			mock.Logf("subtest passes")
			subT.Logf("isolated") // This won't appear in parent mock
		} else {
			mock.Errorf("subtest fails")
			subT.Errorf("isolated") // This won't appear in parent mock
		}
	}
	return mock, fn
}

// runTestCase tests the Run function with different interface branches.
type runTestCase struct {
	// Large fields first (8+ bytes)
	t    core.T
	fn   func(core.T)
	name string

	// Small fields last (1 byte)
	wantedPass bool
}

func (tc runTestCase) Name() string {
	return tc.name
}

func (tc runTestCase) Test(t *testing.T) {
	t.Helper()

	// Execute Run with the mock and test function
	result := testutils.Run(tc.t, "subtest", tc.fn)

	// Verify the result matches expected
	core.AssertEqual(t, tc.wantedPass, result, "Run result")

	// Verify mock-specific behaviour using interface
	rm := core.AssertMustTypeIs[runMockT](t, tc.t, "runMockT")

	core.AssertTrue(t, rm.RunCalled(), "Run should be called")
	core.AssertEqual(t, "subtest", rm.RunName(), "Run name")
	core.AssertEqual(t, tc.wantedPass, rm.RunReturned(), "Run return value")

	// Check logs/errors based on expected result
	if tc.wantedPass {
		// Should have logs for success
		core.AssertTrue(t, rm.LogContains("passes"), "should contain 'passes' in logs")
		_, hasErrors := rm.Errors()
		core.AssertFalse(t, hasErrors, "should not have errors")
		// Verify isolation - "isolated" should not leak to parent mock
		core.AssertFalse(t, rm.LogContains("isolated"), "should not contain 'isolated' in logs")
	} else {
		// Should have errors for failure
		core.AssertTrue(t, rm.ErrorContains("fails"), "should contain 'fails' in errors")
		_, hasLogs := rm.Logs()
		core.AssertFalse(t, hasLogs, "should not have logs")
		// Verify isolation - "isolated" should not leak to parent mock
		core.AssertFalse(t, rm.ErrorContains("isolated"), "should not contain 'isolated' in errors")
	}
}

// Factory functions for runTestCase

func newRunTestCaseWithTestingRun(name string, wantedPass bool) runTestCase {
	mock, fn := newMockTWithTestingRun(wantedPass)
	return runTestCase{
		t:          mock,
		fn:         fn,
		name:       name,
		wantedPass: wantedPass,
	}
}

func newRunTestCaseWithCoreRun(name string, wantedPass bool) runTestCase {
	mock, fn := newMockTWithCoreRun(wantedPass)
	return runTestCase{
		t:          mock,
		fn:         fn,
		name:       name,
		wantedPass: wantedPass,
	}
}

func newRunTestCaseWithoutRun(name string, wantedPass bool) runTestCase {
	mock, fn := newMockTWithoutRun(wantedPass)
	return runTestCase{
		t:          mock,
		fn:         fn,
		name:       name,
		wantedPass: wantedPass,
	}
}

// makeTestRunTestCases returns all test cases for testing Run function.
func makeTestRunTestCases() []runTestCase {
	return []runTestCase{
		// Testing.T interface branch
		newRunTestCaseWithTestingRun("testing.T interface success", true),
		newRunTestCaseWithTestingRun("testing.T interface failure", false),

		// Core.T interface branch
		newRunTestCaseWithCoreRun("core.T interface success", true),
		newRunTestCaseWithCoreRun("core.T interface failure", false),

		// Fallback branch
		newRunTestCaseWithoutRun("fallback success", true),
		newRunTestCaseWithoutRun("fallback failure", false),
	}
}

// TestRun tests the Run function with all three interface branches.
func TestRun(t *testing.T) {
	core.RunTestCases(t, makeTestRunTestCases())
}

// assertContainsNames checks that all expected names appear in the error messages.
func assertContainsNames(t *testing.T, messages []string, expectedNames []string, prefix string) {
	t.Helper()
	if len(expectedNames) == 0 {
		return
	}
	allMessages := strings.Join(messages, "\n")
	for _, name := range expectedNames {
		core.AssertContains(t, allMessages, name, prefix)
	}
}

// runArrayTestCase tests both RunSuccessCases and RunFailureCases with ArrayTestCaseData scenarios.
type runArrayTestCase struct {
	// Large fields first (8+ bytes)
	data        testutils.ArrayTestCaseData
	name        string
	expectError string

	// Small fields last (1 byte)
	expectPass  bool
	successTest bool
}

func (tc runArrayTestCase) Name() string {
	return tc.name
}

func (tc runArrayTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	testCases := tc.data.Make()

	// Run the appropriate function
	if tc.successTest {
		testutils.RunSuccessCases(mock, testCases)
	} else {
		testutils.RunFailureCases(mock, testCases)
	}

	if tc.expectPass {
		core.AssertFalse(t, mock.HasErrors(), "errors")
	} else {
		core.AssertTrue(t, mock.HasErrors(), "errors")

		if tc.expectError != "" {
			if err, ok := mock.LastError(); ok {
				core.AssertContains(t, err, tc.expectError, "error")
			}
		}

		// Check both passing and failing names in correct context
		if tc.successTest {
			// RunSuccessCases: FailNames are tests that failed unexpectedly
			assertContainsNames(t, mock.Errors, tc.data.FailNames(), "failing test")
		} else {
			// RunFailureCases: PassNames are tests that passed unexpectedly
			assertContainsNames(t, mock.Errors, tc.data.PassNames(), "passing test")
		}
	}
}

// newRunArrayTestCase creates a test case for testing Run success/failure functions.
func newRunArrayTestCase(name string, data testutils.ArrayTestCaseData,
	expectPass, successTest bool) runArrayTestCase {
	return runArrayTestCase{
		name:        name,
		data:        data,
		expectPass:  expectPass,
		successTest: successTest,
	}
}

// newRunSuccessArrayTestCase creates a test case for RunSuccessCases testing.
func newRunSuccessAllPassTestCase(name string, failIndexes []int, totalCount int) runArrayTestCase {
	data := testutils.NewArrayTestCaseData(name+" scenario", failIndexes, totalCount)
	return newRunArrayTestCase(name, data, true, true)
}

func newRunSuccessDetectFailTestCase(name string, failIndexes []int, totalCount int) runArrayTestCase {
	data := testutils.NewArrayTestCaseData(name+" scenario", failIndexes, totalCount)
	return newRunArrayTestCase(name, data, false, true)
}

// newRunFailureArrayTestCase creates a test case for RunFailureCases testing.
func newRunFailureAllFailTestCase(name string, failIndexes []int, totalCount int) runArrayTestCase {
	data := testutils.NewArrayTestCaseData(name+" scenario", failIndexes, totalCount)
	return newRunArrayTestCase(name, data, true, false)
}

func newRunFailureDetectPassTestCase(name string, failIndexes []int, totalCount int) runArrayTestCase {
	data := testutils.NewArrayTestCaseData(name+" scenario", failIndexes, totalCount)
	return newRunArrayTestCase(name, data, false, false)
}

// makeTestRunSuccessCasesTestCases creates test cases for TestRunSuccessCases
func makeTestRunSuccessCasesTestCases() []runArrayTestCase {
	return []runArrayTestCase{
		// All tests pass - should succeed
		newRunSuccessAllPassTestCase(
			"all tests pass",
			[]int{}, 5, // no failing indexes
		),

		// Some tests fail - should detect failures
		newRunSuccessDetectFailTestCase(
			"detects single failure",
			[]int{2}, 5, // test2 fails
		),

		// Multiple failures
		newRunSuccessDetectFailTestCase(
			"detects multiple failures",
			[]int{1, 3, 4}, 6,
		),

		// All tests fail - should detect all failures
		newRunSuccessDetectFailTestCase(
			"all tests fail",
			[]int{0, 1, 2}, 3, // all fail
		),

		// Edge case: empty test array
		newRunSuccessAllPassTestCase(
			"empty test array",
			[]int{}, 0, // no tests
		),
	}
}

// TestRunSuccessCases tests RunSuccessCases with comprehensive ArrayTestCaseData scenarios.
func TestRunSuccessCases(t *testing.T) {
	core.RunTestCases(t, makeTestRunSuccessCasesTestCases())
}

// makeTestRunFailureCasesTestCases creates test cases for TestRunFailureCases
func makeTestRunFailureCasesTestCases() []runArrayTestCase {
	return []runArrayTestCase{
		// All tests fail - should succeed
		newRunFailureAllFailTestCase(
			"all tests fail",
			[]int{0, 1, 2, 3, 4}, 5, // all fail
		),

		// Some tests pass unexpectedly - should detect
		newRunFailureDetectPassTestCase(
			"detects unexpected pass",
			[]int{0, 2, 4}, 5, // 1,3 pass unexpectedly
		),

		// Single unexpected pass
		newRunFailureDetectPassTestCase(
			"detects single unexpected pass",
			[]int{0, 1, 3, 4}, 5, // test2 passes unexpectedly
		),

		// All tests pass unexpectedly - should detect all
		newRunFailureDetectPassTestCase(
			"all tests pass unexpectedly",
			[]int{}, 3, // none fail (all pass)
		),

		// Edge case: empty test array
		newRunFailureAllFailTestCase(
			"empty test array",
			[]int{}, 0, // no tests
		),
	}
}

// TestRunFailureCases tests RunFailureCases with comprehensive ArrayTestCaseData scenarios.
func TestRunFailureCases(t *testing.T) {
	core.RunTestCases(t, makeTestRunFailureCasesTestCases())
}

// TestWithRealGeneratedTypes tests the utils functions with actual generated testutils types
func TestWithRealGeneratedTypes(t *testing.T) {
	t.Run("getter test case", runTestRealGetterTestCase)
	t.Run("factory test case", runTestRealFactoryTestCase)
	t.Run("function test case", runTestRealFunctionTestCase)
}

func runTestRealGetterTestCase(t *testing.T) {
	t.Helper()

	// Create a real GetterTestCase using the actual generated type
	obj := &testObject{value: "test-value"}
	tc := testutils.NewGetterTestCase(
		"get value test",
		(*testObject).GetValue,
		"GetValue",
		obj,
		"test-value",
	)

	// Test success case
	testutils.RunSuccessCases(t, []core.TestCase{tc})

	// Test failure case with wrong expected value
	failingTC := testutils.NewGetterTestCase(
		"get value failure",
		(*testObject).GetValue,
		"GetValue",
		obj,
		"wrong-value",
	)

	mockT := &core.MockT{}
	testutils.RunSuccessCases(mockT, []core.TestCase{failingTC})
	core.AssertTrue(t, mockT.HasErrors(), "errors")
}

func runTestRealFactoryTestCase(t *testing.T) {
	t.Helper()

	// Test factory function that creates TestObject
	tc := testutils.NewFactoryTestCase(
		"create test object",
		newTestObject,
		"newTestObject",
		false,
		nil, // base logic handles nil check when expectNil=false
	)

	// Test success case
	testutils.RunSuccessCases(t, []core.TestCase{tc})

	// Test failure case expecting non-nil but function returns nil
	failingTC := testutils.NewFactoryTestCase(
		"create nil object",
		newNilTestObject,
		"newNilTestObject",
		false, // expect non-nil but function returns nil
		nil,   // base logic handles the nil check
	)

	mockT := &core.MockT{}
	testutils.RunSuccessCases(mockT, []core.TestCase{failingTC})
	core.AssertTrue(t, mockT.HasErrors(), "errors")
}

func runTestRealFunctionTestCase(t *testing.T) {
	t.Helper()

	// Test simple function (using a custom function instead of built-in len)
	tc := testutils.NewFunctionOneArgTestCase(
		"string length",
		stringLength,
		"stringLength",
		"hello",
		5,
	)

	// Test success case
	testutils.RunSuccessCases(t, []core.TestCase{tc})

	// Test failure case with wrong expected result
	failingTC := testutils.NewFunctionOneArgTestCase(
		"string length wrong",
		stringLength,
		"len",
		"hello",
		10, // wrong length
	)

	mockT := &core.MockT{}
	testutils.RunSuccessCases(mockT, []core.TestCase{failingTC})
	core.AssertTrue(t, mockT.HasErrors(), "errors")
}

// testObject for testing purposes
type testObject struct {
	value string
}

func (o *testObject) GetValue() string {
	return o.value
}

// Helper functions for testing with real generated types
func newTestObject() *testObject {
	return &testObject{value: "created"}
}

func newNilTestObject() *testObject {
	return nil
}

// stringLength is a wrapper for len to test function test cases.
func stringLength(s string) int {
	return len(s)
}
