# Agent Documentation for x/testutils

## Overview

The `testutils` package provides generic test case types and factories for
testing Go methods and functions. It implements a comprehensive testing
framework with generic TestCase types that follow the `core.TestCase`
interface, supporting the three main testing patterns based on function
signatures and enabling type-safe, reusable test code.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Testing Patterns

- **Method Testing**: Tests methods that take `*T` as a parameter.
- **Factory Testing**: Tests functions that return `*T`
  (constructors/factories).
- **Function Testing**: Tests functions that return comparable values.

### Generic TestCase Types

- **Getter TestCases**: For methods returning values (`GetterTestCase`,
  `GetterOKTestCase`, `GetterErrorTestCase`).
- **Factory TestCases**: For constructor functions (`FactoryTestCase`,
  `FactoryOKTestCase`, `FactoryErrorTestCase`).
- **Function TestCases**: For pure functions (`FunctionTestCase`,
  `FunctionOKTestCase`, `FunctionErrorTestCase`).
- **Error TestCases**: For methods/functions that only return errors
  (`ErrorTestCase`).

### Main Files

- `doc.go`: Package documentation and usage examples.
- `generic.go`: Core generic type definitions and interfaces.
- `generic_gen.go`: Generated TestCase implementations (auto-generated).
- `generic_gen.sh`: Code generation script for TestCase variants.
- `generic_test.go`: Comprehensive test suite.
- `utils.go`: Testing utilities for validating TestCase implementations.
- `utils_test.go`: Tests for the testing utilities.

## Core TestCase Interface

All generated TestCase types implement the `core.TestCase` interface directly.
This design ensures test cases work seamlessly with both standard `*testing.T`
and the more flexible `core.T` interface.

### Testing Utilities

The package provides utility functions for testing TestCase implementations:

- **`RunSuccessCases(t core.T, cases []core.TestCase)`**: Validates that test
  cases pass when expected.
- **`RunFailureCases(t core.T, cases []core.TestCase)`**: Validates that test
  cases fail when expected.
- **`Run(t core.T, name string, fn func(core.T))`**: Creates sub-tests that
  work with both `*testing.T` and `core.T`.

These utilities use `core.MockT` to isolate test execution and verify
outcomes without propagating failures.

## Architecture Notes

The package uses extensive Go generics to provide type-safe testing
utilities while maintaining flexibility. Key design patterns:

1. **Three-Pattern Architecture**: Distinct patterns for methods, factories,
   and functions based on signature differences.
2. **Argument Count Variants**: Support for 0-5 arguments through generated
   code.
3. **Factory Pattern**: Mandatory factory functions for all TestCase types
   following field alignment principles.
4. **Type Safety**: Compile-time safety through generics and interface
   compliance.
5. **Optional Validation**: Type validation for factory-created objects
   through `typeTest` functions.

### TestCase Code Generation

The package uses code generation to create variants for different argument
counts:

- `generic_gen.sh`: A shell script that generates TestCase types and
  factories.
- Supports up to 5 arguments for methods, factories, and functions.
- Maintains consistent naming patterns across all generated types.
- Follows the field alignment and factory function requirements.

## Development Commands

For common development commands and workflow, see the
[root AGENTS.md](../AGENTS.md).

### Code Generation

```bash
# Regenerate TestCase variants (if needed)
./generic_gen.sh > generic_gen.go
```

## Testing Patterns

The package follows the mandatory testing patterns from `TESTING.md` and serves
as a comprehensive example of proper TestCase implementation:

### TestCase Interface Compliance

All TestCase types implement the `core.TestCase` interface:

```go
// Interface validation (MANDATORY)
var _ core.TestCase = GetterTestCase[struct{}, string]{}
var _ core.TestCase = FactoryTestCase[struct{}]{}
var _ core.TestCase = FunctionTestCase[string]{}
```

### Factory Functions (MANDATORY)

All TestCase types have factory functions for creation:

```go
// Factory function for readable parameter order
func NewGetterTestCase[T, V comparable](name string, method GetterMethod[T, V],
    methodName string, receiver *T, want V) GetterTestCase[T, V] {
    return GetterTestCase[T, V]{
        name:       name,
        method:     method,
        methodName: methodName,
        receiver:   receiver,
        want:       want,
    }
}
```

### RunTestCases Usage (MANDATORY)

Tests use `core.RunTestCases` for execution:

```go
func TestExample(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewGetterTestCase(...),
        testutils.NewFactoryTestCase(...),
    }
    core.RunTestCases(t, testCases)
}
```

### Actual Test Patterns in testutils

The package itself demonstrates proper testing patterns:

```go
func TestGetterTestCase(t *testing.T) {
    obj := &testStruct{value: "test"}

    tests := []core.TestCase{
        NewGetterTestCase("get value returns correct result",
            (*testStruct).GetValue, "GetValue", obj, "test"),
        NewGetterTestCase("get empty returns zero value",
            (*testStruct).GetValue, "GetValue", &testStruct{}, ""),
    }

    core.RunTestCases(t, tests)
}

func TestFactoryTestCase(t *testing.T) {
    tests := []core.TestCase{
        NewFactoryTestCase("create valid object",
            newTestStruct, "newTestStruct", false, validateTestStruct),
        NewFactoryOneArgTestCase("create with parameter",
            newTestStructWithValue, "newTestStructWithValue", "param", false,
            nil),
    }

    core.RunTestCases(t, tests)
}
```

### Test Generation Verification

The package includes tests that verify code generation:

```go
func TestGeneratedTypes(t *testing.T) {
    // Verify all argument variants exist and work correctly
    tests := []core.TestCase{
        NewGetterTestCase("no args", (*testStruct).GetValue, "GetValue", obj,
            "test"),
        NewGetterOneArgTestCase("1 arg", (*testStruct).GetWithArg,
            "GetWithArg", obj, "arg", "result"),
        NewGetterTwoArgsTestCase("2 args", (*testStruct).GetWithArgs,
            "GetWithArgs", obj, "a", "b", "result"),
        // ... up to 5 arguments
    }

    core.RunTestCases(t, tests)
}
```

## Common Usage Patterns

### Method Testing

```go
// Test different methods within the same context
testCases := []core.TestCase{
    NewGetterTestCase("get name",
        (*Person).GetName, "GetName", person, "John"),
    NewGetterTwoArgsTestCase("calculate with args",
        (*Calculator).Add, "Add", calc, 5, 10, 15),
    NewGetterOKTestCase("lookup key",
        (*Map).Get, "Get", m, "key", "value", true),
    NewGetterErrorTestCase("parse value",
        (*Parser).Parse, "Parse", parser, "expected", false, nil),
}
core.RunTestCases(t, testCases)
```

### Factory Testing

```go
// Test different factory functions within the same context
testCases := []core.TestCase{
    NewFactoryTestCase("create user",
        NewUser, "NewUser", false, validateUser),
    NewFactoryTwoArgsTestCase("create with args",
        NewConnection, "NewConnection", "host", 8080, false, validateConnection),
    NewFactoryOKOneArgTestCase("try create",
        TryNewResource, "TryNewResource", "config", false, true, validateResource),
    NewFactoryErrorOneArgTestCase("create with validation",
        NewValidatedUser, "NewValidatedUser", "data", false, nil, validateUser),
}
core.RunTestCases(t, testCases)
```

### Function Testing

```go
// Test different functions within the same context
testCases := []core.TestCase{
    NewFunctionTestCase("calculate sum",
        CalculateSum, "CalculateSum", 42),
    NewFunctionTwoArgsTestCase("format string",
        FormatMessage, "FormatMessage", "template", "value", "result"),
    NewFunctionOKOneArgTestCase("try parse",
        TryParseInt, "TryParseInt", "123", 123, true),
    NewFunctionErrorTwoArgsTestCase("validate and process",
        ProcessData, "ProcessData", "input", "flags", "output", false, nil),
}
core.RunTestCases(t, testCases)
```

### Type Validation Functions

```go
// Validation function for factory tests
func validateUser(t core.T, user *User) bool {
    t.Helper()
    if !core.AssertNotNil(t, user, "user") {
        return false
    }
    if !core.AssertNotEqual(t, "", user.Name, "user name") {
        return false
    }
    return core.AssertNotEqual(t, "", user.Email, "user email")
}

// Use in factory test sets
testCases := []core.TestCase{
    NewFactoryTestCase("create valid user",
        NewUser, "NewUser", false, validateUser),
    NewFactoryOneArgTestCase("create user with name",
        NewUserWithName, "NewUserWithName", "John", false, validateUser),
}
core.RunTestCases(t, testCases)
```

## Function Type Definitions

The package defines comprehensive function type patterns:

### Method Types (take `*T` as parameter)

```go
type GetterMethod[T, V any] func(*T) V
type GetterOKMethod[T, V any] func(*T) (V, bool)
type GetterErrorMethod[T, V any] func(*T) (V, error)
type ErrorMethod[T any] func(*T) error

// With arguments (1-5)
type GetterMethod1[T, A1, V any] func(*T, A1) V
type GetterMethod2[T, A1, A2, V any] func(*T, A1, A2) V
// ... up to GetterMethod5
```

### Factory Types (return `*T`)

```go
type Factory[T any] func() *T
type FactoryOK[T any] func() (*T, bool)
type FactoryError[T any] func() (*T, error)

// With arguments (1-5)
type Factory1[A1, T any] func(A1) *T
type Factory2[A1, A2, T any] func(A1, A2) *T
// ... up to Factory5
```

### Function Types (return comparable `V`)

```go
type Function[V any] func() V
type FunctionOK[V any] func() (V, bool)
type FunctionError[V any] func() (V, error)

// With arguments (1-5)
type Function1[A1, V any] func(A1) V
type Function2[A1, A2, V any] func(A1, A2) V
// ... up to Function5
```

## Field Alignment and Memory Optimisation

All TestCase types follow field alignment principles:

```go
type GetterTestCase[T, V any] struct {
    // Large fields first (8+ bytes)
    method     GetterMethod[T, V]  // 8 bytes (function pointer)
    receiver   *T                  // 8 bytes (pointer)
    want       V                   // Variable size
    name       string              // 16 bytes (string header)
    methodName string              // 16 bytes (string header)

    // Small fields last
    // (none in this example)
}
```

Factory functions decouple logical parameter order from memory layout:

```go
// Logical parameter order for readability
func NewGetterTestCase[T, V comparable](name string, method GetterMethod[T, V],
    methodName string, receiver *T, want V) GetterTestCase[T, V] {
    // Memory-optimised field assignment
    return GetterTestCase[T, V]{
        method:     method,     // Memory: first
        receiver:   receiver,   // Memory: second
        want:       want,       // Memory: third
        name:       name,       // Memory: fourth
        methodName: methodName, // Memory: fifth
    }
}
```

## Error Handling Patterns

The package handles different error scenarios:

### Testing Error Conditions

```go
// Test different error scenarios within the same context
testCases := []core.TestCase{
    // Method should return error
    NewGetterErrorTestCase("invalid input",
        (*Parser).Parse, "Parse", parser, "", true, nil),
    // Factory should fail
    NewFactoryErrorOneArgTestCase("invalid config",
        NewConnection, "NewConnection", "invalid", true, nil, nil),
    // Function should succeed
    NewFunctionErrorTwoArgsTestCase("valid processing",
        ProcessData, "ProcessData", "valid", "flags", "result", false, nil),
}
core.RunTestCases(t, testCases)
```

### Validation Function Examples

```go
// Validation function for additional business logic checks
func validateConnection(t core.T, conn *Connection) bool {
    t.Helper()
    // NOTE: No nil check needed - factory logic handles nil/not-nil automatically
    if !core.AssertTrue(t, conn.IsConnected(), "connection established") {
        return false
    }
    return core.AssertTrue(t, conn.Timeout >= 0, "timeout non-negative")
}
```

## Performance Considerations

- **Generic Types**: Compile-time type safety with zero runtime overhead.
- **Field Alignment**: Memory-optimised struct layouts reduce memory usage.
- **Factory Functions**: No runtime cost, only compilation-time convenience.
- **Code Generation**: Eliminates repetitive code while maintaining type safety.

## Integration with Core Testing

The package integrates seamlessly with `darvaza.org/core` testing utilities:

```go
import (
    "testing"

    "darvaza.org/core"
    "darvaza.org/x/testutils"
)

func TestMyCode(t *testing.T) {
    // Mix different TestCase types
    testCases := []core.TestCase{
        testutils.NewGetterTestCase(...),
        testutils.NewFactoryTestCase(...),
        testutils.NewFunctionTestCase(...),
    }

    // Use core.RunTestCases for execution
    core.RunTestCases(t, testCases)
}
```

## Dependencies

- `darvaza.org/core`: Core utilities and TestCase interface.
- Standard library only.

## Code Quality Standards

The package adheres to the strict quality standards:

- **Cognitive Complexity**: ≤7 for all functions.
- **Cyclomatic Complexity**: ≤10 for all functions.
- **Function Length**: ≤40 lines maximum.
- **Linting Compliance**: Passes all revive and golangci-lint checks.
- **Test Coverage**: Comprehensive coverage for all generated types.

## See Also

- [Package README](README.md) for detailed API documentation and examples.
- [Root AGENTS.md](../AGENTS.md) for mono-repo overview and development
  commands.
- [Root TESTING.md](../../core/TESTING.md) for testing pattern
  requirements.
