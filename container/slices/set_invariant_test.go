package slices_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/slices"
)

// TestCustomSet_SortInvariantMaintenance verifies sorted order maintained after all operations
func TestCustomSet_SortInvariantMaintenance(t *testing.T) {
	intCmp := func(a, b int) int { return a - b }
	set, err := slices.NewCustomSet(intCmp)
	core.AssertNoError(t, err, "create set")

	r := rand.New(rand.NewSource(42))

	// Perform random Add/Remove operations
	for i := range 100 {
		value := r.Intn(50)

		if i%3 == 0 && set.Len() > 0 {
			// Remove operation
			values := set.Export()
			toRemove := values[r.Intn(len(values))]
			set.Remove(toRemove)
		} else {
			// Add operation
			set.Add(value)
		}

		// After each operation, verify:
		values := set.Export()

		// 1. Values() returns sorted slice
		if !sort.IntsAreSorted(values) {
			t.Fatalf("Values not sorted after operation %d: %v", i, values)
		}

		// 2. No duplicates exist
		for j := 1; j < len(values); j++ {
			if values[j] == values[j-1] {
				t.Fatalf("Duplicate found after operation %d: %d", i, values[j])
			}
		}

		// 3. Contains() uses binary search correctly
		for _, v := range values {
			if !set.Contains(v) {
				t.Fatalf("Contains(%d) returned false after operation %d", v, i)
			}
		}

		// 4. Values not in set return false
		notInSet := -1
		for _, v := range values {
			if v == notInSet {
				notInSet = -2 // Find a value definitely not in set
			}
		}
		if set.Contains(notInSet) {
			t.Fatalf("Contains(%d) returned true for non-existent value", notInSet)
		}
	}
}

// TestOrderedSet_BinarySearchBoundaries tests searching edge cases
func TestOrderedSet_BinarySearchBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		search   int
		expected bool
	}{
		{
			name:     "empty set",
			initial:  []int{},
			search:   42,
			expected: false,
		},
		{
			name:     "single element - found",
			initial:  []int{42},
			search:   42,
			expected: true,
		},
		{
			name:     "single element - not found",
			initial:  []int{42},
			search:   41,
			expected: false,
		},
		{
			name:     "element smaller than all",
			initial:  []int{10, 20, 30},
			search:   5,
			expected: false,
		},
		{
			name:     "element larger than all",
			initial:  []int{10, 20, 30},
			search:   35,
			expected: false,
		},
		{
			name:     "element in middle",
			initial:  []int{10, 20, 30, 40, 50},
			search:   30,
			expected: true,
		},
		{
			name:     "element between existing",
			initial:  []int{10, 20, 30, 40, 50},
			search:   25,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set := slices.NewOrderedSet(tc.initial...)
			result := set.Contains(tc.search)
			core.AssertEqual(t, tc.expected, result, "contains result")

			// Verify sorted order is maintained
			values := set.Export()
			core.AssertTrue(t, sort.IntsAreSorted(values), "values sorted")
		})
	}
}

// TestCustomSet_ComplexTypeSort tests sort invariant with complex comparison
func TestCustomSet_ComplexTypeSort(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	// Sort by age, then by name
	personCmp := func(a, b Person) int {
		if a.Age != b.Age {
			return a.Age - b.Age
		}
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	}

	set, err := slices.NewCustomSet(personCmp)
	core.AssertNoError(t, err, "create person set")

	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 30},
		{"Diana", 25},
		{"Eve", 35},
	}

	// Add in random order
	for _, p := range people {
		set.Add(p)
	}

	// Verify sorted order
	values := set.Export()
	core.AssertEqual(t, 5, len(values), "all people added")

	// Check sort order: should be by age first, then name
	expected := []Person{
		{"Bob", 25},
		{"Diana", 25},
		{"Alice", 30},
		{"Charlie", 30},
		{"Eve", 35},
	}

	for i, p := range expected {
		core.AssertEqual(t, p.Name, values[i].Name, fmt.Sprintf("person %d name", i))
		core.AssertEqual(t, p.Age, values[i].Age, fmt.Sprintf("person %d age", i))
	}

	// Test that duplicates are not added
	added := set.Add(Person{"Alice", 30})
	core.AssertEqual(t, 0, added, "duplicate not added")
	core.AssertEqual(t, 5, set.Len(), "length unchanged after duplicate")
}

// TestCustomSet_InsertionPointCorrectness verifies binary search insertion points
func TestCustomSet_InsertionPointCorrectness(t *testing.T) {
	intCmp := func(a, b int) int { return a - b }
	set, err := slices.NewCustomSet(intCmp, 10, 20, 30, 40, 50)
	core.AssertNoError(t, err, "create set")

	testCases := []struct {
		value    int
		expected []int
	}{
		{5, []int{5, 10, 20, 30, 40, 50}},   // Insert at beginning
		{15, []int{10, 15, 20, 30, 40, 50}}, // Insert in middle
		{55, []int{10, 20, 30, 40, 50, 55}}, // Insert at end
		{25, []int{10, 20, 25, 30, 40, 50}}, // Insert between elements
		{35, []int{10, 20, 30, 35, 40, 50}}, // Another middle insertion
	}

	for _, tc := range testCases {
		// Clone the original set for each test
		testSet := set.Clone().(*slices.CustomSet[int])
		testSet.Add(tc.value)

		values := testSet.Export()
		core.AssertSliceEqual(t, tc.expected, values, fmt.Sprintf("after adding %d", tc.value))
		core.AssertTrue(t, sort.IntsAreSorted(values), fmt.Sprintf("sorted after adding %d", tc.value))
	}
}

// TestSet_RandomOperationsInvariant performs random operations and checks invariants
func TestSet_RandomOperationsInvariant(t *testing.T) {
	set := slices.NewOrderedSet[int]()
	r := rand.New(rand.NewSource(123))

	// Track what should be in the set
	reference := make(map[int]bool)

	for i := range 500 {
		op := r.Intn(4)
		value := r.Intn(100)

		switch op {
		case 0, 1: // Add (50% chance)
			added := set.Add(value)
			if _, exists := reference[value]; exists {
				core.AssertEqual(t, 0, added, fmt.Sprintf("Add(%d) should return 0 for duplicate", value))
			} else {
				core.AssertEqual(t, 1, added, fmt.Sprintf("Add(%d) should return 1 for new value", value))
				reference[value] = true
			}

		case 2: // Delete (25% chance)
			deleted := set.Remove(value)
			if reference[value] {
				core.AssertEqual(t, 1, deleted, fmt.Sprintf("Delete(%d) should return 1", value))
				delete(reference, value)
			} else {
				core.AssertEqual(t, 0, deleted, fmt.Sprintf("Delete(%d) should return 0", value))
			}

		case 3: // Contains check (25% chance)
			contains := set.Contains(value)
			expected := reference[value]
			core.AssertEqual(t, expected, contains, fmt.Sprintf("Contains(%d)", value))
		}

		// Verify invariants every 50 operations
		if i%50 == 0 {
			values := set.Export()

			// Check sorted
			core.AssertTrue(t, sort.IntsAreSorted(values), fmt.Sprintf("sorted at op %d", i))

			// Check no duplicates
			seen := make(map[int]bool)
			for _, v := range values {
				core.AssertFalse(t, seen[v], fmt.Sprintf("duplicate %d at op %d", v, i))
				seen[v] = true
			}

			// Check length matches
			core.AssertEqual(t, len(reference), set.Len(), fmt.Sprintf("length at op %d", i))

			// Check all reference values are in set
			for v := range reference {
				core.AssertTrue(t, set.Contains(v), fmt.Sprintf("missing %d at op %d", v, i))
			}

			// Check no extra values in set
			for _, v := range values {
				core.AssertTrue(t, reference[v], fmt.Sprintf("extra %d at op %d", v, i))
			}
		}
	}
}
