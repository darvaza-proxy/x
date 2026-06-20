package slices_test

import (
	"fmt"

	"darvaza.org/x/container/slices"
)

// printLn / printF wrap fmt.Println / fmt.Printf to drop the
// byte-count and error returns that revive's unhandled-error
// rule would otherwise flag in every Example body.
func printLn(args ...any)               { _, _ = fmt.Println(args...) }
func printF(format string, args ...any) { _, _ = fmt.Printf(format, args...) }

func ExampleNewOrderedSet() {
	// Create a set of integers
	set := slices.NewOrderedSet(3, 1, 4, 1, 5, 9, 2, 6)

	// The set automatically deduplicates and sorts
	printLn("Length:", set.Len())
	printLn("Contains 4:", set.Contains(4))
	printLn("Contains 7:", set.Contains(7))

	// Export returns a sorted slice
	values := set.Export()
	printLn("Values:", values)

	// Output:
	// Length: 7
	// Contains 4: true
	// Contains 7: false
	// Values: [1 2 3 4 5 6 9]
}

func ExampleCustomSet_Add() {
	set := slices.NewOrderedSet[int]()

	// Add single value
	added := set.Add(5)
	printLn("Added 5:", added)

	// Add multiple values
	added = set.Add(3, 7, 1, 9)
	printLn("Added 4 values:", added)

	// Try to add duplicates
	added = set.Add(5, 3, 11)
	printLn("Added with duplicates:", added)

	printLn("Final set:", set.Export())

	// Output:
	// Added 5: 1
	// Added 4 values: 4
	// Added with duplicates: 1
	// Final set: [1 3 5 7 9 11]
}

func ExampleCustomSet_Remove() {
	set := slices.NewOrderedSet(1, 2, 3, 4, 5)

	// Remove existing value
	removed := set.Remove(3)
	printLn("Removed 3:", removed)

	// Remove non-existing value
	removed = set.Remove(10)
	printLn("Removed 10:", removed)

	// Remove multiple values
	removed = set.Remove(1, 5)
	printLn("Removed 1 and 5:", removed)

	printLn("Remaining:", set.Export())

	// Output:
	// Removed 3: 1
	// Removed 10: 0
	// Removed 1 and 5: 2
	// Remaining: [2 4]
}

func ExampleCustomSet_ForEach() {
	set := slices.NewOrderedSet(10, 20, 30, 40, 50)

	// Sum all values
	sum := 0
	set.ForEach(func(v int) bool {
		sum += v
		return true // continue
	})
	printLn("Sum:", sum)

	// Find first value > 25
	var found int
	set.ForEach(func(v int) bool {
		if v > 25 {
			found = v
			return false // stop
		}
		return true
	})
	printLn("First > 25:", found)

	// Output:
	// Sum: 150
	// First > 25: 30
}

func ExampleNewCustomSet() {
	// Custom type with comparison function
	type Person struct {
		Name string
		ID   int
	}

	// Compare by ID
	cmpByID := func(a, b Person) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	}

	// Create set with custom comparison
	set, err := slices.NewCustomSet(cmpByID,
		Person{ID: 3, Name: "Charlie"},
		Person{ID: 1, Name: "Alice"},
		Person{ID: 2, Name: "Bob"},
		Person{ID: 1, Name: "Alice Duplicate"}, // Will be ignored
	)
	if err != nil {
		panic(err)
	}

	printLn("People count:", set.Len())

	// Export in sorted order (by ID)
	people := set.Export()
	for _, p := range people {
		printF("ID=%d: %s\n", p.ID, p.Name)
	}

	// Output:
	// People count: 3
	// ID=1: Alice
	// ID=2: Bob
	// ID=3: Charlie
}

func ExampleCustomSet_Clone() {
	original := slices.NewOrderedSet(1, 2, 3)

	// Create a clone
	clone := original.Clone()

	// Modify original
	original.Add(4)

	// Clone is independent
	printLn("Original:", original.Export())
	printLn("Clone:", clone.Export())

	// Output:
	// Original: [1 2 3 4]
	// Clone: [1 2 3]
}

func ExampleCustomSet_Reserve() {
	set := slices.NewOrderedSet[int]()

	// Reserve capacity for 100 elements
	reserved := set.Reserve(100)
	printLn("Reserved for 100:", reserved)

	available, total := set.Cap()
	printF("Capacity: %d available, %d total\n", available, total)

	// Add some elements
	set.Add(1, 2, 3, 4, 5)

	// Trim excess capacity
	trimmed := set.Trim()
	printLn("Trimmed:", trimmed)

	available, total = set.Cap()
	printF("After trim: %d available, %d total\n", available, total)

	// Output:
	// Reserved for 100: true
	// Capacity: 100 available, 100 total
	// Trimmed: true
	// After trim: 0 available, 5 total
}
