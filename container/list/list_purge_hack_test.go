package list_test

// cspell:ignore stdlist

import (
	stdlist "container/list"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
)

// TestPurgeForceFailure creates a raw list with mixed types then casts to list.List[int]
func TestPurgeForceFailure(t *testing.T) {
	// Create a raw list.List with mixed types
	rawList := stdlist.New()
	rawList.PushBack(1)
	rawList.PushBack("not an int")
	rawList.PushBack(2)
	rawList.PushBack(3.14)
	rawList.PushBack(3)
	rawList.PushBack(struct{}{})

	// Cast the raw list to list.List[int]
	typedList := (*list.List[int])(rawList)

	// Now Purge should actually remove the non-int elements
	removed := typedList.Purge()

	// Should have removed 3 elements (string, float, struct)
	core.AssertEqual(t, 3, removed, "removed count")
	core.AssertEqual(t, 3, rawList.Len(), "remaining elements")

	// Verify remaining elements are all ints
	for e := rawList.Front(); e != nil; e = e.Next() {
		_, ok := e.Value.(int)
		core.AssertTrue(t, ok, "remaining element is int")
	}
}

// TestPurgeUnsafePointerHack creates raw list first then casts
func TestPurgeUnsafePointerHack(t *testing.T) {
	// Create a raw list with mixed types
	raw := stdlist.New()
	raw.PushBack(1)
	raw.PushBack(2)
	raw.PushBack(3)
	raw.PushBack("wrong")
	raw.PushBack(false)

	// Cast to list.List[int]
	intList := (*list.List[int])(raw)

	// Call Purge
	removed := intList.Purge()
	core.AssertEqual(t, 2, removed, "removed non-int elements")

	// Check what's left
	values := intList.Values()
	core.AssertEqual(t, 3, len(values), "only ints remain")
	for _, v := range values {
		core.AssertTrue(t, v > 0 && v <= 3, "valid int value")
	}
}

// TestPurgeWithActualWrongTypes directly tests with wrong types in underlying list
func TestPurgeWithActualWrongTypes(t *testing.T) {
	// Create a raw list with mixed types
	raw := stdlist.New()
	raw.PushBack("valid string 1")
	raw.PushBack(123) // wrong type: int
	raw.PushBack("valid string 2")
	raw.PushBack(true) // wrong type: bool
	raw.PushBack("valid string 3")
	raw.PushBack(3.14159) // wrong type: float64

	// Cast to list.List[string]
	strList := (*list.List[string])(raw)

	// Now call Purge - it should remove the non-string elements
	removed := strList.Purge()

	// Should remove 3 elements (int, bool, float)
	core.AssertEqual(t, 3, removed, "removed non-string elements")
	core.AssertEqual(t, 3, strList.Len(), "only strings remain")

	// Verify remaining are strings
	values := strList.Values()
	core.AssertEqual(t, 3, len(values), "three strings")
	core.AssertEqual(t, "valid string 1", values[0], "first string")
	core.AssertEqual(t, "valid string 2", values[1], "second string")
	core.AssertEqual(t, "valid string 3", values[2], "third string")
}
