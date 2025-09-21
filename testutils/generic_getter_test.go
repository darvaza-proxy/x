package testutils_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// ============================================================================
// GetterTestCase
// ============================================================================

// newTestStructGetValueTestCase creates a test case for TestStruct.GetValue
func newTestStructGetValueTestCase(name string, ts *TestStruct, expected string) core.TestCase {
	return testutils.NewGetterTestCase(name, (*TestStruct).GetValue, "GetValue", ts, expected)
}

// newTestStructGetCountTestCase creates a test case for TestStruct.GetCount
func newTestStructGetCountTestCase(name string, ts *TestStruct, expected int) core.TestCase {
	return testutils.NewGetterTestCase(name, (*TestStruct).GetCount, "GetCount", ts, expected)
}

// newTestStructIsEnabledTestCase creates a test case for TestStruct.IsEnabled
func newTestStructIsEnabledTestCase(name string, ts *TestStruct, expected bool) core.TestCase {
	return testutils.NewGetterTestCase(name, (*TestStruct).IsEnabled, "IsEnabled", ts, expected)
}

// makeTestGetterTestCaseSuccessCases creates success test cases for basic getter methods.
func makeTestGetterTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// String getter edge cases
		newTestStructGetValueTestCase("get value normal", &TestStruct{value: "hello"}, "hello"),
		newTestStructGetValueTestCase("get value empty", &TestStruct{value: ""}, ""),
		newTestStructGetValueTestCase("get value with spaces", &TestStruct{value: "hello world"}, "hello world"),
		newTestStructGetValueTestCase("get value with special chars", &TestStruct{value: "@#$%^&*()"}, "@#$%^&*()"),
		newTestStructGetValueTestCase("get value unicode", &TestStruct{value: "こんにちは"}, "こんにちは"),
		// Integer getter edge cases
		newTestStructGetCountTestCase("get count positive", &TestStruct{count: 42}, 42),
		newTestStructGetCountTestCase("get count zero", &TestStruct{count: 0}, 0),
		newTestStructGetCountTestCase("get count negative", &TestStruct{count: -5}, -5),
		newTestStructGetCountTestCase("get count max int", &TestStruct{count: 2147483647}, 2147483647),
		newTestStructGetCountTestCase("get count min int", &TestStruct{count: -2147483648}, -2147483648),
		// Boolean getter edge cases
		newTestStructIsEnabledTestCase("is enabled true", &TestStruct{enabled: true}, true),
		newTestStructIsEnabledTestCase("is enabled false", &TestStruct{enabled: false}, false),
	}
}

// makeTestGetterTestCaseFailureCases creates failure test cases for basic getter methods
func makeTestGetterTestCaseFailureCases() []core.TestCase {
	// Named instances for clarity
	helloInstance := &TestStruct{value: "hello", count: 42, enabled: true}
	emptyInstance := &TestStruct{value: "", count: 0, enabled: false}
	negativeInstance := &TestStruct{value: "test", count: -10, enabled: true}

	// Comprehensive wrong expectation scenarios
	return []core.TestCase{
		// String getter wrong expectations
		newTestStructGetValueTestCase("wrong value expectation", helloInstance, "wrong"),
		newTestStructGetValueTestCase("expect empty but get value", helloInstance, ""),
		newTestStructGetValueTestCase("expect value but get empty", emptyInstance, "expected"),
		// Integer getter wrong expectations
		newTestStructGetCountTestCase("wrong count expectation", helloInstance, 100),
		newTestStructGetCountTestCase("expect zero but get positive", helloInstance, 0),
		newTestStructGetCountTestCase("expect positive but get negative", negativeInstance, 10),
		// Boolean getter wrong expectations
		newTestStructIsEnabledTestCase("expect true but get false", emptyInstance, true),
		newTestStructIsEnabledTestCase("expect false but get true", helloInstance, false),
	}
}

func TestGetterTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOneArgTestCase
// ============================================================================

// newTestStructGetValueWithArgTestCase creates a test case for TestStruct.GetValueWithArg
func newTestStructGetValueWithArgTestCase(name string, ts *TestStruct, suffix, expected string) core.TestCase {
	return testutils.NewGetterOneArgTestCase(name, (*TestStruct).GetValueWithArg, "GetValueWithArg", ts, suffix, expected)
}

// makeTestGetterOneArgTestCaseSuccessCases creates success test cases for GetValueWithArg
func makeTestGetterOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Standard concatenation scenarios
		newTestStructGetValueWithArgTestCase("append suffix",
			&TestStruct{value: "hello"}, " world", "hello world"),
		newTestStructGetValueWithArgTestCase("append to empty",
			&TestStruct{value: ""}, "test", "test"),
		newTestStructGetValueWithArgTestCase("append empty string",
			&TestStruct{value: "base"}, "", "base"),
		// Edge case scenarios
		newTestStructGetValueWithArgTestCase("append special chars",
			&TestStruct{value: "prefix"}, "@#$", "prefix@#$"),
		newTestStructGetValueWithArgTestCase("append unicode",
			&TestStruct{value: "hello"}, "こんにちは", "helloこんにちは"),
		newTestStructGetValueWithArgTestCase("both empty",
			&TestStruct{value: ""}, "", ""),
	}
}

// makeTestGetterOneArgTestCaseFailureCases creates failure test cases for GetValueWithArg
func makeTestGetterOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Wrong concatenation expectations
		newTestStructGetValueWithArgTestCase("wrong concatenation",
			&TestStruct{value: "hello"}, " world", "goodbye world"),
		newTestStructGetValueWithArgTestCase("wrong suffix expectation",
			&TestStruct{value: "test"}, "ing", "wrong result"),
		newTestStructGetValueWithArgTestCase("wrong value expectation",
			&TestStruct{value: "foo"}, "bar", "different"),
		// Edge case wrong expectations
		newTestStructGetValueWithArgTestCase("expect empty but get concatenated",
			&TestStruct{value: "base"}, "suffix", ""),
		newTestStructGetValueWithArgTestCase("wrong order expectation",
			&TestStruct{value: "Hello"}, "world", "worldHello"),
	}
}

func TestGetterOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// GetterTwoArgsTestCase
// ============================================================================

// newCombineValuesTestCase creates a test case for the CombineValues method.
func newCombineValuesTestCase(name string, ts *TestStruct, a, b, expected string) core.TestCase {
	return testutils.NewGetterTwoArgsTestCase(name, (*TestStruct).CombineValues, "CombineValues", ts, a, b, expected)
}

// makeTestGetterTwoArgsTestCaseSuccessCases creates success test cases for CombineValues
func makeTestGetterTwoArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCombineValuesTestCase("combine with prefix", &TestStruct{value: "prefix:"}, "a", "b", "prefix:ab"),
		newCombineValuesTestCase("combine with empty prefix", &TestStruct{value: ""}, "x", "y", "xy"),
		newCombineValuesTestCase("combine empty strings", &TestStruct{value: "base"}, "", "", "base"),
		newCombineValuesTestCase("combine special chars", &TestStruct{value: "test"}, "@#$", "%^&", "test@#$%^&"),
		newCombineValuesTestCase("combine unicode chars", &TestStruct{value: "hello"}, "こんにちは", "世界", "helloこんにちは世界"),
		newCombineValuesTestCase("combine with spaces", &TestStruct{value: "base"}, " start", " end", "base start end"),
		newCombineValuesTestCase("combine numbers as strings", &TestStruct{value: "num"}, "123", "456", "num123456"),
		newCombineValuesTestCase("combine long strings", &TestStruct{value: "very-long-prefix"},
			"very-long-middle", "very-long-suffix", "very-long-prefixvery-long-middlevery-long-suffix"),
	}
}

// makeTestGetterTwoArgsTestCaseFailureCases creates failure test cases for CombineValues
func makeTestGetterTwoArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCombineValuesTestCase("wrong combination", &TestStruct{value: "prefix:"}, "a", "b", "wrong"),
		newCombineValuesTestCase("expects empty", &TestStruct{value: "test"}, "1", "2", ""),
		newCombineValuesTestCase("wrong order expectation", &TestStruct{value: "base"}, "first", "second", "secondfirstbase"),
		newCombineValuesTestCase("wrong prefix expectation", &TestStruct{value: "test"}, "a", "b", "wrongab"),
		newCombineValuesTestCase("partial result wrong", &TestStruct{value: "prefix"}, "a", "b", "prefix-a-b"),
	}
}

func TestGetterTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterThreeArgsTestCase
// ============================================================================

// newComplexMethodTestCase creates a test case for the ComplexMethod.
func newComplexMethodTestCase(name string, ts *TestStruct, a, b, c, expected string) core.TestCase {
	return testutils.NewGetterThreeArgsTestCase(name, (*TestStruct).ComplexMethod, "ComplexMethod", ts, a, b, c, expected)
}

// makeTestGetterThreeArgsTestCaseSuccessCases creates success test cases for ComplexMethod
func makeTestGetterThreeArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newComplexMethodTestCase("normal format", &TestStruct{value: "test"}, "a", "b", "c", "test:a+b=c"),
		newComplexMethodTestCase("empty value", &TestStruct{value: ""}, "x", "y", "z", ":x+y=z"),
		newComplexMethodTestCase("with spaces", &TestStruct{value: "foo"}, "1", "2", "3", "foo:1+2=3"),
		newComplexMethodTestCase("special characters", &TestStruct{value: "base"}, "@", "#", "$", "base:@+#=$"),
		newComplexMethodTestCase("unicode characters", &TestStruct{value: "こんにちは"}, "世", "界", "日", "こんにちは:世+界=日"),
		newComplexMethodTestCase("empty arguments", &TestStruct{value: "prefix"}, "", "", "", "prefix:+="),
		newComplexMethodTestCase("mixed empty args", &TestStruct{value: "test"}, "a", "", "c", "test:a+=c"),
		newComplexMethodTestCase("long values", &TestStruct{value: "very-long-base-value"},
			"very-long-arg1", "very-long-arg2", "very-long-arg3",
			"very-long-base-value:very-long-arg1+very-long-arg2=very-long-arg3"),
	}
}

// makeTestGetterThreeArgsTestCaseFailureCases creates failure test cases for ComplexMethod
func makeTestGetterThreeArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newComplexMethodTestCase("wrong format expectation", &TestStruct{value: "test"}, "a", "b", "c", "wrong"),
		newComplexMethodTestCase("wrong prefix", &TestStruct{value: "foo"}, "x", "y", "z", "bar:x+y=z"),
		newComplexMethodTestCase("wrong separator", &TestStruct{value: "test"}, "a", "b", "c", "test-a+b=c"),
		newComplexMethodTestCase("wrong operator", &TestStruct{value: "test"}, "a", "b", "c", "test:a-b=c"),
		newComplexMethodTestCase("wrong equals", &TestStruct{value: "test"}, "a", "b", "c", "test:a+b:c"),
	}
}

func TestGetterThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterFourArgsTestCase
// ============================================================================

// newCombineFourValuesTestCase creates a test case for the CombineFourValues method.
func newCombineFourValuesTestCase(name string, ts *TestStruct, a, b, c, d,
	expected string) core.TestCase {
	return testutils.NewGetterFourArgsTestCase(name, (*TestStruct).CombineFourValues,
		"CombineFourValues", ts, a, b, c, d, expected)
}

// makeTestGetterFourArgsTestCaseSuccessCases creates success test cases for CombineFourValues
func makeTestGetterFourArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCombineFourValuesTestCase("normal format", &TestStruct{value: "base"}, "a", "b", "c", "d", "base:a-b-c-d"),
		newCombineFourValuesTestCase("empty base", &TestStruct{value: ""}, "1", "2", "3", "4", ":1-2-3-4"),
		newCombineFourValuesTestCase("special chars", &TestStruct{value: "!"}, "@", "#", "$", "%", "!:@-#-$-%"),
		newCombineFourValuesTestCase("unicode chars", &TestStruct{value: "日"}, "本", "語", "で", "す", "日:本-語-で-す"),
		newCombineFourValuesTestCase("empty arguments", &TestStruct{value: "prefix"}, "", "", "", "", "prefix:---"),
		newCombineFourValuesTestCase("mixed empty args", &TestStruct{value: "test"}, "a", "", "c", "", "test:a--c-"),
		newCombineFourValuesTestCase("numbers as strings", &TestStruct{value: "num"}, "1", "2", "3", "4", "num:1-2-3-4"),
		newCombineFourValuesTestCase("long values", &TestStruct{value: "very-long-base"},
			"very-long-a", "very-long-b", "very-long-c", "very-long-d",
			"very-long-base:very-long-a-very-long-b-very-long-c-very-long-d"),
	}
}

// makeTestGetterFourArgsTestCaseFailureCases creates failure test cases for CombineFourValues
func makeTestGetterFourArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCombineFourValuesTestCase("wrong format", &TestStruct{value: "base"}, "a", "b", "c", "d", "wrong"),
		newCombineFourValuesTestCase("wrong value", &TestStruct{value: "x"}, "a", "b", "c", "d", "wrong:a-b-c-d"),
		newCombineFourValuesTestCase("wrong separator", &TestStruct{value: "base"}, "a", "b", "c", "d", "base:a+b+c+d"),
		newCombineFourValuesTestCase("wrong prefix", &TestStruct{value: "test"}, "a", "b", "c", "d", "wrong:a-b-c-d"),
		newCombineFourValuesTestCase("partial match wrong", &TestStruct{value: "base"}, "a", "b", "c", "d", "base:a-b-c"),
	}
}

func TestGetterFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterFiveArgsTestCase
// ============================================================================

// newCombineFiveValuesTestCase creates a test case for the CombineFiveValues method.
func newCombineFiveValuesTestCase(name string, ts *TestStruct, a, b, c, d, e,
	expected string) core.TestCase {
	return testutils.NewGetterFiveArgsTestCase(name, (*TestStruct).CombineFiveValues,
		"CombineFiveValues", ts, a, b, c, d, e, expected)
}

// makeTestGetterFiveArgsTestCaseSuccessCases creates success test cases for CombineFiveValues
func makeTestGetterFiveArgsTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newCombineFiveValuesTestCase("normal format", &TestStruct{value: "root"}, "a", "b", "c", "d", "e", "root:a/b/c/d/e"),
		newCombineFiveValuesTestCase("empty root", &TestStruct{value: ""}, "1", "2", "3", "4", "5", ":1/2/3/4/5"),
		newCombineFiveValuesTestCase("numbers", &TestStruct{value: "x"}, "1", "2", "3", "4", "5", "x:1/2/3/4/5"),
		newCombineFiveValuesTestCase("special characters", &TestStruct{value: "base"},
			"@", "#", "$", "%", "^", "base:@/#/$/%/^"),
		newCombineFiveValuesTestCase("unicode path", &TestStruct{value: "ファイル"},
			"パス", "テス", "ト", "デー", "タ", "ファイル:パス/テス/ト/デー/タ"),
		newCombineFiveValuesTestCase("all empty arguments", &TestStruct{value: "base"}, "", "", "", "", "", "base:////"),
		newCombineFiveValuesTestCase("path like structure", &TestStruct{value: "usr"},
			"local", "bin", "test", "app", "exe", "usr:local/bin/test/app/exe"),
		newCombineFiveValuesTestCase("mixed types as strings", &TestStruct{value: "data"},
			"123", "true", "null", "3.14", "end", "data:123/true/null/3.14/end"),
	}
}

// makeTestGetterFiveArgsTestCaseFailureCases creates failure test cases for CombineFiveValues
func makeTestGetterFiveArgsTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newCombineFiveValuesTestCase("wrong format", &TestStruct{value: "root"}, "a", "b", "c", "d", "e", "wrong"),
		newCombineFiveValuesTestCase("wrong value", &TestStruct{value: "x"}, "a", "b", "c", "d", "e", "wrong:a/b/c/d/e"),
		newCombineFiveValuesTestCase("wrong separator", &TestStruct{value: "root"},
			"a", "b", "c", "d", "e", "root:a-b-c-d-e"),
		newCombineFiveValuesTestCase("wrong root", &TestStruct{value: "test"}, "a", "b", "c", "d", "e", "wrong:a/b/c/d/e"),
		newCombineFiveValuesTestCase("partial path wrong", &TestStruct{value: "root"},
			"a", "b", "c", "d", "e", "root:a/b/c/d"),
	}
}

func TestGetterFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKTestCase
// ============================================================================

// newTestStructGetCountOKTestCase creates a test case for TestStruct.GetCountOK
func newTestStructGetCountOKTestCase(name string, ts *TestStruct, expectedCount int, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKTestCase(name, (*TestStruct).GetCountOK, "GetCountOK", ts, expectedCount, expectedOK)
}

// makeTestGetterOKTestCaseSuccessCases creates success test cases for GetCountOK
//
// GetCountOK follows Go's (value, ok) pattern where:
// - When ok=true: value should be meaningful and tested
// - When ok=false: value should be ignored (zero value expected but not enforced)
func makeTestGetterOKTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// Positive counts return true (count > 0)
		newTestStructGetCountOKTestCase("positive count", &TestStruct{count: 42}, 42, true),
		newTestStructGetCountOKTestCase("count of 1", &TestStruct{count: 1}, 1, true),
		newTestStructGetCountOKTestCase("large positive count", &TestStruct{count: 9999}, 9999, true),
		// Zero and negative counts return false (count <= 0)
		newTestStructGetCountOKTestCase("zero count", &TestStruct{count: 0}, 0, false),
		newTestStructGetCountOKTestCase("negative count", &TestStruct{count: -10}, -10, false),
		newTestStructGetCountOKTestCase("large negative count", &TestStruct{count: -9999}, -9999, false),
	}
}

// makeTestGetterOKTestCaseFailureCases creates failure test cases for GetCountOK
//
// For (value, ok) patterns, failure cases should focus on:
// 1. OK mismatch: wrong boolean expectation
// 2. Value mismatch when ok=true: wrong value when value should be meaningful
// 3. When ok=false: value is ignored, so only test ok mismatch
func makeTestGetterOKTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// OK mismatch - these are the primary failures for OK patterns
		newTestStructGetCountOKTestCase("expect false but get true", &TestStruct{count: 42}, 42, false),
		newTestStructGetCountOKTestCase("expect true but get false", &TestStruct{count: 0}, 0, true),
		// Value mismatch when OK is true (value matters when ok=true)
		newTestStructGetCountOKTestCase("wrong value when ok true", &TestStruct{count: 42}, 100, true),
		// Note: When ok=false, value should be ignored per Go conventions
		// So wrong value + correct ok=false should PASS, not fail
	}
}

func TestGetterOKTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKOneArgTestCase
// ============================================================================

func newTestStructFindValueOKTestCase(name string, ts *TestStruct, search,
	expectedValue string, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKOneArgTestCase(name, (*TestStruct).FindValueOK,
		"FindValueOK", ts, search, expectedValue, expectedOK)
}

func makeTestGetterOKOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newTestStructFindValueOKTestCase("exact match found", &TestStruct{value: "test"}, "test", "test", true),
		newTestStructFindValueOKTestCase("no match empty search", &TestStruct{value: "test"}, "", "", false),
		newTestStructFindValueOKTestCase("no match different value", &TestStruct{value: "hello"}, "world", "", false),
		newTestStructFindValueOKTestCase("empty value no match", &TestStruct{value: ""}, "test", "", false),
		newTestStructFindValueOKTestCase("empty value exact match", &TestStruct{value: ""}, "", "", true),
		newTestStructFindValueOKTestCase("special chars match", &TestStruct{value: "@#$"}, "@#$", "@#$", true),
		newTestStructFindValueOKTestCase("unicode match", &TestStruct{value: "こんにちは"}, "こんにちは", "こんにちは", true),
		newTestStructFindValueOKTestCase("case sensitive no match", &TestStruct{value: "Test"}, "test", "", false),
	}
}

func makeTestGetterOKOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newTestStructFindValueOKTestCase("expect match but no match", &TestStruct{value: "hello"}, "world", "world", true),
		newTestStructFindValueOKTestCase("expect no match but match", &TestStruct{value: "test"}, "test", "", false),
		newTestStructFindValueOKTestCase("wrong value when match", &TestStruct{value: "test"}, "test", "wrong", true),
		newTestStructFindValueOKTestCase("wrong ok expectation", &TestStruct{value: "hello"}, "world", "", true),
	}
}

func TestGetterOKOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKTwoArgsTestCase
// ============================================================================

// newTestStructGetterOKTwoArgsTestCase creates a test case for two-argument getter OK testing
func newTestStructGetterOKTwoArgsTestCase(name string, ts *TestStruct, minVal int, maxVal int,
	expectedValue bool, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKTwoArgsTestCase(name, (*TestStruct).CheckRange, "CheckRange",
		ts, minVal, maxVal, expectedValue, expectedOK)
}

// makeTestGetterOKTwoArgsTestCaseSuccessCases creates success test cases
func makeTestGetterOKTwoArgsTestCaseSuccessCases() []core.TestCase {
	tsInRange := &TestStruct{value: "test", count: 5, enabled: true}   // count=5, positive
	tsOutRange := &TestStruct{value: "test", count: 15, enabled: true} // count=15, positive
	tsZeroCount := &TestStruct{value: "test", count: 0, enabled: true} // count=0, not positive
	tsNegCount := &TestStruct{value: "test", count: -1, enabled: true} // count=-1, negative

	return []core.TestCase{
		// In range, positive count
		newTestStructGetterOKTwoArgsTestCase("in range positive count", tsInRange, 1, 10, true, true),
		// Out of range, positive count
		newTestStructGetterOKTwoArgsTestCase("out of range positive count", tsOutRange, 1, 10, false, true),
		// In range, zero count (not positive)
		newTestStructGetterOKTwoArgsTestCase("in range zero count", tsZeroCount, 0, 10, true, false),
		// Out of range, negative count
		newTestStructGetterOKTwoArgsTestCase("out of range negative count", tsNegCount, 1, 10, false, false),
		// Edge case - exact boundary
		newTestStructGetterOKTwoArgsTestCase("exact min boundary", tsInRange, 5, 10, true, true),
		newTestStructGetterOKTwoArgsTestCase("exact max boundary", tsInRange, 1, 5, true, true),
	}
}

// makeTestGetterOKTwoArgsTestCaseFailureCases creates failure test cases
func makeTestGetterOKTwoArgsTestCaseFailureCases() []core.TestCase {
	tsInRange := &TestStruct{value: "test", count: 5, enabled: true}   // count=5, positive
	tsOutRange := &TestStruct{value: "test", count: 15, enabled: true} // count=15, positive

	return []core.TestCase{
		// Wrong value expectation
		newTestStructGetterOKTwoArgsTestCase("wrong value expectation", tsInRange, 1, 10, false, true), // Should be true
		// Wrong ok expectation
		newTestStructGetterOKTwoArgsTestCase("wrong ok expectation", tsOutRange, 1, 10, false, false), // Should be ok=true
		// Both wrong
		newTestStructGetterOKTwoArgsTestCase("both expectations wrong", tsInRange, 1, 10,
			false, false), // Should be true, true
	}
}

func TestGetterOKTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKThreeArgsTestCase
// ============================================================================

// newTestStructGetterOKThreeArgsTestCase creates a test case for three-argument getter OK testing
func newTestStructGetterOKThreeArgsTestCase(name string, ts *TestStruct, a, b, c int,
	expectedValue int, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKThreeArgsTestCase(name, (*TestStruct).CheckThreeValues, "CheckThreeValues",
		ts, a, b, c, expectedValue, expectedOK)
}

// makeTestGetterOKThreeArgsTestCaseSuccessCases creates success test cases
func makeTestGetterOKThreeArgsTestCaseSuccessCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum5 := &TestStruct{value: "test", count: 5, enabled: true}   // count=5
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// Sum matches count (ok=true)
		newTestStructGetterOKThreeArgsTestCase("sum matches count", tsSum10, 2, 3, 5, 10, true),
		newTestStructGetterOKThreeArgsTestCase("sum matches smaller count", tsSum5, 1, 1, 3, 5, true),
		newTestStructGetterOKThreeArgsTestCase("sum matches zero", tsSum0, 0, 0, 0, 0, true),
		// Sum doesn't match count (ok=false)
		newTestStructGetterOKThreeArgsTestCase("sum doesn't match", tsSum10, 1, 2, 3, 6, false),
		newTestStructGetterOKThreeArgsTestCase("sum too large", tsSum5, 10, 10, 10, 30, false),
		// Edge cases
		newTestStructGetterOKThreeArgsTestCase("negative sum", tsSum0, -1, -2, -3, -6, false),
	}
}

// makeTestGetterOKThreeArgsTestCaseFailureCases creates failure test cases
func makeTestGetterOKThreeArgsTestCaseFailureCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum5 := &TestStruct{value: "test", count: 5, enabled: true}   // count=5

	return []core.TestCase{
		// Wrong value expectation - sum is correct but expectation is wrong
		newTestStructGetterOKThreeArgsTestCase("wrong value expectation", tsSum10, 2, 3, 5, 100, true), // Should be 10
		// Wrong ok expectation - sum matches but expects wrong ok
		newTestStructGetterOKThreeArgsTestCase("wrong ok expectation", tsSum5, 1, 1, 3, 5, false), // Should be ok=true
		// Both wrong
		newTestStructGetterOKThreeArgsTestCase("both expectations wrong", tsSum10, 2, 3, 5, 50, false), // Should be 10, true
	}
}

func TestGetterOKThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKFourArgsTestCase
// ============================================================================

// newTestStructGetterOKFourArgsTestCase creates a test case for four-argument getter OK testing
func newTestStructGetterOKFourArgsTestCase(name string, ts *TestStruct, a, b, c, d int,
	expectedValue int, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKFourArgsTestCase(name, (*TestStruct).CheckFourValues, "CheckFourValues",
		ts, a, b, c, d, expectedValue, expectedOK)
}

// makeTestGetterOKFourArgsTestCaseSuccessCases creates success test cases
func makeTestGetterOKFourArgsTestCaseSuccessCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// Sum matches count (ok=true)
		newTestStructGetterOKFourArgsTestCase("sum matches count", tsSum10, 2, 3, 2, 3, 10, true),
		newTestStructGetterOKFourArgsTestCase("sum matches larger count", tsSum20, 5, 5, 5, 5, 20, true),
		newTestStructGetterOKFourArgsTestCase("sum matches zero", tsSum0, 0, 0, 0, 0, 0, true),
		// Sum doesn't match count (ok=false)
		newTestStructGetterOKFourArgsTestCase("sum doesn't match", tsSum10, 1, 2, 3, 5, 11, false),
		newTestStructGetterOKFourArgsTestCase("sum too large", tsSum10, 10, 10, 10, 10, 40, false),
		// Edge cases
		newTestStructGetterOKFourArgsTestCase("negative sum", tsSum0, -1, -2, -3, -4, -10, false),
	}
}

// makeTestGetterOKFourArgsTestCaseFailureCases creates failure test cases
func makeTestGetterOKFourArgsTestCaseFailureCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum15 := &TestStruct{value: "test", count: 15, enabled: true} // count=15

	return []core.TestCase{
		// Wrong value expectation - sum is correct but expectation is wrong
		newTestStructGetterOKFourArgsTestCase("wrong value expectation", tsSum10, 2, 2, 3, 3, 100, true), // Should be 10
		// Wrong ok expectation - sum matches but expects wrong ok
		newTestStructGetterOKFourArgsTestCase("wrong ok expectation", tsSum15, 5, 5, 2, 3, 15, false), // Should be ok=true
		// Both wrong
		newTestStructGetterOKFourArgsTestCase("both expectations wrong", tsSum10, 2, 2, 3, 3, 50, false),
	}
}

func TestGetterOKFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterOKFiveArgsTestCase
// ============================================================================

// newTestStructGetterOKFiveArgsTestCase creates a test case for five-argument getter OK testing
func newTestStructGetterOKFiveArgsTestCase(name string, ts *TestStruct, a, b, c, d, e int,
	expectedValue int, expectedOK bool) core.TestCase {
	return testutils.NewGetterOKFiveArgsTestCase(name, (*TestStruct).CheckFiveValues, "CheckFiveValues",
		ts, a, b, c, d, e, expectedValue, expectedOK)
}

// makeTestGetterOKFiveArgsTestCaseSuccessCases creates success test cases
func makeTestGetterOKFiveArgsTestCaseSuccessCases() []core.TestCase {
	tsSum15 := &TestStruct{value: "test", count: 15, enabled: true} // count=15
	tsSum25 := &TestStruct{value: "test", count: 25, enabled: true} // count=25
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// Sum matches count (ok=true)
		newTestStructGetterOKFiveArgsTestCase("sum matches count", tsSum15, 1, 2, 3, 4, 5, 15, true),
		newTestStructGetterOKFiveArgsTestCase("sum matches larger count", tsSum25, 5, 5, 5, 5, 5, 25, true),
		newTestStructGetterOKFiveArgsTestCase("sum matches zero", tsSum0, 0, 0, 0, 0, 0, 0, true),
		// Sum doesn't match count (ok=false)
		newTestStructGetterOKFiveArgsTestCase("sum doesn't match", tsSum15, 1, 2, 3, 4, 6, 16, false),
		newTestStructGetterOKFiveArgsTestCase("sum too large", tsSum15, 10, 10, 10, 10, 10, 50, false),
		// Edge cases
		newTestStructGetterOKFiveArgsTestCase("negative sum", tsSum0, -1, -2, -3, -4, -5, -15, false),
	}
}

// makeTestGetterOKFiveArgsTestCaseFailureCases creates failure test cases
func makeTestGetterOKFiveArgsTestCaseFailureCases() []core.TestCase {
	tsSum15 := &TestStruct{value: "test", count: 15, enabled: true} // count=15
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20

	return []core.TestCase{
		// Wrong value expectation - sum is correct but expectation is wrong
		newTestStructGetterOKFiveArgsTestCase("wrong value expectation", tsSum15, 1, 2, 3, 4, 5, 100, true), // Should be 15
		// Wrong ok expectation - sum matches but expects wrong ok
		newTestStructGetterOKFiveArgsTestCase("wrong ok expectation", tsSum20, 4, 4, 4, 4, 4, 20, false), // Should be ok=true
		// Both wrong
		newTestStructGetterOKFiveArgsTestCase("both expectations wrong", tsSum15, 1, 2, 3, 4, 5, 50, false),
	}
}

func TestGetterOKFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterOKFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterOKFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorTestCase
// ============================================================================

// newTestStructGetValueErrorTestCase creates a test case for TestStruct.GetValueError
func newTestStructGetValueErrorTestCase(name string, ts *TestStruct, expectedValue string,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorTestCase(name, (*TestStruct).GetValueError, "GetValueError", ts,
		expectedValue, expectError, errorIs)
}

// makeTestGetterErrorTestCaseSuccessCases creates success test cases for GetValueError
//
// GetValueError follows Go's (value, error) pattern where:
// - When err=nil: value should be meaningful and tested
// - When err!=nil: value should be ignored (zero value expected but not enforced)
// - errorIs=nil means "any error acceptable"
func makeTestGetterErrorTestCaseSuccessCases() []core.TestCase {
	validStruct := &TestStruct{value: "test"}
	emptyStruct := &TestStruct{value: ""}
	structWithCannotBeEmptyError := &TestStruct{err: ErrValueCannotBeEmpty}
	structWithNegativeError := &TestStruct{err: ErrCountNegative}

	return []core.TestCase{
		// No error cases
		newTestStructGetValueErrorTestCase("no error with value",
			validStruct, "test", false, nil),
		newTestStructGetValueErrorTestCase("no error empty value",
			emptyStruct, "", false, nil),

		// Specific error cases
		newTestStructGetValueErrorTestCase("specific error expected",
			structWithCannotBeEmptyError, "", true, ErrValueCannotBeEmpty),
		newTestStructGetValueErrorTestCase("different specific error",
			structWithNegativeError, "", true, ErrCountNegative),

		// Any error cases (errorIs == nil)
		newTestStructGetValueErrorTestCase("any error accepted - empty error",
			structWithCannotBeEmptyError, "", true, nil),
		newTestStructGetValueErrorTestCase("any error accepted - negative error",
			structWithNegativeError, "", true, nil),
	}
}

// makeTestGetterErrorTestCaseFailureCases creates failure test cases for GetValueError
//
// For (value, error) patterns, failure cases should focus on:
// 1. Error expectation mismatch: wrong error vs nil expectation
// 2. Value mismatch when err=nil: wrong value when value should be meaningful
// 3. Wrong specific error when expecting different specific error
func makeTestGetterErrorTestCaseFailureCases() []core.TestCase {
	validStruct := &TestStruct{value: "test"}
	structWithCannotBeEmptyError := &TestStruct{err: ErrValueCannotBeEmpty}
	structWithNegativeError := &TestStruct{err: ErrCountNegative}

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructGetValueErrorTestCase("expect error but none",
			validStruct, "", true, nil),
		newTestStructGetValueErrorTestCase("expect no error but get one",
			structWithCannotBeEmptyError, "test", false, nil),

		// Wrong specific error type
		newTestStructGetValueErrorTestCase("expect specific error but get different",
			structWithCannotBeEmptyError, "", true, ErrCountNegative),
		newTestStructGetValueErrorTestCase("expect different specific error",
			structWithNegativeError, "", true, ErrValueCannotBeEmpty),

		// Value mismatch when error is nil
		newTestStructGetValueErrorTestCase("wrong value when no error",
			validStruct, "wrong", false, nil),
	}
}

func TestGetterErrorTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorOneArgTestCase
// ============================================================================

func newTestStructParseValueErrorTestCase(name string, ts *TestStruct, input,
	expectedValue string, expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorOneArgTestCase(name, (*TestStruct).ParseValueError,
		"ParseValueError", ts, input, expectedValue, expectError, errorIs)
}

func makeTestGetterErrorOneArgTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		newTestStructParseValueErrorTestCase("parse valid input",
			&TestStruct{value: "base"}, "test", "base:test", false, nil),
		newTestStructParseValueErrorTestCase("parse with empty base", &TestStruct{value: ""}, "input", ":input", false, nil),
		newTestStructParseValueErrorTestCase("parse with special chars",
			&TestStruct{value: "prefix"}, "@#$", "prefix:@#$", false, nil),
		newTestStructParseValueErrorTestCase("parse with unicode",
			&TestStruct{value: "hello"}, "こんにちは", "hello:こんにちは", false, nil),
		newTestStructParseValueErrorTestCase("empty input specific error",
			&TestStruct{value: "test"}, "", "", true, ErrValueEmpty),
		newTestStructParseValueErrorTestCase("invalid input specific error",
			&TestStruct{value: "test"}, "invalid", "", true, ErrValueCannotBeEmpty),
		newTestStructParseValueErrorTestCase("empty input any error", &TestStruct{value: "test"}, "", "", true, nil),
		newTestStructParseValueErrorTestCase("invalid input any error", &TestStruct{value: "test"}, "invalid", "", true, nil),
	}
}

func makeTestGetterErrorOneArgTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		newTestStructParseValueErrorTestCase("expect error but get none", &TestStruct{value: "test"}, "valid", "", true, nil),
		newTestStructParseValueErrorTestCase("expect no error but get one",
			&TestStruct{value: "test"}, "", "result", false, nil),
		newTestStructParseValueErrorTestCase("wrong value when no error",
			&TestStruct{value: "test"}, "input", "wrong", false, nil),
		newTestStructParseValueErrorTestCase("wrong specific error",
			&TestStruct{value: "test"}, "", "", true, ErrCountNegative),
	}
}

func TestGetterErrorOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorTwoArgsTestCase
// ============================================================================

// newTestStructGetterErrorTwoArgsTestCase creates a test case for two-argument getter error testing
func newTestStructGetterErrorTwoArgsTestCase(name string, ts *TestStruct, minVal, maxVal int,
	expectedValue bool, expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorTwoArgsTestCase(name, (*TestStruct).ValidateRange, "ValidateRange",
		ts, minVal, maxVal, expectedValue, expectError, errorIs)
}

// makeTestGetterErrorTwoArgsTestCaseSuccessCases creates success test cases
func makeTestGetterErrorTwoArgsTestCaseSuccessCases() []core.TestCase {
	tsInRange := &TestStruct{value: "test", count: 5, enabled: true}   // count=5
	tsOutRange := &TestStruct{value: "test", count: 15, enabled: true} // count=15
	tsExactMin := &TestStruct{value: "test", count: 1, enabled: true}  // count=1
	tsExactMax := &TestStruct{value: "test", count: 10, enabled: true} // count=10

	return []core.TestCase{
		// No error cases
		newTestStructGetterErrorTwoArgsTestCase("in range validation", tsInRange, 1, 10, true, false, nil),
		newTestStructGetterErrorTwoArgsTestCase("exact min boundary", tsExactMin, 1, 10, true, false, nil),
		newTestStructGetterErrorTwoArgsTestCase("exact max boundary", tsExactMax, 1, 10, true, false, nil),
		newTestStructGetterErrorTwoArgsTestCase("out of range validation", tsOutRange, 1, 10, false, false, nil),

		// Specific error cases
		newTestStructGetterErrorTwoArgsTestCase("min greater than max", tsInRange, 10, 5, false, true, ErrMinGreaterThanMax),

		// Any error cases (errorIs == nil)
		newTestStructGetterErrorTwoArgsTestCase("any error accepted", tsInRange, 15, 5, false, true, nil),
	}
}

// makeTestGetterErrorTwoArgsTestCaseFailureCases creates failure test cases
func makeTestGetterErrorTwoArgsTestCaseFailureCases() []core.TestCase {
	tsInRange := &TestStruct{value: "test", count: 5, enabled: true} // count=5

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructGetterErrorTwoArgsTestCase("expect error but none", tsInRange, 1, 10, false, true, nil),
		newTestStructGetterErrorTwoArgsTestCase("expect no error but get one", tsInRange, 10, 5, true, false, nil),

		// Wrong specific error type
		newTestStructGetterErrorTwoArgsTestCase("wrong error type", tsInRange, 10, 5, false, true, ErrNegativeValues),

		// Value mismatch when error is nil
		newTestStructGetterErrorTwoArgsTestCase("wrong value when no error", tsInRange, 1, 10, false, false, nil),
	}
}

func TestGetterErrorTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorThreeArgsTestCase
// ============================================================================

// newTestStructGetterErrorThreeArgsTestCase creates a test case for three-argument getter error testing
func newTestStructGetterErrorThreeArgsTestCase(name string, ts *TestStruct, a, b, c int,
	expectedValue int, expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorThreeArgsTestCase(name, (*TestStruct).ValidateThree, "ValidateThree",
		ts, a, b, c, expectedValue, expectError, errorIs)
}

// makeTestGetterErrorThreeArgsTestCaseSuccessCases creates success test cases
func makeTestGetterErrorThreeArgsTestCaseSuccessCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructGetterErrorThreeArgsTestCase("valid positive args", tsSum10, 2, 3, 5, 10, false, nil),
		newTestStructGetterErrorThreeArgsTestCase("zero args", tsSum0, 0, 0, 0, 0, false, nil),
		newTestStructGetterErrorThreeArgsTestCase("valid sum mismatch", tsSum10, 1, 2, 3, 6, false, nil),

		// Specific error cases
		newTestStructGetterErrorThreeArgsTestCase("negative args error", tsSum10, -1, 2, 3, 0, true, ErrNegativeValues),
		newTestStructGetterErrorThreeArgsTestCase("multiple negative args", tsSum10, -1, -2, -3, 0, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructGetterErrorThreeArgsTestCase("any error accepted", tsSum10, -5, 2, 3, 0, true, nil),
	}
}

// makeTestGetterErrorThreeArgsTestCaseFailureCases creates failure test cases
func makeTestGetterErrorThreeArgsTestCaseFailureCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructGetterErrorThreeArgsTestCase("expect error but none", tsSum10, 2, 3, 5, 10, true, nil),
		newTestStructGetterErrorThreeArgsTestCase("expect no error but get one", tsSum10, -1, 2, 3, 0, false, nil),

		// Wrong specific error type
		newTestStructGetterErrorThreeArgsTestCase("wrong error type", tsSum10, -1, 2, 3, 0, true, ErrMinGreaterThanMax),

		// Value mismatch when error is nil
		newTestStructGetterErrorThreeArgsTestCase("wrong value when no error", tsSum10, 2, 3, 5, 100, false, nil),
	}
}

func TestGetterErrorThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorFourArgsTestCase
// ============================================================================

// newTestStructGetterErrorFourArgsTestCase creates a test case for four-argument getter error testing
func newTestStructGetterErrorFourArgsTestCase(name string, ts *TestStruct, a, b, c, d int,
	expectedValue int, expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorFourArgsTestCase(name, (*TestStruct).ValidateFourValues, "ValidateFourValues",
		ts, a, b, c, d, expectedValue, expectError, errorIs)
}

// makeTestGetterErrorFourArgsTestCaseSuccessCases creates success test cases
func makeTestGetterErrorFourArgsTestCaseSuccessCases() []core.TestCase {
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructGetterErrorFourArgsTestCase("valid positive args", tsSum20, 2, 4, 6, 8, 20, false, nil),
		newTestStructGetterErrorFourArgsTestCase("zero args", tsSum0, 0, 0, 0, 0, 0, false, nil),
		newTestStructGetterErrorFourArgsTestCase("valid sum mismatch", tsSum20, 1, 2, 3, 4, 10, false, nil),

		// Specific error cases
		newTestStructGetterErrorFourArgsTestCase("negative args error", tsSum20, -1, 2, 3, 4, 0, true, ErrNegativeValues),
		newTestStructGetterErrorFourArgsTestCase("multiple negative args", tsSum20, -1, -2, 3, 4, 0, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructGetterErrorFourArgsTestCase("any error accepted", tsSum20, -5, 2, 3, 4, 0, true, nil),
	}
}

// makeTestGetterErrorFourArgsTestCaseFailureCases creates failure test cases
func makeTestGetterErrorFourArgsTestCaseFailureCases() []core.TestCase {
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructGetterErrorFourArgsTestCase("expect error but none", tsSum20, 2, 4, 6, 8, 20, true, nil),
		newTestStructGetterErrorFourArgsTestCase("expect no error but get one", tsSum20, -1, 2, 3, 4, 0, false, nil),

		// Wrong specific error type
		newTestStructGetterErrorFourArgsTestCase("wrong error type", tsSum20, -1, 2, 3, 4, 0, true, ErrMinGreaterThanMax),

		// Value mismatch when error is nil
		newTestStructGetterErrorFourArgsTestCase("wrong value when no error", tsSum20, 2, 4, 6, 8, 100, false, nil),
	}
}

func TestGetterErrorFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// GetterErrorFiveArgsTestCase
// ============================================================================

// newTestStructGetterErrorFiveArgsTestCase creates a test case for five-argument getter error testing
func newTestStructGetterErrorFiveArgsTestCase(name string, ts *TestStruct, a, b, c, d, e int,
	expectedValue int, expectError bool, errorIs error) core.TestCase {
	return testutils.NewGetterErrorFiveArgsTestCase(name, (*TestStruct).ValidateFiveValues, "ValidateFiveValues",
		ts, a, b, c, d, e, expectedValue, expectError, errorIs)
}

// makeTestGetterErrorFiveArgsTestCaseSuccessCases creates success test cases
func makeTestGetterErrorFiveArgsTestCaseSuccessCases() []core.TestCase {
	tsSum25 := &TestStruct{value: "test", count: 25, enabled: true} // count=25
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructGetterErrorFiveArgsTestCase("valid positive args", tsSum25, 3, 4, 5, 6, 7, 25, false, nil),
		newTestStructGetterErrorFiveArgsTestCase("zero args", tsSum0, 0, 0, 0, 0, 0, 0, false, nil),
		newTestStructGetterErrorFiveArgsTestCase("valid sum mismatch", tsSum25, 1, 2, 3, 4, 5, 15, false, nil),

		// Specific error cases
		newTestStructGetterErrorFiveArgsTestCase("negative args error", tsSum25, -1, 2, 3, 4, 5, 0, true, ErrNegativeValues),
		newTestStructGetterErrorFiveArgsTestCase("multiple negative args", tsSum25,
			-1, -2, 3, 4, 5, 0, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructGetterErrorFiveArgsTestCase("any error accepted", tsSum25, -5, 2, 3, 4, 5, 0, true, nil),
	}
}

// makeTestGetterErrorFiveArgsTestCaseFailureCases creates failure test cases
func makeTestGetterErrorFiveArgsTestCaseFailureCases() []core.TestCase {
	tsSum25 := &TestStruct{value: "test", count: 25, enabled: true} // count=25

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructGetterErrorFiveArgsTestCase("expect error but none", tsSum25, 3, 4, 5, 6, 7, 25, true, nil),
		newTestStructGetterErrorFiveArgsTestCase("expect no error but get one", tsSum25, -1, 2, 3, 4, 5, 0, false, nil),

		// Wrong specific error type
		newTestStructGetterErrorFiveArgsTestCase("wrong error type", tsSum25, -1, 2, 3, 4, 5, 0, true, ErrMinGreaterThanMax),

		// Value mismatch when error is nil
		newTestStructGetterErrorFiveArgsTestCase("wrong value when no error", tsSum25, 3, 4, 5, 6, 7, 100, false, nil),
	}
}

func TestGetterErrorFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestGetterErrorFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestGetterErrorFiveArgsTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorTestCase
// ============================================================================

// newTestStructValidateTestCase creates a test case for TestStruct.Validate
func newTestStructValidateTestCase(name string, ts *TestStruct, expectError bool, expectedError error) core.TestCase {
	return testutils.NewErrorTestCase(name, (*TestStruct).Validate, "Validate", ts, expectError, expectedError)
}

// makeTestErrorTestCaseSuccessCases creates success test cases for Validate method
//
// Validate follows Go's error-only pattern where:
// - Methods return nil for success, non-nil error for failure
// - errorIs=nil means "any error acceptable"
func makeTestErrorTestCaseSuccessCases() []core.TestCase {
	return []core.TestCase{
		// No error cases
		newTestStructValidateTestCase("valid struct no error",
			&TestStruct{value: "valid"}, false, nil),
		newTestStructValidateTestCase("non-empty value no error",
			&TestStruct{value: "test"}, false, nil),

		// Specific error cases
		newTestStructValidateTestCase("empty value specific error",
			&TestStruct{value: ""}, true, ErrValueEmpty),

		// Any error cases (errorIs == nil)
		newTestStructValidateTestCase("empty value any error",
			&TestStruct{value: ""}, true, nil),
	}
}

// makeTestErrorTestCaseFailureCases creates failure test cases for Validate method
//
// For error-only patterns, failure cases should focus on:
// 1. Error expectation mismatch: expect error but get nil, or vice versa
// 2. Wrong error type: expect specific error but get different error
func makeTestErrorTestCaseFailureCases() []core.TestCase {
	return []core.TestCase{
		// Error expectation mismatch
		newTestStructValidateTestCase("expect error but get none",
			&TestStruct{value: "valid"}, true, nil),
		newTestStructValidateTestCase("expect no error but get one",
			&TestStruct{value: ""}, false, nil),

		// Wrong specific error type
		newTestStructValidateTestCase("wrong error type",
			&TestStruct{value: ""}, true, ErrCountNegative),
	}
}

func TestErrorTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorOneArgTestCase
// ============================================================================

// newTestStructErrorOneArgTestCase creates a test case for one-argument error testing
func newTestStructErrorOneArgTestCase(name string, ts *TestStruct, value string,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewErrorOneArgTestCase(name, (*TestStruct).ValidateOne, "ValidateOne",
		ts, value, expectError, errorIs)
}

// makeTestErrorOneArgTestCaseSuccessCases creates success test cases
func makeTestErrorOneArgTestCaseSuccessCases() []core.TestCase {
	tsValid := &TestStruct{value: "test", count: 5, enabled: true}     // count=5 (positive)
	tsNegative := &TestStruct{value: "test", count: -1, enabled: true} // count=-1 (negative)

	return []core.TestCase{
		// No error cases
		newTestStructErrorOneArgTestCase("valid value positive count", tsValid, "valid", false, nil),
		newTestStructErrorOneArgTestCase("valid value negative count", tsNegative, "valid", true, ErrCountNegative),

		// Specific error cases
		newTestStructErrorOneArgTestCase("empty value error", tsValid, "", true, ErrValueEmpty),

		// Any error cases (errorIs == nil)
		newTestStructErrorOneArgTestCase("any error accepted", tsValid, "", true, nil),
		newTestStructErrorOneArgTestCase("any count error accepted", tsNegative, "valid", true, nil),
	}
}

// makeTestErrorOneArgTestCaseFailureCases creates failure test cases
func makeTestErrorOneArgTestCaseFailureCases() []core.TestCase {
	tsValid := &TestStruct{value: "test", count: 5, enabled: true}     // count=5 (positive)
	tsNegative := &TestStruct{value: "test", count: -1, enabled: true} // count=-1 (negative)

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructErrorOneArgTestCase("expect error but none", tsValid, "valid", true, nil),
		newTestStructErrorOneArgTestCase("expect no error but get one", tsValid, "", false, nil),

		// Wrong specific error type
		newTestStructErrorOneArgTestCase("wrong error type", tsValid, "", true, ErrMinGreaterThanMax),
		newTestStructErrorOneArgTestCase("wrong error type for count", tsNegative, "valid", true, ErrValueEmpty),
	}
}

func TestErrorOneArgTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorOneArgTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorOneArgTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorTwoArgsTestCase
// ============================================================================

// newTestStructErrorTwoArgsTestCase creates a test case for two-argument error testing
func newTestStructErrorTwoArgsTestCase(name string, ts *TestStruct, minVal, maxVal int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewErrorTwoArgsTestCase(name, (*TestStruct).ValidateTwoArgs, "ValidateTwoArgs",
		ts, minVal, maxVal, expectError, errorIs)
}

// makeTestErrorTwoArgsTestCaseSuccessCases creates success test cases
func makeTestErrorTwoArgsTestCaseSuccessCases() []core.TestCase {
	tsValid := &TestStruct{value: "test", count: 5, enabled: true} // count=5

	return []core.TestCase{
		// No error cases
		newTestStructErrorTwoArgsTestCase("valid range no error", tsValid, 1, 10, false, nil),
		newTestStructErrorTwoArgsTestCase("equal range no error", tsValid, 5, 5, false, nil),

		// Specific error cases
		newTestStructErrorTwoArgsTestCase("min greater than max", tsValid, 10, 5, true, ErrMinGreaterThanMax),

		// Any error cases (errorIs == nil)
		newTestStructErrorTwoArgsTestCase("any error accepted", tsValid, 15, 5, true, nil),
	}
}

// makeTestErrorTwoArgsTestCaseFailureCases creates failure test cases
func makeTestErrorTwoArgsTestCaseFailureCases() []core.TestCase {
	tsValid := &TestStruct{value: "test", count: 5, enabled: true} // count=5

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructErrorTwoArgsTestCase("expect error but none", tsValid, 1, 10, true, nil),
		newTestStructErrorTwoArgsTestCase("expect no error but get one", tsValid, 10, 5, false, nil),

		// Wrong specific error type
		newTestStructErrorTwoArgsTestCase("wrong error type", tsValid, 10, 5, true, ErrNegativeValues),
	}
}

func TestErrorTwoArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorTwoArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorTwoArgsTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorThreeArgsTestCase
// ============================================================================

// newTestStructErrorThreeArgsTestCase creates a test case for three-argument error testing
func newTestStructErrorThreeArgsTestCase(name string, ts *TestStruct, a, b, c int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewErrorThreeArgsTestCase(name, (*TestStruct).ValidateThreeArgs, "ValidateThreeArgs",
		ts, a, b, c, expectError, errorIs)
}

// makeTestErrorThreeArgsTestCaseSuccessCases creates success test cases
func makeTestErrorThreeArgsTestCaseSuccessCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructErrorThreeArgsTestCase("valid positive args matching", tsSum10, 2, 3, 5, false, nil),
		newTestStructErrorThreeArgsTestCase("valid positive args not matching", tsSum10, 1, 2, 3, true, ErrDivisionByZero),
		newTestStructErrorThreeArgsTestCase("zero args matching", tsSum0, 0, 0, 0, false, nil),

		// Specific error cases
		newTestStructErrorThreeArgsTestCase("negative args error", tsSum10, -1, 2, 3, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructErrorThreeArgsTestCase("any error accepted", tsSum10, -5, 2, 3, true, nil),
	}
}

// makeTestErrorThreeArgsTestCaseFailureCases creates failure test cases
func makeTestErrorThreeArgsTestCaseFailureCases() []core.TestCase {
	tsSum10 := &TestStruct{value: "test", count: 10, enabled: true} // count=10

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructErrorThreeArgsTestCase("expect error but none", tsSum10, 2, 3, 5, true, nil),
		newTestStructErrorThreeArgsTestCase("expect no error but get one", tsSum10, -1, 2, 3, false, nil),

		// Wrong specific error type
		newTestStructErrorThreeArgsTestCase("wrong error type", tsSum10, -1, 2, 3, true, ErrMinGreaterThanMax),
	}
}

func TestErrorThreeArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorThreeArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorThreeArgsTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorFourArgsTestCase
// ============================================================================

// newTestStructErrorFourArgsTestCase creates a test case for four-argument error testing
func newTestStructErrorFourArgsTestCase(name string, ts *TestStruct, a, b, c, d int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewErrorFourArgsTestCase(name, (*TestStruct).ValidateFourArgs, "ValidateFourArgs",
		ts, a, b, c, d, expectError, errorIs)
}

// makeTestErrorFourArgsTestCaseSuccessCases creates success test cases
func makeTestErrorFourArgsTestCaseSuccessCases() []core.TestCase {
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructErrorFourArgsTestCase("valid positive args matching", tsSum20, 2, 4, 6, 8, false, nil),
		newTestStructErrorFourArgsTestCase("valid positive args not matching", tsSum20, 1, 2, 3, 4, true, ErrDivisionByZero),
		newTestStructErrorFourArgsTestCase("zero args matching", tsSum0, 0, 0, 0, 0, false, nil),

		// Specific error cases
		newTestStructErrorFourArgsTestCase("negative args error", tsSum20, -1, 2, 3, 4, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructErrorFourArgsTestCase("any error accepted", tsSum20, -5, 2, 3, 4, true, nil),
	}
}

// makeTestErrorFourArgsTestCaseFailureCases creates failure test cases
func makeTestErrorFourArgsTestCaseFailureCases() []core.TestCase {
	tsSum20 := &TestStruct{value: "test", count: 20, enabled: true} // count=20

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructErrorFourArgsTestCase("expect error but none", tsSum20, 2, 4, 6, 8, true, nil),
		newTestStructErrorFourArgsTestCase("expect no error but get one", tsSum20, -1, 2, 3, 4, false, nil),

		// Wrong specific error type
		newTestStructErrorFourArgsTestCase("wrong error type", tsSum20, -1, 2, 3, 4, true, ErrMinGreaterThanMax),
	}
}

func TestErrorFourArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorFourArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorFourArgsTestCaseFailureCases())
	})
}

// ============================================================================
// ErrorFiveArgsTestCase
// ============================================================================

// newTestStructErrorFiveArgsTestCase creates a test case for five-argument error testing
func newTestStructErrorFiveArgsTestCase(name string, ts *TestStruct, a, b, c, d, e int,
	expectError bool, errorIs error) core.TestCase {
	return testutils.NewErrorFiveArgsTestCase(name, (*TestStruct).ValidateFiveArgs, "ValidateFiveArgs",
		ts, a, b, c, d, e, expectError, errorIs)
}

// makeTestErrorFiveArgsTestCaseSuccessCases creates success test cases
func makeTestErrorFiveArgsTestCaseSuccessCases() []core.TestCase {
	tsSum25 := &TestStruct{value: "test", count: 25, enabled: true} // count=25
	tsSum0 := &TestStruct{value: "test", count: 0, enabled: true}   // count=0

	return []core.TestCase{
		// No error cases
		newTestStructErrorFiveArgsTestCase("valid positive args matching", tsSum25, 3, 4, 5, 6, 7, false, nil),
		newTestStructErrorFiveArgsTestCase("valid positive args not matching", tsSum25,
			1, 2, 3, 4, 5, true, ErrDivisionByZero),
		newTestStructErrorFiveArgsTestCase("zero args matching", tsSum0, 0, 0, 0, 0, 0, false, nil),

		// Specific error cases
		newTestStructErrorFiveArgsTestCase("negative args error", tsSum25, -1, 2, 3, 4, 5, true, ErrNegativeValues),

		// Any error cases (errorIs == nil)
		newTestStructErrorFiveArgsTestCase("any error accepted", tsSum25, -5, 2, 3, 4, 5, true, nil),
	}
}

// makeTestErrorFiveArgsTestCaseFailureCases creates failure test cases
func makeTestErrorFiveArgsTestCaseFailureCases() []core.TestCase {
	tsSum25 := &TestStruct{value: "test", count: 25, enabled: true} // count=25

	return []core.TestCase{
		// Error expectation mismatch
		newTestStructErrorFiveArgsTestCase("expect error but none", tsSum25, 3, 4, 5, 6, 7, true, nil),
		newTestStructErrorFiveArgsTestCase("expect no error but get one", tsSum25, -1, 2, 3, 4, 5, false, nil),

		// Wrong specific error type
		newTestStructErrorFiveArgsTestCase("wrong error type", tsSum25, -1, 2, 3, 4, 5, true, ErrMinGreaterThanMax),
	}
}

func TestErrorFiveArgsTestCase(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		testutils.RunSuccessCases(t, makeTestErrorFiveArgsTestCaseSuccessCases())
	})
	t.Run("failure cases", func(t *testing.T) {
		testutils.RunFailureCases(t, makeTestErrorFiveArgsTestCaseFailureCases())
	})
}
