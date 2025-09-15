package slices

import (
	"testing"

	"darvaza.org/core"
)

// Tests for private methods that need internal package access

func TestCustomSet_PrivateSearchBoundaries(t *testing.T) {
	t.Run("search empty slice", runTestCustomSetPrivateSearchEmpty)
}

func runTestCustomSetPrivateSearchEmpty(t *testing.T) {
	t.Helper()
	set := NewOrderedSet[int]() // Use factory
	i, found := set.search(0, 5)
	core.AssertFalse(t, found, "search in empty set")
	core.AssertEqual(t, 0, i, "index")
}

func TestCustomSet_PrivateDoRemoveOne(t *testing.T) {
	t.Run("doRemoveOne from empty set", runTestCustomSetPrivateDoRemoveOneEmpty)
}

func runTestCustomSetPrivateDoRemoveOneEmpty(t *testing.T) {
	t.Helper()
	set := NewOrderedSet[int]() // Use factory for empty set
	i, removed := set.doRemoveOne(0, 5)
	core.AssertFalse(t, removed, "remove from empty")
	core.AssertEqual(t, 0, i, "index")
}

func runTestDoTrimZeroCapacity(t *testing.T) {
	t.Helper()
	// Create a set with capacity but no elements to trigger the capacity==0 case
	set := NewOrderedSet[int]()
	set.Reserve(5) // Give it capacity
	// Now it has capacity but no elements - perfect for testing capacity==0 branch
	changed := set.doTrim(0)
	core.AssertTrue(t, changed, "trim to zero")
	core.AssertEqual(t, 0, len(set.s), "length")
	core.AssertEqual(t, 0, cap(set.s), "capacity")
}

func runTestDoTrimSameCapacity(t *testing.T) {
	t.Helper()
	set := NewOrderedSet(1, 2, 3)
	l := set.Len()
	_, c := set.Cap()
	set.mu.Lock()
	changed := set.doTrim(c)
	set.mu.Unlock()
	core.AssertFalse(t, changed, "trim same capacity")
	core.AssertEqual(t, l, set.Len(), "length unchanged")
}

func runTestDoTrimReduceCapacity(t *testing.T) {
	t.Helper()
	set := NewOrderedSet(1, 2, 3)
	set.Reserve(10)
	set.mu.Lock()
	changed := set.doTrim(3)
	set.mu.Unlock()
	core.AssertTrue(t, changed, "trim changed")
	core.AssertEqual(t, 3, set.Len(), "length")
	_, c := set.Cap()
	core.AssertEqual(t, 3, c, "capacity")
}

func TestCustomSet_PrivateDoTrim(t *testing.T) {
	t.Run("doTrim with zero capacity", runTestDoTrimZeroCapacity)
	t.Run("doTrim with same capacity", runTestDoTrimSameCapacity)
	t.Run("doTrim reduces capacity", runTestDoTrimReduceCapacity)
}
