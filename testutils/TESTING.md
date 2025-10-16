# Testing Guidelines for testutils Package

This document outlines the testing patterns and practices for using the `testutils`
package in your projects, building upon the `darvaza.org/core` testing framework
and following the mandatory testing requirements from the parent `TESTING.md`.

**Two distinct usage contexts:**

1. **Normal Usage**: Using testutils TestCase types in your projects.
2. **Meta-Testing**: Testing the testutils package itself (internal only).

**⚠️ WARNING**: The patterns in this document are tools, not rules. If you find
yourself contorting your tests to fit these patterns, STOP. Write a simple test
instead. Clear, straightforward tests are always better than clever
abstractions.

## Overview

The `testutils` package provides generic test case factories for common testing
patterns. **The core purpose is composability** - enabling you to create reusable
test functions like `makeUserTestCases()` that test multiple methods for any
object instance, avoiding monolithic Test() functions.

The package implements three main testing approaches:

1. **Method Testing**: Testing methods that take `*T` as a parameter.
2. **Factory Testing**: Testing functions that return `*T` (constructors).
3. **Function Testing**: Testing functions that return comparable values.

## Core Testing Principles

### Clarity Over Patterns

**The primary goal of TestCase is CLARITY, not pattern compliance.**

Before using any TestCase pattern, ask yourself:

- Does this make the test easier to understand?
- Would a simple test function be clearer?
- Am I adding abstraction that obscures what's being tested?

Remember: **Over-engineering tests makes them harder to maintain.** If you're
struggling to make your test fit a pattern, you probably shouldn't use that
pattern.

### Understanding testutils and core.TestCase

The `testutils` package provides **pre-built implementations** of the
`core.TestCase` interface. This relationship is important to understand:

1. **testutils types ARE core.TestCase implementations**: Every TestCase type
   in testutils (GetterTestCase, FactoryTestCase, etc.) implements the
   `core.TestCase` interface.

2. **Use with core.RunTestCases**: You still use `core.RunTestCases()` to
   execute testutils TestCases - they're fully compatible.

3. **Convenience layer**: testutils provides pre-built patterns for common
   testing scenarios, eliminating boilerplate code.

4. **Composability**: testutils.TestCases help automate complex testing by
   allowing you to compose `[]core.TestCase{}` covering multiple methods for
   the same object instead of huge Test() functions with specific Assert calls.

5. **Custom TestCase option**: When testutils patterns do not fit your needs,
   you can create custom types that implement `core.TestCase`.

## testutils TestCase Types

**RECOMMENDED USAGE**: Create custom factory functions that wrap testutils to
make your test files easier to read and maintain. Hide testutils complexity
behind domain-specific interfaces.

The core benefit of testutils is **composability** - enabling you to compose
`[]core.TestCase{}` covering multiple methods for the same object instead of
huge Test() functions with specific Assert calls.

```go
// ✅ PREFERRED - Domain-specific factory makes tests readable
func newUserGetNameTestCase(name string, user *User, expected string) core.TestCase {
    return testutils.NewGetterTestCase(name, (*User).GetName, "GetName", user, expected)
}

func newUserGetEmailTestCase(name string, user *User, expected string) core.TestCase {
    return testutils.NewGetterTestCase(name, (*User).GetEmail, "GetEmail", user, expected)
}

func newUserGetRoleTestCase(name string, user *User, expected string) core.TestCase {
    return testutils.NewGetterTestCase(name, (*User).GetRole, "GetRole", user, expected)
}

// Test usage becomes clean and readable:
tests := []core.TestCase{
    newUserGetNameTestCase("user has correct name", user, "Alice"),
    newUserGetEmailTestCase("user has correct email", user, "alice@example.com"),
    newUserGetRoleTestCase("user has correct role", user, "admin"),
}

// ❌ AVOID - Raw testutils calls clutter test files
testutils.NewGetterTestCase("user has correct name", (*User).GetName,
    "GetName", user, "Alice")

// ❌ AVOID - Monolithic Test() function with many specific assertions
func TestUserMonolithic(t *testing.T) {
    user := &User{Name: "Alice", Email: "alice@test.com", Role: "admin"}

    // Testing many methods in one huge function
    core.AssertEqual(t, "Alice", user.GetName(), "name")
    core.AssertEqual(t, "alice@test.com", user.GetEmail(), "email")
    core.AssertEqual(t, "admin", user.GetRole(), "role")
    core.AssertTrue(t, user.IsAdmin(), "admin status")
    core.AssertNotNil(t, user.GetPermissions(), "permissions")
    core.AssertEqual(t, 5, len(user.GetPermissions()), "permission count")
    // ... many more assertions - becomes unwieldy
}

// ✅ PREFERRED - Composed TestCase array covering multiple methods
func TestUserComposed(t *testing.T) {
    t.Run("admin", func(t *testing.T) {
        user := &User{Name: "Alice", Email: "alice@test.com", Role: "admin"}
        core.RunTestCases(t, makeUserComposedTestCases(user, "Alice", "alice@test.com", "admin"))
    })

    t.Run("guest", func(t *testing.T) {
        user := &User{Name: "Bob", Email: "bob@test.com", Role: "guest"}
        core.RunTestCases(t, makeUserComposedTestCases(user, "Bob", "bob@test.com", "guest"))
    })
}

// Reusable test composition for any user
func makeUserComposedTestCases(user *User, expectedName, expectedEmail, expectedRole string) []core.TestCase {
    return []core.TestCase{
        newUserGetNameTestCase("user has correct name", user, expectedName),
        newUserGetEmailTestCase("user has correct email", user, expectedEmail),
        newUserGetRoleTestCase("user has correct role", user, expectedRole),
        newUserValidationTestCase("user passes validation", user, true, nil),
        newUserPermissionTestCase("admin has permissions", user, 5),
    }
}
```

### Method Testing (operates on existing objects)

For testing methods that take `*T` as a parameter:

#### GetterTestCase - Simple getter methods

```go
// Testing different getter methods within the same context
instance := &MyStruct{value: "test", name: "example"}
testCases := []core.TestCase{
    testutils.NewGetterTestCase("GetValue returns correct value",
        (*MyStruct).GetValue, "GetValue", instance, "test"),
    testutils.NewGetterTestCase("GetName returns correct name",
        (*MyStruct).GetName, "GetName", instance, "example"),
    testutils.NewGetterTestCase("GetID returns correct ID",
        (*MyStruct).GetID, "GetID", instance, 42),
}
core.RunTestCases(t, testCases)
```

#### GetterOKTestCase - Methods returning (value, bool)

```go
// Testing different getter OK methods within the same context
cache := &Cache{data: map[string]string{"key": "value", "name": "test"}}
testCases := []core.TestCase{
    testutils.NewGetterOKOneArgTestCase("Get returns existing value",
        (*Cache).Get, "Get", cache, "key", "value", true),
    testutils.NewGetterOKOneArgTestCase("Get returns false for missing key",
        (*Cache).Get, "Get", cache, "missing", "", false),
    testutils.NewGetterOKTestCase("HasData returns true when populated",
        (*Cache).HasData, "HasData", cache, true),
}
core.RunTestCases(t, testCases)
```

#### GetterErrorTestCase - Methods returning (value, error)

```go
// Testing different getter error methods within the same context
parser := &Parser{config: defaultConfig}
testCases := []core.TestCase{
    testutils.NewGetterErrorOneArgTestCase("Parse handles valid input",
        (*Parser).Parse, "Parse", parser, "valid", Result{Value: "parsed"}, false, nil),
    testutils.NewGetterErrorOneArgTestCase("Parse returns error for invalid input",
        (*Parser).Parse, "Parse", parser, "invalid", Result{}, true, nil),
    testutils.NewGetterErrorOneArgTestCase("Validate checks input format",
        (*Parser).Validate, "Validate", parser, "format", true, false, nil),
}
core.RunTestCases(t, testCases)
```

#### ErrorTestCase - Methods returning only error

```go
// Testing different error methods within the same context
writer := &Writer{closed: true}
testCases := []core.TestCase{
    testutils.NewErrorTestCase("Flush fails on closed writer",
        (*Writer).Flush, "Flush", writer, true, nil),
    testutils.NewErrorOneArgTestCase("Write fails on closed writer",
        (*Writer).Write, "Write", writer, "data", true, nil),
    testutils.NewErrorTestCase("Close succeeds on closed writer",
        (*Writer).Close, "Close", writer, false, nil),
}
core.RunTestCases(t, testCases)
```

### Factory Testing (creates new objects)

For testing functions that return `*T`:

**Important**: Factory TestCases automatically handle nil/not-nil testing based
on the `expectNil` parameter. Type validation functions should **only** check
business logic and field values, not nil conditions.

#### FactoryTestCase - Simple factory

```go
// Testing different factory functions within the same context
testCases := []core.TestCase{
    testutils.NewFactoryOneArgTestCase("NewConfig creates valid config",
        NewConfig, "NewConfig", "/path/to/config", false, validateConfig),
    testutils.NewFactoryTestCase("NewDefaultConfig creates with defaults",
        NewDefaultConfig, "NewDefaultConfig", false, validateConfig),
    testutils.NewFactoryTwoArgsTestCase("NewConfigWithTimeout creates with settings",
        NewConfigWithTimeout, "NewConfigWithTimeout", "/path", 30, false, validateConfig),
}
core.RunTestCases(t, testCases)

func validateConfig(t core.T, config *Config) bool {
    t.Helper()
    // Factory logic handles nil/not-nil automatically
    return core.AssertNotEqual(t, "", config.Path, "config path")
}
```

#### FactoryOKTestCase - Factory returning (*T, bool)

```go
// Testing different factory OK functions within the same context
testCases := []core.TestCase{
    testutils.NewFactoryOKOneArgTestCase("TryParse succeeds with valid data",
        TryParse, "TryParse", validData, false, true, validateObject),
    testutils.NewFactoryOKOneArgTestCase("TryParseFromFile loads from file",
        TryParseFromFile, "TryParseFromFile", "file.txt", false, true, validateObject),
    testutils.NewFactoryOKTestCase("TryCreateDefault creates default instance",
        TryCreateDefault, "TryCreateDefault", false, true, validateObject),
}
core.RunTestCases(t, testCases)
```

#### FactoryErrorTestCase - Factory returning (*T, error)

```go
// Testing different factory error functions within the same context
testCases := []core.TestCase{
    testutils.NewFactoryErrorOneArgTestCase("NewClient creates with valid config",
        NewClient, "NewClient", validConfig, false, nil, validateClient),
    testutils.NewFactoryErrorOneArgTestCase("NewClient fails with invalid config",
        NewClient, "NewClient", invalidConfig, true, nil, nil),
    testutils.NewFactoryErrorTwoArgsTestCase("NewClientWithTimeout creates with settings",
        NewClientWithTimeout, "NewClientWithTimeout", validConfig, 30, false, nil, validateClient),
}
core.RunTestCases(t, testCases)
```

### Function Testing (returns comparable values)

For testing standalone functions that return comparable values:

#### FunctionTestCase - Simple function

```go
// Testing different functions within the same context
testCases := []core.TestCase{
    testutils.NewFunctionOneArgTestCase("CalculateHash produces expected hash",
        CalculateHash, "CalculateHash", []byte("test data"), "expected-hash"),
    testutils.NewFunctionTwoArgsTestCase("FormatString combines values",
        FormatString, "FormatString", "Hello", "World", "Hello World"),
    testutils.NewFunctionTestCase("GetConstant returns expected value",
        GetConstant, "GetConstant", 42),
}
core.RunTestCases(t, testCases)
```

#### FunctionOKTestCase - Function returning (value, bool)

```go
// Testing different function OK methods within the same context
testCases := []core.TestCase{
    testutils.NewFunctionOKOneArgTestCase("LookupValue finds existing key",
        LookupValue, "LookupValue", "existing-key", "expected-value", true),
    testutils.NewFunctionOKOneArgTestCase("TryParseInt converts valid string",
        TryParseInt, "TryParseInt", "123", 123, true),
    testutils.NewFunctionOKTestCase("GetDefaultValue returns default",
        GetDefaultValue, "GetDefaultValue", "default", true),
}
core.RunTestCases(t, testCases)
```

#### FunctionErrorTestCase - Function returning (value, error)

```go
// Testing different function error methods within the same context
testCases := []core.TestCase{
    testutils.NewFunctionErrorOneArgTestCase("ParseInt handles valid input",
        ParseInt, "ParseInt", "123", 123, false, nil),
    testutils.NewFunctionErrorOneArgTestCase("ParseInt handles invalid input",
        ParseInt, "ParseInt", "not-a-number", 0, true, nil),
    testutils.NewFunctionErrorTwoArgsTestCase("ValidateRange checks boundaries",
        ValidateRange, "ValidateRange", 5, 10, true, false, nil),
}
core.RunTestCases(t, testCases)
```

## Multiple Arguments Support

testutils provides TestCase types following a systematic pattern:

**Pattern**: `{Base}{Args}TestCase[TypeParams]`

**Base types:**

- **Getter**, **GetterOK**, **GetterError**, **Error** (for methods on `*T`)
- **Factory**, **FactoryOK**, **FactoryError** (for functions returning `*T`)
- **Function**, **FunctionOK**, **FunctionError** (for functions returning `V`)

**Argument variants:**

- **(none)**, **OneArg**, **TwoArgs**, **ThreeArgs**, **FourArgs**, **FiveArgs**

**Examples:**

- `GetterTestCase[T, V]` - method with no arguments
- `GetterTwoArgsTestCase[T, A1, A2, V]` - method with 2 arguments
- `FactoryErrorOneArgTestCase[A1, T]` - factory with 1 argument returning `(*T, error)`

**Coverage**: 10 base patterns × 6 argument variants = 60 TestCase types.

## Decision Framework: When to Use TestCase

**CRITICAL PRINCIPLE**: Always start with the simplest approach that achieves your testing goals. Use TestCase when you have **2+ similar test scenarios with identical logic** to avoid duplicate code.

### Decision Checklist

**Use this checklist to decide:**

```text
❓ Do I have 2+ test scenarios with identical test logic?
   NO  → Use simple test
   YES → Continue checking

❓ Are all test scenarios testing the SAME function/method?
   NO  → Use simple test or t.Run() with named functions
   YES → Continue checking

❓ Do scenarios differ only in input/output data?
   NO  → Use simple test (different logic = different tests)
   YES → Continue checking

❓ Does the test fit getter/factory/function patterns?
   NO  → Use simple test or custom TestCase
   YES → Continue checking

❓ Does TestCase make the test CLEARER than simple assertions?
   NO  → Use simple test
   YES → Consider testutils TestCase

❓ Can I create a domain-specific factory function?
   NO  → Use simple test
   YES → Use testutils TestCase with domain wrapper
```

### Testing Pattern Priority

**Always start with the simplest approach:**

1. **Simple test functions** - Default choice for most tests.
2. **t.Run() with named functions** - When you have different test behaviours.
3. **testutils TestCase types** - ONLY when you have 2+ similar test scenarios.
4. **Custom TestCase implementations** - ONLY when complexity is justified.

**Golden Rule**: Start with simple tests. Only move to more complex patterns when the simple approach becomes genuinely insufficient.

## When NOT to Use TestCase

**CRITICAL PRINCIPLE**: TestCase is a tool for specific scenarios, not a universal testing solution. Most tests should NOT use TestCase patterns.

### Scenarios Where TestCase is INAPPROPRIATE

#### 1. Single or Minimal Test Cases

**Problem**: TestCase overhead for 1-2 scenarios adds complexity without benefit.

```go
// ❌ WRONG - TestCase overkill for single test
func TestSingleCase(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewGetterTestCase("get value", (*Object).GetValue,
            "GetValue", obj, "expected"),
    }
    core.RunTestCases(t, testCases)
}

// ✅ RIGHT - Simple and clear
func TestSingleCase(t *testing.T) {
    obj := &Object{value: "expected"}
    core.AssertEqual(t, "expected", obj.GetValue(), "value")
}
```

**Rule**: Use TestCase only when you have 2+ similar test scenarios with identical logic.

#### 2. External Resource Management

**Problem**: Shared external resources (servers, databases) do not fit table-driven patterns well.

```go
// ❌ PROBLEMATIC - External resources shared across TestCases
func TestWithExternalResources(t *testing.T) {
    server := startTestServer()
    defer server.Stop()
    db := createTestDatabase()
    defer db.Close()

    // All test cases share same external resources
    testCases := []core.TestCase{...}
    core.RunTestCases(t, testCases)
}

// ✅ BETTER - Simple test manages external resources directly
func TestWithExternalResources(t *testing.T) {
    server := startTestServer()
    defer server.Stop()
    client := setupClientWithAuth(server.URL)
    result := client.GetData()
    core.AssertEqual(t, "expected", result, "data")
}
```

**Note**: TestCase structs CAN have setup methods. The issue is external resource sharing, not setup itself.

**Rule**: Use simple tests for shared external resource management.

#### 3. Unique Validation Logic Per Test

**Problem**: Each test needs different validation approaches.

```go
// ❌ WRONG - Forcing TestCase when validation differs
func TestDifferentValidations(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewGetterTestCase("parse JSON", (*Parser).Parse,
            "Parse", parser, complexResult), // Needs JSON structure validation
        testutils.NewGetterTestCase("parse XML", (*Parser).Parse,
            "Parse", parser, xmlResult),     // Needs XML schema validation
    }
}

// ✅ RIGHT - Each test handles its unique validation
func TestDifferentValidations(t *testing.T) {
    t.Run("parse JSON", func(t *testing.T) {
        result := parser.Parse(jsonInput)
        core.AssertEqual(t, "value", result.Data["key"], "JSON key")
    })
    t.Run("parse XML", func(t *testing.T) {
        result := parser.Parse(xmlInput)
        core.AssertEqual(t, "namespace", result.Namespace, "XML namespace")
    })
}
```

**Rule**: Use simple tests when each case requires unique validation logic.

#### 4. Testing Different Behaviours

**Problem**: Tests verify different functionality, not data variations.

```go
// ❌ WRONG - Different behaviours forced into TestCase
func TestDifferentBehaviours(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewGetterTestCase("get name", (*User).GetName, "GetName", user, "Alice"),
        testutils.NewGetterTestCase("is admin", (*User).IsAdmin, "IsAdmin", user, true),
        testutils.NewGetterTestCase("last login", (*User).LastLogin, "LastLogin", user, timestamp),
    }
}

// ✅ RIGHT - Separate tests for different behaviours
func TestUserProperties(t *testing.T) {
    user := &User{Name: "Alice", Role: "admin", LastLogin: timestamp}
    core.AssertEqual(t, "Alice", user.GetName(), "name")
    core.AssertTrue(t, user.IsAdmin(), "admin status")
    core.AssertEqual(t, timestamp, user.LastLogin(), "last login")
}
```

**Rule**: Use TestCase only when testing the SAME behaviour with different data.

#### 5. Integration or End-to-End Tests

**Problem**: TestCase patterns do not fit complex workflows.

```go
// ✅ RIGHT - Simple test models workflow naturally
func TestUserWorkflow(t *testing.T) {
    user, err := CreateUser("test@example.com")
    core.AssertNoError(t, err, "create user")
    defer DeleteUser(user.ID)

    token, err := Authenticate(user.Email, "password")
    core.AssertNoError(t, err, "authenticate")

    result, err := ProcessData(token, testData)
    core.AssertNoError(t, err, "process data")
    core.AssertEqual(t, "processed", result.Status, "process status")
}
```

**Rule**: Use simple tests for integration testing and complex workflows.

#### 6. Performance or Benchmark Tests

**Problem**: TestCase adds overhead to performance measurements.

```go
// ❌ WRONG - TestCase overhead affects benchmarks
func BenchmarkWithTestCase(b *testing.B) {
    testCases := []core.TestCase{
        testutils.NewFunctionTestCase("calculate", Calculate, "Calculate", 42),
    }
    // TestCase overhead skews benchmark results
}

// ✅ RIGHT - Direct function calls for accurate benchmarks
func BenchmarkCalculate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := Calculate(inputData)
        _ = result // Prevent compiler optimisation
    }
}
```

**Rule**: Never use TestCase patterns in benchmark tests.

#### 7. Tests Requiring Mocks or Dependencies

**Problem**: TestCase does not accommodate dependency injection well.

```go
// ✅ RIGHT - Simple test handles dependency injection
func TestWithMocks(t *testing.T) {
    mockDB := &MockDatabase{}
    service := &Service{db: mockDB}
    mockDB.ExpectQuery("SELECT").WillReturn("result")

    result := service.GetData()
    core.AssertEqual(t, "expected", result, "data")
    mockDB.AssertExpectations(t)
}
```

**Rule**: Use simple tests when mocks or dependency injection are required.

### Common Anti-patterns to Avoid

#### 1. TestCase for Single Assertions

```go
// ❌ ANTI-PATTERN
func TestSingleValue(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewGetterTestCase("get value", (*Object).GetValue,
            "GetValue", obj, "test"),
    }
    core.RunTestCases(t, testCases)
}

// ✅ CORRECT
func TestSingleValue(t *testing.T) {
    obj := &Object{value: "test"}
    core.AssertEqual(t, "test", obj.GetValue(), "value")
}
```

#### 2. Forcing Unrelated Tests into TestCase

```go
// ❌ ANTI-PATTERN - Different functions forced together
testCases := []core.TestCase{
    testutils.NewGetterTestCase("user name", (*User).GetName, "GetName", user, "Alice"),
    testutils.NewFunctionTestCase("calculate hash", CalculateHash, "CalculateHash", "abc123"), // Unrelated!
}

// ✅ CORRECT - Separate tests for different concerns
func TestUserName(t *testing.T) {
    core.AssertEqual(t, "Alice", user.GetName(), "name")
}
func TestCalculateHash(t *testing.T) {
    core.AssertEqual(t, "abc123", CalculateHash("input"), "hash")
}
```

#### 3. TestCase When Custom Validation is Needed

```go
// ❌ ANTI-PATTERN - Complex validation in TestCase
func TestComplexValidation(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewFactoryTestCase("create complex", NewComplex, "NewComplex",
            false, func(t core.T, c *Complex) bool {
                // Complex validation logic that's hard to reuse
                if c.Field1 != "expected" { return false }
                // ... many more checks
                return true
            }),
    }
}

// ✅ CORRECT - Custom validation in simple test
func TestComplexCreation(t *testing.T) {
    result := NewComplex()
    core.AssertNotNil(t, result, "result")
    core.AssertEqual(t, "expected", result.Field1, "field1")
    core.AssertEqual(t, 3, len(result.Array), "array length")
}
```

#### 4. Function Fields in TestCase Structs (FORBIDDEN)

```go
// ❌ ANTI-PATTERN - Test function as field
type badTestCase struct {
    name     string
    testFunc func(*testing.T) // FORBIDDEN - delegating test logic to field
}

func (tc badTestCase) Test(t *testing.T) {
    tc.testFunc(t) // ANTI-PATTERN - Test() delegating to field function
}

// ✅ CORRECT - Test logic in Test() method, data in fields
type goodTestCase struct {
    name     string
    input    string
    expected string
    wantErr  bool
}

func (tc goodTestCase) Test(t *testing.T) {
    t.Helper()
    // Test logic lives HERE in Test() method
    result, err := ProcessInput(tc.input)
    if tc.wantErr {
        core.AssertError(t, err, "process error")
        return
    }
    core.AssertNoError(t, err, "process")
    core.AssertEqual(t, tc.expected, result, "result")
}
```

**Rule**: TestCase structs contain only data. Test() methods contain test logic. Never use function fields for delegation.

## Type Validation Function Design Guide

**CRITICAL PRINCIPLE**: Validation functions should focus ONLY on business logic and internal object consistency. The testutils factory logic automatically handles nil/not-nil testing based on the `expectNil` parameter.

### Design Philosophy

**What validation functions ARE for:**

- **Internal object consistency**: Ensuring field relationships are valid.
- **Business logic verification**: Domain-specific rules and constraints.
- **State consistency**: Ensuring object fields are in valid states.
- **Cross-field validation**: Checking relationships between fields.

**What validation functions are NOT for:**

- **Basic nil checking**: Factory logic handles this automatically.
- **Method testing**: We have method TestCase types for testing behaviour.
- **Type assertion**: Go's type system ensures correctness.
- **Framework testing**: Testing the factory itself, not the object.

### Core Design Principles

#### 1. Internal Consistency Focus

**Validate internal field relationships, not external behaviour:**

```go
// ✅ GOOD - Internal consistency validation
func validateUser(t core.T, user *User) bool {
    t.Helper()
    // Factory logic handles nil/not-nil automatically

    // Internal consistency: User fields must be coherent
    if !core.AssertNotEqual(t, "", user.Name, "user name") {
        return false
    }

    // Cross-field consistency: Email format must be valid
    if !core.AssertTrue(t, strings.Contains(user.Email, "@"), "email format") {
        return false
    }

    // Business constraint: Age must be reasonable
    return core.AssertTrue(t, user.Age >= 0 && user.Age <= 150, "age range")
}

// ❌ BAD - Testing method behaviour, not object consistency
func validateUserBad(t core.T, user *User) bool {
    t.Helper()

    // This tests GetName() method behaviour, not object consistency
    if !core.AssertEqual(t, "Alice", user.GetName(), "name method") {
        return false
    }

    // This tests IsValid() method behaviour, not internal state
    return core.AssertTrue(t, user.IsValid(), "validity method")
}
```

#### 2. Early Return vs Continued Checking

**Choose based on validation purpose:**

```go
// ✅ EARLY RETURN - For dependent validations
func validateConfigWithDependencies(t core.T, config *Config) bool {
    t.Helper()
    if !core.AssertNotEqual(t, "", config.DatabaseURL, "database URL") {
        return false // Cannot validate the connection without a URL
    }
    if !core.AssertTrue(t, config.IsValidDatabaseURL(), "database URL format") {
        return false
    }
    return core.AssertTrue(t, config.CanConnect(), "database connection")
}

// ✅ CONTINUED CHECKING - For independent validations
func validateUserProfile(t core.T, profile *UserProfile) bool {
    t.Helper()
    valid := true
    if !core.AssertNotEqual(t, "", profile.FirstName, "first name") {
        valid = false
    }
    if !core.AssertNotEqual(t, "", profile.LastName, "last name") {
        valid = false
    }
    return valid
}
```

#### 3. Structure Validation for Complex Types

**Validate nested structures and collections systematically:**

```go
// ✅ COMPREHENSIVE - Complex structure validation
func validateOrder(t core.T, order *Order) bool {
    t.Helper()
    // Factory logic handles nil/not-nil automatically

    // Validate top-level fields
    if !core.AssertNotEqual(t, "", order.ID, "order ID") {
        return false
    }

    if !core.AssertTrue(t, order.Total >= 0, "order total non-negative") {
        return false
    }

    // Validate collection structure
    if !core.AssertTrue(t, len(order.Items) > 0, "order has items") {
        return false
    }

    // Validate each item in collection
    for i, item := range order.Items {
        if !core.AssertNotEqual(t, "", item.Name, fmt.Sprintf("item %d name", i)) {
            return false
        }
        if !core.AssertTrue(t, item.Price >= 0, fmt.Sprintf("item %d price", i)) {
            return false
        }
    }

    // Validate nested object consistency
    if order.Customer != nil {
        if !core.AssertNotEqual(t, "", order.Customer.Email, "customer email") {
            return false
        }
        // Cross-field consistency between order and customer
        if !core.AssertEqual(t, order.CustomerID, order.Customer.ID, "customer ID consistency") {
            return false
        }
    }

    return true
}
```

#### 4. Performance Considerations

**Choose validation strategies based on performance requirements:**

```go
// High-performance validation - minimal allocations
func validateConfigFast(t core.T, config *Config) bool {
    t.Helper()
    if config.Host == "" {
        t.Errorf("config host empty")
        return false
    }
    return config.Port > 0 && config.Port <= 65535
}

// Comprehensive validation - detailed reporting
func validateConfigDetailed(t core.T, config *Config) bool {
    t.Helper()
    valid := true
    if !core.AssertNotEqual(t, "", config.Host, "host") {
        valid = false
    }
    if !core.AssertTrue(t, config.Port > 0 && config.Port <= 65535, "port range") {
        valid = false
    }
    return valid
}
```

### Advanced Validation Patterns

#### 1. Conditional Validation

**Validate based on object state or field values:**

```go
func validateUserAccount(t core.T, account *UserAccount) bool {
    t.Helper()
    valid := true

    if !core.AssertNotEqual(t, "", account.Username, "username") {
        valid = false
    }

    // Conditional validation based on account type
    if account.Type == "premium" {
        if !core.AssertTrue(t, account.CreditLimit > 0, "premium credit limit") {
            valid = false
        }
    }
    if account.Type == "business" {
        if !core.AssertTrue(t, len(account.TeamMembers) > 0, "business team members") {
            valid = false
        }
    }

    return valid
}
```

#### 2. Collection Validation with Different Strategies

**Validate collections using appropriate strategies:**

```go
// Strategy 1: Early return for critical collection properties
func validateOrderStrict(t core.T, order *Order) bool {
    t.Helper()
    if !core.AssertTrue(t, len(order.Items) > 0, "order has items") {
        return false
    }
    for i, item := range order.Items {
        if !core.AssertTrue(t, item.Price > 0, fmt.Sprintf("item %d positive price", i)) {
            return false
        }
    }
    return true
}

// Strategy 2: Continued checking for independent properties
func validateOrderPermissive(t core.T, order *Order) bool {
    t.Helper()
    valid := true
    if !core.AssertNotEqual(t, "", order.CustomerID, "customer ID") {
        valid = false
    }
    if len(order.Items) > 0 {
        for i, item := range order.Items {
            if !core.AssertNotEqual(t, "", item.Name, fmt.Sprintf("item %d name", i)) {
                valid = false
            }
        }
    }
    return valid
}
```

### Validation Function Anti-patterns

#### 1. Redundant Nil Checking

```go
// ❌ ANTI-PATTERN - Factory handles nil checking
func validateUserBad(t core.T, user *User) bool {
    t.Helper()
    // DON'T DO THIS - factory logic handles nil checking
    if !core.AssertNotNil(t, user, "user") {
        return false
    }
    return core.AssertNotEqual(t, "", user.Name, "name")
}

// ✅ CORRECT - Focus on internal consistency only
func validateUser(t core.T, user *User) bool {
    t.Helper()
    return core.AssertNotEqual(t, "", user.Name, "name")
}
```

#### 2. Testing Method Behaviour

```go
// ❌ ANTI-PATTERN - Testing method behaviour, not object consistency
func validateUserBad(t core.T, user *User) bool {
    t.Helper()

    // This tests GetName() method, not object internal state
    if !core.AssertEqual(t, "expected", user.GetName(), "name method") {
        return false
    }

    // This tests IsValid() method, not internal consistency
    return core.AssertTrue(t, user.IsValid(), "validity method")
}

// ✅ CORRECT - Test object internal consistency
func validateUser(t core.T, user *User) bool {
    t.Helper()
    if !core.AssertNotEqual(t, "", user.Name, "name field") {
        return false
    }
    if user.Type == "admin" && !core.AssertTrue(t, len(user.Permissions) > 0, "admin permissions") {
        return false
    }
    return true
}
```

#### 3. Complex Logic Better Suited for Simple Tests

```go
// ❌ ANTI-PATTERN - Complex validation in factory test
func validateComplexUserBad(t core.T, user *User) bool {
    t.Helper()
    // This is too complex for a factory validation function
    if user.Type == "admin" {
        if len(user.Permissions) < 5 {
            return false
        }
        // ... more complex logic
    }
    return true
}

// ✅ CORRECT - Use simple test for complex validation
func TestCreateComplexUser(t *testing.T) {
    adminUser := NewUser("admin", "admin@test.com", "admin")
    core.AssertNotNil(t, adminUser, "admin user")
    core.AssertEqual(t, "admin", adminUser.Type, "user type")
    core.AssertTrue(t, len(adminUser.Permissions) >= 5, "admin permissions")
}
```

### When NOT to Use Validation Functions

**Skip validation functions when:**

1. **Factory always succeeds**: If the factory cannot fail in meaningful ways.
2. **Simple success validation**: When you only care that the factory returns non-nil.
3. **Complex validation needed**: When validation logic is more complex than the factory itself.
4. **Method testing focus**: When you're testing method behaviour, not object consistency.

```go
// Example: Simple factory that always works
testCases := []core.TestCase{
    testutils.NewFactoryTestCase("create simple object",
        NewSimpleObject, "NewSimpleObject", false, nil), // No validation needed
}
```

## Error Testing Strategy Guide

### Error Testing Philosophy

**Comprehensive error testing requires both success and failure scenarios with specific and general error validation.**

Error test cases support a specific error validation through the `errorIs`
parameter for precise error checking. This enables complete coverage of error
handling logic in TestCase implementations.

### Core Error Testing Strategies

#### 1. Complete Error Coverage Pattern

**Test both success and all failure modes:**

```go
var ErrInvalidInput = errors.New("invalid input")
var ErrTooLarge = errors.New("value too large")

func NewValidatedStruct(value int) (*MyStruct, error) {
    if value < 0 {
        return nil, ErrInvalidInput
    }
    if value > 1000 {
        return nil, ErrTooLarge
    }
    return &MyStruct{Value: value}, nil
}

// Test different error scenarios within the same context
testCases := []core.TestCase{
    // Success cases
    testutils.NewFactoryErrorOneArgTestCase("accepts valid values",
        NewValidatedStruct, "NewValidatedStruct", 10, false, nil, validateMyStruct),
    testutils.NewFactoryErrorOneArgTestCase("accepts boundary values",
        NewValidatedStruct, "NewValidatedStruct", 1000, false, nil, validateMyStruct),

    // Specific error cases
    testutils.NewFactoryErrorOneArgTestCase("rejects negative values",
        NewValidatedStruct, "NewValidatedStruct", -1, true, ErrInvalidInput, nil),
    testutils.NewFactoryErrorOneArgTestCase("rejects large values",
        NewValidatedStruct, "NewValidatedStruct", 1001, true, ErrTooLarge, nil),

    // Any error cases (errorIs == nil) - CRITICAL for coverage
    testutils.NewFactoryErrorOneArgTestCase("rejects invalid input any error",
        NewValidatedStruct, "NewValidatedStruct", -1, true, nil, nil),
}
core.RunTestCases(t, testCases)
```

#### 2. Error Hierarchy Testing

**Test error inheritance and wrapping:**

```go
// Test error types and hierarchy
testCases := []core.TestCase{
    testutils.NewFunctionErrorOneArgTestCase("wraps network errors",
        ProcessRequest, "ProcessRequest", "invalid-url", "", true, ErrNetworkFailure),
    testutils.NewFunctionErrorOneArgTestCase("wraps timeout errors",
        ProcessRequest, "ProcessRequest", "slow-endpoint", "", true, ErrTimeout),
    testutils.NewFunctionErrorOneArgTestCase("wraps any network error",
        ProcessRequest, "ProcessRequest", "network-fail", "", true, nil),
}
core.RunTestCases(t, testCases)
```

#### 3. Error Message vs Error Type Testing

**Test specific errors when type matters, nil for general error detection:**

```go
testCases := []core.TestCase{
    // Type-only testing (most common)
    testutils.NewFactoryErrorOneArgTestCase("rejects invalid input",
        NewValidatedStruct, "NewValidatedStruct", -1, true, ErrInvalidInput, nil),

    // Any error testing - validates error detection without type constraints
    testutils.NewFactoryErrorOneArgTestCase("detects any validation error",
        NewValidatedStruct, "NewValidatedStruct", -1, true, nil, nil),
}
core.RunTestCases(t, testCases)

// When you need to test error message content, use simple tests
func TestErrorMessage(t *testing.T) {
    _, err := NewValidatedStruct(-1)
    core.AssertError(t, err, "validation error")
    core.AssertContains(t, err.Error(), "negative", "error message content")
}
```

#### 4. Error Context and State Testing

**Test errors that depend on object state or context:**

```go
// Test error conditions that depend on object state
func TestStateBasedErrors(t *testing.T) {
    closedWriter := &Writer{closed: true}
    openWriter := &Writer{closed: false}

    testCases := []core.TestCase{
        testutils.NewErrorTestCase("write to closed writer fails",
            (*Writer).Write, "Write", closedWriter, true, ErrWriterClosed),
        testutils.NewErrorTestCase("write to open writer succeeds",
            (*Writer).Write, "Write", openWriter, false, nil),
        testutils.NewErrorTestCase("close already closed writer succeeds",
            (*Writer).Close, "Close", closedWriter, false, nil),
    }
    core.RunTestCases(t, testCases)
}
```

### Advanced Error Testing Patterns

#### 1. Conditional Error Testing

**Test errors that depend on specific conditions:**

```go
func TestConditionalErrors(t *testing.T) {
    testCases := []core.TestCase{
        // Permission-based errors
        testutils.NewGetterErrorOneArgTestCase("admin can access",
            (*Service).GetAdminData, "GetAdminData", adminService, "data", "result", false, nil),
        testutils.NewGetterErrorOneArgTestCase("user cannot access",
            (*Service).GetAdminData, "GetAdminData", userService, "data", "", true, ErrPermissionDenied),

        // State-based errors
        testutils.NewErrorTestCase("connected service succeeds",
            (*Service).Ping, "Ping", connectedService, false, nil),
        testutils.NewErrorTestCase("disconnected service fails",
            (*Service).Ping, "Ping", disconnectedService, true, ErrNotConnected),
    }
    core.RunTestCases(t, testCases)
}
```

#### 2. Timeout and Resource Error Testing

**Test functions that can timeout or exhaust resources:**

```go
testCases := []core.TestCase{
    testutils.NewFunctionErrorTwoArgsTestCase("operation succeeds within timeout",
        ProcessWithTimeout, "ProcessWithTimeout", "fast-data", 1000, "result", false, nil),
    testutils.NewFunctionErrorTwoArgsTestCase("operation times out",
        ProcessWithTimeout, "ProcessWithTimeout", "slow-data", 100, "", true, ErrTimeout),
    testutils.NewFunctionErrorTwoArgsTestCase("timeout any error",
        ProcessWithTimeout, "ProcessWithTimeout", "slow-data", 100, "", true, nil),
}
```

#### 3. Error Recovery and Retry Testing

**Test functions that handle error recovery:**

```go
testCases := []core.TestCase{
    testutils.NewFunctionErrorOneArgTestCase("retry succeeds after failure",
        RetryOperation, "RetryOperation", "transient-fail", "success", false, nil),
    testutils.NewFunctionErrorOneArgTestCase("retry fails permanently",
        RetryOperation, "RetryOperation", "permanent-fail", "", true, ErrPermanentFailure),
    testutils.NewFunctionErrorOneArgTestCase("max retries exceeded",
        RetryOperation, "RetryOperation", "max-retries", "", true, ErrMaxRetriesExceeded),
}
```

### Error Testing Anti-patterns

#### 1. Incomplete Error Coverage

```go
// ❌ ANTI-PATTERN - Only testing specific errors
func TestIncompleteErrorCoverage(t *testing.T) {
    testCases := []core.TestCase{
        // Missing any-error cases (errorIs == nil)
        testutils.NewFactoryErrorOneArgTestCase("specific error only",
            NewValidated, "NewValidated", "invalid", true, ErrSpecific, nil),
        // This does not test the errorIs == nil branch in TestCase logic
    }
    core.RunTestCases(t, testCases)
}

// ✅ CORRECT - Complete error coverage
func TestCompleteErrorCoverage(t *testing.T) {
    testCases := []core.TestCase{
        // Success case
        testutils.NewFactoryErrorOneArgTestCase("valid input",
            NewValidated, "NewValidated", "valid", false, nil, validateObj),
        // Specific error case
        testutils.NewFactoryErrorOneArgTestCase("specific error",
            NewValidated, "NewValidated", "invalid", true, ErrSpecific, nil),
        // Any error case - CRITICAL for complete coverage
        testutils.NewFactoryErrorOneArgTestCase("any error",
            NewValidated, "NewValidated", "invalid", true, nil, nil),
    }
    core.RunTestCases(t, testCases)
}
```

#### 2. Testing Only Specific Errors

```go
// ❌ ANTI-PATTERN - Missing general error detection
testCases := []core.TestCase{
    testutils.NewFactoryErrorOneArgTestCase("empty value",
        NewUser, "NewUser", "", true, ErrEmptyName, nil),
    // Missing errorIs == nil cases
}

// ✅ CORRECT - Include both specific and general error cases
testCases := []core.TestCase{
    testutils.NewFactoryErrorOneArgTestCase("empty name specific",
        NewUser, "NewUser", "", true, ErrEmptyName, nil),
    testutils.NewFactoryErrorOneArgTestCase("empty name any error",
        NewUser, "NewUser", "", true, nil, nil),
}
```

#### 3. Mixing Error Types in Same Test

```go
// ❌ ANTI-PATTERN - Mixing unrelated error conditions
testCases := []core.TestCase{
    testutils.NewFactoryErrorOneArgTestCase("network error",
        CreateConnection, "CreateConnection", "invalid-host", true, ErrNetwork, nil),
    testutils.NewFactoryErrorOneArgTestCase("validation error",
        CreateUser, "CreateUser", "", true, ErrValidation, nil), // Unrelated function!
}

// ✅ CORRECT - Group related error conditions
func TestConnectionErrors(t *testing.T) {
    testCases := []core.TestCase{
        testutils.NewFactoryErrorOneArgTestCase("network error",
            CreateConnection, "CreateConnection", "invalid-host", true, ErrNetwork, nil),
        testutils.NewFactoryErrorOneArgTestCase("any connection error",
            CreateConnection, "CreateConnection", "invalid-host", true, nil, nil),
    }
    core.RunTestCases(t, testCases)
}
```

### Meta-Testing Error Validation

**For testing TestCase implementations themselves (testutils package development):**

```go
// Meta-testing: Verify error TestCase implementations work correctly
func TestGetterErrorTestCase(t *testing.T) {
    t.Run("success cases", func(t *testing.T) {
        testutils.RunSuccessCases(t, makeGetterErrorSuccessCases())
    })
    t.Run("failure cases", func(t *testing.T) {
        testutils.RunFailureCases(t, makeGetterErrorFailureCases())
    })
}

func makeGetterErrorSuccessCases() []core.TestCase {
    return []core.TestCase{
        // TestCase should pass when expectations match reality
        testutils.NewGetterErrorOneArgTestCase("valid input no error",
            (*Parser).Parse, "Parse", validParser, "input", "result", false, nil),
        testutils.NewGetterErrorOneArgTestCase("invalid input specific error",
            (*Parser).Parse, "Parse", errorParser, "input", "", true, ErrParseError),
        testutils.NewGetterErrorOneArgTestCase("invalid input any error",
            (*Parser).Parse, "Parse", errorParser, "input", "", true, nil),
    }
}

func makeGetterErrorFailureCases() []core.TestCase {
    return []core.TestCase{
        // TestCase should fail when expectations do not match reality
        testutils.NewGetterErrorOneArgTestCase("expect error but get none",
            (*Parser).Parse, "Parse", validParser, "input", "", true, nil),
        testutils.NewGetterErrorOneArgTestCase("expect no error but get one",
            (*Parser).Parse, "Parse", errorParser, "input", "result", false, nil),
        testutils.NewGetterErrorOneArgTestCase("wrong specific error",
            (*Parser).Parse, "Parse", errorParser, "input", "", true, ErrWrongType),
    }
}
```

### Error Testing Best Practices

#### 1. Comprehensive Coverage Strategy

**Include all three error testing scenarios for complete coverage:**

**Why all three are essential:**

- **Success case**: Validates normal operation path.
- **Specific error case**: Validates exact error type matching logic.
- **Any error case**: Validates general error detection logic.
- **Complete coverage**: Exercises all branches in TestCase.Test() error handling.

#### 2. Error Test Organisation

**Group related error conditions together:**

```go
func TestNetworkOperations(t *testing.T) {
    t.Run("connection errors", func(t *testing.T) {
        core.RunTestCases(t, makeConnectionErrorTestCases())
    })
    t.Run("timeout errors", func(t *testing.T) {
        core.RunTestCases(t, makeTimeoutErrorTestCases())
    })
    t.Run("protocol errors", func(t *testing.T) {
        core.RunTestCases(t, makeProtocolErrorTestCases())
    })
}

func makeConnectionErrorTestCases() []core.TestCase {
    return []core.TestCase{
        newNetworkTestCase("connection succeeds", "valid-host", false, nil),
        newNetworkTestCase("connection fails specific", "invalid-host", true, ErrConnectionFailed),
        newNetworkTestCase("connection fails any error", "invalid-host", true, nil),
    }
}
```

#### 3. Error Test Naming

**Use descriptive names that explain the error condition:**

```go
// ✅ EXCELLENT - Clear error condition descriptions
testCases := []core.TestCase{
    newValidationTestCase("accepts valid email format", "user@domain.com", false, nil),
    newValidationTestCase("rejects empty email specific error", "", true, ErrEmptyEmail),
    newValidationTestCase("rejects invalid format specific error", "invalid", true, ErrInvalidFormat),
    newValidationTestCase("detects empty email any error", "", true, nil),
    newValidationTestCase("detects invalid format any error", "invalid", true, nil),
}

// ❌ BAD - Vague error descriptions
testCases := []core.TestCase{
    newValidationTestCase("test1", "user@domain.com", false, nil),
    newValidationTestCase("error case", "", true, ErrEmpty),
    newValidationTestCase("bad input", "invalid", true, nil),
}
```

### Summary: Error Testing Strategy

**For complete error testing in your projects:**

1. **Test all paths**: Success, specific errors, and any-error scenarios.
2. **Group related errors**: Organise by error category or function.
3. **Use descriptive names**: Make error conditions clear from test names.
4. **Include coverage cases**: Both `errorIs == specificError` and `errorIs == nil`.
5. **Separate complex error logic**: Use simple tests when error handling is complex.

## Meta-Testing vs Normal Usage

**This section covers two distinct usage patterns for testutils.**

### Meta-Testing (testutils package development)

**Meta-testing is for testing TestCase implementations themselves.** This is used internally by the testutils package to verify that TestCase types correctly handle success and failure detection.

#### Meta-Testing Utilities

**Complete enumeration of meta-testing functions:**

```go
// Core meta-testing utilities
testutils.RunSuccessCases(t core.T, cases []core.TestCase)
testutils.RunFailureCases(t core.T, cases []core.TestCase)
testutils.Run(t core.T, name string, fn func(core.T))
testutils.RunTest(name string, testFunc func(*testing.T)) bool

// Meta-testing infrastructure types
testutils.NewArrayTestCaseData(name string, failIndexes []int, totalCount int) *ArrayTestCaseData
testutils.NewDummyTestCase(name string, shouldPass bool) DummyTestCase
```

**Usage context**: These utilities are for testutils package development only, not for normal application testing.

#### RunSuccessCases and RunFailureCases

**Core utilities for meta-testing TestCase implementations:**

```go
// Meta-testing: Verify GetterTestCase works correctly
func TestGetterTestCase(t *testing.T) {
    t.Run("success cases", func(t *testing.T) {
        testutils.RunSuccessCases(t, makeGetterSuccessCases())
    })
    t.Run("failure cases", func(t *testing.T) {
        testutils.RunFailureCases(t, makeGetterFailureCases())
    })
}

func makeGetterSuccessCases() []core.TestCase {
    return []core.TestCase{
        // These TestCases should PASS when run - they have correct expectations
        testutils.NewGetterTestCase("correct expectation",
            (*TestStruct).GetValue, "GetValue", &TestStruct{value: "test"}, "test"),
        testutils.NewGetterTestCase("another correct expectation",
            (*TestStruct).GetCount, "GetCount", &TestStruct{count: 42}, 42),
    }
}

func makeGetterFailureCases() []core.TestCase {
    return []core.TestCase{
        // These TestCases should FAIL when run - they have wrong expectations
        testutils.NewGetterTestCase("wrong expectation",
            (*TestStruct).GetValue, "GetValue", &TestStruct{value: "test"}, "wrong"),
        testutils.NewGetterTestCase("another wrong expectation",
            (*TestStruct).GetCount, "GetCount", &TestStruct{count: 42}, 100),
    }
}
```

#### ArrayTestCaseData for Controlled Testing

**Generate test arrays with specific pass/fail patterns:**

```go
// Meta-testing: Test array handling with controlled failures
func TestArrayHandling(t *testing.T) {
    // Create test scenario: tests 1,3,5 should fail, others pass
    data := testutils.NewArrayTestCaseData("mixed results", []int{1, 3, 5}, 8)
    testCases := data.Make()

    // Verify the pattern works as expected
    core.AssertEqual(t, 8, len(testCases), "test count")
    core.AssertEqual(t, []int{1, 3, 5}, data.FailIndexes(), "fail indexes")
    core.AssertEqual(t, []int{0, 2, 4, 6, 7}, data.PassIndexes(), "pass indexes")

    // Run the tests to verify pass/fail behaviour
    passCount := 0
    failCount := 0
    for _, tc := range testCases {
        if testutils.RunTest(tc.Name(), tc.Test) {
            passCount++
        } else {
            failCount++
        }
    }

    core.AssertEqual(t, 5, passCount, "pass count")
    core.AssertEqual(t, 3, failCount, "fail count")
}
```

#### DummyTestCase for Meta-Testing Infrastructure

**Use DummyTestCase for testing meta-testing utilities:**

```go
func TestRunSuccessCases(t *testing.T) {
    // Create cases that should pass
    passingCases := []core.TestCase{
        testutils.NewDummyTestCase("pass1", true),
        testutils.NewDummyTestCase("pass2", true),
    }

    // These should all succeed when run via RunSuccessCases
    testutils.RunSuccessCases(t, passingCases)
}

func TestRunFailureCases(t *testing.T) {
    // Create cases that should fail
    failingCases := []core.TestCase{
        testutils.NewDummyTestCase("fail1", false),
        testutils.NewDummyTestCase("fail2", false),
    }

    // These should all fail when run via RunFailureCases
    testutils.RunFailureCases(t, failingCases)
}
```

### Normal Usage (all other projects)

**Normal usage is for testing your application code using testutils TestCase types.**

#### Pattern 1: Single Test Set

```go
func makeUserValidationTestCases() []core.TestCase {
    return []core.TestCase{
        newUserValidateTestCase("valid user", validUser, false, nil),
        newUserValidateTestCase("invalid user", invalidUser, true, ErrValidation),
        newUserValidateTestCase("invalid user any error", invalidUser, true, nil),
    }
}

func TestUserValidation(t *testing.T) {
    core.RunTestCases(t, makeUserValidationTestCases())
}
```

#### Pattern 2: Multiple Test Sets with t.Run

```go
func TestUserMethods(t *testing.T) {
    t.Run("validation", func(t *testing.T) {
        core.RunTestCases(t, makeUserValidationTestCases())
    })
    t.Run("getters", func(t *testing.T) {
        core.RunTestCases(t, makeUserGetterTestCases())
    })
    t.Run("factories", func(t *testing.T) {
        core.RunTestCases(t, makeUserFactoryTestCases())
    })
}
```

#### Pattern 3: Systematic Success/Failure Testing for Custom TestCase Types

**When creating custom TestCase implementations, use the systematic pattern:**

```go
// Custom TestCase implementation
type parseComplexTestCase struct {
    name           string
    input          string
    expectedValue  string
    expectedTokens []string
    wantErr        bool
}

func (tc parseComplexTestCase) Name() string {
    return tc.name
}

func (tc parseComplexTestCase) Test(t *testing.T) {
    t.Helper()

    result, err := ParseComplex(tc.input)

    if tc.wantErr {
        core.AssertError(t, err, "parse error")
        return
    }

    core.AssertNoError(t, err, "parse")
    core.AssertEqual(t, tc.expectedValue, result.Value, "value")
    core.AssertSliceEqual(t, tc.expectedTokens, result.Tokens, "tokens")
}

// MANDATORY: Factory function
func newParseComplexTestCase(name, input, expectedValue string,
    expectedTokens []string, wantErr bool) parseComplexTestCase {
    return parseComplexTestCase{
        name:           name,
        input:          input,
        expectedValue:  expectedValue,
        expectedTokens: expectedTokens,
        wantErr:        wantErr,
    }
}

// MANDATORY: Systematic success/failure testing
func TestParseComplexTestCase(t *testing.T) {
    t.Run("success cases", func(t *testing.T) {
        testutils.RunSuccessCases(t, makeParseComplexSuccessCases())
    })
    t.Run("failure cases", func(t *testing.T) {
        testutils.RunFailureCases(t, makeParseComplexFailureCases())
    })
}

func makeParseComplexSuccessCases() []core.TestCase {
    return []core.TestCase{
        // These should pass when run - correct expectations
        newParseComplexTestCase("valid input", "input", "result", []string{"input"}, false),
        newParseComplexTestCase("invalid input", "bad", "", nil, true),
    }
}

func makeParseComplexFailureCases() []core.TestCase {
    return []core.TestCase{
        // These should fail when run - wrong expectations
        newParseComplexTestCase("wrong expectation", "input", "wrong", []string{"input"}, false),
        newParseComplexTestCase("wrong error expectation", "bad", "result", nil, false),
    }
}
```

**Key Distinction:**

- **Meta-testing**: Testing the TestCase implementations using RunSuccessCases/RunFailureCases.
- **Normal usage**: Using TestCase implementations to test your application code.

## Comprehensive Example

```go
package example_test

import (
    "testing"

    "darvaza.org/core"
    "darvaza.org/x/testutils"
)

// Specific factory functions for Calculator methods
func newCalculatorCreateTestCase(name string, initial int) core.TestCase {
    return testutils.NewFactoryOneArgTestCase(name,
        NewCalculator, "NewCalculator", initial, false, nil)
}

func newCalculatorGetValueTestCase(name string, calc *Calculator,
    expected int) core.TestCase {
    return testutils.NewGetterTestCase(name, (*Calculator).GetValue,
        "GetValue", calc, expected)
}

func newCalculatorAddTestCase(name string, calc *Calculator, add,
    expected int) core.TestCase {
    return testutils.NewGetterOneArgTestCase(name, (*Calculator).Add,
        "Add", calc, add, expected)
}

// Shared test logic for any calculator
func makeCalculatorTestCases(prefix string, calc *Calculator,
    initialValue int) []core.TestCase {
    return []core.TestCase{
        newCalculatorGetValueTestCase(prefix+" get initial value", calc,
            initialValue),
        newCalculatorAddTestCase(prefix+" add positive", calc, 5,
            initialValue+5),
        newCalculatorAddTestCase(prefix+" add negative", calc, -3,
            initialValue+2),
    }
}

// Test multiple calculators with shared logic
func TestCalculator(t *testing.T) {
    t.Run("factory", func(t *testing.T) {
        tests := []core.TestCase{
            newCalculatorCreateTestCase("create with initial value", 10),
            newCalculatorCreateTestCase("create with zero", 0),
        }
        core.RunTestCases(t, tests)
    })

    t.Run("calculator with value 10", func(t *testing.T) {
        calc := NewCalculator(10)
        core.RunTestCases(t, makeCalculatorTestCases("calculator 10", calc, 10))
    })

    t.Run("calculator with value 100", func(t *testing.T) {
        calc := NewCalculator(100)
        tests := makeCalculatorTestCases("calculator 100", calc, 100)
        core.RunTestCases(t, tests)
    })
}
```

## TestCase Compliance (When Using Table-Driven Tests)

**When using TestCase interface for table-driven tests, ALL files must meet
these 6 requirements:**

1. **TestCase Interface Validations**: `var _ TestCase = ...` declarations
   for all test case types.
2. **Factory Functions**: All TestCase types have `new{TestContext}TestCase()`
   functions (enables field alignment + logical parameters).
3. **Factory Usage**: All test case declarations use factory functions
   (no naked struct literals).
4. **RunTestCases Usage**: Test functions use `RunTestCases(t, cases)`
   instead of manual loops.
5. **Anonymous Functions**: No `t.Run("name", func(t *testing.T) { ... })`
   patterns longer than 3 lines.
6. **Test Case List Factories**: Complex test case lists use
   `makeTestFunctionTestCases()` factory functions.

These requirements apply **ONLY** to table-driven tests using TestCase, not
to standard t.Run() patterns.

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

## Factory Naming Conventions (MANDATORY)

### Test Case Factories

**Individual TestCase Factories**: Named after `new{TestContext}TestCase`

```go
func newParseURLTestCase(name, input string, expected *url.URL,
    wantErr bool) parseURLTestCase {
    return parseURLTestCase{
        name:     name,
        input:    input,
        expected: expected,
        wantErr:  wantErr,
    }
}

func newValidationTestCase(name, input string, wantErr bool) validationTestCase {
    return validationTestCase{
        name:    name,
        input:   input,
        wantErr: wantErr,
    }
}
```

### Test Set Factories

**Test Set Factories**: Named after `make{TestFunction}TestCases`, return `[]core.TestCase` for composability

```go
// Pattern: make{TestFunction}TestCases() []core.TestCase
func makeHTTPClientTestCases() []core.TestCase
func makeParseURLTestCases() []core.TestCase
func makeValidationTestCases() []core.TestCase
```

**Parameterised Test Set Factories**: Include parameter context

```go
func makeValidationTestCases(fieldName string) []core.TestCase
func makeUserTestCases(userType string) []core.TestCase
```

## Anonymous Function Limits (MANDATORY)

**Rule: Anonymous functions in `t.Run` are allowed ONLY if they are 3 lines
or shorter.**

```go
// ✅ ALLOWED - Short anonymous function (≤3 lines)
t.Run("nil input", func(t *testing.T) {
    result := ProcessInput(nil)
    core.AssertNil(t, result, "result")
})

// ❌ NEVER DO THIS - Long anonymous function (>3 lines)
t.Run("complex test", func(t *testing.T) {
    setup := createTestData()
    result := ProcessComplex(setup)
    validateResult(t, result)
    cleanUpTestData(setup)
})

// ✅ CORRECT - Extract to named function
func runTestComplexScenario(t *testing.T) {
    t.Helper()
    setup := createTestData()
    result := ProcessComplex(setup)
    validateResult(t, result)
    cleanUpTestData(setup)
}

func TestComplexFeature(t *testing.T) {
    t.Run("complex test", runTestComplexScenario)
}
```

## Best Practices

### 1. Choose the Right Pattern

**Always start with the simplest approach that achieves your testing goals:**

```go
// ✅ SINGLE TEST - Use simple function
func TestCalculateArea(t *testing.T) {
    area := CalculateArea(5, 10)
    core.AssertEqual(t, 50, area, "area")
}

// ✅ DIFFERENT BEHAVIOURS - Use t.Run() with named functions
func TestFileOperations(t *testing.T) {
    file := setupTestFile()
    defer file.Close()

    t.Run("read", runTestFileRead)
    t.Run("write", runTestFileWrite)
    t.Run("seek", runTestFileSeek)
}

// ✅ SAME BEHAVIOUR, DIFFERENT DATA - Use TestCase with custom factories
func TestPasswordValidation(t *testing.T) {
    testCases := []core.TestCase{
        newPasswordTestCase("valid strong password", "SecurePass123!", true, nil),
        newPasswordTestCase("too short", "123", false, ErrTooShort),
        newPasswordTestCase("no special chars", "Password123", false, ErrNoSpecialChars),
    }
    core.RunTestCases(t, testCases)
}
```

### 2. Creating Custom Test Factories (RECOMMENDED)

**Hide testutils complexity behind domain-specific interfaces:**

```go
// ✅ PREFERRED - Domain-specific wrappers
func newUserValidationTestCase(name, email string, expectValid bool,
    expectedErr error) core.TestCase {
    return testutils.NewFactoryErrorTwoArgsTestCase(name,
        NewUserWithValidation, "NewUserWithValidation",
        name, email, !expectValid, expectedErr, validateUser)
}

func newUserGetterTestCase(name string, user *User, field string,
    expected string) core.TestCase {
    var method func(*User) string
    switch field {
    case "name":  method = (*User).GetName
    case "email": method = (*User).GetEmail
    case "role":  method = (*User).GetRole
    }
    return testutils.NewGetterTestCase(name, method, field, user, expected)
}

// Usage becomes clean and readable
testCases := []core.TestCase{
    newUserValidationTestCase("valid user", "john@test.com", true, nil),
    newUserValidationTestCase("invalid email", "not-email", false, ErrInvalidEmail),
    newUserGetterTestCase("get user name", user, "name", "John"),
    newUserGetterTestCase("get user email", user, "email", "john@test.com"),
}
core.RunTestCases(t, testCases)

// ❌ AVOID - Raw testutils calls clutter tests
testutils.NewFactoryErrorTwoArgsTestCase("valid user",
    NewUserWithValidation, "NewUserWithValidation",
    "john", "john@test.com", false, nil, validateUser)
```

#### Advanced Pattern: Shared Test Logic with makeTestCases

```go
// Create reusable test sets for common scenarios
func makeUserCRUDTestCases(prefix string, userType string) []core.TestCase {
    return []core.TestCase{
        newUserCreateTestCase(prefix+" create", userType, true, nil),
        newUserUpdateTestCase(prefix+" update", userType, true, nil),
        newUserDeleteTestCase(prefix+" delete", userType, true, nil),
    }
}

func makeUserValidationTestCases(userType string) []core.TestCase {
    return []core.TestCase{
        newUserValidationTestCase("valid "+userType, "test@test.com", true, nil),
        newUserValidationTestCase("invalid "+userType+" email", "invalid", false, ErrInvalidEmail),
        newUserValidationTestCase("empty "+userType+" name", "", false, ErrEmptyName),
    }
}

// Compose test sets for comprehensive testing
func TestUserManagement(t *testing.T) {
    t.Run("admin users", func(t *testing.T) {
        var allTests []core.TestCase
        allTests = append(allTests, makeUserCRUDTestCases("admin", "admin")...)
        allTests = append(allTests, makeUserValidationTestCases("admin")...)
        core.RunTestCases(t, allTests)
    })

    t.Run("regular users", func(t *testing.T) {
        var allTests []core.TestCase
        allTests = append(allTests, makeUserCRUDTestCases("regular", "user")...)
        allTests = append(allTests, makeUserValidationTestCases("user")...)
        core.RunTestCases(t, allTests)
    })
}
```

### 3. Validation Functions

See Type Validation Function Design Guide above for complete patterns and examples.

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`](https://pkg.go.dev/darvaza.org/core).

## Compliance with darvaza.org Standards

The testutils package follows all mandatory patterns from `darvaza.org/core`:

- **TestCase Interface**: All types implement `core.TestCase`.
- **Factory Functions**: Mandatory factory functions for all TestCase creation.
- **RunTestCases Usage**: Designed for use with `core.RunTestCases()`.
- **Field Alignment**: Memory-optimised struct layouts.
- **Linting Compliance**: Meets all complexity and quality requirements.

## Summary

**Key principles for using testutils effectively:**

1. **Start simple**: Most tests should use simple assertions, not TestCase patterns
2. **Use TestCase only when appropriate**: 2+ similar scenarios with identical logic
3. **Create domain-specific factories**: Hide testutils complexity behind readable interfaces
4. **Focus on clarity**: The pattern should make tests easier to understand, not harder
5. **Follow the decision framework**: Use the checklist to determine the right approach
6. **Validation functions check internal consistency only**: Factory logic handles technical concerns
7. **Maintain strict compliance**: Follow all mandatory patterns when using TestCase
8. **Apply systematic testing**: Use makeXSuccessCases/FailureCases for custom TestCase implementations
9. **Understand meta-testing**: Use RunSuccessCases/RunFailureCases for testing TestCase implementations

**Remember**: These patterns are tools to improve test clarity and maintainability. When they do not serve that purpose, use simpler approaches.
