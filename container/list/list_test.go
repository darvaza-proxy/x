package list

import (
	"testing"
)

func TestForEachBug(t *testing.T) {
	// Create a new list
	l := New[int]()

	// Add two items
	l.PushBack(1)
	l.PushBack(2)

	// Verify length is correct
	if l.Len() != 2 {
		t.Fatalf("Expected length 2, got %d", l.Len())
	}

	// Test Values() method (which uses ForEach internally)
	values := l.Values()
	if len(values) != 2 {
		t.Errorf("Values() returned %d items, expected 2", len(values))
		t.Logf("Values: %v", values)
	}

	// Test ForEach directly
	var items []int
	l.ForEach(func(v int) bool {
		items = append(items, v)
		return true // Continue iteration
	})

	if len(items) != 2 {
		t.Errorf("ForEach collected %d items, expected 2", len(items))
		t.Logf("ForEach items: %v", items)
	}

	// Manual verification using underlying list
	count := 0
	sysList := l.Sys()
	for e := sysList.Front(); e != nil; e = e.Next() {
		count++
		if v, ok := e.Value.(int); ok {
			t.Logf("Manual iteration [%d]: %d", count-1, v)
		}
	}

	if count != 2 {
		t.Errorf("Manual iteration found %d items, expected 2", count)
	}
}

func TestForEachEarlyTermination(t *testing.T) {
	// Test that ForEach can be terminated early by returning false
	l := New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}

	var collected []int
	l.ForEach(func(v int) bool {
		collected = append(collected, v)
		return v < 3 // Stop after 3
	})

	if len(collected) != 3 {
		t.Errorf("Expected to collect 3 items before stopping, got %d", len(collected))
	}

	if collected[len(collected)-1] != 3 {
		t.Errorf("Expected last item to be 3, got %d", collected[len(collected)-1])
	}
}

func TestValuesMethod(t *testing.T) {
	// Test that Values() returns all elements
	l := New[string]()
	expected := []string{"first", "second", "third"}

	for _, v := range expected {
		l.PushBack(v)
	}

	values := l.Values()
	if len(values) != len(expected) {
		t.Errorf("Values() returned %d items, expected %d", len(values), len(expected))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Values()[%d] = %q, expected %q", i, v, expected[i])
		}
	}
}

func TestEmptyList(t *testing.T) {
	// Test ForEach on empty list
	l := New[int]()

	count := 0
	l.ForEach(func(_ int) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("ForEach on empty list called callback %d times, expected 0", count)
	}

	values := l.Values()
	if len(values) != 0 {
		t.Errorf("Values() on empty list returned %d items, expected 0", len(values))
	}
}

type TestStruct struct {
	ID   int
	Name string
}

func makeTestList() (list *List[*TestStruct], first, second *TestStruct) {
	list = New[*TestStruct]()
	first = &TestStruct{ID: 1, Name: "first"}
	second = &TestStruct{ID: 2, Name: "second"}
	list.PushBack(first)
	list.PushBack(second)
	return list, first, second
}

func testPointerListLength(t *testing.T) {
	l, _, _ := makeTestList()
	if l.Len() != 2 {
		t.Fatalf("Expected length 2, got %d", l.Len())
	}
}

func testPointerListValues(t *testing.T) {
	l, item1, item2 := makeTestList()
	values := l.Values()
	if len(values) != 2 {
		t.Errorf("Values() returned %d items, expected 2", len(values))
	}
	if len(values) >= 1 && values[0] != item1 {
		t.Errorf("First value mismatch: got %+v, expected %+v", values[0], item1)
	}
	if len(values) >= 2 && values[1] != item2 {
		t.Errorf("Second value mismatch: got %+v, expected %+v", values[1], item2)
	}
}

func testPointerListForEach(t *testing.T) {
	l, item1, item2 := makeTestList()
	var items []*TestStruct
	l.ForEach(func(v *TestStruct) bool {
		items = append(items, v)
		t.Logf("ForEach item: %+v", v)
		return true
	})

	if len(items) != 2 {
		t.Errorf("ForEach collected %d items, expected 2", len(items))
	}
	if len(items) >= 1 && items[0] != item1 {
		t.Errorf("First item mismatch: got %+v, expected %+v", items[0], item1)
	}
	if len(items) >= 2 && items[1] != item2 {
		t.Errorf("Second item mismatch: got %+v, expected %+v", items[1], item2)
	}
}

func TestForEachWithPointers(t *testing.T) {
	t.Run("length", testPointerListLength)
	t.Run("Values", testPointerListValues)
	t.Run("ForEach", testPointerListForEach)
}
