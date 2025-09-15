package list_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/list"
)

func TestForEachBug(t *testing.T) {
	// Create a new list
	l := list.New[int]()

	// Add two items
	l.PushBack(1)
	l.PushBack(2)

	// Verify length is correct
	core.AssertMustEqual(t, 2, l.Len(), "list length")

	// Test Values() method (which uses ForEach internally)
	values := l.Values()
	core.AssertEqual(t, 2, len(values), "values count")

	// Test ForEach directly
	var items []int
	l.ForEach(func(v int) bool {
		items = append(items, v)
		return true // Continue iteration
	})

	core.AssertEqual(t, 2, len(items), "ForEach items count")

	// Manual verification using underlying list
	count := 0
	sysList := l.Sys()
	for e := sysList.Front(); e != nil; e = e.Next() {
		count++
	}

	core.AssertEqual(t, 2, count, "manual iteration count")
}

func TestForEachEarlyTermination(t *testing.T) {
	// Test that ForEach can be terminated early by returning false
	l := list.New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}

	var collected []int
	l.ForEach(func(v int) bool {
		collected = append(collected, v)
		return v < 3 // Stop after 3
	})

	core.AssertEqual(t, 3, len(collected), "early termination count")

	core.AssertEqual(t, 3, collected[len(collected)-1], "last collected item")
}

func TestValuesMethod(t *testing.T) {
	// Test that Values() returns all elements
	l := list.New[string]()
	expected := []string{"first", "second", "third"}

	for _, v := range expected {
		l.PushBack(v)
	}

	values := l.Values()
	core.AssertEqual(t, len(expected), len(values), "values count")

	for i, v := range values {
		core.AssertEqual(t, expected[i], v, "value at index %d", i)
	}
}

func TestEmptyList(t *testing.T) {
	// Test ForEach on empty list
	l := list.New[int]()

	count := 0
	l.ForEach(func(_ int) bool {
		count++
		return true
	})

	core.AssertEqual(t, 0, count, "empty list ForEach count")

	values := l.Values()
	core.AssertEqual(t, 0, len(values), "empty list values count")
}

type TestStruct struct {
	Name string
	ID   int
}

func makeTestList() (l *list.List[*TestStruct], first, second *TestStruct) {
	l = list.New[*TestStruct]()
	first = &TestStruct{ID: 1, Name: "first"}
	second = &TestStruct{ID: 2, Name: "second"}
	l.PushBack(first)
	l.PushBack(second)
	return l, first, second
}

func runTestPointerListLength(t *testing.T) {
	t.Helper()
	l, _, _ := makeTestList()
	core.AssertMustEqual(t, 2, l.Len(), "list length")
}

func runTestPointerListValues(t *testing.T) {
	t.Helper()
	l, item1, item2 := makeTestList()
	values := l.Values()
	core.AssertEqual(t, 2, len(values), "values count")
	if len(values) >= 1 {
		core.AssertSame(t, item1, values[0], "first value")
	}
	if len(values) >= 2 {
		core.AssertSame(t, item2, values[1], "second value")
	}
}

func runTestPointerListForEach(t *testing.T) {
	t.Helper()
	l, item1, item2 := makeTestList()
	var items []*TestStruct
	l.ForEach(func(v *TestStruct) bool {
		items = append(items, v)
		return true
	})

	core.AssertEqual(t, 2, len(items), "ForEach items count")
	if len(items) >= 1 {
		core.AssertSame(t, item1, items[0], "first item")
	}
	if len(items) >= 2 {
		core.AssertSame(t, item2, items[1], "second item")
	}
}

func TestForEachWithPointers(t *testing.T) {
	t.Run("length", runTestPointerListLength)
	t.Run("Values", runTestPointerListValues)
	t.Run("ForEach", runTestPointerListForEach)
}
