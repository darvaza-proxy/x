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

func TestCustomSet_New(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")

	other := set.New()
	core.AssertMustNotNil(t, other, "New")
	core.AssertEqual(t, 0, other.Len(), "len")
}

func TestCustomSet_InitCustomSet(t *testing.T) {
	var set slices.CustomSet[int]

	core.AssertMustNoError(t, slices.InitCustomSet(&set, cmpInt, 1, 2, 3), "InitCustomSet")
	core.AssertEqual(t, 3, set.Len(), "len")

	// Test init on already initialized set
	core.AssertError(t, slices.InitCustomSet(&set, cmpInt), "re-init")
}

var _ core.TestCase = customSetContainsTestCase{}

// customSetContainsTestCase checks CustomSet.Contains against a shared set.
type customSetContainsTestCase struct {
	set   *slices.CustomSet[int]
	name  string
	value int
	want  bool
}

func newCustomSetContainsTestCase(name string, s *slices.CustomSet[int],
	value int, want bool) customSetContainsTestCase {
	return customSetContainsTestCase{
		set:   s,
		name:  name,
		value: value,
		want:  want,
	}
}

func (tc customSetContainsTestCase) Name() string {
	return tc.name
}

func (tc customSetContainsTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.set.Contains(tc.value), "Contains(%d)", tc.value)
}

func customSetContainsTestCases() []customSetContainsTestCase {
	s := slices.MustCustomSet(cmpInt, 1, 3, 5, 7, 9)
	return []customSetContainsTestCase{
		newCustomSetContainsTestCase("present 1", s, 1, true),
		newCustomSetContainsTestCase("present 3", s, 3, true),
		newCustomSetContainsTestCase("present 5", s, 5, true),
		newCustomSetContainsTestCase("present 7", s, 7, true),
		newCustomSetContainsTestCase("present 9", s, 9, true),
		newCustomSetContainsTestCase("absent 2", s, 2, false),
		newCustomSetContainsTestCase("absent 4", s, 4, false),
		newCustomSetContainsTestCase("absent 10", s, 10, false),
	}
}

func TestCustomSet_Contains(t *testing.T) {
	core.RunTestCases(t, customSetContainsTestCases())
}

func TestCustomSet_Clone(t *testing.T) {
	original, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")

	clone := original.Clone()
	core.AssertMustNotNil(t, clone, "Clone")

	// Verify same elements
	core.AssertEqual(t, original.Len(), clone.Len(), "len")
	for i := 0; i < original.Len(); i++ {
		origVal, _ := original.GetByIndex(i)
		cloneVal, _ := clone.(*slices.CustomSet[int]).GetByIndex(i)
		core.AssertEqual(t, origVal, cloneVal, "element %d", i)
	}

	// Verify independence
	original.Add(4)
	core.AssertFalse(t, clone.Contains(4), "clone independence")
}

func TestCustomSet_Clear(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3, 4, 5)
	core.AssertMustNoError(t, err, "NewCustomSet")
	core.AssertMustEqual(t, 5, set.Len(), "initial len")

	set.Clear()
	core.AssertEqual(t, 0, set.Len(), "len after clear")

	// Verify capacity remains
	_, total := set.Cap()
	core.AssertNotEqual(t, 0, total, "capacity preserved")
}

func TestCustomSet_Purge(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 5, 3, 1, 4, 2)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Returned elements should be sorted
	core.AssertSliceEqual(t, core.S(1, 2, 3, 4, 5), set.Purge(), "purged")

	// Verify set is empty with no capacity
	core.AssertEqual(t, 0, set.Len(), "len after purge")
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
	core.AssertEqual(t, 15, sum, "sum")

	// Test early termination
	var count int
	set.ForEach(func(_ int) bool {
		count++
		return count < 3
	})
	core.AssertEqual(t, 3, count, "early termination count")

	// A nil callback is a no-op.
	core.AssertNoPanic(t, func() { set.ForEach(nil) }, "nil callback")
}

func TestCustomSet_Reserve(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")

	core.AssertTrue(t, set.Reserve(10), "Reserve(10)")

	available, _ := set.Cap()
	core.AssertTrue(t, available >= 10, "available capacity")

	// Reserve less than current should return false
	core.AssertFalse(t, set.Reserve(5), "Reserve(5) when already >= 10")
}

func TestCustomSet_Grow(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")

	initialAvailable, _ := set.Cap()

	core.AssertTrue(t, set.Grow(5), "Grow(5)")

	newAvailable, _ := set.Cap()
	core.AssertTrue(t, newAvailable >= initialAvailable+5, "available capacity")
}

func TestCustomSet_Trim(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Add elements and then grow capacity
	set.Add(1, 2, 3)
	set.Grow(20)

	available, total := set.Cap()
	core.AssertMustTrue(t, available > 0, "available after grow")
	core.AssertMustTrue(t, total > set.Len(), "total after grow")

	// Trim excess capacity
	core.AssertTrue(t, set.Trim(), "Trim()")

	newAvailable, newTotal := set.Cap()
	core.AssertEqual(t, set.Len(), newTotal, "total after trim")
	core.AssertEqual(t, 0, newAvailable, "available after trim")
}

func TestInitOrderedSet(t *testing.T) {
	var set slices.CustomSet[int]

	core.AssertMustNoError(t, slices.InitOrderedSet(&set, 3, 1, 4, 1, 5), "InitOrderedSet")

	// Should have deduplicated and sorted
	core.AssertSliceEqual(t, core.S(1, 3, 4, 5), set.Export(), "values")
}

func TestCustomSet_GetByIndex_Error(t *testing.T) {
	set, err := slices.NewCustomSet(cmpInt, 1, 2, 3)
	core.AssertMustNoError(t, err, "NewCustomSet")

	// Test negative index
	_, ok := set.GetByIndex(-1)
	core.AssertFalse(t, ok, "GetByIndex(-1)")

	// Test out of bounds
	_, ok = set.GetByIndex(3)
	core.AssertFalse(t, ok, "GetByIndex(3) of 3 elements")
}

func TestCustomSet_Nil(t *testing.T) {
	var set *slices.CustomSet[int]

	// Test nil receiver methods
	core.AssertFalse(t, set.Contains(1), "nil Contains")
	core.AssertEqual(t, 0, set.Len(), "nil Len")

	available, total := set.Cap()
	core.AssertEqual(t, 0, available, "nil Cap available")
	core.AssertEqual(t, 0, total, "nil Cap total")

	core.AssertEqual(t, 0, set.Add(1), "nil Add")
	core.AssertEqual(t, 0, set.Remove(1), "nil Remove")
	core.AssertEqual(t, 0, len(set.Export()), "nil Export")

	core.AssertNil(t, set.New(), "nil New")
	core.AssertNil(t, set.Clone(), "nil Clone")
	core.AssertEqual(t, 0, len(set.Purge()), "nil Purge")
	core.AssertFalse(t, set.Reserve(10), "nil Reserve")
	core.AssertFalse(t, set.Grow(10), "nil Grow")
	core.AssertFalse(t, set.Trim(), "nil Trim")

	// Clear and ForEach are no-ops on a nil receiver.
	core.AssertNoPanic(t, func() {
		set.Clear()
		set.ForEach(func(int) bool { return true })
	}, "nil Clear/ForEach")
}

func TestMustCustomSet_Panic(t *testing.T) {
	core.AssertPanic(t, func() {
		_ = slices.MustCustomSet[int](nil)
	}, core.ErrInvalid, "MustCustomSet(nil)")
}

func testInitNilReceiver(t *testing.T) {
	var set *slices.CustomSet[int]
	core.AssertErrorIs(t, slices.InitCustomSet(set, cmpInt), core.ErrNilReceiver, "nil receiver")
}

func testInitNilComparison(t *testing.T) {
	var set slices.CustomSet[int]
	core.AssertErrorIs(t, slices.InitCustomSet(&set, nil), core.ErrInvalid, "nil comparison")
}

func TestCustomSet_InitErrors(t *testing.T) {
	t.Run("nil receiver", testInitNilReceiver)
	t.Run("nil comparison", testInitNilComparison)
}

var _ core.TestCase = uninitialisedTestCase{}

// uninitialisedTestCase checks that an operation on a zero-value CustomSet
// (nil comparison, non-nil receiver) panics, covering the not-initialised
// arms reached when NewCustomSet/InitCustomSet are bypassed.
type uninitialisedTestCase struct {
	op   func(*slices.CustomSet[int])
	name string
}

func newUninitialisedTestCase(name string,
	op func(*slices.CustomSet[int])) uninitialisedTestCase {
	return uninitialisedTestCase{
		op:   op,
		name: name,
	}
}

func (tc uninitialisedTestCase) Name() string {
	return tc.name
}

func (tc uninitialisedTestCase) Test(t *testing.T) {
	t.Helper()
	var set slices.CustomSet[int]
	core.AssertPanic(t, func() { tc.op(&set) }, nil, tc.name)
}

func uninitialisedTestCases() []uninitialisedTestCase {
	return []uninitialisedTestCase{
		newUninitialisedTestCase("New", func(s *slices.CustomSet[int]) { _ = s.New() }),
		newUninitialisedTestCase("Clone", func(s *slices.CustomSet[int]) { _ = s.Clone() }),
		newUninitialisedTestCase("Add", func(s *slices.CustomSet[int]) { s.Add(1) }),
		newUninitialisedTestCase("Remove", func(s *slices.CustomSet[int]) { s.Remove(1) }),
	}
}

func TestCustomSet_Uninitialised(t *testing.T) {
	core.RunTestCases(t, uninitialisedTestCases())
}

func TestCustomSet_ContainsEmpty(t *testing.T) {
	set := slices.MustCustomSet(cmpInt)
	core.AssertFalse(t, set.Contains(1), "empty Contains")
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
	core.AssertEqual(t, 3, set.Len(), "len")

	// Should be sorted by ID
	first, _ := set.GetByIndex(0)
	core.AssertEqual(t, 1, first.ID, "first ID")
}
