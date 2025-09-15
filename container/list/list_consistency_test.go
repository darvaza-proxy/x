package list_test

import (
	"container/list"
	"testing"

	"darvaza.org/core"
	xlist "darvaza.org/x/container/list"
)

// TestList_MixedTypeHandling verifies lists skip wrong-type items when present
func TestList_MixedTypeHandling(t *testing.T) {
	// Create a raw list with mixed types
	rawList := list.New()
	rawList.PushBack("string1")
	rawList.PushBack(42) // wrong type
	rawList.PushBack("string2")
	rawList.PushBack(3.14) // wrong type
	rawList.PushBack("string3")
	rawList.PushBack(true) // wrong type
	rawList.PushBack("string4")

	// Cast to typed List[string]
	typedList := (*xlist.List[string])(rawList)

	// Test ForEach - should only see correct types
	var collected []string
	typedList.ForEach(func(s string) bool {
		collected = append(collected, s)
		return true // continue iteration
	})
	core.AssertEqual(t, 4, len(collected), "collected string count")
	core.AssertSliceEqual(t, core.S("string1", "string2", "string3", "string4"), collected, "collected strings")

	// Test Values - should only return correct types
	values := typedList.Values()
	core.AssertEqual(t, 4, len(values), "values count")
	core.AssertSliceEqual(t, core.S("string1", "string2", "string3", "string4"), values, "values")

	// Test Len - should count all elements (including wrong types)
	core.AssertEqual(t, 7, typedList.Len(), "total element count")

	// Test Front/Back methods
	frontVal, hasFirst := typedList.Front()
	core.AssertTrue(t, hasFirst, "has first element")
	core.AssertEqual(t, "string1", frontVal, "first value")

	// Test Back method
	backVal, hasLast := typedList.Back()
	core.AssertTrue(t, hasLast, "has last element")
	core.AssertEqual(t, "string4", backVal, "last value")

	// Test Purge removes exactly the wrong-type count
	removed := typedList.Purge()
	core.AssertEqual(t, 3, removed, "purged count") // 42, 3.14, true

	// After purge, Len should match typed element count
	core.AssertEqual(t, 4, typedList.Len(), "len after purge")

	// Verify all remaining elements are correct type
	values = typedList.Values()
	core.AssertSliceEqual(t, core.S("string1", "string2", "string3", "string4"), values, "values after purge")
}

// TestList_EmptyWithWrongTypes verifies behavior when list has only wrong-type elements
func TestList_EmptyWithWrongTypes(t *testing.T) {
	// Create raw list with only wrong types
	rawList := list.New()
	rawList.PushBack(42)
	rawList.PushBack(3.14)
	rawList.PushBack(true)

	typedList := (*xlist.List[string])(rawList)

	// Front and Back should return false (no matching types)
	_, hasFirst := typedList.Front()
	core.AssertFalse(t, hasFirst, "no first on wrong-type list")
	_, hasLast := typedList.Back()
	core.AssertFalse(t, hasLast, "no last on wrong-type list")

	// Values should be empty
	values := typedList.Values()
	core.AssertEqual(t, 0, len(values), "values from wrong-type list")

	// ForEach should not iterate
	count := 0
	typedList.ForEach(func(_ string) bool {
		count++
		return true
	})
	core.AssertEqual(t, 0, count, "ForEach iteration count")

	// Len still counts all elements
	core.AssertEqual(t, 3, typedList.Len(), "len with wrong types")

	// Purge should remove all
	removed := typedList.Purge()
	core.AssertEqual(t, 3, removed, "purged all wrong types")
	core.AssertEqual(t, 0, typedList.Len(), "len after purge all")
}

// TestList_BasicOperations verifies basic list operations maintain consistency
func TestList_BasicOperations(t *testing.T) {
	l := xlist.New[int](1, 2, 3, 4, 5)

	// Test Values() returns correct order
	values := l.Values()
	core.AssertSliceEqual(t, core.S(1, 2, 3, 4, 5), values, "initial values")

	// Test PushFront and PushBack
	l.PushFront(0)
	l.PushBack(6)
	values = l.Values()
	core.AssertSliceEqual(t, core.S(0, 1, 2, 3, 4, 5, 6), values, "after push operations")

	// Test Front and Back directly (no Pop methods available)
	frontVal, ok := l.Front()
	core.AssertTrue(t, ok, "has front")
	core.AssertEqual(t, 0, frontVal, "front value")

	backVal, ok := l.Back()
	core.AssertTrue(t, ok, "has back")
	core.AssertEqual(t, 6, backVal, "back value")

	// Verify values are in correct order
	values = l.Values()
	core.AssertSliceEqual(t, core.S(0, 1, 2, 3, 4, 5, 6), values, "final values")
}

// TestList_RemovalConsistency verifies removal operations maintain consistency
func TestList_RemovalConsistency(t *testing.T) {
	l := xlist.New[string]("a", "b", "c", "d", "e")

	// Test PopFirstMatchFn to remove specific elements
	val, ok := l.PopFirstMatchFn(func(s string) bool { return s == "a" })
	core.AssertTrue(t, ok, "PopFirstMatchFn found 'a'")
	core.AssertEqual(t, "a", val, "popped value")
	core.AssertSliceEqual(t, core.S("b", "c", "d", "e"), l.Values(), "after removing 'a'")

	// Remove last element
	val, ok = l.PopFirstMatchFn(func(s string) bool { return s == "e" })
	core.AssertTrue(t, ok, "PopFirstMatchFn found 'e'")
	core.AssertEqual(t, "e", val, "popped value")
	core.AssertSliceEqual(t, core.S("b", "c", "d"), l.Values(), "after removing 'e'")

	// Remove all remaining elements
	l.DeleteMatchFn(func(_ string) bool { return true })

	// List should be empty now
	core.AssertEqual(t, 0, l.Len(), "empty after delete all")
	_, hasFirst := l.Front()
	core.AssertFalse(t, hasFirst, "no front after empty")
	_, hasLast := l.Back()
	core.AssertFalse(t, hasLast, "no back after empty")

	// Try to pop from empty list
	_, ok = l.PopFirstMatchFn(func(_ string) bool { return true })
	core.AssertFalse(t, ok, "PopFirstMatchFn from empty")
}

// TestList_CloneIndependence verifies cloned lists are independent
func TestList_CloneIndependence(t *testing.T) {
	original := xlist.New[int](1, 2, 3, 4, 5)

	// Clone the list
	clone := original.Clone()

	// Verify initial equality
	core.AssertSliceEqual(t, original.Values(), clone.Values(), "clone matches original")

	// Modify clone
	clone.PushBack(6)
	clone.PushFront(0)

	// Original should be unchanged
	core.AssertSliceEqual(t, core.S(1, 2, 3, 4, 5), original.Values(), "original unchanged")
	core.AssertSliceEqual(t, core.S(0, 1, 2, 3, 4, 5, 6), clone.Values(), "clone modified")

	// Modify original by removing first and last
	original.PopFirstMatchFn(func(i int) bool { return i == 1 })
	original.PopFirstMatchFn(func(i int) bool { return i == 5 })

	// Clone should be unaffected
	core.AssertSliceEqual(t, core.S(2, 3, 4), original.Values(), "original modified")
	core.AssertSliceEqual(t, core.S(0, 1, 2, 3, 4, 5, 6), clone.Values(), "clone unchanged")

	// Test that modifications to one don't affect the other
	original.PopFirstMatchFn(func(i int) bool { return i == 2 })
	original.PopFirstMatchFn(func(i int) bool { return i == 4 })
	// Clone should still have all its elements
	core.AssertSliceEqual(t, core.S(0, 1, 2, 3, 4, 5, 6), clone.Values(), "clone unaffected by original removals")

	clone.PopFirstMatchFn(func(i int) bool { return i == 0 })
	clone.PopFirstMatchFn(func(i int) bool { return i == 6 })
	// Original should maintain its state
	core.AssertSliceEqual(t, core.S(3), original.Values(), "original unaffected by clone removals")
}
