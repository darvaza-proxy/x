package testutils_test

import (
	"errors"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// ============================================================================
// FunctionTestCase specific test functions
// ============================================================================

// Functions specific to function testing (not shared with other test files)

// ============================================================================
// FunctionTestCase
// ============================================================================

// GetConstant function for testing FunctionTestCase (no args, returns V)
func GetConstant() int {
	return 42
}

// newGetConstantTestCase creates a test case for GetConstant function
func newGetConstantTestCase(name string, expected int) core.TestCase {
	return testutils.NewFunctionTestCase(name, GetConstant, "GetConstant", expected)
}

// makeTestFunctionTestCaseSuccessCases creates comprehensive success test cases for basic function testing
func makeTestFunctionTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test actual function from common
		newGetConstantTestCase("get constant 42", 42),
		// Test different functions with predictable outputs
		testutils.NewFunctionTestCase("zero function",
			func() int { return 0 }, "zeroFunc", 0),
		testutils.NewFunctionTestCase("negative function",
			func() int { return -100 }, "negFunc", -100),
		testutils.NewFunctionTestCase("string function",
			func() string { return "test" }, "stringFunc", "test"),
		testutils.NewFunctionTestCase("bool function true",
			func() bool { return true }, "boolFunc", true),
		testutils.NewFunctionTestCase("bool function false",
			func() bool { return false }, "boolFunc", false),
	}
}

// makeTestFunctionTestCaseFailureCases creates comprehensive failure test cases for basic function testing
func makeTestFunctionTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong expected value
		newGetConstantTestCase("expects wrong value", 100), // Expects 100 but GetConstant returns 42
	}
}

func TestFunctionTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOneArgTestCase
// ============================================================================

// makeTestFunctionOneArgTestCaseSuccessCases creates success test cases for one-argument function testing
func makeTestFunctionOneArgTestCaseSuccessCases() []core.TestCase {
	doubleNumber := func(n int) int { return n * 2 }
	squareNumber := func(n int) int { return n * n }

	return []core.TestCase{
		// Test double function
		testutils.NewFunctionOneArgTestCase("double 5",
			doubleNumber, "doubleNumber", 5, 10),
		testutils.NewFunctionOneArgTestCase("double 0",
			doubleNumber, "doubleNumber", 0, 0),
		testutils.NewFunctionOneArgTestCase("double negative",
			doubleNumber, "doubleNumber", -3, -6),
		// Test square function
		testutils.NewFunctionOneArgTestCase("square 4",
			squareNumber, "squareNumber", 4, 16),
		testutils.NewFunctionOneArgTestCase("square negative",
			squareNumber, "squareNumber", -5, 25),
	}
}

// makeTestFunctionOneArgTestCaseFailureCases creates failure test cases for one-argument function testing
func makeTestFunctionOneArgTestCaseFailureCases() []core.TestCase {
	doubleNumber := func(n int) int { return n * 2 }

	return []core.TestCase{
		// Test wrong expected value
		testutils.NewFunctionOneArgTestCase("expects wrong value",
			doubleNumber, "doubleNumber", 5, 100), // Expects 100 but double(5) returns 10
	}
}

func TestFunctionOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionTwoArgsTestCase
// ============================================================================

// Functions for testing FunctionTwoArgsTestCase (2 args, returns V)
func AddTwoNumbers(a, b int) int {
	return a + b
}

func ConcatenateStrings(a, b string) string {
	return a + b
}

// newAddTwoNumbersTestCase creates a test case for AddTwoNumbers function
func newAddTwoNumbersTestCase(name string, a, b, expected int) core.TestCase {
	return testutils.NewFunctionTwoArgsTestCase(name, AddTwoNumbers, "AddTwoNumbers", a, b, expected)
}

// newConcatenateStringsTestCase creates a test case for ConcatenateStrings function
func newConcatenateStringsTestCase(name string, a, b, expected string) core.TestCase {
	return testutils.NewFunctionTwoArgsTestCase(name, ConcatenateStrings, "ConcatenateStrings", a, b, expected)
}

// makeTestFunctionTwoArgsTestCaseSuccessCases creates success test cases for two-argument function testing
func makeTestFunctionTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test addition function from common
		newAddTwoNumbersTestCase("add positive numbers", 5, 7, 12),
		newAddTwoNumbersTestCase("add zero", 10, 0, 10),
		newAddTwoNumbersTestCase("add negatives", -5, -3, -8),
		newAddTwoNumbersTestCase("add mixed signs", 10, -4, 6),
		// Test string concatenation function
		newConcatenateStringsTestCase("concatenate hello world", "Hello", "World", "HelloWorld"),
		newConcatenateStringsTestCase("concatenate empty strings", "", "", ""),
		newConcatenateStringsTestCase("concatenate with empty", "test", "", "test"),
	}
}

// makeTestFunctionTwoArgsTestCaseFailureCases creates failure test cases for two-argument function testing
func makeTestFunctionTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong expected values
		newAddTwoNumbersTestCase("expects wrong sum", 5, 7, 100), // Expects 100 but 5+7=12
		newConcatenateStringsTestCase("expects wrong concatenation", "Hello", "World", "WorldHello"),
	}
}

func TestFunctionTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionThreeArgsTestCase
// ============================================================================

// Function for testing FunctionThreeArgsTestCase (3 args, returns V)
func ComplexFunction(a, b, c int) int {
	return a*100 + b*10 + c
}

// newComplexFunctionTestCase creates a test case for ComplexFunction function
func newComplexFunctionTestCase(name string, a, b, c, expected int) core.TestCase {
	return testutils.NewFunctionThreeArgsTestCase(name, ComplexFunction, "ComplexFunction", a, b, c, expected)
}

// makeTestFunctionThreeArgsTestCaseSuccessCases creates success test cases for three-argument function testing
func makeTestFunctionThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test ComplexFunction from common
		newComplexFunctionTestCase("calculate correctly", 1, 2, 3, 123),
		newComplexFunctionTestCase("calculate with zeros", 0, 0, 5, 5),
		newComplexFunctionTestCase("calculate all zeros", 0, 0, 0, 0),
		newComplexFunctionTestCase("calculate negative", 1, 2, -3, 117),
		newComplexFunctionTestCase("calculate large numbers", 9, 8, 7, 987),
	}
}

// makeTestFunctionThreeArgsTestCaseFailureCases creates failure test cases for three-argument function testing
func makeTestFunctionThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong expected values
		newComplexFunctionTestCase("expects wrong calculation", 1, 2, 3, 999),
		newComplexFunctionTestCase("expects wrong zero result", 0, 0, 0, 42),
	}
}

func TestFunctionThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionFourArgsTestCase
// ============================================================================

// Functions for testing FunctionFourArgsTestCase (4 args, returns V)
func FourArgFunction(a, b, c, d int) int {
	return a*1000 + b*100 + c*10 + d
}

func ConcatenateFourStrings(a, b, c, d string) string {
	return a + "-" + b + "-" + c + "-" + d
}

// newFourArgFunctionTestCase creates a test case for FourArgFunction function
func newFourArgFunctionTestCase(name string, a, b, c, d int, expected int) core.TestCase {
	return testutils.NewFunctionFourArgsTestCase(name, FourArgFunction, "FourArgFunction", a, b, c, d, expected)
}

// newConcatenateFourStringsTestCase creates a test case for ConcatenateFourStrings function
func newConcatenateFourStringsTestCase(name string, a, b, c, d, expected string) core.TestCase {
	return testutils.NewFunctionFourArgsTestCase(name, ConcatenateFourStrings, "ConcatenateFourStrings",
		a, b, c, d, expected)
}

// makeTestFunctionFourArgsTestCaseSuccessCases creates success test cases for four-argument function testing
func makeTestFunctionFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test FourArgFunction from common
		newFourArgFunctionTestCase("standard calc", 1, 2, 3, 4, 1234),
		newFourArgFunctionTestCase("with zeros", 0, 0, 0, 5, 5),
		newFourArgFunctionTestCase("all zeros", 0, 0, 0, 0, 0),
		// Test ConcatenateFourStrings function
		newConcatenateFourStringsTestCase("concatenate four strings", "a", "b", "c", "d", "a-b-c-d"),
		newConcatenateFourStringsTestCase("concatenate with empty", "", "x", "", "y", "-x--y"),
		newConcatenateFourStringsTestCase("concatenate all empty", "", "", "", "", "---"),
	}
}

// makeTestFunctionFourArgsTestCaseFailureCases creates failure test cases for four-argument function testing
func makeTestFunctionFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newFourArgFunctionTestCase("wrong calculation", 1, 2, 3, 4, 9999),
		newConcatenateFourStringsTestCase("wrong result", "a", "b", "c", "d", "wrong"),
	}
}

func TestFunctionFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionFiveArgsTestCase
// ============================================================================

// Functions for testing FunctionFiveArgsTestCase (5 args, returns V)
func FiveArgFunction(a, b, c, d, e int) int {
	return a*10000 + b*1000 + c*100 + d*10 + e
}

func ConcatenateFiveStrings(a, b, c, d, e string) string {
	return a + "/" + b + "/" + c + "/" + d + "/" + e
}

// newFiveArgFunctionTestCase creates a test case for FiveArgFunction function
func newFiveArgFunctionTestCase(name string, a, b, c, d, e int, expected int) core.TestCase {
	return testutils.NewFunctionFiveArgsTestCase(name, FiveArgFunction, "FiveArgFunction", a, b, c, d, e, expected)
}

// newConcatenateFiveStringsTestCase creates a test case for ConcatenateFiveStrings function
func newConcatenateFiveStringsTestCase(name string, a, b, c, d, e, expected string) core.TestCase {
	return testutils.NewFunctionFiveArgsTestCase(name, ConcatenateFiveStrings, "ConcatenateFiveStrings",
		a, b, c, d, e, expected)
}

// makeTestFunctionFiveArgsTestCaseSuccessCases creates success test cases for five-argument function testing
func makeTestFunctionFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test FiveArgFunction from common
		newFiveArgFunctionTestCase("standard calc", 1, 2, 3, 4, 5, 12345),
		newFiveArgFunctionTestCase("with zeros", 0, 0, 0, 0, 7, 7),
		newFiveArgFunctionTestCase("all zeros", 0, 0, 0, 0, 0, 0),
		// Test ConcatenateFiveStrings function
		newConcatenateFiveStringsTestCase("concatenate five strings", "a", "b", "c", "d", "e", "a/b/c/d/e"),
		newConcatenateFiveStringsTestCase("concatenate with empty", "", "x", "", "y", "z", "/x//y/z"),
		newConcatenateFiveStringsTestCase("concatenate all empty", "", "", "", "", "", "////"),
	}
}

// makeTestFunctionFiveArgsTestCaseFailureCases creates failure test cases for five-argument function testing
func makeTestFunctionFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newFiveArgFunctionTestCase("wrong calculation", 1, 2, 3, 4, 5, 99999),
		newConcatenateFiveStringsTestCase("wrong result", "a", "b", "c", "d", "e", "wrong"),
	}
}

func TestFunctionFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKTestCase
// ============================================================================

// Function for testing FunctionOKTestCase (no args, returns V, bool)
func GetConstantOK() (int, bool) {
	return 42, true
}

// newGetConstantOKTestCase creates a test case for GetConstantOK function
func newGetConstantOKTestCase(name string, expectedValue int, expectedOK bool) core.TestCase {
	return testutils.NewFunctionOKTestCase(name, GetConstantOK, "GetConstantOK", expectedValue, expectedOK)
}

// makeTestFunctionOKTestCaseSuccessCases creates success test cases for function OK testing
func makeTestFunctionOKTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Test GetConstantOK from common
		newGetConstantOKTestCase("constant ok returns true", 42, true),
		// Test various function return patterns
		testutils.NewFunctionOKTestCase("function returns false",
			func() (int, bool) { return 0, false }, "failFunc", 0, false),
		testutils.NewFunctionOKTestCase("negative with true",
			func() (int, bool) { return -10, true }, "negFunc", -10, true),
		testutils.NewFunctionOKTestCase("string function ok",
			func() (string, bool) { return "test", true }, "stringOKFunc", "test", true),
		testutils.NewFunctionOKTestCase("string function not ok",
			func() (string, bool) { return "", false }, "stringFailFunc", "", false),
	}
}

// makeTestFunctionOKTestCaseFailureCases creates failure test cases for function OK testing
func makeTestFunctionOKTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong ok expectation
		newGetConstantOKTestCase("expects wrong ok", 42, false), // Expects false but GetConstantOK returns true
		// Test wrong value expectation
		newGetConstantOKTestCase("expects wrong value", 100, true), // Expects 100 but GetConstantOK returns 42
	}
}

func TestFunctionOKTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKOneArgTestCase
// ============================================================================

// Functions for testing FunctionOKOneArgTestCase (1 arg, returns V, bool)
func ParseInt(s string) (int, bool) {
	switch s {
	case "zero":
		return 0, true
	case "one":
		return 1, true
	case "two":
		return 2, true
	default:
		return 0, false
	}
}

func FindValue(key string) (string, bool) {
	values := map[string]string{
		"foo":   "bar",
		"hello": "world",
	}
	val, ok := values[key]
	return val, ok
}

// newParseIntTestCase creates a test case for ParseInt function
func newParseIntTestCase(name string, input string, expectedValue int, expectedOK bool) core.TestCase {
	return testutils.NewFunctionOKOneArgTestCase(name, ParseInt, "ParseInt", input, expectedValue, expectedOK)
}

// newFindValueTestCase creates a test case for FindValue function
func newFindValueTestCase(name string, key string, expectedValue string, expectedOK bool) core.TestCase {
	return testutils.NewFunctionOKOneArgTestCase(name, FindValue, "FindValue", key, expectedValue, expectedOK)
}

// makeTestFunctionOKOneArgTestCaseSuccessCases creates success test cases for one-argument function OK testing
func makeTestFunctionOKOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful parsing using dedicated factory
		newParseIntTestCase("parse one", "one", 1, true),
		newParseIntTestCase("parse two", "two", 2, true),
		newParseIntTestCase("parse zero", "zero", 0, true),
		// Failed parsing
		newParseIntTestCase("parse invalid", "invalid", 0, false),
		newParseIntTestCase("parse empty", "", 0, false),
		// Additional edge cases
		newParseIntTestCase("parse unknown word", "unknown", 0, false),
		// FindValue function tests
		newFindValueTestCase("find foo", "foo", "bar", true),
		newFindValueTestCase("find hello", "hello", "world", true),
		newFindValueTestCase("find missing", "missing", "", false),
		newFindValueTestCase("find empty key", "", "", false),
	}
}

// makeTestFunctionOKOneArgTestCaseFailureCases creates failure test cases for one-argument function OK testing
func makeTestFunctionOKOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong value expectation
		newParseIntTestCase("expects wrong value", "one", 10, true), // Expects 10 but ParseInt("one") returns 1
		// Test wrong ok expectation
		newParseIntTestCase("expects wrong ok", "one", 1, false), // Expects false but ParseInt("one") returns true
		// FindValue function failure cases
		newFindValueTestCase("expects wrong value", "foo", "wrong", true),
		newFindValueTestCase("expects wrong ok", "foo", "bar", false),
	}
}

func TestFunctionOKOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKTwoArgsTestCase
// ============================================================================

// Function for testing FunctionOKTwoArgsTestCase (2 args, returns V, bool)
func CheckTwoConditions(a, b bool) (allTrue, anyTrue bool) {
	return a && b, a || b
}

// newCheckTwoConditionsTestCase creates a test case for CheckTwoConditions function
func newCheckTwoConditionsTestCase(name string, a, b bool, expectedAllTrue, expectedAnyTrue bool) core.TestCase {
	return testutils.NewFunctionOKTwoArgsTestCase(name, CheckTwoConditions, "CheckTwoConditions",
		a, b, expectedAllTrue, expectedAnyTrue)
}

// makeTestFunctionOKTwoArgsTestCaseSuccessCases creates success test cases for two-argument function OK testing
func makeTestFunctionOKTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCheckTwoConditionsTestCase("both true", true, true, true, true),
		newCheckTwoConditionsTestCase("one true", true, false, false, true),
		newCheckTwoConditionsTestCase("both false", false, false, false, false),
	}
}

// makeTestFunctionOKTwoArgsTestCaseFailureCases creates failure test cases for two-argument function OK testing
func makeTestFunctionOKTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCheckTwoConditionsTestCase("wrong all expectation", true, true, false, true),
		newCheckTwoConditionsTestCase("wrong any expectation", true, true, true, false),
	}
}

func TestFunctionOKTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKThreeArgsTestCase
// ============================================================================

// Function for testing FunctionOKThreeArgsTestCase (3 args, returns V, bool)
func CheckThreeConditions(a, b, c bool) (allTrue, anyTrue bool) {
	allTrue = a && b && c
	anyTrue = a || b || c
	return allTrue, anyTrue
}

// newCheckThreeConditionsTestCase creates a test case for CheckThreeConditions function
func newCheckThreeConditionsTestCase(name string, a, b, c bool, expectedAllTrue, expectedAnyTrue bool) core.TestCase {
	return testutils.NewFunctionOKThreeArgsTestCase(name, CheckThreeConditions, "CheckThreeConditions",
		a, b, c, expectedAllTrue, expectedAnyTrue)
}

// makeTestFunctionOKThreeArgsTestCaseSuccessCases creates success test cases for three-argument function OK testing
func makeTestFunctionOKThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCheckThreeConditionsTestCase("all true", true, true, true, true, true),
		newCheckThreeConditionsTestCase("one true", true, false, false, false, true),
		newCheckThreeConditionsTestCase("all false", false, false, false, false, false),
	}
}

// makeTestFunctionOKThreeArgsTestCaseFailureCases creates failure test cases for three-argument function OK testing
func makeTestFunctionOKThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCheckThreeConditionsTestCase("wrong all expectation", true, true, true, false, true),
		newCheckThreeConditionsTestCase("wrong any expectation", true, false, false, false, false),
	}
}

func TestFunctionOKThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKFourArgsTestCase
// ============================================================================

// Function for testing FunctionOKFourArgsTestCase (4 args, returns V, bool)
func CheckFourConditions(a, b, c, d bool) (allTrue, anyTrue bool) {
	allTrue = a && b && c && d
	anyTrue = a || b || c || d
	return allTrue, anyTrue
}

// newCheckFourConditionsTestCase creates a test case for CheckFourConditions function
func newCheckFourConditionsTestCase(name string, a, b, c, d bool, expectedAllTrue, expectedAnyTrue bool) core.TestCase {
	return testutils.NewFunctionOKFourArgsTestCase(name, CheckFourConditions, "CheckFourConditions",
		a, b, c, d, expectedAllTrue, expectedAnyTrue)
}

// makeTestFunctionOKFourArgsTestCaseSuccessCases creates success test cases for four-argument function OK testing
func makeTestFunctionOKFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCheckFourConditionsTestCase("all true", true, true, true, true, true, true),
		newCheckFourConditionsTestCase("one true", true, false, false, false, false, true),
		newCheckFourConditionsTestCase("all false", false, false, false, false, false, false),
	}
}

// makeTestFunctionOKFourArgsTestCaseFailureCases creates failure test cases for four-argument function OK testing
func makeTestFunctionOKFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCheckFourConditionsTestCase("wrong all expectation", true, true, true, true, false, true),
		newCheckFourConditionsTestCase("wrong any expectation", true, false, false, false, false, false),
	}
}

func TestFunctionOKFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionOKFiveArgsTestCase
// ============================================================================

// Function for testing FunctionOKFiveArgsTestCase (5 args, returns V, bool)
func CheckFiveConditions(a, b, c, d, e bool) (allTrue, anyTrue bool) {
	allTrue = a && b && c && d && e
	anyTrue = a || b || c || d || e
	return allTrue, anyTrue
}

// newCheckFiveConditionsTestCase creates a test case for CheckFiveConditions function
func newCheckFiveConditionsTestCase(name string, a, b, c, d, e bool,
	expectedAllTrue, expectedAnyTrue bool) core.TestCase {
	return testutils.NewFunctionOKFiveArgsTestCase(name, CheckFiveConditions, "CheckFiveConditions",
		a, b, c, d, e, expectedAllTrue, expectedAnyTrue)
}

// makeTestFunctionOKFiveArgsTestCaseSuccessCases creates success test cases for five-argument function OK testing
func makeTestFunctionOKFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCheckFiveConditionsTestCase("all true", true, true, true, true, true, true, true),
		newCheckFiveConditionsTestCase("one true", true, false, false, false, false, false, true),
		newCheckFiveConditionsTestCase("all false", false, false, false, false, false, false, false),
	}
}

// makeTestFunctionOKFiveArgsTestCaseFailureCases creates failure test cases for five-argument function OK testing
func makeTestFunctionOKFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCheckFiveConditionsTestCase("wrong all expectation", true, true, true, true, true, false, true),
		newCheckFiveConditionsTestCase("wrong any expectation", true, false, false, false, false, false, false),
	}
}

func TestFunctionOKFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionOKFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionOKFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorTestCase
// ============================================================================

// Function for testing FunctionErrorTestCase (no args, returns V, error)
func GetConstantError() (int, error) {
	return 42, nil
}

// newGetConstantErrorTestCase creates a test case for GetConstantError function
func newGetConstantErrorTestCase(name string, expectedValue int, expectError bool,
	errorIs error) core.TestCase {
	return testutils.NewFunctionErrorTestCase(name, GetConstantError, "GetConstantError",
		expectedValue, expectError, errorIs)
}

// makeTestFunctionErrorTestCaseSuccessCases creates success test cases for function error testing
func makeTestFunctionErrorTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error case using dedicated factory
		newGetConstantErrorTestCase("constant no error", 42, false, nil),
		// Function that returns error
		testutils.NewFunctionErrorTestCase("function with error",
			func() (int, error) { return 0, errors.New("failed") },
			"errorFunc", 0, true, nil),
		// Specific error testing
		testutils.NewFunctionErrorTestCase("function with specific error",
			func() (int, error) { return 0, ErrDivisionByZero },
			"specificErrorFunc", 0, true, ErrDivisionByZero),
		// Any error testing (errorIs == nil)
		testutils.NewFunctionErrorTestCase("function any error",
			func() (int, error) { return 0, ErrDivisionByZero },
			"anyErrorFunc", 0, true, nil),
		// Different values no error
		testutils.NewFunctionErrorTestCase("negative no error",
			func() (int, error) { return -5, nil },
			"negFunc", -5, false, nil),
	}
}

// makeTestFunctionErrorTestCaseFailureCases creates failure test cases for function error testing
func makeTestFunctionErrorTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test expecting error but none occurs
		// Expects error but GetConstantError returns nil
		newGetConstantErrorTestCase("expects error but none", 42, true, nil),
		// Test wrong specific error
		testutils.NewFunctionErrorTestCase("wrong error type",
			func() (int, error) { return 0, ErrDivisionByZero },
			"wrongErrorFunc", 0, true, ErrCountNegative),
		// Test expects no error but gets one
		testutils.NewFunctionErrorTestCase("expects no error but gets one",
			func() (int, error) { return 0, errors.New("failed") },
			"unexpectedErrorFunc", 0, false, nil),
	}
}

func TestFunctionErrorTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorOneArgTestCase
// ============================================================================

// Function for testing FunctionErrorOneArgTestCase (1 arg, returns V, error)
func ValidateStringSimple(s string) (bool, error) {
	if s == "" {
		return false, ErrValueEmpty
	}
	return true, nil
}

// newValidateStringSimpleTestCase creates a test case for ValidateStringSimple function
func newValidateStringSimpleTestCase(name string, input string, expectedValue bool,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewFunctionErrorOneArgTestCase(name, ValidateStringSimple, "ValidateStringSimple",
		input, expectedValue, expectError, errorIs)
}

// makeTestFunctionErrorOneArgTestCaseSuccessCases creates success test cases for one-argument function error testing
func makeTestFunctionErrorOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases
		newValidateStringSimpleTestCase("valid string no error", "valid", true, false, nil),
		newValidateStringSimpleTestCase("non-empty string no error", "test", true, false, nil),
		// Error cases
		newValidateStringSimpleTestCase("empty string specific error", "", false, true, ErrValueEmpty),
		newValidateStringSimpleTestCase("empty string any error", "", false, true, nil),
	}
}

// makeTestFunctionErrorOneArgTestCaseFailureCases creates failure test cases for one-argument function error testing
func makeTestFunctionErrorOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong error expectation
		newValidateStringSimpleTestCase("expect error but none", "valid", false, true, nil),
		// Wrong specific error
		newValidateStringSimpleTestCase("wrong error type", "", false, true, ErrCountNegative),
	}
}

func TestFunctionErrorOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorTwoArgsTestCase
// ============================================================================

// Functions for testing FunctionErrorTwoArgsTestCase (2 args, returns V, error)
func DivideNumbers(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

func ValidateLength(s string, minLen int) (bool, error) {
	if minLen < 0 {
		return false, ErrMinLengthNegative
	}
	return len(s) >= minLen, nil
}

// newDivideNumbersTestCase creates a test case for DivideNumbers function
func newDivideNumbersTestCase(name string, a, b, expectedValue int, expectError bool,
	expectedError error) core.TestCase {
	return testutils.NewFunctionErrorTwoArgsTestCase(name, DivideNumbers, "DivideNumbers",
		a, b, expectedValue, expectError, expectedError)
}

// newValidateLengthTestCase creates a test case for ValidateLength function
func newValidateLengthTestCase(name string, s string, minLen int, expectedValue bool,
	expectError bool, expectedError error) core.TestCase {
	return testutils.NewFunctionErrorTwoArgsTestCase(name, ValidateLength, "ValidateLength",
		s, minLen, expectedValue, expectError, expectedError)
}

// makeTestFunctionErrorTwoArgsTestCaseSuccessCases creates success test cases for two-argument function error testing
func makeTestFunctionErrorTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Successful division using dedicated factory
		newDivideNumbersTestCase("divide 10 by 2", 10, 2, 5, false, nil),
		newDivideNumbersTestCase("divide negative", -20, 4, -5, false, nil),
		newDivideNumbersTestCase("divide by 1", 7, 1, 7, false, nil),
		// Division by zero error - specific
		newDivideNumbersTestCase("divide by zero specific error", 5, 0, 0, true, ErrDivisionByZero),
		// Division by zero error - any error (errorIs == nil)
		newDivideNumbersTestCase("divide by zero any error", 10, 0, 0, true, nil),
		// ValidateLength function testing
		newValidateLengthTestCase("validate length valid", "hello", 3, true, false, nil),
		newValidateLengthTestCase("validate length too short", "hi", 5, false, false, nil),
		newValidateLengthTestCase("validate length negative min", "test", -1, false, true, ErrMinLengthNegative),
	}
}

// makeTestFunctionErrorTwoArgsTestCaseFailureCases creates failure test cases for two-argument function error testing
func makeTestFunctionErrorTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Test wrong error type expectation
		newDivideNumbersTestCase("wrong error type", 5, 0, 0, true, ErrCountNegative), // Should be ErrDivisionByZero
		// Test expects error but none
		newDivideNumbersTestCase("expects error but none", 10, 2, 5, true, nil), // Expects error but division succeeds
		// Test expects no error but gets one
		// Expects no error but division by zero
		newDivideNumbersTestCase("expects no error but gets one", 10, 0, 0, false, nil),
	}
}

func TestFunctionErrorTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorThreeArgsTestCase
// ============================================================================

// Function for testing FunctionErrorThreeArgsTestCase (3 args, returns V, error)
func ValidateThreeNumbers(a, b, c int) (int, error) {
	if a < 0 || b < 0 || c < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c, nil
}

// newValidateThreeNumbersTestCase creates a test case for ValidateThreeNumbers function
func newValidateThreeNumbersTestCase(name string, a, b, c, expectedSum int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewFunctionErrorThreeArgsTestCase(name, ValidateThreeNumbers, "ValidateThreeNumbers",
		a, b, c, expectedSum, expectError, errorIs)
}

// makeTestFunctionErrorThreeArgsTestCaseSuccessCases creates success test cases
// for three-argument function error testing
func makeTestFunctionErrorThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newValidateThreeNumbersTestCase("valid numbers no error", 1, 2, 3, 6, false, nil),
		newValidateThreeNumbersTestCase("zero values valid", 0, 0, 0, 0, false, nil),
		newValidateThreeNumbersTestCase("negative first specific error", -1, 2, 3, 0, true, ErrNegativeValues),
		newValidateThreeNumbersTestCase("negative first any error", -1, 2, 3, 0, true, nil),
	}
}

// makeTestFunctionErrorThreeArgsTestCaseFailureCases creates failure test cases
// for three-argument function error testing
func makeTestFunctionErrorThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newValidateThreeNumbersTestCase("expect error but get none", 1, 2, 3, 0, true, nil),
		newValidateThreeNumbersTestCase("wrong specific error", -1, 2, 3, 0, true, ErrCountNegative),
	}
}

func TestFunctionErrorThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorFourArgsTestCase
// ============================================================================

// Function for testing FunctionErrorFourArgsTestCase (4 args, returns V, error)
func ValidateFourNumbers(a, b, c, d int) (int, error) {
	if a < 0 || b < 0 || c < 0 || d < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c + d, nil
}

// newValidateFourNumbersTestCase creates a test case for ValidateFourNumbers function
func newValidateFourNumbersTestCase(name string, a, b, c, d, expectedSum int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewFunctionErrorFourArgsTestCase(name, ValidateFourNumbers, "ValidateFourNumbers",
		a, b, c, d, expectedSum, expectError, errorIs)
}

// makeTestFunctionErrorFourArgsTestCaseSuccessCases creates success test cases for four-argument function error testing
func makeTestFunctionErrorFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newValidateFourNumbersTestCase("valid numbers no error", 1, 2, 3, 4, 10, false, nil),
		newValidateFourNumbersTestCase("zero values valid", 0, 0, 0, 0, 0, false, nil),
		newValidateFourNumbersTestCase("negative first specific error", -1, 2, 3, 4, 0, true, ErrNegativeValues),
		newValidateFourNumbersTestCase("negative number any error", -1, 2, 3, 4, 0, true, nil),
	}
}

// makeTestFunctionErrorFourArgsTestCaseFailureCases creates failure test cases for four-argument function error testing
func makeTestFunctionErrorFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newValidateFourNumbersTestCase("expect error but get none", 1, 2, 3, 4, 0, true, nil),
		newValidateFourNumbersTestCase("wrong specific error", -1, 2, 3, 4, 0, true, ErrCountNegative),
	}
}

func TestFunctionErrorFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// FunctionErrorFiveArgsTestCase
// ============================================================================

// Function for testing FunctionErrorFiveArgsTestCase (5 args, returns V, error)
func ValidateFiveNumbers(a, b, c, d, e int) (int, error) {
	if a < 0 || b < 0 || c < 0 || d < 0 || e < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c + d + e, nil
}

// newValidateFiveNumbersTestCase creates a test case for ValidateFiveNumbers function
func newValidateFiveNumbersTestCase(name string, a, b, c, d, e, expectedSum int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewFunctionErrorFiveArgsTestCase(name, ValidateFiveNumbers, "ValidateFiveNumbers",
		a, b, c, d, e, expectedSum, expectError, errorIs)
}

// makeTestFunctionErrorFiveArgsTestCaseSuccessCases creates success test cases for five-argument function error testing
func makeTestFunctionErrorFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newValidateFiveNumbersTestCase("valid numbers no error", 1, 2, 3, 4, 5, 15, false, nil),
		newValidateFiveNumbersTestCase("zero values valid", 0, 0, 0, 0, 0, 0, false, nil),
		newValidateFiveNumbersTestCase("negative first specific error", -1, 2, 3, 4, 5, 0, true, ErrNegativeValues),
		newValidateFiveNumbersTestCase("negative number any error", -1, 2, 3, 4, 5, 0, true, nil),
	}
}

// makeTestFunctionErrorFiveArgsTestCaseFailureCases creates failure test cases for five-argument function error testing
func makeTestFunctionErrorFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newValidateFiveNumbersTestCase("expect error but get none", 1, 2, 3, 4, 5, 0, true, nil),
		newValidateFiveNumbersTestCase("wrong specific error", -1, 2, 3, 4, 5, 0, true, ErrCountNegative),
	}
}

func TestFunctionErrorFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestFunctionErrorFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestFunctionErrorFiveArgsTestCaseFailureCases())
	})
}
