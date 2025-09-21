package testutils_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// ============================================================================
// FactoryTestCase
// ============================================================================

// newTestStructFactoryTestCase creates a test case for basic factory testing.
func newTestStructFactoryTestCase(name string, expectNil bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryTestCase(name, NewTestStruct, "NewTestStruct", expectNil, typeTest)
}

// makeTestFactoryTestCaseSuccessCases creates comprehensive success test cases for basic factory testing.
func makeTestFactoryTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Non-nil expectations with field validation
		newTestStructFactoryTestCase("create default struct with validation", false, validateTestStruct),
		// Non-nil expectations without validation (base nil check only)
		newTestStructFactoryTestCase("create without validation", false, nil),
		// Nil expectations (testing factories that return nil)
		testutils.NewFactoryTestCase("nil factory returns nil",
			func() *TestStruct { return nil }, "nilFactory", true, nil),
	}
}

// makeTestFactoryTestCaseFailureCases creates comprehensive failure test cases for basic factory testing
func makeTestFactoryTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Expects nil but factory returns non-nil
		newTestStructFactoryTestCase("expects nil but gets non-nil", true, nil),
		// Validation failure
		testutils.NewFactoryTestCase("validation fails",
			func() *TestStruct { return &TestStruct{value: "", count: -1} }, // Invalid struct
			"invalidFactory", false, validateTestStruct),
	}
}

func TestFactoryTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOneArgTestCase
// ============================================================================

// newTestStructFactoryOneArgTestCase creates a test case for one-argument factory testing.
func newTestStructFactoryOneArgTestCase(name string, value string, expectNil bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOneArgTestCase(name, NewTestStructOneArg, "NewTestStructOneArg",
		value, expectNil, typeTest)
}

// makeTestFactoryOneArgTestCaseSuccessCases creates comprehensive success test cases for one-argument factory testing
func makeTestFactoryOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Valid inputs with field validation
		newTestStructFactoryOneArgTestCase("create with custom value", "custom", false, validateTestStruct),
		newTestStructFactoryOneArgTestCase("create with special chars", "!@#$%", false, validateTestStruct),
		newTestStructFactoryOneArgTestCase("create with unicode", "こんにちは", false, validateTestStruct),
		// Valid inputs without validation (base nil check only)
		newTestStructFactoryOneArgTestCase("create with empty value no validation", "", false, nil),
		newTestStructFactoryOneArgTestCase("create with long value", "very-long-value-for-testing", false, nil),
		// Nil expectation scenarios
		testutils.NewFactoryOneArgTestCase("nil factory with arg",
			func(_ string) *TestStruct { return nil }, "nilFactoryOneArg",
			"test", true, nil),
	}
}

// makeTestFactoryOneArgTestCaseFailureCases creates comprehensive failure test cases for one-argument factory testing
func makeTestFactoryOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Validation failure - empty value fails validation
		newTestStructFactoryOneArgTestCase("validation fails empty value", "", false, validateTestStruct),
		// Expects nil but gets non-nil
		newTestStructFactoryOneArgTestCase("expects nil but gets non-nil", "test", true, nil),
	}
}

func TestFactoryOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryTwoArgsTestCase
// ============================================================================

// newTestStructFactoryTwoArgsTestCase creates a test case for two-argument factory testing
func newTestStructFactoryTwoArgsTestCase(name string, value string, count int, expectNil bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryTwoArgsTestCase(name, NewTestStructTwoArgs, "NewTestStructTwoArgs",
		value, count, expectNil, typeTest)
}

// makeTestFactoryTwoArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Valid inputs with field validation using NewTestStructTwoArgs (from common)
		newTestStructFactoryTwoArgsTestCase("create with custom values", "test", 5, false, validateTestStruct),
		newTestStructFactoryTwoArgsTestCase("create with zero count", "value", 0, false, validateTestStruct),
		newTestStructFactoryTwoArgsTestCase("create with large count", "data", 1000, false, validateTestStruct),
		// Valid inputs without validation (base nil check only)
		newTestStructFactoryTwoArgsTestCase("create without validation", "simple", 10, false, nil),
		newTestStructFactoryTwoArgsTestCase("create with empty value no validation", "", -5, false, nil),
		// Nil expectation scenarios
		testutils.NewFactoryTwoArgsTestCase("nil factory with two args",
			func(_, _ int) *TestStruct { return nil }, "nilFactoryTwoArgs",
			1, 2, true, nil),
	}
}

// makeTestFactoryTwoArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Validation failure - empty value fails validation
		newTestStructFactoryTwoArgsTestCase("validation fails empty value", "", 5, false, validateTestStruct),
		// Validation failure - negative count fails validation
		newTestStructFactoryTwoArgsTestCase("validation fails negative count", "test", -1, false, validateTestStruct),
		// Expects nil but gets non-nil
		newTestStructFactoryTwoArgsTestCase("expects nil but gets non-nil", "test", 10, true, nil),
		// Expects non-nil but gets nil
		testutils.NewFactoryTwoArgsTestCase("expects non-nil but gets nil",
			func(_, _ int) *TestStruct { return nil }, "nilFactory", 1, 2, false, nil),
	}
}

func TestFactoryTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryThreeArgsTestCase
// ============================================================================

// newTestStructFactoryThreeArgsTestCase creates a test case for three-argument factory testing
func newTestStructFactoryThreeArgsTestCase(name string, value string, count int, enabled bool,
	expectNil bool, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryThreeArgsTestCase(name, NewTestStructThreeArgs, "NewTestStructThreeArgs",
		value, count, enabled, expectNil, typeTest)
}

// makeTestFactoryThreeArgsTestCaseSuccessCases creates comprehensive success test cases
// for three-argument factory testing
func makeTestFactoryThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Non-nil expectations with field validation
		newTestStructFactoryThreeArgsTestCase("creates instance with all args",
			"test", 5, true, false, validateTestStruct),
		newTestStructFactoryThreeArgsTestCase("creates with disabled",
			"value", 10, false, false, validateTestStruct),
		// Non-nil expectations without validation (base nil check only)
		newTestStructFactoryThreeArgsTestCase("creates with empty value",
			"", 0, false, false, nil),
		newTestStructFactoryThreeArgsTestCase("creates with negative count",
			"test", -5, true, false, nil),
		// Nil expectation
		testutils.NewFactoryThreeArgsTestCase("nil factory three args",
			func(string, int, bool) *TestStruct { return nil }, "nilFactoryThreeArgs",
			"any", 0, false, true, nil),
	}
}

// makeTestFactoryThreeArgsTestCaseFailureCases creates comprehensive failure test cases
// for three-argument factory testing
func makeTestFactoryThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Expects nil but factory returns non-nil
		newTestStructFactoryThreeArgsTestCase("expects nil but gets instance",
			"test", 5, true, true, nil),
		// Validation failure
		newTestStructFactoryThreeArgsTestCase("validation fails empty value",
			"", 5, true, false, validateTestStruct),
		newTestStructFactoryThreeArgsTestCase("validation fails negative count",
			"test", -1, true, false, validateTestStruct),
	}
}

func TestFactoryThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryFourArgsTestCase
// ============================================================================

// newTestStructFactoryFourArgsTestCase creates a test case for four-argument factory testing
func newTestStructFactoryFourArgsTestCase(name string, value string, count int, enabled bool,
	err error, expectNil bool, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryFourArgsTestCase(name, NewTestStructFourArgs, "NewTestStructFourArgs",
		value, count, enabled, err, expectNil, typeTest)
}

// makeTestFactoryFourArgsTestCaseSuccessCases creates comprehensive success test cases
// for four-argument factory testing
func makeTestFactoryFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Non-nil expectations with field validation
		newTestStructFactoryFourArgsTestCase("creates instance with all args",
			"test", 10, true, nil, false, validateTestStruct),
		newTestStructFactoryFourArgsTestCase("creates with error field",
			"value", 5, false, ErrValueCannotBeEmpty, false, validateTestStruct),
		// Non-nil expectations without validation (base nil check only)
		newTestStructFactoryFourArgsTestCase("creates with empty value no validation",
			"", 0, false, nil, false, nil),
		newTestStructFactoryFourArgsTestCase("creates with negative count no validation",
			"test", -1, true, ErrCountNegative, false, nil),
		// Nil expectation
		testutils.NewFactoryFourArgsTestCase("nil factory four args",
			func(string, int, bool, error) *TestStruct { return nil }, "nilFactoryFourArgs",
			"any", 0, false, nil, true, nil),
	}
}

// makeTestFactoryFourArgsTestCaseFailureCases creates comprehensive failure test cases
// for four-argument factory testing
func makeTestFactoryFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Validation failure
		newTestStructFactoryFourArgsTestCase("validation fails empty value",
			"", 10, true, nil, false, validateTestStruct),
		newTestStructFactoryFourArgsTestCase("validation fails negative count",
			"test", -1, false, nil, false, validateTestStruct),
		// Expects nil but gets non-nil
		newTestStructFactoryFourArgsTestCase("expects nil but gets instance",
			"test", 5, true, nil, true, nil),
	}
}

func TestFactoryFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryFiveArgsTestCase
// ============================================================================

// newTestStructFactoryFiveArgsTestCase creates a test case for five-argument factory testing
func newTestStructFactoryFiveArgsTestCase(name string, value string, count int, enabled bool,
	err error, extra int, expectNil bool, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryFiveArgsTestCase(name, NewTestStructFiveArgs, "NewTestStructFiveArgs",
		value, count, enabled, err, extra, expectNil, typeTest)
}

// makeTestFactoryFiveArgsTestCaseSuccessCases creates comprehensive success test cases
// for five-argument factory testing
func makeTestFactoryFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Non-nil expectations with field validation
		newTestStructFactoryFiveArgsTestCase("creates instance with extra",
			"test", 10, true, nil, 5, false, validateTestStruct),
		newTestStructFactoryFiveArgsTestCase("creates with zero extra",
			"value", 1, false, ErrCountNegative, 0, false, validateTestStruct),
		// Non-nil expectations without validation (base nil check only)
		newTestStructFactoryFiveArgsTestCase("creates with negative extra",
			"test", 0, true, nil, -5, false, nil),
		newTestStructFactoryFiveArgsTestCase("creates with large extra",
			"", 5, false, ErrValueEmpty, 1000, false, nil),
		// Nil expectation
		testutils.NewFactoryFiveArgsTestCase("nil factory five args",
			func(string, int, bool, error, int) *TestStruct { return nil }, "nilFactoryFiveArgs",
			"any", 0, false, nil, 0, true, nil),
	}
}

// makeTestFactoryFiveArgsTestCaseFailureCases creates comprehensive failure test cases
// for five-argument factory testing
func makeTestFactoryFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Validation failure - note: NewTestStructFiveArgs adds extra to count, so need to account for that
		newTestStructFactoryFiveArgsTestCase("validation fails empty value",
			"", 10, true, nil, 5, false, validateTestStruct),
		// The count+extra must be >= 0 for validation to pass
		newTestStructFactoryFiveArgsTestCase("validation fails negative final count",
			"test", 1, true, nil, -5, false, validateTestStruct), // 1 + (-5) = -4, fails validation
		// Expects nil but gets non-nil
		newTestStructFactoryFiveArgsTestCase("expects nil but gets instance",
			"test", 10, true, nil, 5, true, nil),
	}
}

func TestFactoryFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKTestCase
// ============================================================================

// makeTestFactoryOKTestCaseSuccessCases creates comprehensive success test cases for factory OK testing
func makeTestFactoryOKTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test NewTestStructOK - returns true with validation
		testutils.NewFactoryOKTestCase("default factory returns true with validation",
			NewTestStructOK, "NewTestStructOK", true, validateTestStruct),
		testutils.NewFactoryOKTestCase("default factory returns true no validation",
			NewTestStructOK, "NewTestStructOK", true, nil),
		// Test NewTestStructOKTrue - returns true
		testutils.NewFactoryOKTestCase("success factory returns true",
			NewTestStructOKTrue, "NewTestStructOKTrue", true, validateTestStruct),
		// Test NewTestStructOKFalse - returns false
		testutils.NewFactoryOKTestCase("false factory returns false",
			NewTestStructOKFalse, "NewTestStructOKFalse", false, nil),
	}
}

// makeTestFactoryOKTestCaseFailureCases creates comprehensive failure test cases for factory OK testing
func makeTestFactoryOKTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but NewTestStructOK returns true
		testutils.NewFactoryOKTestCase("expects false but gets true",
			NewTestStructOK, "NewTestStructOK", false, nil),
		// Wrong ok expectation - expects true but NewTestStructOKFalse returns false
		testutils.NewFactoryOKTestCase("expects true but gets false",
			NewTestStructOKFalse, "NewTestStructOKFalse", true, nil),
		// Validation failure when ok=true
		testutils.NewFactoryOKTestCase("validation fails when ok true",
			func() (*TestStruct, bool) { return &TestStruct{value: "", count: -1}, true }, // Invalid struct but ok=true
			"invalidOKFactory", true, validateTestStruct),
	}
}

func TestFactoryOKTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKOneArgTestCase
// ============================================================================

// newTestStructFactoryOKOneArgTestCase creates a test case for one-argument factory OK testing
func newTestStructFactoryOKOneArgTestCase(name string, value string, expectOK bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOKOneArgTestCase(name, NewTestStructOKOneArg, "NewTestStructOKOneArg",
		value, expectOK, typeTest)
}

// makeTestFactoryOKOneArgTestCaseSuccessCases creates success test cases
func makeTestFactoryOKOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful creation (ok=true) with field validation
		newTestStructFactoryOKOneArgTestCase("valid value returns true", "test", true, validateTestStruct),
		newTestStructFactoryOKOneArgTestCase("unicode value returns true", "こんにちは", true, validateTestStruct),
		newTestStructFactoryOKOneArgTestCase("special chars true", "!@#$", true, validateTestStruct),
		// Successful creation without validation
		newTestStructFactoryOKOneArgTestCase("valid without validation", "simple", true, nil),
		// Failed creation (ok=false)
		newTestStructFactoryOKOneArgTestCase("empty value returns false", "", false, nil),
		// Nil factory returning false
		testutils.NewFactoryOKOneArgTestCase("nil factory returns false",
			func(_ string) (*TestStruct, bool) { return nil, false }, "nilOKFactory",
			"any", false, nil),
	}
}

// makeTestFactoryOKOneArgTestCaseFailureCases creates failure test cases
func makeTestFactoryOKOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but gets true
		newTestStructFactoryOKOneArgTestCase("expects false but gets true", "valid", false, nil),
		// Wrong ok expectation - expects true but gets false
		newTestStructFactoryOKOneArgTestCase("expects true but gets false", "", true, nil),
		// Validation failure when ok=true
		testutils.NewFactoryOKOneArgTestCase("validation fails when ok true",
			func(_ string) (*TestStruct, bool) { return &TestStruct{value: "", count: -1}, true }, // Invalid struct but ok=true
			"invalidOKFactory", "test", true, validateTestStruct),
	}
}

func TestFactoryOKOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKTwoArgsTestCase
// ============================================================================

// newTestStructFactoryOKTwoArgsTestCase creates a test case for two-argument factory OK testing
func newTestStructFactoryOKTwoArgsTestCase(name string, value string, count int, expectOK bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOKTwoArgsTestCase(name, NewTestStructOKTwoArgs, "NewTestStructOKTwoArgs",
		value, count, expectOK, typeTest)
}

// makeTestFactoryOKTwoArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryOKTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful creation (ok=true) with field validation
		newTestStructFactoryOKTwoArgsTestCase("valid args return true", "test", 5, true, validateTestStruct),
		newTestStructFactoryOKTwoArgsTestCase("zero count returns true", "value", 0, true, validateTestStruct),
		newTestStructFactoryOKTwoArgsTestCase("large count returns true", "data", 1000, true, validateTestStruct),
		// Successful creation without validation
		newTestStructFactoryOKTwoArgsTestCase("valid without validation", "simple", 10, true, nil),
		// Failed creation (ok=false) - arguments drive the false return
		newTestStructFactoryOKTwoArgsTestCase("empty value returns false", "", 5, false, nil),
		newTestStructFactoryOKTwoArgsTestCase("negative count returns false", "test", -1, false, nil),
	}
}

// makeTestFactoryOKTwoArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryOKTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but gets true (valid args)
		newTestStructFactoryOKTwoArgsTestCase("expects false but gets true", "valid", 5, false, nil),
		// Wrong ok expectation - expects true but gets false (invalid args)
		newTestStructFactoryOKTwoArgsTestCase("expects true but gets false", "", 5, true, nil),
		newTestStructFactoryOKTwoArgsTestCase("expects true but negative count gives false", "test", -1, true, nil),
	}
}

func TestFactoryOKTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKThreeArgsTestCase
// ============================================================================

// newTestStructFactoryOKThreeArgsTestCase creates a test case for three-argument factory OK testing
func newTestStructFactoryOKThreeArgsTestCase(name string, value string, count int, enabled bool, expectOK bool,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOKThreeArgsTestCase(name, NewTestStructOKThreeArgs, "NewTestStructOKThreeArgs",
		value, count, enabled, expectOK, typeTest)
}

// makeTestFactoryOKThreeArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryOKThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful creation (ok=true) with field validation
		newTestStructFactoryOKThreeArgsTestCase("valid args return true", "test", 5, true, true, validateTestStruct),
		newTestStructFactoryOKThreeArgsTestCase("disabled but valid", "value", 10, false, true, validateTestStruct),
		newTestStructFactoryOKThreeArgsTestCase("zero count valid", "data", 0, true, true, validateTestStruct),
		// Successful creation without validation
		newTestStructFactoryOKThreeArgsTestCase("valid without validation", "simple", 20, false, true, nil),
		// Failed creation (ok=false) - arguments drive the false return
		newTestStructFactoryOKThreeArgsTestCase("empty value returns false", "", 5, true, false, nil),
		newTestStructFactoryOKThreeArgsTestCase("negative count returns false", "test", -1, false, false, nil),
	}
}

// makeTestFactoryOKThreeArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryOKThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but gets true (valid args)
		newTestStructFactoryOKThreeArgsTestCase("expects false but gets true", "valid", 5, true, false, nil),
		// Wrong ok expectation - expects true but gets false (invalid args)
		newTestStructFactoryOKThreeArgsTestCase("expects true but gets false", "", 5, true, true, nil),
		newTestStructFactoryOKThreeArgsTestCase("expects true but negative count gives false", "test", -1, false, true, nil),
	}
}

func TestFactoryOKThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKFourArgsTestCase
// ============================================================================

// newTestStructFactoryOKFourArgsTestCase creates a test case for four-argument factory OK testing
func newTestStructFactoryOKFourArgsTestCase(name string, value string, count int, enabled bool,
	err error, expectOK bool, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOKFourArgsTestCase(name, NewTestStructOKFourArgs, "NewTestStructOKFourArgs",
		value, count, enabled, err, expectOK, typeTest)
}

// makeTestFactoryOKFourArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryOKFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful creation (ok=true) with field validation
		newTestStructFactoryOKFourArgsTestCase("valid args return true", "test", 5, true, nil, true, validateTestStruct),
		newTestStructFactoryOKFourArgsTestCase("disabled but valid", "value", 10, false,
			ErrValueEmpty, true, validateTestStruct),
		newTestStructFactoryOKFourArgsTestCase("zero count valid", "data", 0, true, nil, true, validateTestStruct),
		// Successful creation without validation
		newTestStructFactoryOKFourArgsTestCase("valid without validation", "simple", 20, false, ErrCountNegative, true, nil),
		// Failed creation (ok=false) - arguments drive the false return
		newTestStructFactoryOKFourArgsTestCase("empty value returns false", "", 5, true, nil, false, nil),
		newTestStructFactoryOKFourArgsTestCase("negative count returns false", "test", -1, false, ErrValueEmpty, false, nil),
	}
}

// makeTestFactoryOKFourArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryOKFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but gets true (valid args)
		newTestStructFactoryOKFourArgsTestCase("expects false but gets true", "valid", 5, true, nil, false, nil),
		// Wrong ok expectation - expects true but gets false (invalid args)
		newTestStructFactoryOKFourArgsTestCase("expects true but gets false", "", 5, true, nil, true, nil),
		newTestStructFactoryOKFourArgsTestCase("expects true but negative count gives false",
			"test", -1, false, ErrValueEmpty, true, nil),
	}
}

func TestFactoryOKFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryOKFiveArgsTestCase
// ============================================================================

// newTestStructFactoryOKFiveArgsTestCase creates a test case for five-argument factory OK testing
func newTestStructFactoryOKFiveArgsTestCase(name string, value string, count int, enabled bool,
	err error, extra int, expectOK bool, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryOKFiveArgsTestCase(name, NewTestStructOKFiveArgs, "NewTestStructOKFiveArgs",
		value, count, enabled, err, extra, expectOK, typeTest)
}

// makeTestFactoryOKFiveArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryOKFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful creation (ok=true) with field validation
		newTestStructFactoryOKFiveArgsTestCase("valid args return true", "test", 5, true, nil, 10, true, validateTestStruct),
		newTestStructFactoryOKFiveArgsTestCase("disabled but valid", "value", 10, false,
			ErrValueEmpty, 0, true, validateTestStruct),
		newTestStructFactoryOKFiveArgsTestCase("zero count with positive extra", "data", 0, true,
			nil, 5, true, validateTestStruct),
		// Successful creation without validation
		newTestStructFactoryOKFiveArgsTestCase("valid without validation", "simple", 20, false,
			ErrCountNegative, -10, true, nil),
		// Failed creation (ok=false) - arguments drive the false return
		newTestStructFactoryOKFiveArgsTestCase("empty value returns false", "", 5, true, nil, 10, false, nil),
		newTestStructFactoryOKFiveArgsTestCase("negative count returns false", "test", -1, false,
			ErrValueEmpty, 5, false, nil),
	}
}

// makeTestFactoryOKFiveArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryOKFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong ok expectation - expects false but gets true (valid args)
		newTestStructFactoryOKFiveArgsTestCase("expects false but gets true", "valid", 5, true, nil, 10, false, nil),
		// Wrong ok expectation - expects true but gets false (invalid args)
		newTestStructFactoryOKFiveArgsTestCase("expects true but gets false", "", 5, true, nil, 10, true, nil),
		newTestStructFactoryOKFiveArgsTestCase("expects true but negative count gives false",
			"test", -1, false, ErrValueEmpty, 5, true, nil),
	}
}

func TestFactoryOKFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryOKFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryOKFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorTestCase
// ============================================================================

// newTestStructFactoryErrorTestCase creates a test case for basic factory error testing.
func newTestStructFactoryErrorTestCase(name string, expectError bool, errorIs error,
	typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorTestCase(name, NewTestStructError, "NewTestStructError",
		expectError, errorIs, typeTest)
}

// makeTestFactoryErrorTestCaseSuccessCases creates comprehensive success test cases for basic factory error testing
func makeTestFactoryErrorTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorTestCase("no error case with validation", false, nil, validateTestStruct),
		// No error cases without validation (base nil check only)
		newTestStructFactoryErrorTestCase("no error case no validation", false, nil, nil),
		// Error cases (simulating factories that can return errors)
		testutils.NewFactoryErrorTestCase("error case any error",
			func() (*TestStruct, error) { return nil, ErrValueCannotBeEmpty },
			"errorFactory", true, nil, nil),
		testutils.NewFactoryErrorTestCase("error case specific error",
			func() (*TestStruct, error) { return nil, ErrValueCannotBeEmpty },
			"errorFactory", true, ErrValueCannotBeEmpty, nil),
	}
}

// makeTestFactoryErrorTestCaseFailureCases creates comprehensive failure test cases for basic factory error testing
func makeTestFactoryErrorTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Expects error but none occurs
		newTestStructFactoryErrorTestCase("expects error but none", true, nil, nil),
		// Wrong specific error type
		testutils.NewFactoryErrorTestCase("wrong error type",
			func() (*TestStruct, error) { return nil, ErrValueCannotBeEmpty },
			"errorFactory", true, ErrCountNegative, nil),
		// Expects no error but gets one
		testutils.NewFactoryErrorTestCase("expects no error but gets one",
			func() (*TestStruct, error) { return nil, ErrValueCannotBeEmpty },
			"errorFactory", false, nil, nil),
	}
}

func TestFactoryErrorTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorOneArgTestCase
// ============================================================================

// newTestStructFactoryErrorOneArgTestCase creates a test case for one-argument factory error testing
func newTestStructFactoryErrorOneArgTestCase(name string, value string, expectError bool,
	errorIs error, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorOneArgTestCase(name, NewTestStructErrorOneArg, "NewTestStructErrorOneArg",
		value, expectError, errorIs, typeTest)
}

// makeTestFactoryErrorOneArgTestCaseSuccessCases creates comprehensive success test cases
// for one-argument factory error testing
func makeTestFactoryErrorOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorOneArgTestCase("valid value no error with validation",
			"valid", false, nil, validateTestStruct),
		newTestStructFactoryErrorOneArgTestCase("non-empty value no error",
			"test", false, nil, validateTestStruct),
		// No error cases without validation (base nil check only)
		newTestStructFactoryErrorOneArgTestCase("unicode value no error",
			"こんにちは", false, nil, nil),
		newTestStructFactoryErrorOneArgTestCase("special chars no error",
			"@#$%^&*()", false, nil, nil),
		// Specific error cases
		newTestStructFactoryErrorOneArgTestCase("empty value specific error",
			"", true, ErrValueCannotBeEmpty, nil),
		// Any error cases (errorIs == nil) - important for coverage
		newTestStructFactoryErrorOneArgTestCase("empty value any error",
			"", true, nil, nil),
	}
}

// makeTestFactoryErrorOneArgTestCaseFailureCases creates comprehensive failure test cases
// for one-argument factory error testing
func makeTestFactoryErrorOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong error type
		newTestStructFactoryErrorOneArgTestCase("wrong error type",
			"", true, ErrCountNegative, nil), // Should be ErrValueCannotBeEmpty
		// Expects error but none
		newTestStructFactoryErrorOneArgTestCase("expects error but none",
			"valid", true, nil, nil),
		// Expects no error but gets one
		newTestStructFactoryErrorOneArgTestCase("expects no error but gets one",
			"", false, nil, nil),
	}
}

func TestFactoryErrorOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorTwoArgsTestCase
// ============================================================================

// newTestStructFactoryErrorTwoArgsTestCase creates a test case for two-argument factory error testing
func newTestStructFactoryErrorTwoArgsTestCase(name string, value string, count int,
	expectError bool, errorIs error, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorTwoArgsTestCase(name, NewTestStructErrorTwoArgs, "NewTestStructErrorTwoArgs",
		value, count, expectError, errorIs, typeTest)
}

// makeTestFactoryErrorTwoArgsTestCaseSuccessCases creates comprehensive success test cases
// for two-argument factory error testing
func makeTestFactoryErrorTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorTwoArgsTestCase("valid arguments no error",
			"valid", 5, false, nil, validateTestStruct),
		newTestStructFactoryErrorTwoArgsTestCase("zero count valid",
			"test", 0, false, nil, validateTestStruct),
		// No error cases without validation (base nil check only)
		newTestStructFactoryErrorTwoArgsTestCase("large count no validation",
			"value", 9999, false, nil, nil),
		// Specific error cases
		newTestStructFactoryErrorTwoArgsTestCase("empty value specific error",
			"", 5, true, ErrValueCannotBeEmpty, nil),
		newTestStructFactoryErrorTwoArgsTestCase("negative count specific error",
			"valid", -1, true, ErrCountNegative, nil),
		// Any error cases (errorIs == nil) - important for coverage
		newTestStructFactoryErrorTwoArgsTestCase("empty value any error",
			"", 5, true, nil, nil),
		newTestStructFactoryErrorTwoArgsTestCase("negative count any error",
			"valid", -1, true, nil, nil),
	}
}

// makeTestFactoryErrorTwoArgsTestCaseFailureCases creates comprehensive failure test cases
// for two-argument factory error testing
func makeTestFactoryErrorTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong error type
		newTestStructFactoryErrorTwoArgsTestCase("wrong error type",
			"", 5, true, ErrCountNegative, nil), // Should be ErrValueCannotBeEmpty
		// Expects error but none
		newTestStructFactoryErrorTwoArgsTestCase("expects error but none",
			"valid", 5, true, nil, nil),
		// Expects no error but gets one
		newTestStructFactoryErrorTwoArgsTestCase("expects no error but gets one",
			"", 5, false, nil, nil),
	}
}

func TestFactoryErrorTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorThreeArgsTestCase
// ============================================================================

// newTestStructFactoryErrorThreeArgsTestCase creates a test case for three-argument factory error testing
func newTestStructFactoryErrorThreeArgsTestCase(name string, value string, count int, enabled bool,
	expectError bool, errorIs error, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorThreeArgsTestCase(name, NewTestStructErrorThreeArgs, "NewTestStructErrorThreeArgs",
		value, count, enabled, expectError, errorIs, typeTest)
}

// makeTestFactoryErrorThreeArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryErrorThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorThreeArgsTestCase("valid args no error", "test", 5, true, false, nil, validateTestStruct),
		newTestStructFactoryErrorThreeArgsTestCase("disabled but valid", "value", 10, false, false, nil, validateTestStruct),
		newTestStructFactoryErrorThreeArgsTestCase("zero count valid", "data", 0, true, false, nil, validateTestStruct),
		// No error cases without validation
		newTestStructFactoryErrorThreeArgsTestCase("valid without validation", "simple", 20, false, false, nil, nil),
		// Specific error cases - arguments drive the error return
		newTestStructFactoryErrorThreeArgsTestCase("empty value specific error", "", 5, true,
			true, ErrValueCannotBeEmpty, nil),
		newTestStructFactoryErrorThreeArgsTestCase("negative count specific error", "test", -1,
			false, true, ErrCountNegative, nil),
		// Any error cases (errorIs == nil)
		newTestStructFactoryErrorThreeArgsTestCase("empty value any error", "", 5, true, true, nil, nil),
		newTestStructFactoryErrorThreeArgsTestCase("negative count any error", "test", -1, false, true, nil, nil),
	}
}

// makeTestFactoryErrorThreeArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryErrorThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong error type - expects different error than what arguments produce
		newTestStructFactoryErrorThreeArgsTestCase("wrong error type", "", 5, true, true,
			ErrCountNegative, nil), // Should be ErrValueCannotBeEmpty
		// Expects error but none - valid arguments don't produce error
		newTestStructFactoryErrorThreeArgsTestCase("expects error but none", "valid", 5, true, true, nil, nil),
		// Expects no error but gets one - invalid arguments produce error
		newTestStructFactoryErrorThreeArgsTestCase("expects no error but gets one", "", 5, true, false, nil, nil),
		newTestStructFactoryErrorThreeArgsTestCase("expects no error but negative count gives one",
			"test", -1, false, false, nil, nil),
	}
}

func TestFactoryErrorThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorFourArgsTestCase
// ============================================================================

// newTestStructFactoryErrorFourArgsTestCase creates a test case for four-argument factory error testing
func newTestStructFactoryErrorFourArgsTestCase(name string, value string, count int, enabled bool,
	err error, expectError bool, errorIs error, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorFourArgsTestCase(name, NewTestStructErrorFourArgs, "NewTestStructErrorFourArgs",
		value, count, enabled, err, expectError, errorIs, typeTest)
}

// makeTestFactoryErrorFourArgsTestCaseSuccessCases creates comprehensive success test cases
// for four-argument factory error testing
func makeTestFactoryErrorFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorFourArgsTestCase("valid params no error with validation",
			"test", 5, true, nil, false, nil, validateTestStruct),
		newTestStructFactoryErrorFourArgsTestCase("valid with error field",
			"value", 10, false, ErrValueEmpty, false, nil, validateTestStruct),
		// No error cases without validation (base nil check only)
		newTestStructFactoryErrorFourArgsTestCase("zero count no validation",
			"test", 0, true, nil, false, nil, nil),
		// Specific error cases
		newTestStructFactoryErrorFourArgsTestCase("empty value specific error",
			"", 5, true, nil, true, ErrValueCannotBeEmpty, nil),
		newTestStructFactoryErrorFourArgsTestCase("negative count specific error",
			"test", -1, true, nil, true, ErrCountNegative, nil),
		// Any error cases (errorIs == nil) - important for coverage
		newTestStructFactoryErrorFourArgsTestCase("empty value any error",
			"", 5, true, nil, true, nil, nil),
		newTestStructFactoryErrorFourArgsTestCase("negative count any error",
			"test", -1, false, ErrValueEmpty, true, nil, nil),
	}
}

// makeTestFactoryErrorFourArgsTestCaseFailureCases creates comprehensive failure test cases
// for four-argument factory error testing
func makeTestFactoryErrorFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Expects error but none
		newTestStructFactoryErrorFourArgsTestCase("expects error but none",
			"test", 5, true, nil, true, nil, nil),
		// Wrong specific error type
		newTestStructFactoryErrorFourArgsTestCase("wrong error type",
			"", 5, true, nil, true, ErrCountNegative, nil), // Should be ErrValueCannotBeEmpty
		// Expects no error but gets one
		newTestStructFactoryErrorFourArgsTestCase("expects no error but gets one",
			"", 5, true, nil, false, nil, nil),
	}
}

func TestFactoryErrorFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FactoryErrorFiveArgsTestCase
// ============================================================================

// newTestStructFactoryErrorFiveArgsTestCase creates a test case for five-argument factory error testing
func newTestStructFactoryErrorFiveArgsTestCase(name string, value string, count int, enabled bool, err error, extra int,
	expectError bool, errorIs error, typeTest func(core.T, *TestStruct) bool) core.TestCase {
	return testutils.NewFactoryErrorFiveArgsTestCase(name, NewTestStructErrorFiveArgs, "NewTestStructErrorFiveArgs",
		value, count, enabled, err, extra, expectError, errorIs, typeTest)
}

// makeTestFactoryErrorFiveArgsTestCaseSuccessCases creates success test cases
func makeTestFactoryErrorFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases with field validation
		newTestStructFactoryErrorFiveArgsTestCase("valid args no error", "test", 5, true, nil, 10,
			false, nil, validateTestStruct),
		newTestStructFactoryErrorFiveArgsTestCase("disabled but valid", "value", 10, false,
			ErrValueEmpty, 0, false, nil, validateTestStruct),
		newTestStructFactoryErrorFiveArgsTestCase("zero count with positive extra", "data", 0, true,
			nil, 5, false, nil, validateTestStruct),
		// No error cases without validation
		newTestStructFactoryErrorFiveArgsTestCase("valid without validation", "simple", 20, false,
			ErrCountNegative, -10, false, nil, nil),
		// Specific error cases - arguments drive the error return
		newTestStructFactoryErrorFiveArgsTestCase("empty value specific error", "", 5, true, nil, 10,
			true, ErrValueCannotBeEmpty, nil),
		newTestStructFactoryErrorFiveArgsTestCase("negative count specific error", "test", -1, false,
			ErrValueEmpty, 5, true, ErrCountNegative, nil),
		// Any error cases (errorIs == nil)
		newTestStructFactoryErrorFiveArgsTestCase("empty value any error", "", 5, true, nil, 10, true, nil, nil),
		newTestStructFactoryErrorFiveArgsTestCase("negative count any error", "test", -1, false,
			ErrValueEmpty, 5, true, nil, nil),
	}
}

// makeTestFactoryErrorFiveArgsTestCaseFailureCases creates failure test cases
func makeTestFactoryErrorFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong error type - expects different error than what arguments produce
		newTestStructFactoryErrorFiveArgsTestCase("wrong error type", "", 5, true, nil, 10, true,
			ErrCountNegative, nil), // Should be ErrValueCannotBeEmpty
		// Expects error but none - valid arguments don't produce error
		newTestStructFactoryErrorFiveArgsTestCase("expects error but none", "valid", 5, true, nil, 10, true, nil, nil),
		// Expects no error but gets one - invalid arguments produce error
		newTestStructFactoryErrorFiveArgsTestCase("expects no error but gets one", "", 5, true, nil, 10, false, nil, nil),
		newTestStructFactoryErrorFiveArgsTestCase("expects no error but negative count gives one",
			"test", -1, false, ErrValueEmpty, 5, false, nil, nil),
	}
}

func TestFactoryErrorFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFactoryErrorFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFactoryErrorFiveArgsTestCaseFailureCases())
	})
}
