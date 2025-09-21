# `darvaza.org/x/testutils`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/testutils.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/testutils
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/testutils
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/testutils
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=testutils
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/testutils
[socket-link]: https://socket.dev/go/package/darvaza.org/x/testutils

## Overview

The `testutils` package provides generic test case types and factories for
testing Go methods and functions. It implements a comprehensive testing
framework with generic TestCase types that follow the `core.TestCase`
interface, supporting three main testing patterns for different function
signatures.

## Key Features

* **Generic TestCase Types**: Type-safe test cases for methods, factories,
  and functions.
* **Factory Functions**: Comprehensive factory patterns for readable test
  creation.
* **Multiple Argument Support**: Support for functions with 0-5 arguments.
* **Three Testing Patterns**: Specialised support for methods, factories,
  and pure functions.
* **Integration with core**: Built on `darvaza.org/core` testing utilities.
* **Type Validation**: Optional type validation for factory-created objects.

## Package Structure

The `testutils` package consists of:

* **Core Files**:
  * `doc.go`: Package documentation and overview.
  * `generic.go`: Core type definitions and interfaces.
  * `run.go`: Test runner functions with interface adaptation.
  * `utils.go`: Testing utilities for validating TestCase implementations.

* **Test Files**:
  * `example_test.go`: Comprehensive usage examples and documentation.
  * `generic_common_test.go`: Test fixtures and factory functions for testing.
  * `generic_factory_test.go`: Factory TestCase comprehensive test suite.
  * `generic_function_test.go`: Function TestCase comprehensive test suite.
  * `generic_getter_test.go`: Getter TestCase comprehensive test suite.
  * `run_test.go`: Test runner function validation.
  * `utils_test.go`: Tests for the testing utilities.
  * `utils_private_test.go`: Private utility function unit tests.

* **Generated Files**:
  * `generic_gen.go`: Auto-generated TestCase implementations for 0-5 arguments.
  * `generic_gen.sh`: Code generation script for creating argument variants.

The package uses code generation to provide comprehensive support for functions
with varying argument counts while maintaining type safety and consistency.

## Testing Patterns

The package supports three distinct testing patterns based on function
signatures:

### 1. METHOD TESTING

Testing methods that take `*T` as a parameter:

* Use `method`/`methodName` fields.
* No `typeTest` validation needed (object already exists).
* **Examples**: `GetterTestCase`, `ErrorTestCase`, `GetterOKTestCase`,
  `GetterErrorTestCase`.

### 2. FACTORY TESTING

Testing functions that return `*T` (constructors/factories):

* Use `fn`/`funcName` fields.
* Optional `typeTest` validation for returned objects.
* **Examples**: `FactoryTestCase`, `FactoryErrorTestCase`, `FactoryOKTestCase`.

### 3. FUNCTION TESTING

Testing functions that return comparable values (not `*T`):

* Use `fn`/`funcName` fields.
* Tests pure functions returning comparable values.
* **Examples**: `FunctionTestCase`, `FunctionOKTestCase`,
  `FunctionErrorTestCase`.

## Function Type Patterns

### Method Types (take `*T` as parameter)

* `GetterMethod[T, V]`: `func(*T) V`
* `GetterOKMethod[T, V]`: `func(*T) (V, bool)`
* `GetterErrorMethod[T, V]`: `func(*T) (V, error)`
* `ErrorMethod[T]`: `func(*T) error`
* And variants with 1, 2, 3, 4, or 5 arguments.

### Factory Types (return `*T`)

* `Factory[T]`: `func() *T`
* `FactoryOK[T]`: `func() (*T, bool)`
* `FactoryError[T]`: `func() (*T, error)`
* And variants with 1, 2, 3, 4, or 5 arguments.

### Function Types (return comparable `V`)

* `Function[V]`: `func() V`
* `FunctionOK[V]`: `func() (V, bool)`
* `FunctionError[V]`: `func() (V, error)`
* And variants with 1, 2, 3, 4, or 5 arguments.

## Core TestCase Interface

All generated TestCase types implement the `core.TestCase` interface directly.
This design ensures test cases work seamlessly with both standard `*testing.T`
and the more flexible `core.T` interface.

## Testing Utilities

The package provides utilities for validating TestCase implementations and executing tests:

### RunTest

Executes a single test function using sandboxed `testing.RunTests`:

```go
passed := RunTest("test name", func(t *testing.T) {
    // Test code here using *testing.T
})
```

### RunSuccessCases

Validates that test cases pass when expected:

```go
RunSuccessCases(t, []core.TestCase{
    testutils.NewGetterTestCase("valid test", (*MyType).GetValue, "GetValue",
        instance, expectedValue),
})
```

### RunFailureCases

Validates that test cases fail when expected:

```go
RunFailureCases(t, []core.TestCase{
    testutils.NewGetterTestCase("should fail", (*MyType).GetValue, "GetValue",
        instance, wrongValue),
})
```

### Run

Creates sub-tests compatible with both `*testing.T` and `core.T` using interface detection:

```go
Run(t, "sub-test name", func(t core.T) {
    // Test code here
})
```

### Meta-Testing Infrastructure

For testing TestCase implementations themselves:

```go
// Create test scenarios with controlled pass/fail patterns
data := NewArrayTestCaseData("scenario", []int{1, 3}, 5) // tests 1,3 fail
testCases := data.Make()

// Create mock for isolated testing
mockT := &core.MockT{}

// Validate that expected failures are detected
RunSuccessCases(mockT, testCases)
```

## Basic Usage Examples

### Testing Multiple Methods

```go
// Testing different methods within the same context
instance := &MyStruct{value: "test", name: "example"}
testCases := []core.TestCase{
    NewGetterTestCase("GetValue returns correct value",
        (*MyStruct).GetValue, "GetValue", instance, "test"),
    NewGetterTestCase("GetName returns correct name",
        (*MyStruct).GetName, "GetName", instance, "example"),
    NewGetterOneArgTestCase("SetValue updates value",
        (*MyStruct).SetValue, "SetValue", instance, "new", "new"),
}
core.RunTestCases(t, testCases)
```

### Testing Multiple Factories

```go
// Testing different factory functions within the same context
testCases := []core.TestCase{
    NewFactoryTestCase("NewMyStruct creates default instance",
        NewMyStruct, "NewMyStruct", false, validateMyStruct),
    NewFactoryOneArgTestCase("NewMyStructWithValue creates with parameter",
        NewMyStructWithValue, "NewMyStructWithValue", "param", false, validateMyStruct),
    NewFactoryTwoArgsTestCase("NewMyStructComplete creates full instance",
        NewMyStructComplete, "NewMyStructComplete", "value", "name", false, validateMyStruct),
}
core.RunTestCases(t, testCases)
```

### Testing Multiple Functions

```go
// Testing different functions within the same context
testCases := []core.TestCase{
    NewFunctionTestCase("GetConstant returns expected value",
        GetConstant, "GetConstant", 42),
    NewFunctionTwoArgsTestCase("AddNumbers adds correctly",
        AddNumbers, "AddNumbers", 10, 5, 15),
    NewFunctionOneArgTestCase("DoubleValue doubles input",
        DoubleValue, "DoubleValue", 7, 14),
}
core.RunTestCases(t, testCases)
```

### Using with core.RunTestCases

```go
testCases := []core.TestCase{
    NewGetterTestCase(...),
    NewFactoryTestCase(...),
    NewFunctionTestCase(...),
}
core.RunTestCases(t, testCases)
```

## Available TestCase Types

### Method Testing

* **`GetterTestCase[T, V]`**: Tests getter methods that return a value.
* **`GetterOKTestCase[T, V]`**: Tests getter methods that return
  `(value, bool)`.
* **`GetterErrorTestCase[T, V]`**: Tests getter methods that return
  `(value, error)`.
* **`ErrorTestCase[T]`**: Tests methods that return only an error.

### Factory Testing

* **`FactoryTestCase[T]`**: Tests factory functions that return `*T`.
* **`FactoryOKTestCase[T]`**: Tests factory functions that return
  `(*T, bool)`.
* **`FactoryErrorTestCase[T]`**: Tests factory functions that return
  `(*T, error)`.

### Function Testing

* **`FunctionTestCase[V]`**: Tests functions that return a comparable value.
* **`FunctionOKTestCase[V]`**: Tests functions that return `(value, bool)`.
* **`FunctionErrorTestCase[V]`**: Tests functions that return `(value, error)`.

Each type supports variants with 1-5 arguments (e.g., `GetterOneArgTestCase`,
`GetterTwoArgsTestCase`, etc.).

## Factory Functions

All TestCase types include comprehensive factory functions for easy creation:

### Basic Factories

```go
// Method testing
func NewGetterTestCase[T any, V comparable](name string, method GetterMethod[T, V],
    methodName string, receiver *T, want V) GetterTestCase[T, V]

// Factory testing
func NewFactoryTestCase[T any](name string, fn Factory[T], funcName string,
    expectNil bool, typeTest TypeTestFunc[T]) FactoryTestCase[T]

// Function testing
func NewFunctionTestCase[V comparable](name string, fn Function[V],
    funcName string, want V) FunctionTestCase[V]
```

### Argument Variants

Each TestCase type has variants for different argument counts:

```go
// 1 argument variants
func NewGetterOneArgTestCase[T, A1 any, V comparable](name string,
    method GetterMethodOneArg[T, A1, V], methodName string, receiver *T,
    arg1 A1, want V) GetterOneArgTestCase[T, A1, V]

// 2 argument variants
func NewGetterTwoArgsTestCase[T, A1, A2 any, V comparable](name string,
    method GetterMethodTwoArgs[T, A1, A2, V], methodName string, receiver *T,
    arg1 A1, arg2 A2, want V) GetterTwoArgsTestCase[T, A1, A2, V]

// Up to 5 arguments supported
```

### Error Testing Variants

For functions that can fail:

```go
// Error-returning methods
func NewGetterErrorTestCase[T any, V comparable](name string,
    method GetterErrorMethod[T, V], methodName string, item *T,
    expected V, expectError bool, errorIs error) GetterErrorTestCase[T, V]

// Error-returning factories
func NewFactoryErrorTestCase[T any](name string, fn FactoryError[T],
    funcName string, expectError bool, errorIs error,
    typeTest TypeTestFunc[T]) FactoryErrorTestCase[T]

// Error-returning functions
func NewFunctionErrorTestCase[V comparable](name string, fn FunctionError[V],
    funcName string, expected V, expectError bool, errorIs error) FunctionErrorTestCase[V]
```

## Type Validation

Factory test cases support optional type validation through `TypeTestFunc[T]`
functions for **additional domain-specific assertions only**:

**Important**: Factory TestCases automatically handle nil/not-nil testing based
on the `expectNil` parameter. Type validation functions should **only** check
business logic and field values, not nil conditions.

```go
// Validation function for additional business logic checks only
func validateMyStruct(t core.T, obj *MyStruct) bool {
    t.Helper()
    // NOTE: No nil check needed - factory logic handles nil/not-nil automatically
    if !core.AssertNotEqual(t, "", obj.Value, "object value") {
        return false
    }
    if !core.AssertTrue(t, obj.Count >= 0, "count non-negative") {
        return false
    }
    return true
}

// Usage with validation in test sets
testCases := []core.TestCase{
    NewFactoryTestCase("create valid struct", NewMyStruct, "NewMyStruct",
        false, validateMyStruct), // expectNil=false + validation
    NewFactoryTestCase("create returns nil", NewFailingFactory, "NewFailingFactory",
        true, nil), // expectNil=true, no validation needed
}
core.RunTestCases(t, testCases)
```

## Error Testing

Error test cases support specific error validation through the `errorIs`
parameter:

```go
var ErrInvalidInput = errors.New("invalid input")

func NewValidatedStruct(value int) (*MyStruct, error) {
    if value < 0 {
        return nil, ErrInvalidInput
    }
    return &MyStruct{Value: value}, nil
}

// Test different error scenarios within the same context
testCases := []core.TestCase{
    // Success case
    NewFactoryErrorOneArgTestCase("accepts valid values",
        NewValidatedStruct, "NewValidatedStruct", 10, false, nil, validateMyStruct),
    // Specific error case
    NewFactoryErrorOneArgTestCase("rejects negative values",
        NewValidatedStruct, "NewValidatedStruct", -1, true, ErrInvalidInput, nil),
    // Standard library error case
    // cspell:ignore strconv Atoi
    NewFactoryErrorOneArgTestCase("rejects invalid string",
        strconv.Atoi, "Atoi", "not-a-number", true, strconv.ErrSyntax, nil),
}
core.RunTestCases(t, testCases)
```

## Comprehensive Example

```go
package example_test

import (
    "testing"

    "darvaza.org/core"
    "darvaza.org/x/testutils"
)

type Calculator struct {
    value int
}

func NewCalculator(initial int) *Calculator {
    return &Calculator{value: initial}
}

func (c *Calculator) Add(n int) int {
    c.value += n
    return c.value
}

func (c *Calculator) GetValue() int {
    return c.value
}

func TestCalculator(t *testing.T) {
    testCases := []core.TestCase{
        // Test factory
        testutils.NewFactoryOneArgTestCase("create calculator",
            NewCalculator, "NewCalculator", 10, false, nil),

        // Test getter method
        testutils.NewGetterTestCase("get initial value",
            (*Calculator).GetValue, "GetValue", NewCalculator(42), 42),

        // Test method with argument
        testutils.NewGetterOneArgTestCase("add numbers",
            (*Calculator).Add, "Add", NewCalculator(10), 5, 15),
    }

    core.RunTestCases(t, testCases)
}
```

## Argument Count Support

All TestCase types support functions with varying argument counts:

* **Base types**: No additional arguments.
* **OneArg**: 1 additional argument.
* **TwoArgs**: 2 additional arguments.
* **ThreeArgs**: 3 additional arguments.
* **FourArgs**: 4 additional arguments.
* **FiveArgs**: 5 additional arguments.

This covers the vast majority of function signatures encountered in Go code.

## Integration with Testing Framework

The package follows the mandatory testing patterns from `darvaza.org/core`:

1. **TestCase Interface**: All types implement `core.TestCase`.
2. **Factory Functions**: Mandatory factory functions for all TestCase creation.
3. **RunTestCases Usage**: Designed for use with `core.RunTestCases()`.
4. **Compliance**: Meets all linting and complexity requirements.

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core
[core-testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md

## Installation

```bash
go get darvaza.org/x/testutils
```

## Development

For development guidelines, architecture notes, and AI agent instructions,
see [AGENTS.md](AGENTS.md).

## See Also

* [TESTING.md](TESTING.md) for testing pattern guidelines.
* [AGENTS.md](AGENTS.md) for development notes and architecture.
* [Core TESTING.md][core-testing] for core testing requirements.

## Licence

This project is licensed under the MIT Licence. See [LICENCE.txt](LICENCE.txt)
for details.
