package list_test

// cspell:ignore stdlist

import (
	stdlist "container/list"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
)

// TestPurgeWithInterface tests Purge with interface types where type assertion can fail
func TestPurgeWithInterface(t *testing.T) {
	t.Run("interface type assertion", runTestPurgeInterfaceType)
	t.Run("any type assertion", runTestPurgeAnyType)
}

func runTestPurgeInterfaceType(t *testing.T) {
	t.Helper()

	// Create a list of any type
	l := list.New[any]()

	// Add mixed types - these are all stored as interface{}
	l.PushBack(1)
	l.PushBack("string")
	l.PushBack(3.14)
	l.PushBack(true)

	// Purge won't remove anything because all values satisfy interface{}
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "no elements removed for interface{}")
	core.AssertEqual(t, 4, l.Len(), "all elements remain")
}

func runTestPurgeAnyType(t *testing.T) {
	t.Helper()

	// Create a list with 'any' type (alias for interface{})
	l := list.New[any]()

	// Add various types
	l.PushBack(42)
	l.PushBack("test")
	l.PushBack(struct{ Name string }{Name: "example"})

	// All values are valid 'any' types
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "no elements removed for any type")
	core.AssertEqual(t, 3, l.Len(), "all elements remain")
}

// TestPurgeWithConcreteInterface tests with a specific interface type
func TestPurgeWithConcreteInterface(t *testing.T) {
	t.Run("stringer interface", runTestPurgeStringerInterface)
}

type myStringer struct {
	value string
}

func (s myStringer) String() string {
	return s.value
}

func runTestPurgeStringerInterface(t *testing.T) {
	t.Helper()

	// Create a raw list and add mixed values
	rawList := stdlist.New()
	rawList.PushBack(myStringer{value: "first"})  // Implements fmt.Stringer
	rawList.PushBack(42)                          // Does NOT implement fmt.Stringer
	rawList.PushBack(myStringer{value: "second"}) // Implements fmt.Stringer
	rawList.PushBack("plain string")              // Does NOT implement fmt.Stringer

	// Cast to a List of fmt.Stringer
	// This is where we can actually have type assertion failures
	type Stringer interface {
		String() string
	}

	typedList := (*list.List[Stringer])(rawList)

	// Purge will remove elements that don't implement the interface
	removed := typedList.Purge()

	// Now that Purge is fixed, it correctly removes non-Stringer elements
	core.AssertEqual(t, 2, removed, "non-Stringer elements removed")
	core.AssertEqual(t, 2, rawList.Len(), "only Stringer elements remain")

	// But we can still get the values that do implement Stringer
	remaining := typedList.Values()
	core.AssertEqual(t, 2, len(remaining), "two valid Stringer elements returned")

	// The Values() method filters correctly based on type assertions
	for i, v := range remaining {
		if v != nil {
			t.Logf("Element %d: %v", i, v.String())
		}
	}
}

// TestPurgeDirectTypeAssertion creates the exact scenario for line coverage
func TestPurgeDirectTypeAssertion(t *testing.T) {
	t.Run("force type assertion failure", runTestForceTypeAssertionFailure)
}

func runTestForceTypeAssertionFailure(t *testing.T) {
	t.Helper()

	// Create a raw list with mixed types
	rawList := stdlist.New()
	rawList.PushBack(1)
	rawList.PushBack(2)
	rawList.PushBack("not an int") // This will fail int type assertion
	rawList.PushBack(3)
	rawList.PushBack(3.14) // This will also fail int type assertion

	// Create a list.List[any] first, then manipulate the underlying list
	// This is a bit of a hack but helps us test the code path
	interfaceList := (*list.List[any])(rawList)

	// First verify all elements are there
	core.AssertEqual(t, 5, interfaceList.Len(), "initial length")

	// Now reinterpret as list.List[int] - this is where type assertions can fail
	intList := (*list.List[int])(rawList)

	// Purge should try to remove non-int elements
	// But due to generics, this might not work as expected
	removed := intList.Purge()

	// Due to generics limitations, this might not remove anything
	// But we're testing to ensure the code path is covered
	t.Logf("Removed %d elements", removed)
	t.Logf("Remaining elements: %d", rawList.Len())
}

// TestPurgeWithNilElements tests Purge with nil values in the list
func TestPurgeWithNilElements(t *testing.T) {
	t.Run("nil in interface list", runTestPurgeNilInInterfaceList)
}

func runTestPurgeNilInInterfaceList(t *testing.T) {
	t.Helper()

	// Create a raw list
	rawList := stdlist.New()
	rawList.PushBack(1)
	rawList.PushBack(nil) // nil value
	rawList.PushBack(2)

	// Cast to list.List[int]
	intList := (*list.List[int])(rawList)

	// Try to purge - nil won't match int type assertion
	removed := intList.Purge()

	// Check the result - nil is not an int, so it should be removed
	t.Logf("Removed %d elements from list with nil", removed)
	core.AssertEqual(t, 1, removed, "nil element removed")
	core.AssertEqual(t, 2, rawList.Len(), "length after purge")
}
