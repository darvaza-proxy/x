package list

import (
	"container/list"
	"testing"

	"darvaza.org/core"
)

// Tests that access unexported methods and helper functions

// Test helper functions for complete coverage
func TestListHelpersCoverage(t *testing.T) {
	// Test newGetAllMatchElement with nil out parameter
	t.Run("newGetAllMatchElement nil out", runTestNewGetAllMatchElementNilOut)

	// Test newGetMatchElement with nil parameters - BOTH nil
	t.Run("newGetMatchElement nil params", runTestNewGetMatchElementNilParams)

	// Test newGetMatchElement with nil out parameter only - hits line 386
	t.Run("newGetMatchElement nil out hits line 386", runTestNewGetMatchElementNilOutHitsLine386)

	// Test newMatchElement with nil function
	t.Run("newMatchElement nil fn", runTestNewMatchElementNilFn)

	// Test newMatchElement with type mismatch to cover the else branch
	t.Run("newMatchElement type mismatch", runTestNewMatchElementTypeMismatch)

	t.Run("newMatchElement nil function", runTestNewMatchElementNilCoverage)
	t.Run("backward iteration", runTestBackwardIterationCoverage)
	t.Run("get all match element", runTestGetAllMatchElementCoverage)
	t.Run("newMatchElement type assertion", runTestNewMatchElementTypeAssertion)
}

func runTestNewGetAllMatchElementNilOut(t *testing.T) {
	t.Helper()
	// This will cover the nil check branch in newGetAllMatchElement
	out, cb := newGetAllMatchElement[int](nil, nil)
	core.AssertNotNil(t, out, "out pointer should be allocated")
	core.AssertNotNil(t, cb, "callback should be returned")
}

func runTestNewGetMatchElementNilParams(t *testing.T) {
	t.Helper()
	// This will cover the nil check branches in newGetMatchElement
	match, out, cb := newGetMatchElement[int](nil, nil, nil)
	core.AssertNotNil(t, match, "match pointer should be allocated")
	core.AssertNotNil(t, out, "out pointer should be allocated")
	core.AssertNotNil(t, cb, "callback should be returned")

	// Actually execute the callback to hit the code path
	rawList := list.New()
	el := rawList.PushBack(42)
	result := cb(el, 42)
	core.AssertEqual(t, false, result, "should continue iteration")
}

func runTestNewGetMatchElementNilOutHitsLine386(t *testing.T) {
	t.Helper()
	var match *list.Element
	// This should hit the "out == nil" branch specifically (line 386)
	_, out, cb := newGetMatchElement[int](nil, &match, nil)
	core.AssertNotNil(t, out, "out pointer should be allocated")
	core.AssertNotNil(t, cb, "callback should be returned")

	// Execute callback to ensure line 386-388 path is taken
	rawList := list.New()
	el := rawList.PushBack(99)
	result := cb(el, 99)
	core.AssertFalse(t, result, "callback should stop on match")
	core.AssertEqual(t, 99, *out, "value stored in allocated out pointer")
}

func runTestNewMatchElementNilFn(t *testing.T) {
	t.Helper()
	cb := newMatchElement[int](nil)
	core.AssertNotNil(t, cb, "callback should be returned")

	// Create a dummy element to test the callback
	rawList := list.New()
	el := rawList.PushBack(42)
	result := cb(el)
	core.AssertEqual(t, false, result, "should return false for continue iteration")
}

func runTestNewMatchElementTypeMismatch(t *testing.T) {
	t.Helper()
	// Create a raw list with mixed types
	rawList := list.New()
	rawList.PushBack(1)
	rawList.PushBack("string") // This won't match T=int

	// Create callback for int type
	cb := newMatchElement[int](func(_ *list.Element, _ int) bool {
		return true // continue
	})

	// Test the string element - should not match
	stringEl := rawList.Back()
	result := cb(stringEl)
	core.AssertEqual(t, false, result, "type mismatch returns false")
}

func runTestNewMatchElementNilCoverage(t *testing.T) {
	t.Helper()
	l := New(1, 2, 3)

	// Test with nil match function in helper
	elems := l.unsafeGetAllMatchElement(nil)
	core.AssertEqual(t, 3, len(elems), "nil matches all")

	// Test FirstMatchFn with backward iteration and nil
	elem, val, found := l.unsafeFirstMatchBackwardElement(nil)
	core.AssertTrue(t, found, "found with nil match")
	core.AssertNotNil(t, elem, "element not nil")
	core.AssertEqual(t, 3, val, "last value with backward iteration")
}

func runTestBackwardIterationCoverage(t *testing.T) {
	t.Helper()
	l := New(1, 2, 3, 4, 5)

	// Find first even number searching backward
	elem, val, found := l.unsafeFirstMatchBackwardElement(func(v int) bool {
		return v%2 == 0
	})
	core.AssertTrue(t, found, "found")
	core.AssertNotNil(t, elem, "element")
	core.AssertEqual(t, 4, val, "last even number")

	// Test not found case
	elem2, val2, found2 := l.unsafeFirstMatchBackwardElement(func(v int) bool {
		return v > 10
	})
	core.AssertFalse(t, found2, "not found")
	core.AssertNil(t, elem2, "no element")
	core.AssertEqual(t, 0, val2, "zero value")
}

func runTestGetAllMatchElementCoverage(t *testing.T) {
	t.Helper()
	l := New(1, 2, 3, 4, 5, 6)

	// Get all even numbers
	elems := l.unsafeGetAllMatchElement(func(v int) bool {
		return v%2 == 0
	})
	core.AssertEqual(t, 3, len(elems), "even count")

	// Get all with nil function
	elems = l.unsafeGetAllMatchElement(nil)
	core.AssertEqual(t, 6, len(elems), "all elements")

	// Empty list
	empty := New[int]()
	elems = empty.unsafeGetAllMatchElement(func(_ int) bool { return true })
	core.AssertEqual(t, 0, len(elems), "empty list")
}

func runTestNewMatchElementTypeAssertion(t *testing.T) {
	t.Helper()

	// Test newMatchElement with type assertion failure
	fn := func(_ *list.Element, _ int) bool {
		// This should never be called for wrong types
		t.Fatal("should not be called for wrong type")
		return false
	}

	cb := newMatchElement(fn)

	// Create element with wrong type
	rawList := list.New()
	stringEl := rawList.PushBack("not an int")

	// This should hit the type assertion failure path
	result := cb(stringEl)
	core.AssertFalse(t, result, "type mismatch returns false")

	// Test with nil function too
	cb2 := newMatchElement[int](nil)
	result2 := cb2(stringEl)
	core.AssertFalse(t, result2, "nil function with wrong type returns false")
}

// Test unsafe operations for coverage
func TestListUnsafeOperationsCoverage(t *testing.T) {
	t.Run("unsafeForEachElement", runTestUnsafeForEachElementCoverage)
	t.Run("unsafeForEachBackwardElement", runTestUnsafeForEachBackwardElementCoverage)
}

func runTestUnsafeForEachElementCoverage(t *testing.T) {
	t.Helper()
	l := New(1, 2, 3, 4, 5)

	// Test normal iteration
	var collected []int
	l.unsafeForEachElement(func(_ *list.Element, v int) bool {
		collected = append(collected, v)
		return true // continue
	})
	core.AssertSliceEqual(t, []int{1, 2, 3, 4, 5}, collected, "forward order")

	// Test early termination
	collected = nil
	l.unsafeForEachElement(func(_ *list.Element, v int) bool {
		collected = append(collected, v)
		return v < 3 // stop after 3
	})
	core.AssertSliceEqual(t, []int{1, 2, 3}, collected, "early termination")

	// Test with nil function
	l.unsafeForEachElement(nil)
	core.AssertEqual(t, 5, l.Len(), "unchanged after nil function")
}

func runTestUnsafeForEachBackwardElementCoverage(t *testing.T) {
	t.Helper()
	l := New(1, 2, 3, 4, 5)

	// Test normal backward iteration
	var collected []int
	l.unsafeForEachBackwardElement(func(_ *list.Element, v int) bool {
		collected = append(collected, v)
		return true // continue
	})
	core.AssertSliceEqual(t, []int{5, 4, 3, 2, 1}, collected, "backward order")

	// Test early termination
	collected = nil
	l.unsafeForEachBackwardElement(func(_ *list.Element, v int) bool {
		collected = append(collected, v)
		return v > 3 // stop after going below 3
	})
	core.AssertSliceEqual(t, []int{5, 4, 3}, collected, "early termination")

	// Test with nil function
	l.unsafeForEachBackwardElement(nil)
	core.AssertEqual(t, 5, l.Len(), "unchanged after nil function")
}

// Test unsafeFirstMatchElement with no matching elements
func TestUnsafeFirstMatchElementNoMatch(t *testing.T) {
	l := New(1, 3, 5, 7, 9)

	// Look for an even number (none exist)
	elem, val, found := l.unsafeFirstMatchElement(func(v int) bool {
		return v%2 == 0
	})

	core.AssertFalse(t, found, "no match found")
	core.AssertNil(t, elem, "no element returned")
	core.AssertEqual(t, 0, val, "zero value returned")
}
