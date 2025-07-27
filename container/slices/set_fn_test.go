package slices

import (
	"testing"
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

// Assert helpers
func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

//revive:disable-next-line:flag-parameter
func assertFalse(t *testing.T, v bool, msg string) {
	t.Helper()
	if v {
		t.Errorf("%s: expected false, got true", msg)
	}
}

func TestCustomSet_New(t *testing.T) {
	set, err := NewCustomSet(cmpInt)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}
	other := set.New()

	if other == nil {
		t.Fatal("New() returned nil")
	}

	// Verify it's empty
	if other.Len() != 0 {
		t.Errorf("New() created set with %d elements, expected 0", other.Len())
	}
}

func TestCustomSet_InitCustomSet(t *testing.T) {
	var set CustomSet[int]

	if err := InitCustomSet(&set, cmpInt, 1, 2, 3); err != nil {
		t.Fatalf("InitCustomSet failed: %v", err)
	}

	if set.Len() != 3 {
		t.Errorf("InitCustomSet created set with %d elements, expected 3", set.Len())
	}

	// Test init on already initialized set
	if err := InitCustomSet(&set, cmpInt); err == nil {
		t.Error("InitCustomSet should fail on already initialized set")
	}
}

func TestCustomSet_Contains(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 1, 3, 5, 7, 9)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

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
		if got := set.Contains(tc.value); got != tc.expect {
			t.Errorf("Contains(%d) = %v, expected %v", tc.value, got, tc.expect)
		}
	}
}

func TestCustomSet_Clone(t *testing.T) {
	original, err := NewCustomSet(cmpInt, 1, 2, 3)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}
	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone() returned nil")
	}

	// Verify same elements
	if clone.Len() != original.Len() {
		t.Errorf("Clone has %d elements, original has %d", clone.Len(), original.Len())
	}

	for i := 0; i < original.Len(); i++ {
		origVal, _ := original.GetByIndex(i)
		cloneVal, _ := clone.(*CustomSet[int]).GetByIndex(i)
		if origVal != cloneVal {
			t.Errorf("Clone element %d: got %d, expected %d", i, cloneVal, origVal)
		}
	}

	// Verify independence
	original.Add(4)
	if clone.Contains(4) {
		t.Error("Clone should not be affected by original modification")
	}
}

func TestCustomSet_Clear(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	if set.Len() != 5 {
		t.Fatalf("Initial set has %d elements, expected 5", set.Len())
	}

	set.Clear()

	if set.Len() != 0 {
		t.Errorf("Clear() left %d elements, expected 0", set.Len())
	}

	// Verify capacity remains
	available, total := set.Cap()
	if total == 0 {
		t.Error("Clear() should preserve capacity")
	}
	_ = available // unused
}

func TestCustomSet_Purge(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 5, 3, 1, 4, 2)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	elements := set.Purge()

	// Verify returned elements
	if len(elements) != 5 {
		t.Errorf("Purge() returned %d elements, expected 5", len(elements))
	}

	// Elements should be sorted
	expected := []int{1, 2, 3, 4, 5}
	for i, v := range elements {
		if v != expected[i] {
			t.Errorf("Purge() element %d: got %d, expected %d", i, v, expected[i])
		}
	}

	// Verify set is empty with no capacity
	if set.Len() != 0 {
		t.Errorf("Purge() left %d elements, expected 0", set.Len())
	}

	_, total := set.Cap()
	if total != 0 {
		t.Errorf("Purge() left capacity %d, expected 0", total)
	}
}

func TestCustomSet_ForEach(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	// Test full iteration
	var sum int
	set.ForEach(func(v int) bool {
		sum += v
		return true
	})

	if sum != 15 {
		t.Errorf("ForEach sum = %d, expected 15", sum)
	}

	// Test early termination
	var count int
	set.ForEach(func(_ int) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Errorf("ForEach early termination count = %d, expected 3", count)
	}
}

func TestCustomSet_Reserve(t *testing.T) {
	set, err := NewCustomSet(cmpInt)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	// Reserve capacity
	if !set.Reserve(10) {
		t.Error("Reserve(10) returned false")
	}

	available, _ := set.Cap()
	if available < 10 {
		t.Errorf("After Reserve(10), available capacity = %d, expected >= 10", available)
	}

	// Reserve less than current should return false
	if set.Reserve(5) {
		t.Error("Reserve(5) should return false when capacity is already >= 10")
	}
}

func TestCustomSet_Grow(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 1, 2, 3)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	initialAvailable, _ := set.Cap()

	// Grow capacity
	if !set.Grow(5) {
		t.Error("Grow(5) returned false")
	}

	newAvailable, _ := set.Cap()
	if newAvailable < initialAvailable+5 {
		t.Errorf("After Grow(5), available capacity increased by %d, expected >= 5",
			newAvailable-initialAvailable)
	}
}

func TestCustomSet_Trim(t *testing.T) {
	set, err := NewCustomSet(cmpInt)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	// Add elements and then grow capacity
	set.Add(1, 2, 3)
	set.Grow(20)

	available, total := set.Cap()
	if available <= 0 {
		t.Fatalf("After Grow(20), available capacity = %d, expected > 0", available)
	}
	if total <= set.Len() {
		t.Fatalf("After Grow(20), total capacity = %d, expected > %d", total, set.Len())
	}

	// Trim excess capacity
	if !set.Trim() {
		t.Error("Trim() returned false")
	}

	newAvailable, newTotal := set.Cap()
	if newTotal != set.Len() {
		t.Errorf("After Trim(), total capacity = %d, expected %d", newTotal, set.Len())
	}
	if newAvailable != 0 {
		t.Errorf("After Trim(), available capacity = %d, expected 0", newAvailable)
	}
}

func TestInitOrderedSet(t *testing.T) {
	var set CustomSet[int]

	if err := InitOrderedSet(&set, 3, 1, 4, 1, 5); err != nil {
		t.Fatalf("InitOrderedSet failed: %v", err)
	}

	// Should have deduplicated and sorted
	if set.Len() != 4 {
		t.Errorf("InitOrderedSet created set with %d elements, expected 4", set.Len())
	}

	// Verify sorted order
	expected := []int{1, 3, 4, 5}
	for i, exp := range expected {
		if v, _ := set.GetByIndex(i); v != exp {
			t.Errorf("Element %d: got %d, expected %d", i, v, exp)
		}
	}
}

func TestCustomSet_GetByIndex_Error(t *testing.T) {
	set, err := NewCustomSet(cmpInt, 1, 2, 3)
	assertNoError(t, err, "NewCustomSet")

	// Test negative index
	_, ok := set.GetByIndex(-1)
	assertFalse(t, ok, "GetByIndex(-1) should return false")

	// Test out of bounds
	_, ok = set.GetByIndex(3)
	assertFalse(t, ok, "GetByIndex(3) with 3 elements should return false")
}

func TestCustomSet_Nil(t *testing.T) {
	var set *CustomSet[int]

	// Test nil receiver methods
	if set.Contains(1) {
		t.Error("nil set Contains should return false")
	}

	if set.Len() != 0 {
		t.Error("nil set Len should return 0")
	}

	available, total := set.Cap()
	if available != 0 || total != 0 {
		t.Error("nil set Cap should return 0, 0")
	}

	if set.Add(1) != 0 {
		t.Error("nil set Add should return 0")
	}

	if set.Remove(1) != 0 {
		t.Error("nil set Remove should return 0")
	}

	elements := set.Export()
	if len(elements) != 0 {
		t.Error("nil set Export should return empty slice")
	}
}

func TestMustCustomSet_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustCustomSet should panic with nil comparison function")
		}
	}()

	// This should panic
	_ = MustCustomSet[int](nil)
}

func TestCustomSet_Complex(t *testing.T) {
	// Test with a custom type
	type Person struct {
		ID   int
		Name string
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

	set, err := NewCustomSet(cmpPerson,
		Person{ID: 3, Name: "Charlie"},
		Person{ID: 1, Name: "Alice"},
		Person{ID: 2, Name: "Bob"},
		Person{ID: 1, Name: "Alice Duplicate"},
	)
	if err != nil {
		t.Fatalf("NewCustomSet failed: %v", err)
	}

	// Should have 3 unique persons (by ID)
	if set.Len() != 3 {
		t.Errorf("Set has %d persons, expected 3", set.Len())
	}

	// Should be sorted by ID
	first, _ := set.GetByIndex(0)
	if first.ID != 1 {
		t.Errorf("First person ID = %d, expected 1", first.ID)
	}
}
