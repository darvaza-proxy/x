package slices_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/slices"
)

// Compile-time verification
var _ core.TestCase = removeLastElementTestCase{}

func TestCustomSet_NilReceiver(t *testing.T) {
	t.Run("New returns nil", runTestCustomSetNilReceiverNew)
	t.Run("Clone returns nil", runTestCustomSetNilReceiverClone)
	t.Run("Clear does not panic", runTestCustomSetNilReceiverClear)
	t.Run("Purge returns nil", runTestCustomSetNilReceiverPurge)
	t.Run("Reserve returns false", runTestCustomSetNilReceiverReserve)
	t.Run("Grow returns false", runTestCustomSetNilReceiverGrow)
	t.Run("Trim returns false", runTestCustomSetNilReceiverTrim)
}

func runTestCustomSetNilReceiverNew(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	var result slices.Set[int]
	core.AssertMustNoPanic(t, func() {
		result = s.New()
	}, "New on nil receiver")

	core.AssertNil(t, result, "New on nil receiver")
}

func runTestCustomSetNilReceiverClone(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Verify nil receiver returns nil without panicking
	result := s.Clone()
	core.AssertNil(t, result, "Clone on nil receiver")

	// Verify cloning nil doesn't create a new set
	core.AssertNil(t, s, "receiver remains nil")
	// Clone of nil should be nil, not an empty set
	if result != nil {
		t.Error("Clone of nil should return nil, not empty set")
	}
}

func runTestCustomSetNilReceiverClear(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Clear on nil should be a no-op, not panic
	s.Clear()

	// Verify receiver is still nil (no allocation occurred)
	core.AssertNil(t, s, "receiver remains nil after Clear")

	// Verify subsequent operations still handle nil correctly
	core.AssertEqual(t, 0, s.Len(), "nil set has zero length")
	core.AssertFalse(t, s.Contains(42), "nil set contains nothing")
}

func runTestCustomSetNilReceiverPurge(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Purge on nil should return nil
	result := s.Purge()
	core.AssertNil(t, result, "Purge on nil receiver")

	// Verify no side effects
	core.AssertNil(t, s, "receiver remains nil")
	// Purge returns removed elements; nil set has nothing to remove
	if result != nil {
		t.Error("Purge of nil should return nil slice")
	}
}

func runTestCustomSetNilReceiverReserve(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Reserve on nil should return false (cannot reserve)
	result := s.Reserve(10)
	core.AssertFalse(t, result, "Reserve on nil receiver returns false")

	// Verify no allocation occurred
	core.AssertNil(t, s, "receiver remains nil")

	// Verify capacity-related operations are consistent
	available, total := s.Cap()
	core.AssertEqual(t, 0, available, "nil set has zero available capacity")
	core.AssertEqual(t, 0, total, "nil set has zero total capacity")
	core.AssertFalse(t, s.Grow(5), "Grow also returns false on nil")
}

func runTestCustomSetNilReceiverGrow(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Grow on nil should return false (cannot grow)
	result := s.Grow(10)
	core.AssertFalse(t, result, "Grow on nil receiver returns false")

	// Verify no allocation occurred
	core.AssertNil(t, s, "receiver remains nil")

	// Verify capacity remains zero
	available, total := s.Cap()
	core.AssertEqual(t, 0, available, "nil set available capacity unchanged")
	core.AssertEqual(t, 0, total, "nil set total capacity unchanged")
	// Growing by 0 should also return false
	core.AssertFalse(t, s.Grow(0), "Grow(0) on nil also returns false")
}

func runTestCustomSetNilReceiverTrim(t *testing.T) {
	t.Helper()
	var s *slices.CustomSet[int]
	// Trim on nil should return false (nothing to trim)
	result := s.Trim()
	core.AssertFalse(t, result, "Trim on nil receiver returns false")

	// Verify no side effects
	core.AssertNil(t, s, "receiver remains nil")

	// Verify Trim is idempotent on nil
	result2 := s.Trim()
	core.AssertFalse(t, result2, "repeated Trim also returns false")
	available, total := s.Cap()
	core.AssertEqual(t, 0, available, "available capacity remains zero")
	core.AssertEqual(t, 0, total, "total capacity remains zero")
}

func runTestInitCustomSetNil(t *testing.T) {
	t.Helper()
	var nilSet *slices.CustomSet[int]
	err := slices.InitCustomSet(nilSet, nil, 1, 2, 3)
	core.AssertError(t, err, "InitCustomSet with nil set")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error type")
}

func runTestInitCustomSetNilCmp(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	err := slices.InitCustomSet(&set, nil, 1, 2, 3)
	core.AssertError(t, err, "InitCustomSet with nil cmp")
	core.AssertContains(t, err.Error(), "comparison function is required", "error message")
}

func runTestForEachNilFunction(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet(1, 2, 3)

	// ForEach with nil function should be a safe no-op
	set.ForEach(nil) // Should not panic

	// Verify set is unchanged
	core.AssertEqual(t, 3, set.Len(), "set length unchanged")
	core.AssertSliceEqual(t, core.S(1, 2, 3), set.Export(), "set values unchanged")

	// Verify set still functions normally after nil ForEach
	count := 0
	set.ForEach(func(_ int) bool {
		count++
		return true
	})
	core.AssertEqual(t, 3, count, "all elements visited")
}

func runTestCloneEmptyWithCapacity(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet[int]()
	set.Reserve(10)
	clone := set.Clone()
	core.AssertNotNil(t, clone, "clone")
	core.AssertEqual(t, 0, clone.Len(), "clone length")
}

func TestCustomSet_EdgeCases(t *testing.T) {
	t.Run("InitCustomSet with nil set", runTestInitCustomSetNil)
	t.Run("InitCustomSet with nil cmp", runTestInitCustomSetNilCmp)
	t.Run("ForEach with nil function", runTestForEachNilFunction)
	t.Run("Clone empty set with capacity", runTestCloneEmptyWithCapacity)
}

// removeLastElementTestCase tests removing the last element scenarios
type removeLastElementTestCase struct {
	name     string
	initial  []int
	remove   int
	expected []int
}

func (tc removeLastElementTestCase) Name() string {
	return tc.name
}

func (tc removeLastElementTestCase) Test(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet(tc.initial...)
	n := set.Remove(tc.remove)
	if len(tc.expected) < len(tc.initial) {
		core.AssertEqual(t, 1, n, "removed count")
	} else {
		core.AssertEqual(t, 0, n, "removed count")
	}
	values := set.Export()
	core.AssertSliceEqual(t, tc.expected, values, "final set")
}

// newRemoveLastElementTestCase creates a new removeLastElementTestCase
func newRemoveLastElementTestCase(name string, initial []int, remove int,
	expected []int) removeLastElementTestCase {
	return removeLastElementTestCase{
		name:     name,
		initial:  initial,
		remove:   remove,
		expected: expected,
	}
}

func TestCustomSet_RemoveLastElement(t *testing.T) {
	tests := []removeLastElementTestCase{
		newRemoveLastElementTestCase(
			"remove only element",
			[]int{5},
			5,
			[]int{},
		),
		newRemoveLastElementTestCase(
			"remove last in sequence",
			[]int{1, 2, 3, 4, 5},
			5,
			[]int{1, 2, 3, 4},
		),
		newRemoveLastElementTestCase(
			"remove first element",
			[]int{1, 2, 3},
			1,
			[]int{2, 3},
		),
		newRemoveLastElementTestCase(
			"remove middle element",
			[]int{1, 3, 5},
			3,
			[]int{1, 5},
		),
	}

	core.RunTestCases(t, tests)
}
