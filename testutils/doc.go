// Package testutils provides generic test case types and factories for testing Go methods and functions.
//
// This package implements a comprehensive testing framework with generic TestCase types that follow
// the core.TestCase interface. It supports three main patterns:
//
// # METHOD TESTING
//
// Testing methods that take *T as a parameter:
//   - Use method/methodName fields.
//   - No typeTest validation needed (object already exists).
//   - Examples: GetterTestCase, ErrorTestCase, GetterOKTestCase, GetterErrorTestCase
//
// # FACTORY TESTING
//
// Testing functions that return *T (constructors/factories):
//   - Use fn/funcName fields.
//   - Optional typeTest validation for returned objects.
//   - Examples: FactoryTestCase, FactoryErrorTestCase, FactoryOKTestCase
//
// # FUNCTION TESTING
//
// Testing functions that return comparable values (not *T):
//   - Use fn/funcName fields.
//   - Tests pure functions returning comparable values.
//   - Examples: FunctionTestCase, FunctionOKTestCase, FunctionErrorTestCase
//
// # Function Type Patterns
//
// Method types (take *T as parameter):
//   - GetterMethod[T, V]: func(*T) V
//   - GetterOKMethod[T, V]: func(*T) (V, bool)
//   - GetterErrorMethod[T, V]: func(*T) (V, error)
//   - ErrorMethod[T]: func(*T) error
//   - And variants with 1, 2, 3, 4, or 5 arguments.
//
// Factory types (return *T):
//   - Factory[T]: func() *T
//   - FactoryOK[T]: func() (*T, bool)
//   - FactoryError[T]: func() (*T, error)
//   - And variants with 1, 2, 3, 4, or 5 arguments.
//
// Function types (return comparable V):
//   - Function[V]: func() V
//   - FunctionOK[V]: func() (V, bool)
//   - FunctionError[V]: func() (V, error)
//   - And variants with 1, 2, 3, 4, or 5 arguments.
//
// # Usage Examples
//
// Testing different methods within the same context:
//
//	instance := &MyStruct{value: "test", name: "example"}
//	testCases := []core.TestCase{
//	    NewGetterTestCase("GetValue returns correct value",
//	        (*MyStruct).GetValue, "GetValue", instance, "test"),
//	    NewGetterTestCase("GetName returns correct name",
//	        (*MyStruct).GetName, "GetName", instance, "example"),
//	    NewGetterOneArgTestCase("SetValue updates value",
//	        (*MyStruct).SetValue, "SetValue", instance, "new", "new"),
//	}
//	core.RunTestCases(t, testCases)
//
// Testing different error-handling methods within the same context:
//
//	testCases := []core.TestCase{
//	    NewGetterErrorOneArgTestCase("ParseValue processes input",
//	        (*MyStruct).ParseValue, "ParseValue", instance, "valid", "result", false, nil),
//	    NewGetterErrorOneArgTestCase("ValidateFormat checks format",
//	        (*MyStruct).ValidateFormat, "ValidateFormat", instance, "format", true, false, nil),
//	    NewErrorTestCase("Save persists data",
//	        (*MyStruct).Save, "Save", instance, false, nil),
//	}
//	core.RunTestCases(t, testCases)
//
// Testing different factory functions within the same context:
//
//	testCases := []core.TestCase{
//	    NewFactoryTestCase("NewMyStruct creates default instance",
//	        NewMyStruct, "NewMyStruct", false, validateMyStruct),
//	    NewFactoryOneArgTestCase("NewMyStructWithValue creates with parameter",
//	        NewMyStructWithValue, "NewMyStructWithValue", "param", false, validateMyStruct),
//	    NewFactoryTwoArgsTestCase("NewMyStructComplete creates full instance",
//	        NewMyStructComplete, "NewMyStructComplete", "value", "name", false, validateMyStruct),
//	}
//	core.RunTestCases(t, testCases)
//
// Testing different error-handling factories within the same context:
//
//	testCases := []core.TestCase{
//	    NewFactoryErrorOneArgTestCase("NewMyStructFromString creates from string",
//	        NewMyStructFromString, "NewMyStructFromString", "valid", false, nil, validateMyStruct),
//	    NewFactoryErrorOneArgTestCase("NewMyStructFromFile loads from file",
//	        NewMyStructFromFile, "NewMyStructFromFile", "file.txt", false, nil, validateMyStruct),
//	    NewFactoryErrorTwoArgsTestCase("NewMyStructWithValidation creates with checks",
//	        NewMyStructWithValidation, "NewMyStructWithValidation", "data", "rules", false, nil, validateMyStruct),
//	}
//	core.RunTestCases(t, testCases)
//
// For testing the same function with multiple inputs, use dedicated factories:
//
//	func newValidatedStructTestCase(name string, value int, expectError bool,
//	    expectedError error) core.TestCase {
//	    return NewFactoryErrorOneArgTestCase(name, NewValidatedStruct,
//	        "NewValidatedStruct", value, expectError, expectedError, validateMyStruct)
//	}
//
//	testCases := []core.TestCase{
//	    newValidatedStructTestCase("accepts valid input", 10, false, nil),
//	    newValidatedStructTestCase("rejects negative values", -1, true, ErrInvalidValue),
//	    newValidatedStructTestCase("rejects zero values", 0, true, ErrInvalidValue),
//	}
//	core.RunTestCases(t, testCases)
//
// Testing different functions within the same context:
//
//	testCases := []core.TestCase{
//	    NewFunctionTestCase("GetConstant returns expected value",
//	        GetConstant, "GetConstant", 42),
//	    NewFunctionTwoArgsTestCase("AddNumbers adds correctly",
//	        AddNumbers, "AddNumbers", 10, 5, 15),
//	    NewFunctionOneArgTestCase("DoubleValue doubles input",
//	        DoubleValue, "DoubleValue", 7, 14),
//	}
//	core.RunTestCases(t, testCases)
//
// Testing different error-handling functions within the same context:
//
//	testCases := []core.TestCase{
//	    NewFunctionErrorOneArgTestCase("ParseInt converts strings to integers",
//	        ParseInt, "ParseInt", "123", 123, false, nil),
//	    NewFunctionErrorOneArgTestCase("ValidateEmail checks email format",
//	        ValidateEmail, "ValidateEmail", "test@example.com", true, false, nil),
//	    NewFunctionErrorTwoArgsTestCase("FormatTemplate processes templates",
//	        FormatTemplate, "FormatTemplate", "Hello {{name}}", "World", "Hello World", false, nil),
//	}
//	core.RunTestCases(t, testCases)
//
// Using with core.RunTestCases:
//
//	testCases := []core.TestCase{
//	    NewGetterTestCase(...),
//	    NewFactoryTestCase(...),
//	    NewFunctionTestCase(...),
//	}
//	core.RunTestCases(t, testCases)
package testutils
