package list_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
)

func TestNewWithValues(t *testing.T) {
	values := core.S(1, 2, 3, 4, 5)
	l := list.New(values...)

	core.AssertSliceEqual(t, values, l.Values(), "values")
}

func testFrontBackEmpty(t *testing.T) {
	l := list.New[string]()
	_, frontOK := l.Front()
	core.AssertFalse(t, frontOK, "front present")
	_, backOK := l.Back()
	core.AssertFalse(t, backOK, "back present")
}

func testFrontBackSingle(t *testing.T) {
	l := list.New("only")
	front, ok := l.Front()
	core.AssertTrue(t, ok, "front present")
	core.AssertEqual(t, "only", front, "front")
	back, ok := l.Back()
	core.AssertTrue(t, ok, "back present")
	core.AssertEqual(t, "only", back, "back")
}

func testFrontBackMultiple(t *testing.T) {
	l := list.New("first", "middle", "last")
	front, ok := l.Front()
	core.AssertTrue(t, ok, "front present")
	core.AssertEqual(t, "first", front, "front")
	back, ok := l.Back()
	core.AssertTrue(t, ok, "back present")
	core.AssertEqual(t, "last", back, "back")
}

func TestFrontBack(t *testing.T) {
	t.Run("empty list", testFrontBackEmpty)
	t.Run("single element", testFrontBackSingle)
	t.Run("multiple elements", testFrontBackMultiple)
}

func TestPushFrontBack(t *testing.T) {
	l := list.New[int]()

	// Build: [3, 1, 2, 4]
	l.PushBack(1)
	l.PushBack(2)
	l.PushFront(3)
	l.PushBack(4)

	core.AssertSliceEqual(t, core.S(3, 1, 2, 4), l.Values(), "values")
}

func testDeleteMatchFnNone(t *testing.T) {
	l := list.New(1, 3, 5, 7)
	// Try to delete even numbers (none exist)
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	core.AssertEqual(t, 4, l.Len(), "length")
}

func testDeleteMatchFnSome(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Delete even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	core.AssertSliceEqual(t, core.S(1, 3, 5), l.Values(), "values")
}

func testDeleteMatchFnAll(t *testing.T) {
	l := list.New(2, 4, 6, 8)
	// Delete all even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	core.AssertEqual(t, 0, l.Len(), "length")
}

func TestDeleteMatchFn(t *testing.T) {
	t.Run("delete none", testDeleteMatchFnNone)
	t.Run("delete some", testDeleteMatchFnSome)
	t.Run("delete all", testDeleteMatchFnAll)
}

func testPopFirstMatchFnFound(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Pop first even number
	v, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	core.AssertTrue(t, ok, "popped")
	core.AssertEqual(t, 2, v, "popped value")
	// Verify it was removed
	core.AssertSliceEqual(t, core.S(1, 3, 4, 5), l.Values(), "remaining")
}

func testPopFirstMatchFnNotFound(t *testing.T) {
	l := list.New(1, 3, 5)
	// Try to pop even number (none exist)
	_, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	core.AssertFalse(t, ok, "not found")
	// List should be unchanged
	core.AssertEqual(t, 3, l.Len(), "length")
}

func TestPopFirstMatchFn(t *testing.T) {
	t.Run("found", testPopFirstMatchFnFound)
	t.Run("not found", testPopFirstMatchFnNotFound)
}

func TestMoveToBackFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first even number to back
	l.MoveToBackFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})

	core.AssertSliceEqual(t, core.S(1, 3, 4, 5, 2), l.Values(), "values")
}

func TestMoveToFrontFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first number > 3 to front
	l.MoveToFrontFirstMatchFn(func(n int) bool {
		return n > 3
	})

	core.AssertSliceEqual(t, core.S(4, 1, 2, 3, 5), l.Values(), "values")
}

func testFirstMatchFnFound(t *testing.T) {
	l := list.New("apple", "banana", "cherry", "date")
	// Find first string with length > 5
	v, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	core.AssertTrue(t, ok, "found")
	core.AssertEqual(t, "banana", v, "match")
}

func testFirstMatchFnNotFound(t *testing.T) {
	l := list.New("a", "b", "c")
	// Find string with length > 5
	_, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	core.AssertFalse(t, ok, "not found")
}

func TestFirstMatchFn(t *testing.T) {
	t.Run("found", testFirstMatchFnFound)
	t.Run("not found", testFirstMatchFnNotFound)
}

func TestZero(t *testing.T) {
	l := list.New[int]()
	core.AssertEqual(t, 0, l.Zero(), "int zero")

	ls := list.New[string]()
	core.AssertEqual(t, "", ls.Zero(), "string zero")

	type custom struct {
		b string
		a int
	}
	lc := list.New[custom]()
	z := lc.Zero()
	core.AssertEqual(t, 0, z.a, "custom zero a")
	core.AssertEqual(t, "", z.b, "custom zero b")
}

func TestClone(t *testing.T) {
	original := list.New(1, 2, 3)
	cloned := original.Clone()

	// Verify they have same values
	core.AssertSliceEqual(t, original.Values(), cloned.Values(), "clone values")

	// Verify they are independent
	cloned.PushBack(4)
	core.AssertNotEqual(t, original.Len(), cloned.Len(), "independent length")
}

func testCopyFilterTransform(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5, 6)
	// Copy only even numbers, tripled
	copied := l.Copy(func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 3, true
		}
		return 0, false
	})

	core.AssertSliceEqual(t, core.S(6, 12, 18), copied.Values(), "values")
}

func testCopyAll(t *testing.T) {
	l := list.New("a", "b", "c")
	copied := l.Copy(func(v string) (string, bool) {
		return v + v, true // double each string
	})

	core.AssertSliceEqual(t, core.S("aa", "bb", "cc"), copied.Values(), "values")
}

func TestCopy(t *testing.T) {
	t.Run("filter and transform", testCopyFilterTransform)
	t.Run("copy all", testCopyAll)
}

func TestPurge(t *testing.T) {
	// Purge is specifically for removing elements that don't match the type
	// This is more relevant when dealing with interface{} conversions
	// For a generic List[T], all elements should already be of type T
	l := list.New[int]()

	// Add some values
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Since all elements are already int, purge should remove nothing
	core.AssertEqual(t, 0, l.Purge(), "removed")
	core.AssertEqual(t, 3, l.Len(), "length")
}

func TestNilList(t *testing.T) {
	var l *list.List[int]

	// Test nil-safe methods
	core.AssertEqual(t, 0, l.Len(), "nil length")
	core.AssertNil(t, l.Sys(), "nil Sys")

	_, frontOK := l.Front()
	core.AssertFalse(t, frontOK, "nil Front")

	_, backOK := l.Back()
	core.AssertFalse(t, backOK, "nil Back")

	core.AssertEqual(t, 0, len(l.Values()), "nil Values")

	// These should not panic
	core.AssertNoPanic(t, func() {
		l.PushFront(1)
		l.PushBack(2)
		l.ForEach(func(int) bool { return true })
		l.DeleteMatchFn(func(int) bool { return true })
	}, "nil mutators")
}
