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
	core.AssertMustEqual(t, 2, l.Len(), "length")

	// Test Values() method (which uses ForEach internally)
	core.AssertEqual(t, 2, len(l.Values()), "Values count")

	// Test ForEach directly
	var items []int
	l.ForEach(func(v int) bool {
		items = append(items, v)
		return true // Continue iteration
	})
	core.AssertEqual(t, 2, len(items), "ForEach count")

	// Manual verification using underlying list
	count := 0
	for e := l.Sys().Front(); e != nil; e = e.Next() {
		count++
	}
	core.AssertEqual(t, 2, count, "manual count")
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

	core.AssertMustEqual(t, 3, len(collected), "collected count")
	core.AssertEqual(t, 3, collected[len(collected)-1], "last item")
}

func TestValuesMethod(t *testing.T) {
	// Test that Values() returns all elements
	l := list.New[string]()
	expected := core.S("first", "second", "third")

	for _, v := range expected {
		l.PushBack(v)
	}

	core.AssertSliceEqual(t, expected, l.Values(), "values")
}

func TestEmptyList(t *testing.T) {
	// Test ForEach on empty list
	l := list.New[int]()

	count := 0
	l.ForEach(func(_ int) bool {
		count++
		return true
	})

	core.AssertEqual(t, 0, count, "callback count")
	core.AssertEqual(t, 0, len(l.Values()), "Values count")
}

type TestStruct struct {
	Name string
	ID   int
}

func makeTestList() (out *list.List[*TestStruct], first, second *TestStruct) {
	out = list.New[*TestStruct]()
	first = &TestStruct{ID: 1, Name: "first"}
	second = &TestStruct{ID: 2, Name: "second"}
	out.PushBack(first)
	out.PushBack(second)
	return out, first, second
}

func testPointerListLength(t *testing.T) {
	l, _, _ := makeTestList()
	core.AssertMustEqual(t, 2, l.Len(), "length")
}

func testPointerListValues(t *testing.T) {
	l, item1, item2 := makeTestList()
	core.AssertSliceEqual(t, core.S(item1, item2), l.Values(), "values")
}

func testPointerListForEach(t *testing.T) {
	l, item1, item2 := makeTestList()
	var items []*TestStruct
	l.ForEach(func(v *TestStruct) bool {
		items = append(items, v)
		return true
	})

	core.AssertSliceEqual(t, core.S(item1, item2), items, "items")
}

func TestForEachWithPointers(t *testing.T) {
	t.Run("length", testPointerListLength)
	t.Run("Values", testPointerListValues)
	t.Run("ForEach", testPointerListForEach)
}
