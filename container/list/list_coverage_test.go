package list_test

// cspell:ignore stdlist

import (
	stdlist "container/list"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
)

// Test Purge function with various scenarios to achieve 100% coverage
func TestListPurgeComprehensive(t *testing.T) {
	t.Run("normal list", runTestListPurgeNormalList)
	t.Run("empty list", runTestListPurgeEmptyList)
	t.Run("nil list", runTestListPurgeNilList)

	// Test with mixed types to ensure Purge removes wrong types
	t.Run("mixed types", runTestListPurgeMixedTypes)

	// Additional test to force coverage of the Purge nil path
	t.Run("purge with nil underlying list", runTestListPurgeNilUnderlying)

	// Test to hit line 291 closing brace - Purge with actual removals
	t.Run("purge with actual removals hits closing brace", runTestListPurgeActualRemovals)
}

// Test Clone with nil list
func TestListCloneNilReceiver(t *testing.T) {
	var l *list.List[int]
	cloned := l.Clone()
	core.AssertNotNil(t, cloned, "clone of nil returns new list")
	core.AssertEqual(t, 0, cloned.Len(), "cloned list is empty")
}

// Test Copy with nil function to trigger default function assignment
func TestListCopyNilFunction(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)

	// Test Copy with nil function - should use default identity function
	copied := l.Copy(nil)
	core.AssertNotNil(t, copied, "copy result")
	core.AssertEqual(t, 5, copied.Len(), "all elements copied")
	core.AssertSliceEqual(t, []int{1, 2, 3, 4, 5}, copied.Values(), "values match")

	// Verify it's a new list, not the same
	core.AssertNotSame(t, l, copied, "different list instances")
}

// Test Copy on nil list
func TestListCopyNilList(t *testing.T) {
	var l *list.List[int]
	copied := l.Copy(nil)
	core.AssertNotNil(t, copied, "copy of nil list returns new empty list")
	core.AssertEqual(t, 0, copied.Len(), "empty list")
}

// Test Copy with filter function that rejects some elements
func TestListCopyWithFilter(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Only copy even numbers
	copied := l.Copy(func(v int) (int, bool) {
		return v, v%2 == 0
	})
	core.AssertNotNil(t, copied, "copied list should not be nil")
	core.AssertEqual(t, 2, copied.Len(), "copied list should have 2 even numbers")
}

// Test Copy to hit line 318 return true
func TestListCopyReturnTrue(t *testing.T) {
	l := list.New(10, 20, 30)
	copied := l.Copy(func(v int) (int, bool) {
		// Transform and include all
		return v * 2, true
	})
	core.AssertEqual(t, 3, copied.Len(), "all elements copied")
	values := copied.Values()
	expected := []int{20, 40, 60}
	core.AssertEqual(t, len(expected), len(values), "same length")
	for i, v := range expected {
		core.AssertEqual(t, v, values[i], "value at index %d", i)
	}
}

// Test DeleteMatchFn with nil for complete coverage
func TestDeleteMatchFnNilCoverage(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	l.DeleteMatchFn(nil) // nil function does nothing
	core.AssertEqual(t, 5, l.Len(), "nothing deleted with nil function")
}

// Test MoveToBack/MoveToFront with nil match function
func TestMoveOperationsNilMatchCoverage(t *testing.T) {
	t.Run("MoveToBack nil", runTestMoveOperationsNilMatchCoverageMoveToBackNil)
	t.Run("MoveToFront nil", runTestMoveOperationsNilMatchCoverageMoveToFrontNil)
}

// Test PopFirstMatchFn with nil match
func TestPopFirstMatchFnNilCoverage(t *testing.T) {
	l := list.New(1, 2, 3)
	v, ok := l.PopFirstMatchFn(nil) // nil function returns false
	core.AssertFalse(t, ok, "not found with nil function")
	core.AssertEqual(t, 0, v, "zero value")
	core.AssertSliceEqual(t, []int{1, 2, 3}, l.Values(), "unchanged with nil function")
}

// Test FirstMatchFn with nil match
func TestFirstMatchFnNilCoverage(t *testing.T) {
	l := list.New("first", "second", "third")
	v, ok := l.FirstMatchFn(nil) // nil function returns false
	core.AssertFalse(t, ok, "not found with nil function")
	core.AssertEqual(t, "", v, "zero value")
	core.AssertEqual(t, 3, l.Len(), "length unchanged")
}

// Test Copy behaviour to understand ForEach return value semantics
func TestListCopyForEachBehaviour(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)

	// Test that Copy processes all elements correctly
	var processed []int
	copied := l.Copy(func(v int) (int, bool) {
		processed = append(processed, v)
		return v * 2, true // transform and include
	})

	core.AssertEqual(t, 5, len(processed), "all elements processed")
	core.AssertEqual(t, 5, copied.Len(), "all elements copied")
	core.AssertSliceEqual(t, []int{1, 2, 3, 4, 5}, processed, "processed in order")
	core.AssertSliceEqual(t, []int{2, 4, 6, 8, 10}, copied.Values(), "transformed values")
}

func runTestListPurgeMixedTypes(t *testing.T) {
	t.Helper()
	rawList := stdlist.New()
	rawList.PushBack(1)
	rawList.PushBack("not int")
	rawList.PushBack(2)
	rawList.PushBack(3.14)

	typedList := (*list.List[int])(rawList)
	removed := typedList.Purge()
	// Purge correctly removes non-int elements
	core.AssertEqual(t, 2, removed, "removed non-int elements")
	core.AssertEqual(t, 2, rawList.Len(), "only ints remain")
}

func runTestListPurgeNormalList(t *testing.T) {
	t.Helper()
	// Test with a properly typed list
	l := list.New(1, 2, 3)
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "no elements removed from properly typed list")
	core.AssertEqual(t, 3, l.Len(), "length unchanged")
}

func runTestListPurgeEmptyList(t *testing.T) {
	t.Helper()
	l := list.New[int]()
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "no elements removed from empty list")
}

func runTestListPurgeNilList(t *testing.T) {
	t.Helper()
	var l *list.List[int]
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "nil list returns 0")
}

func runTestListPurgeNilUnderlying(t *testing.T) {
	t.Helper()
	// Create a List with nil underlying list
	var l list.List[int]
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "purge on uninitialized list")
}

func runTestListPurgeActualRemovals(t *testing.T) {
	t.Helper()
	rawList := stdlist.New()
	rawList.PushBack(1)
	rawList.PushBack("string")
	rawList.PushBack(3)

	typedList := (*list.List[int])(rawList)
	removed := typedList.Purge()
	core.AssertEqual(t, 1, removed, "one non-int removed")
}

// Test Purge with list that has all valid elements (empty remove slice path)
func TestListPurgeNoRemovals(t *testing.T) {
	// Create a properly typed list - all elements are valid
	l := list.New(10, 20, 30, 40, 50)

	// Purge should find no elements to remove
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "no elements removed")
	core.AssertEqual(t, 5, l.Len(), "length unchanged")
	core.AssertSliceEqual(t, []int{10, 20, 30, 40, 50}, l.Values(), "values unchanged")
}

func runTestMoveOperationsNilMatchCoverageMoveToBackNil(t *testing.T) {
	t.Helper()
	l := list.New(1, 2, 3, 4, 5)
	l.MoveToBackFirstMatchFn(nil) // nil function does nothing
	core.AssertSliceEqual(t, []int{1, 2, 3, 4, 5}, l.Values(), "unchanged with nil function")
}

func runTestMoveOperationsNilMatchCoverageMoveToFrontNil(t *testing.T) {
	t.Helper()
	l := list.New(1, 2, 3, 4, 5)
	l.MoveToFrontFirstMatchFn(nil) // nil function does nothing
	core.AssertSliceEqual(t, []int{1, 2, 3, 4, 5}, l.Values(), "unchanged with nil function")
}
