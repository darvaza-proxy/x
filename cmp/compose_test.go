package cmp

import (
	"strconv"
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var _ core.TestCase = composeTestCase[int]{}

// composeTestCase is a generic test case for Compose function
type composeTestCase[T any] struct {
	matcher  Matcher[T]
	input    T
	name     string
	expected bool
}

func newComposeTestCase[T any](name string, input T, expected bool, matcher Matcher[T]) composeTestCase[T] {
	return composeTestCase[T]{
		name:     name,
		input:    input,
		expected: expected,
		matcher:  matcher,
	}
}

func (tc composeTestCase[T]) Name() string {
	return tc.name
}

func (tc composeTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := tc.matcher.Match(tc.input)
	core.AssertEqual(t, tc.expected, result, "Match(%+v)", tc.input)
}

// TestCompose verifies that the Compose function properly creates
// matchers that apply transformations before matching.
func TestCompose(t *testing.T) {
	t.Run("with struct fields", runTestComposeStructFields)
	t.Run("with type transformations", runTestComposeTypeTransformations)
	t.Run("with accessor returning false", runTestComposeAccessorFalse)
	t.Run("nested composition", runTestComposeNested)
	t.Run("complex conditions", runTestComposeComplex)
}

func runTestComposeStructFields(t *testing.T) {
	t.Helper()
	type Person struct {
		Name string
		Age  int
	}

	// Create a matcher for people older than 18
	isAdult := Compose(
		func(p Person) (int, bool) { return p.Age, true },
		MatchGtEq(18),
	)

	tests := []composeTestCase[Person]{
		newComposeTestCase("adult", Person{"Alice", 30}, true, isAdult),
		newComposeTestCase("child", Person{"Bob", 12}, false, isAdult),
		newComposeTestCase("exact boundary", Person{"Charlie", 18}, true, isAdult),
		newComposeTestCase("just under boundary", Person{"David", 17}, false, isAdult),
	}

	core.RunTestCases(t, tests)

	// Create a matcher to check if a person's name starts with 'A'
	nameStartsWithA := Compose(
		func(p Person) (string, bool) { return p.Name, true },
		MatchFunc[string](func(s string) bool {
			return len(s) > 0 && s[0] == 'A'
		}),
	)

	nameTests := []composeTestCase[Person]{
		newComposeTestCase("starts with A", Person{"Alice", 30}, true, nameStartsWithA),
		newComposeTestCase("starts with different letter", Person{"Bob", 25}, false, nameStartsWithA),
		newComposeTestCase("empty name", Person{"", 20}, false, nameStartsWithA),
	}

	core.RunTestCases(t, nameTests)
}

func runTestComposeTypeTransformations(t *testing.T) {
	t.Helper()
	// Convert int to string and check if it has a specific length
	hasThreeDigits := Compose(
		func(i int) (string, bool) { return strconv.Itoa(i), true },
		MatchFunc[string](func(s string) bool {
			return len(s) == 3
		}),
	)

	tests := []composeTestCase[int]{
		newComposeTestCase("three digits", 123, true, hasThreeDigits),
		newComposeTestCase("two digits", 45, false, hasThreeDigits),
		newComposeTestCase("four digits", 1000, false, hasThreeDigits),
		newComposeTestCase("negative three digits", -123, false, hasThreeDigits), // has 4 chars due to minus sign
	}

	core.RunTestCases(t, tests)
}

func runTestComposeAccessorFalse(t *testing.T) {
	t.Helper()
	// Accessor that fails for even numbers
	failsForEven := Compose(
		func(i int) (int, bool) { return i, i%2 != 0 },
		MatchGt(0),
	)

	tests := []composeTestCase[int]{
		newComposeTestCase("odd positive", 3, true, failsForEven),
		newComposeTestCase("odd negative", -5, false, failsForEven),  // fails the Gt(0) check
		newComposeTestCase("even positive", 4, false, failsForEven),  // accessor returns false
		newComposeTestCase("even negative", -2, false, failsForEven), // accessor returns false
		newComposeTestCase("zero", 0, false, failsForEven),           // accessor returns false (0 is even)
	}

	core.RunTestCases(t, tests)
}

func runTestComposeNested(t *testing.T) {
	t.Helper()
	type Address struct {
		Street string
		City   string
		Zip    string
	}

	newAddress := func(street, city, zip string) Address {
		return Address{
			Street: street,
			City:   city,
			Zip:    zip,
		}
	}

	type Person struct {
		Address Address
		Name    string
		Age     int
	}

	newPerson := func(name string, age int, address Address) Person {
		return Person{
			Address: address,
			Name:    name,
			Age:     age,
		}
	}

	// Create a matcher that checks if a person's city starts with 'New'
	isFromNewCity := Compose(
		func(p Person) (Address, bool) { return p.Address, true },
		Compose(
			func(a Address) (string, bool) { return a.City, true },
			MatchFunc[string](func(s string) bool {
				return len(s) >= 3 && s[0:3] == "New"
			}),
		),
	)

	tests := []composeTestCase[Person]{
		newComposeTestCase(
			"from new city",
			newPerson("Alice", 30, newAddress("123 Broadway", "New York", "10001")),
			true,
			isFromNewCity,
		),
		newComposeTestCase(
			"from another new city",
			newPerson("Bob", 25, newAddress("456 Main St", "New Orleans", "70112")),
			true,
			isFromNewCity,
		),
		newComposeTestCase(
			"not from new city",
			newPerson("Charlie", 40, newAddress("789 Oak Dr", "Boston", "02108")),
			false,
			isFromNewCity,
		),
		newComposeTestCase(
			"empty city",
			newPerson("David", 22, newAddress("101 Pine Ave", "", "90210")),
			false,
			isFromNewCity,
		),
	}

	core.RunTestCases(t, tests)
}

func runTestComposeComplex(t *testing.T) {
	t.Helper()
	type Score struct {
		Value int
		Max   int
	}

	// Check if a score is a passing grade (>= 70%)
	isPassingGrade := Compose(
		func(s Score) (float64, bool) {
			if s.Max == 0 {
				return 0, false
			}
			return float64(s.Value) / float64(s.Max), true
		},
		MatchGtEq(0.7),
	)

	tests := []composeTestCase[Score]{
		newComposeTestCase("perfect score", Score{100, 100}, true, isPassingGrade),
		newComposeTestCase("passing score", Score{70, 100}, true, isPassingGrade),
		newComposeTestCase("failing score", Score{69, 100}, false, isPassingGrade),
		newComposeTestCase("zero score", Score{0, 100}, false, isPassingGrade),
		newComposeTestCase("invalid score", Score{50, 0}, false, isPassingGrade), // division by zero avoided
	}

	core.RunTestCases(t, tests)
}

// TestComposePanic verifies that Compose panics when given nil arguments.
func TestComposePanic(t *testing.T) {
	t.Run("nil accessor function", runTestComposePanicNilAccessor)
	t.Run("nil matcher", runTestComposePanicNilMatcher)
}

func runTestComposePanicNilAccessor(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		Compose[int](nil, MatchGtEq(5))
	}, "nil accessor function", "nil accessor")
}

func runTestComposePanicNilMatcher(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		Compose(func(i int) (int, bool) { return i, true }, nil)
	}, "no match condition", "nil matcher")
}
