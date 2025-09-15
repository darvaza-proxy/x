package list_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/list"
)

func TestNewWithValues(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}
	l := list.New(values...)

	core.AssertEqual(t, len(values), l.Len(), "list length")

	result := l.Values()
	for i, v := range result {
		core.AssertEqual(t, values[i], v, "value at index %d", i)
	}
}

func runTestFrontBackEmpty(t *testing.T) {
	t.Helper()
	l := list.New[string]()
	v, ok := l.Front()
	core.AssertEqual(t, false, ok, "front exists")
	core.AssertEqual(t, "", v, "front value")
	v, ok = l.Back()
	core.AssertEqual(t, false, ok, "back exists")
	core.AssertEqual(t, "", v, "back value")
}

func runTestFrontBackSingle(t *testing.T) {
	t.Helper()
	l := list.New("only")
	front, ok := l.Front()
	core.AssertEqual(t, true, ok, "front should exist")
	core.AssertEqual(t, "only", front, "front value")
	back, ok := l.Back()
	core.AssertEqual(t, true, ok, "back should exist")
	core.AssertEqual(t, "only", back, "back value")
}

func runTestFrontBackMultiple(t *testing.T) {
	t.Helper()
	l := list.New("first", "middle", "last")
	front, ok := l.Front()
	core.AssertEqual(t, true, ok, "front should exist")
	core.AssertEqual(t, "first", front, "front value")
	back, ok := l.Back()
	core.AssertEqual(t, true, ok, "back should exist")
	core.AssertEqual(t, "last", back, "back value")
}

func TestFrontBack(t *testing.T) {
	t.Run("empty list", runTestFrontBackEmpty)
	t.Run("single element", runTestFrontBackSingle)
	t.Run("multiple elements", runTestFrontBackMultiple)
}

func TestPushFrontBack(t *testing.T) {
	l := list.New[int]()

	// Build: [3, 1, 2, 4]
	l.PushBack(1)
	l.PushBack(2)
	l.PushFront(3)
	l.PushBack(4)

	expected := []int{3, 1, 2, 4}
	values := l.Values()

	core.AssertMustEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func runTestDeleteMatchFnNone(t *testing.T) {
	t.Helper()
	l := list.New(1, 3, 5, 7)
	// Try to delete even numbers (none exist)
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	core.AssertEqual(t, 4, l.Len(), "list length")
}

func runTestDeleteMatchFnSome(t *testing.T) {
	t.Helper()
	l := list.New(1, 2, 3, 4, 5)
	// Delete even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	expected := []int{1, 3, 5}
	values := l.Values()
	core.AssertMustEqual(t, len(expected), len(values), "values count")
	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func runTestDeleteMatchFnAll(t *testing.T) {
	t.Helper()
	l := list.New(2, 4, 6, 8)
	// Delete all even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	core.AssertEqual(t, 0, l.Len(), "list length")
}

func TestDeleteMatchFn(t *testing.T) {
	t.Run("delete none", runTestDeleteMatchFnNone)
	t.Run("delete some", runTestDeleteMatchFnSome)
	t.Run("delete all", runTestDeleteMatchFnAll)
}

func runTestPopFirstMatchFnFound(t *testing.T) {
	t.Helper()
	l := list.New(1, 2, 3, 4, 5)
	// Pop first even number
	v, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	core.AssertEqual(t, true, ok, "should find match")
	core.AssertEqual(t, 2, v, "popped value")
	// Verify it was removed
	values := l.Values()
	expected := []int{1, 3, 4, 5}
	core.AssertMustEqual(t, len(expected), len(values), "values count")
	for i, val := range values {
		core.AssertEqual(t, expected[i], val, "value at index %d", i)
	}
}

func runTestPopFirstMatchFnNotFound(t *testing.T) {
	t.Helper()
	l := list.New(1, 3, 5)
	// Try to pop even number (none exist)
	v, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	core.AssertEqual(t, false, ok, "should not find match")
	core.AssertEqual(t, 0, v, "zero value when not found")
	// List should be unchanged
	core.AssertEqual(t, 3, l.Len(), "list length")
}

func TestPopFirstMatchFn(t *testing.T) {
	t.Run("found", runTestPopFirstMatchFnFound)
	t.Run("not found", runTestPopFirstMatchFnNotFound)
}

func TestMoveToBackFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first even number to back
	l.MoveToBackFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})

	values := l.Values()
	expected := []int{1, 3, 4, 5, 2}

	core.AssertMustEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func TestMoveToFrontFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first number > 3 to front
	l.MoveToFrontFirstMatchFn(func(n int) bool {
		return n > 3
	})

	values := l.Values()
	expected := []int{4, 1, 2, 3, 5}

	core.AssertMustEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func runTestFirstMatchFnFound(t *testing.T) {
	t.Helper()
	l := list.New("apple", "banana", "cherry", "date")
	// Find first string with length > 5
	v, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	core.AssertEqual(t, true, ok, "should find match")
	core.AssertEqual(t, "banana", v, "found value")
}

func runTestFirstMatchFnNotFound(t *testing.T) {
	t.Helper()
	l := list.New("a", "b", "c")
	// Find string with length > 5
	v, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	core.AssertEqual(t, false, ok, "should not find match")
	core.AssertEqual(t, "", v, "zero value when not found")
}

func TestFirstMatchFn(t *testing.T) {
	t.Run("found", runTestFirstMatchFnFound)
	t.Run("not found", runTestFirstMatchFnNotFound)
}

func TestZero(t *testing.T) {
	l := list.New[int]()
	core.AssertEqual(t, 0, l.Zero(), "zero value")

	ls := list.New[string]()
	core.AssertEqual(t, "", ls.Zero(), "zero value")

	type custom struct {
		b string
		a int
	}
	lc := list.New[custom]()
	z := lc.Zero()
	core.AssertEqual(t, 0, z.a, "zero value field a")
	core.AssertEqual(t, "", z.b, "zero value field b")
}

func TestClone(t *testing.T) {
	original := list.New(1, 2, 3)
	cloned := original.Clone()

	// Verify they have same values
	core.AssertEqual(t, original.Len(), cloned.Len(), "clone length")

	origValues := original.Values()
	cloneValues := cloned.Values()
	for i := range origValues {
		core.AssertEqual(t, origValues[i], cloneValues[i], "value at index %d", i)
	}

	// Verify they are independent
	cloned.PushBack(4)
	core.AssertNotEqual(t, original.Len(), cloned.Len(), "clone independence")
}

func runTestCopyFilterTransform(t *testing.T) {
	t.Helper()
	l := list.New(1, 2, 3, 4, 5, 6)
	// Copy only even numbers, tripled
	copied := l.Copy(func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 3, true
		}
		return 0, false
	})

	expected := []int{6, 12, 18}
	values := copied.Values()

	core.AssertMustEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func runTestCopyAll(t *testing.T) {
	t.Helper()
	l := list.New("a", "b", "c")
	copied := l.Copy(func(v string) (string, bool) {
		return v + v, true // double each string
	})

	expected := []string{"aa", "bb", "cc"}
	values := copied.Values()

	core.AssertMustEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func TestCopy(t *testing.T) {
	t.Run("filter and transform", runTestCopyFilterTransform)
	t.Run("copy all", runTestCopyAll)
}

func TestPurge(t *testing.T) {
	// Purge is specifically for removing elements that don't match the type
	// This is more relevant when dealing with interface{} conversions
	// For a generic list.List[T], all elements should already be of type T
	l := list.New[int]()

	// Add some values
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Since all elements are already int, purge should remove nothing
	removed := l.Purge()
	core.AssertEqual(t, 0, removed, "removed count")

	core.AssertEqual(t, 3, l.Len(), "list length")
}

func TestNilList(t *testing.T) {
	var l *list.List[int]

	// Test nil-safe methods
	core.AssertEqual(t, 0, l.Len(), "nil list length")

	core.AssertNil(t, l.Sys(), "nil list sys")

	_, ok := l.Front()
	core.AssertEqual(t, false, ok, "nil list front")

	_, ok = l.Back()
	core.AssertEqual(t, false, ok, "nil list back")

	values := l.Values()
	core.AssertEqual(t, 0, len(values), "nil list values")

	// These should not panic
	l.PushFront(1)
	l.PushBack(2)
	l.ForEach(func(int) bool { return true })
	l.DeleteMatchFn(func(int) bool { return true })
}
