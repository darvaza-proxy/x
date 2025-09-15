package slices_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/slices"
)

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// No longer need custom assert helpers since we use core

func TestCustomSet_New(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")
	other := set.New()

	core.AssertNotNil(t, other, "New() result")

	// Verify it's empty
	core.AssertEqual(t, 0, other.Len(), "new set length")
}

func TestCustomSet_InitCustomSet(t *testing.T) {
	var set slices.CustomSet[int]

	err := slices.InitCustomSet(&set, cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "InitCustomSet")

	core.AssertEqual(t, 3, set.Len(), "set length")

	// Test init on already initialized set
	err = slices.InitCustomSet(&set, cmpInt)
	core.AssertError(t, err, "InitCustomSet on initialized set should fail")
}

func TestCustomSet_UninitializedPanic(t *testing.T) {
	t.Run("New on uninitialized", runTestCustomSetUninitializedNew)
	t.Run("Clone on uninitialized", runTestCustomSetUninitializedClone)
	t.Run("Add on uninitialized", runTestCustomSetUninitializedAdd)
	t.Run("Remove on uninitialized", runTestCustomSetUninitializedRemove)
}

func runTestCustomSetUninitializedNew(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	core.AssertPanic(t, func() {
		_ = set.New()
	}, nil, "CustomSet not initialised")
}

func runTestCustomSetUninitializedClone(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	core.AssertPanic(t, func() {
		_ = set.Clone()
	}, nil, "CustomSet not initialised")
}

func runTestCustomSetUninitializedAdd(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	core.AssertPanic(t, func() {
		_ = set.Add(1, 2, 3)
	}, nil, "CustomSet not initialised")
}

func runTestCustomSetUninitializedRemove(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	core.AssertPanic(t, func() {
		_ = set.Remove(1)
	}, nil, "CustomSet not initialised")
}

func TestCustomSet_Contains(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 3, 5, 7, 9)
	core.AssertMustNoError(t, err, "NewCustomSet")

	tests := []struct {
		value  int
		expect bool
	}{
		{1, true},
		{3, true},
		{5, true},
		{7, true},
		{9, true},
		{2, false},
		{4, false},
		{10, false},
	}

	for _, tc := range tests {
		got := set.Contains(tc.value)
		core.AssertEqual(t, tc.expect, got, "Contains(%d)", tc.value)
	}
}

func TestCustomSet_Clone(t *testing.T) {
	original, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")
	clone := original.Clone()

	core.AssertNotNil(t, clone, "Clone() result")

	// Verify same elements
	core.AssertEqual(t, original.Len(), clone.Len(), "clone length")

	for i := 0; i < original.Len(); i++ {
		origVal, _ := original.GetByIndex(i)
		cloneVal, _ := clone.(*slices.CustomSet[int]).GetByIndex(i)
		core.AssertEqual(t, origVal, cloneVal, "element at index %d", i)
	}

	// Verify independence
	original.Add(4)
	core.AssertEqual(t, false, clone.Contains(4), "clone independence")
}

func TestCustomSet_Clear(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3, 4, 5)
	core.AssertMustNoError(t, err, "NewCustomSet")

	core.AssertMustEqual(t, 5, set.Len(), "initial set length")

	set.Clear()

	core.AssertEqual(t, 0, set.Len(), "set length after clear")

	// Verify capacity remains
	available, total := set.Cap()
	core.AssertTrue(t, total > 0, "Clear() should preserve capacity")
	_ = available // unused
}

func TestCustomSet_Purge(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 5, 3, 1, 4, 2)
	core.AssertMustNoError(t, err, "NewCustomSet")

	elements := set.Purge()

	// Verify returned elements
	core.AssertEqual(t, 5, len(elements), "purged elements count")

	// Elements should be sorted
	expected := []int{1, 2, 3, 4, 5}
	for i, v := range elements {
		core.AssertEqual(t, expected[i], v, "element at index %d", i)
	}

	// Verify set is empty with no capacity
	core.AssertEqual(t, 0, set.Len(), "set length after purge")

	_, total := set.Cap()
	core.AssertEqual(t, 0, total, "capacity after purge")
}

func TestCustomSet_ForEach(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3, 4, 5)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Test full iteration
	var sum int
	set.ForEach(func(v int) bool {
		sum += v
		return true
	})

	core.AssertEqual(t, 15, sum, "ForEach sum")

	// Test early termination
	var count int
	set.ForEach(func(_ int) bool {
		count++
		return count < 3
	})

	core.AssertEqual(t, 3, count, "ForEach early termination count")
}

func TestCustomSet_Reserve(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Reserve capacity
	core.AssertEqual(t, true, set.Reserve(10), "Reserve(10) result")

	available, _ := set.Cap()
	core.AssertTrue(t, available >= 10, "available capacity should be >= 10")

	// Reserve less than current should return false
	core.AssertEqual(t, false, set.Reserve(5), "Reserve(5) should return false")
}

func TestCustomSet_Grow(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")

	initialAvailable, _ := set.Cap()

	// Grow capacity
	core.AssertEqual(t, true, set.Grow(5), "Grow(5) result")

	newAvailable, _ := set.Cap()
	core.AssertTrue(t, newAvailable >= initialAvailable+5, "capacity should increase by at least 5")
}

func TestCustomSet_Trim(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Add elements and then grow capacity
	set.Add(1, 2, 3)
	set.Grow(20)

	available, total := set.Cap()
	core.AssertMustTrue(t, available > 0, "available capacity should be > 0 after grow")
	core.AssertMustTrue(t, total > set.Len(), "total capacity should be > set length")

	// Trim excess capacity
	core.AssertEqual(t, true, set.Trim(), "Trim() result")

	newAvailable, newTotal := set.Cap()
	core.AssertEqual(t, set.Len(), newTotal, "total capacity after trim")
	core.AssertEqual(t, 0, newAvailable, "available capacity after trim")
}

func TestInitOrderedSet(t *testing.T) {
	var set slices.CustomSet[int]

	err := slices.InitOrderedSet(&set, 3, 1, 4, 1, 5)
	core.AssertMustNoError(t, err, "InitOrderedSet")

	// Should have deduplicated and sorted
	core.AssertEqual(t, 4, set.Len(), "set length after init")

	// Verify sorted order
	expected := []int{1, 3, 4, 5}
	for i, exp := range expected {
		v, _ := set.GetByIndex(i)
		core.AssertEqual(t, exp, v, "element at index %d", i)
	}
}

func TestCustomSet_GetByIndex_Error(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Test negative index
	_, ok := set.GetByIndex(-1)
	core.AssertEqual(t, false, ok, "GetByIndex(-1) should return false")

	// Test out of bounds
	_, ok = set.GetByIndex(3)
	core.AssertEqual(t, false, ok, "GetByIndex(3) with 3 elements should return false")
}

func TestCustomSet_Nil(t *testing.T) {
	var set *slices.CustomSet[int]

	// Test nil receiver methods
	core.AssertEqual(t, false, set.Contains(1), "nil set Contains")

	core.AssertEqual(t, 0, set.Len(), "nil set Len")

	available, total := set.Cap()
	core.AssertEqual(t, 0, available, "nil set available capacity")
	core.AssertEqual(t, 0, total, "nil set total capacity")

	core.AssertEqual(t, 0, set.Add(1), "nil set Add")

	core.AssertEqual(t, 0, set.Remove(1), "nil set Remove")

	elements := set.Export()
	core.AssertEqual(t, 0, len(elements), "nil set Export")
}

func TestMustCustomSet_Panic(t *testing.T) {
	// This should panic with nil comparison function
	core.AssertPanic(t, func() {
		_ = slices.MustCustomSet[int](nil)
	}, nil, "nil comparison function")
}

func TestCustomSet_Complex(t *testing.T) {
	// Test with a custom type
	type Person struct {
		Name string
		ID   int
	}

	cmpPerson := func(a, b Person) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	}

	set, err := slices.NewCustomSet(cmpPerson,
		Person{ID: 3, Name: "Charlie"},
		Person{ID: 1, Name: "Alice"},
		Person{ID: 2, Name: "Bob"},
		Person{ID: 1, Name: "Alice Duplicate"},
	)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Should have 3 unique persons (by ID)
	core.AssertEqual(t, 3, set.Len(), "set length")

	// Should be sorted by ID
	first, _ := set.GetByIndex(0)
	core.AssertEqual(t, 1, first.ID, "first person ID")
}
