package list

import (
	"testing"
)

func TestNewWithValues(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}
	l := New(values...)

	if l.Len() != len(values) {
		t.Errorf("Expected length %d, got %d", len(values), l.Len())
	}

	result := l.Values()
	for i, v := range result {
		if v != values[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, values[i])
		}
	}
}

func testFrontBackEmpty(t *testing.T) {
	l := New[string]()
	if v, ok := l.Front(); ok {
		t.Errorf("Expected no front value, got %v", v)
	}
	if v, ok := l.Back(); ok {
		t.Errorf("Expected no back value, got %v", v)
	}
}

func testFrontBackSingle(t *testing.T) {
	l := New("only")
	front, ok := l.Front()
	if !ok || front != "only" {
		t.Errorf("Front: got (%v, %v), expected (only, true)", front, ok)
	}
	back, ok := l.Back()
	if !ok || back != "only" {
		t.Errorf("Back: got (%v, %v), expected (only, true)", back, ok)
	}
}

func testFrontBackMultiple(t *testing.T) {
	l := New("first", "middle", "last")
	front, ok := l.Front()
	if !ok || front != "first" {
		t.Errorf("Front: got (%v, %v), expected (first, true)", front, ok)
	}
	back, ok := l.Back()
	if !ok || back != "last" {
		t.Errorf("Back: got (%v, %v), expected (last, true)", back, ok)
	}
}

func TestFrontBack(t *testing.T) {
	t.Run("empty list", testFrontBackEmpty)
	t.Run("single element", testFrontBackSingle)
	t.Run("multiple elements", testFrontBackMultiple)
}

func TestPushFrontBack(t *testing.T) {
	l := New[int]()

	// Build: [3, 1, 2, 4]
	l.PushBack(1)
	l.PushBack(2)
	l.PushFront(3)
	l.PushBack(4)

	expected := []int{3, 1, 2, 4}
	values := l.Values()

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}

func testDeleteMatchFnNone(t *testing.T) {
	l := New(1, 3, 5, 7)
	// Try to delete even numbers (none exist)
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	if l.Len() != 4 {
		t.Errorf("Expected length 4, got %d", l.Len())
	}
}

func testDeleteMatchFnSome(t *testing.T) {
	l := New(1, 2, 3, 4, 5)
	// Delete even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	expected := []int{1, 3, 5}
	values := l.Values()
	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}

func testDeleteMatchFnAll(t *testing.T) {
	l := New(2, 4, 6, 8)
	// Delete all even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	if l.Len() != 0 {
		t.Errorf("Expected empty list, got length %d", l.Len())
	}
}

func TestDeleteMatchFn(t *testing.T) {
	t.Run("delete none", testDeleteMatchFnNone)
	t.Run("delete some", testDeleteMatchFnSome)
	t.Run("delete all", testDeleteMatchFnAll)
}

func testPopFirstMatchFnFound(t *testing.T) {
	l := New(1, 2, 3, 4, 5)
	// Pop first even number
	v, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	if !ok || v != 2 {
		t.Errorf("Expected (2, true), got (%d, %v)", v, ok)
	}
	// Verify it was removed
	values := l.Values()
	expected := []int{1, 3, 4, 5}
	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}
	for i, val := range values {
		if val != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, val, expected[i])
		}
	}
}

func testPopFirstMatchFnNotFound(t *testing.T) {
	l := New(1, 3, 5)
	// Try to pop even number (none exist)
	v, ok := l.PopFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})
	if ok {
		t.Errorf("Expected not found, got (%d, true)", v)
	}
	// List should be unchanged
	if l.Len() != 3 {
		t.Errorf("Expected length 3, got %d", l.Len())
	}
}

func TestPopFirstMatchFn(t *testing.T) {
	t.Run("found", testPopFirstMatchFnFound)
	t.Run("not found", testPopFirstMatchFnNotFound)
}

func TestMoveToBackFirstMatchFn(t *testing.T) {
	l := New(1, 2, 3, 4, 5)
	// Move first even number to back
	l.MoveToBackFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})

	values := l.Values()
	expected := []int{1, 3, 4, 5, 2}

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}

func TestMoveToFrontFirstMatchFn(t *testing.T) {
	l := New(1, 2, 3, 4, 5)
	// Move first number > 3 to front
	l.MoveToFrontFirstMatchFn(func(n int) bool {
		return n > 3
	})

	values := l.Values()
	expected := []int{4, 1, 2, 3, 5}

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}

func testFirstMatchFnFound(t *testing.T) {
	l := New("apple", "banana", "cherry", "date")
	// Find first string with length > 5
	v, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	if !ok || v != "banana" {
		t.Errorf("Expected (banana, true), got (%s, %v)", v, ok)
	}
}

func testFirstMatchFnNotFound(t *testing.T) {
	l := New("a", "b", "c")
	// Find string with length > 5
	v, ok := l.FirstMatchFn(func(s string) bool {
		return len(s) > 5
	})
	if ok {
		t.Errorf("Expected not found, got (%s, true)", v)
	}
}

func TestFirstMatchFn(t *testing.T) {
	t.Run("found", testFirstMatchFnFound)
	t.Run("not found", testFirstMatchFnNotFound)
}

func TestZero(t *testing.T) {
	l := New[int]()
	if z := l.Zero(); z != 0 {
		t.Errorf("Expected zero value 0, got %d", z)
	}

	ls := New[string]()
	if z := ls.Zero(); z != "" {
		t.Errorf("Expected zero value empty string, got %q", z)
	}

	type custom struct {
		a int
		b string
	}
	lc := New[custom]()
	z := lc.Zero()
	if z.a != 0 || z.b != "" {
		t.Errorf("Expected zero value {0, \"\"}, got %+v", z)
	}
}

func TestClone(t *testing.T) {
	original := New(1, 2, 3)
	cloned := original.Clone()

	// Verify they have same values
	if original.Len() != cloned.Len() {
		t.Errorf("Clone has different length: %d vs %d", cloned.Len(), original.Len())
	}

	origValues := original.Values()
	cloneValues := cloned.Values()
	for i := range origValues {
		if origValues[i] != cloneValues[i] {
			t.Errorf("Value mismatch at index %d: %v vs %v", i, origValues[i], cloneValues[i])
		}
	}

	// Verify they are independent
	cloned.PushBack(4)
	if original.Len() == cloned.Len() {
		t.Error("Clone modification affected original")
	}
}

func testCopyFilterTransform(t *testing.T) {
	l := New(1, 2, 3, 4, 5, 6)
	// Copy only even numbers, tripled
	copied := l.Copy(func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 3, true
		}
		return 0, false
	})

	expected := []int{6, 12, 18}
	values := copied.Values()

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}

func testCopyAll(t *testing.T) {
	l := New("a", "b", "c")
	copied := l.Copy(func(v string) (string, bool) {
		return v + v, true // double each string
	})

	expected := []string{"aa", "bb", "cc"}
	values := copied.Values()

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Value at index %d: got %s, expected %s", i, v, expected[i])
		}
	}
}

func TestCopy(t *testing.T) {
	t.Run("filter and transform", testCopyFilterTransform)
	t.Run("copy all", testCopyAll)
}

func TestPurge(t *testing.T) {
	// Purge is specifically for removing elements that don't match the type
	// This is more relevant when dealing with interface{} conversions
	// For a generic List[T], all elements should already be of type T
	l := New[int]()

	// Add some values
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Since all elements are already int, purge should remove nothing
	removed := l.Purge()
	if removed != 0 {
		t.Errorf("Expected 0 elements removed, got %d", removed)
	}

	if l.Len() != 3 {
		t.Errorf("Expected length 3 after purge, got %d", l.Len())
	}
}

func TestNilList(t *testing.T) {
	var l *List[int]

	// Test nil-safe methods
	if l.Len() != 0 {
		t.Errorf("Nil list length should be 0, got %d", l.Len())
	}

	if l.Sys() != nil {
		t.Error("Nil list Sys() should return nil")
	}

	if v, ok := l.Front(); ok {
		t.Errorf("Nil list Front() should return (zero, false), got (%v, %v)", v, ok)
	}

	if v, ok := l.Back(); ok {
		t.Errorf("Nil list Back() should return (zero, false), got (%v, %v)", v, ok)
	}

	values := l.Values()
	if len(values) != 0 {
		t.Errorf("Nil list Values() should return empty slice, got %v", values)
	}

	// These should not panic
	l.PushFront(1)
	l.PushBack(2)
	l.ForEach(func(int) bool { return true })
	l.DeleteMatchFn(func(int) bool { return true })
}
