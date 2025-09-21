package testutils_test

import (
	"errors"
	"strconv"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/testutils"
)

// Example structs for demonstration

// Calculator demonstrates the method testing patterns.
type Calculator struct {
	value int
}

// NewCalculator demonstrates the factory testing patterns.
func NewCalculator(initial int) *Calculator {
	return &Calculator{value: initial}
}

// NewCalculatorFromString demonstrates factory with validation
func NewCalculatorFromString(s string) (*Calculator, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &Calculator{value: value}, nil
}

// NewValidatedCalculator demonstrates factory with domain validation
func NewValidatedCalculator(initial int) (*Calculator, error) {
	if initial < 0 {
		return nil, ErrNegativeValue
	}
	if initial > 1000 {
		return nil, ErrValueTooLarge
	}
	return &Calculator{value: initial}, nil
}

// Domain-specific errors for testing errorIs functionality
var (
	ErrNegativeValue = errors.New("calculator value cannot be negative")
	ErrValueTooLarge = errors.New("calculator value too large")
)

// TryNewCalculator demonstrates factory returning (value, bool)
func TryNewCalculator(initial int) (*Calculator, bool) {
	if initial < 0 {
		return nil, false
	}
	return &Calculator{value: initial}, true
}

// GetValue demonstrates getter method (method testing)
func (c *Calculator) GetValue() int {
	return c.value
}

// Add demonstrates method with argument (method testing)
func (c *Calculator) Add(n int) int {
	c.value += n
	return c.value
}

// Multiply demonstrates method with two arguments (method testing)
func (c *Calculator) Multiply(a, b int) int {
	c.value = c.value * a * b
	return c.value
}

// TryDivide demonstrates method returning (value, bool)
func (c *Calculator) TryDivide(divisor int) (int, bool) {
	if divisor == 0 {
		return 0, false
	}
	c.value = c.value / divisor
	return c.value, true
}

// ParseAndSet demonstrates method returning (value, error)
func (c *Calculator) ParseAndSet(s string) (int, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	c.value = value
	return c.value, nil
}

// Reset demonstrates error-only method
func (c *Calculator) Reset() error {
	if c.value == -1 {
		return errors.New("cannot reset calculator in error state")
	}
	c.value = 0
	return nil
}

// Pure functions for function testing

// CalculateSum demonstrates the simple function testing.
func CalculateSum() int {
	return 42
}

// AddNumbers demonstrates function with arguments
func AddNumbers(a, b int) int {
	return a + b
}

// FormatNumber demonstrates function with three arguments
func FormatNumber(value int, prefix, suffix string) string {
	return prefix + strconv.Itoa(value) + suffix
}

// TryParseInt demonstrates function returning (value, bool)
func TryParseInt(s string) (int, bool) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return value, true
}

// ParseFloat demonstrates function returning (value, error)
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Validation functions for factory testing.

func validateCalculator(t core.T, c *Calculator) bool {
	t.Helper()
	// NOTE: No nil check needed - factory logic handles nil/not-nil automatically
	return core.AssertTrue(t, c.value >= -1000, "calculator value in valid range")
}

func validatePositiveCalculator(t core.T, c *Calculator) bool {
	t.Helper()
	// NOTE: No nil check needed - factory logic handles nil/not-nil automatically
	return core.AssertTrue(t, c.value >= 0, "calculator value non-negative")
}

// Custom factory functions for cleaner test code.

// newCalcGetterTest creates a test for Calculator getter methods
func newCalcGetterTest(name string, calc *Calculator, expected int) core.TestCase {
	return testutils.NewGetterTestCase(name,
		(*Calculator).GetValue, "GetValue", calc, expected)
}

// newCalcAddTest creates a test for Calculator.Add method
func newCalcAddTest(name string, calc *Calculator, addValue, expected int) core.TestCase {
	return testutils.NewGetterOneArgTestCase(name,
		(*Calculator).Add, "Add", calc, addValue, expected)
}

// newCalcFactoryTest creates a test for NewCalculator factory
func newCalcFactoryTest(name string, initial int) core.TestCase {
	return testutils.NewFactoryOneArgTestCase(name,
		NewCalculator, "NewCalculator", initial, false, validateCalculator)
}

// newCalcValidatedFactoryTest creates a test for NewValidatedCalculator with error checking
func newCalcValidatedFactoryTest(name string, initial int, expectErr bool,
	expectedErr error) core.TestCase {
	return testutils.NewFactoryErrorOneArgTestCase(name,
		NewValidatedCalculator, "NewValidatedCalculator",
		initial, expectErr, expectedErr, validateCalculator)
}

// newParseTest creates a test for parsing functions
func newParseTest(name string, input string, expected float64,
	shouldFail bool) core.TestCase {
	return testutils.NewFunctionErrorOneArgTestCase(name,
		ParseFloat, "ParseFloat", input, expected, shouldFail, nil)
}

// TestExampleMethodTesting demonstrates testing methods that take *T as parameter
func TestExampleMethodTesting(t *testing.T) {
	calc := NewCalculator(10)

	testCases := []core.TestCase{
		newCalcGetterTest("GetValue returns current value", calc, 10),
		newCalcAddTest("Add increases value", calc, 5, 15),
		testutils.NewGetterTwoArgsTestCase("Multiply with two factors",
			(*Calculator).Multiply, "Multiply", calc, 2, 3, 90),
		testutils.NewGetterOKOneArgTestCase("TryDivide by valid divisor",
			(*Calculator).TryDivide, "TryDivide", calc, 3, 30, true),
		testutils.NewGetterErrorOneArgTestCase("ParseAndSet with valid string",
			(*Calculator).ParseAndSet, "ParseAndSet", calc, "42", 42, false, nil),
		testutils.NewErrorTestCase("Reset successful",
			(*Calculator).Reset, "Reset", calc, false, nil),
	}

	core.RunTestCases(t, testCases)
}

// TestExampleFactoryTesting demonstrates testing functions that return *T
func TestExampleFactoryTesting(t *testing.T) {
	testCases := []core.TestCase{
		newCalcFactoryTest("NewCalculator with zero", 0),
		newCalcFactoryTest("NewCalculator with initial value", 42),
		testutils.NewFactoryOKOneArgTestCase("TryNewCalculator with valid input",
			TryNewCalculator, "TryNewCalculator", 10, true, validateCalculator),
		testutils.NewFactoryErrorOneArgTestCase("NewCalculatorFromString with valid input",
			NewCalculatorFromString, "NewCalculatorFromString", "123", false, nil, validateCalculator),
		testutils.NewFactoryErrorOneArgTestCase("NewCalculatorFromString with invalid input",
			NewCalculatorFromString, "NewCalculatorFromString", "invalid", true, nil, nil),
		newCalcValidatedFactoryTest("NewValidatedCalculator with valid input",
			500, false, nil),
	}

	core.RunTestCases(t, testCases)
}

// TestExampleFunctionTesting demonstrates testing functions that return comparable values
func TestExampleFunctionTesting(t *testing.T) {
	testCases := []core.TestCase{
		testutils.NewFunctionTestCase("CalculateSum returns expected value",
			CalculateSum, "CalculateSum", 42),
		testutils.NewFunctionTwoArgsTestCase("AddNumbers with positive values",
			AddNumbers, "AddNumbers", 10, 15, 25),
		testutils.NewFunctionThreeArgsTestCase("FormatNumber with prefix and suffix",
			FormatNumber, "FormatNumber", 123, "Value: ", " units", "Value: 123 units"),
		testutils.NewFunctionOKOneArgTestCase("TryParseInt with valid string",
			TryParseInt, "TryParseInt", "456", 456, true),
		newParseTest("ParseFloat with valid string", "3.14", 3.14, false),
		newParseTest("ParseFloat with invalid string", "invalid", 0.0, true),
	}

	core.RunTestCases(t, testCases)
}

// TestExampleComprehensive demonstrates a complete test using all patterns
func TestExampleComprehensive(t *testing.T) {
	// This example shows how you would structure a complete test
	// combining all three testing patterns

	// Create test instances
	calc := NewCalculator(100)

	// Combine all test types
	testCases := []core.TestCase{
		// Method testing
		testutils.NewGetterTestCase("calculator returns initial value",
			(*Calculator).GetValue, "GetValue", calc, 100),
		testutils.NewGetterOneArgTestCase("calculator adds correctly",
			(*Calculator).Add, "Add", calc, 50, 150),

		// Factory testing
		testutils.NewFactoryOneArgTestCase("create calculator with value",
			NewCalculator, "NewCalculator", 42, false, validateCalculator),
		testutils.NewFactoryOKOneArgTestCase("try create calculator success",
			TryNewCalculator, "TryNewCalculator", 10, true, validatePositiveCalculator),

		// Function testing
		testutils.NewFunctionTestCase("calculate fixed sum",
			CalculateSum, "CalculateSum", 42),
		testutils.NewFunctionTwoArgsTestCase("add two numbers",
			AddNumbers, "AddNumbers", 20, 30, 50),
	}

	// Run all tests
	core.RunTestCases(t, testCases)
}

// TestExampleErrorHandling demonstrates error testing patterns
func TestExampleErrorHandling(t *testing.T) {
	calc := NewCalculator(-1)

	testCases := []core.TestCase{
		testutils.NewErrorTestCase("Reset fails in error state",
			(*Calculator).Reset, "Reset", calc, true, nil),
		testutils.NewGetterErrorOneArgTestCase("ParseAndSet with invalid string",
			(*Calculator).ParseAndSet, "ParseAndSet", calc, "not-a-number", 0, true, nil),
		testutils.NewFactoryOKOneArgTestCase("TryNewCalculator with negative value",
			TryNewCalculator, "TryNewCalculator", -5, false, nil),
		testutils.NewFunctionErrorOneArgTestCase("ParseFloat with invalid input",
			ParseFloat, "ParseFloat", "not-a-float", 0.0, true, nil),
	}

	core.RunTestCases(t, testCases)
}

// TestExampleTypeValidation demonstrates factory testing with type validation
func TestExampleTypeValidation(t *testing.T) {
	testCases := []core.TestCase{
		testutils.NewFactoryOneArgTestCase("create positive calculator",
			NewCalculator, "NewCalculator", 42, false, validatePositiveCalculator),
		testutils.NewFactoryOneArgTestCase("create calculator without validation",
			NewCalculator, "NewCalculator", 0, false, nil),
	}

	core.RunTestCases(t, testCases)
}

// TestExampleValidationFailure demonstrates testing validation failure cases
func TestExampleValidationFailure(t *testing.T) {
	// Simple factory that returns nil
	createNilCalculator := func() *Calculator { return nil }

	testCases := []core.TestCase{
		// When factory returns nil, we expect nil and validation is not called
		testutils.NewFactoryTestCase("nil calculator returns nil",
			createNilCalculator, "createNilCalculator", true, nil),

		// Test specific validation errors using errorIs
		newCalcValidatedFactoryTest("negative value fails validation",
			-10, true, ErrNegativeValue),
		newCalcValidatedFactoryTest("large value fails validation",
			2000, true, ErrValueTooLarge),
		newCalcValidatedFactoryTest("valid value passes",
			100, false, nil),

		// Test string parsing errors
		testutils.NewFactoryErrorOneArgTestCase("invalid string fails parsing",
			NewCalculatorFromString, "NewCalculatorFromString", "not-a-number", true, strconv.ErrSyntax, nil),
	}

	core.RunTestCases(t, testCases)
}

// makeCalculatorOperationsTestCases creates test cases for calculator operations
func makeCalculatorOperationsTestCases() []core.TestCase {
	calc1 := NewCalculator(100)
	calc2 := NewCalculator(200)
	calc3 := NewCalculator(300)

	return []core.TestCase{
		// Clean, readable, and focused on the test intent
		newCalcGetterTest("calc1 initial value", calc1, 100),
		newCalcGetterTest("calc2 initial value", calc2, 200),
		newCalcGetterTest("calc3 initial value", calc3, 300),

		newCalcAddTest("calc1 add 50", calc1, 50, 150),
		newCalcAddTest("calc2 add -50", calc2, -50, 150),
		newCalcAddTest("calc3 add -150", calc3, -150, 150),
	}
}

// makeFactoryValidationTestCases creates test cases for factory validation
func makeFactoryValidationTestCases() []core.TestCase {
	return []core.TestCase{
		// Much cleaner than repeating all the testutils parameters
		newCalcValidatedFactoryTest("accept minimum", 0, false, nil),
		newCalcValidatedFactoryTest("accept middle", 500, false, nil),
		newCalcValidatedFactoryTest("accept maximum", 1000, false, nil),
		newCalcValidatedFactoryTest("reject negative", -1, true, ErrNegativeValue),
		newCalcValidatedFactoryTest("reject too large", 1001, true, ErrValueTooLarge),
	}
}

// makeParsingOperationsTestCases creates test cases for parsing operations
func makeParsingOperationsTestCases() []core.TestCase {
	return []core.TestCase{
		// Domain-specific and clear
		newParseTest("parse integer", "42", 42.0, false),
		newParseTest("parse decimal", "3.14159", 3.14159, false),
		newParseTest("parse scientific", "1.23e4", 12300.0, false),
		newParseTest("parse invalid", "not-a-number", 0.0, true),
		newParseTest("parse empty", "", 0.0, true),
	}
}

// TestExampleCustomFactories demonstrates the benefit of custom factory functions
// This shows how custom factories make tests more readable and maintainable
func TestExampleCustomFactories(t *testing.T) {
	// Without custom factories, this would be verbose and repetitive
	// With custom factories, the intent is clear and changes are easy

	t.Run("calculator operations", func(t *testing.T) {
		core.RunTestCases(t, makeCalculatorOperationsTestCases())
	})

	t.Run("factory validation", func(t *testing.T) {
		core.RunTestCases(t, makeFactoryValidationTestCases())
	})

	t.Run("parsing operations", func(t *testing.T) {
		core.RunTestCases(t, makeParsingOperationsTestCases())
	})
}
