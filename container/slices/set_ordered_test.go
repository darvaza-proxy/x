package slices_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/slices"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = sortedSetAddTestCase{}
var _ core.TestCase = sortedSetRemoveTestCase{}
var _ core.TestCase = sortedSetTrimNTestCase{}

//revive:disable-next-line:cognitive-complexity
func TestSet(t *testing.T) {
	t.Run("empty", runTestSetEmpty)
	t.Run("no duplicates", runTestSetNoDuplicates)
	t.Run("with duplicates", runTestSetWithDuplicates)
}

func runTestSetEmpty(t *testing.T) {
	t.Helper()
	var s []string
	got := slices.NewOrderedSet(s...)
	core.AssertEqual(t, 0, got.Len(), "empty set length")
}

func runTestSetNoDuplicates(t *testing.T) {
	t.Helper()
	s := []string{"a", "b", "c"}
	got := slices.NewOrderedSet(s...)
	want := []string{"a", "b", "c"}
	core.AssertEqual(t, len(want), got.Len(), "set length")
	for i := range want {
		v, ok := got.GetByIndex(i)
		core.AssertEqual(t, true, ok, "GetByIndex should succeed")
		core.AssertEqual(t, want[i], v, "value at index %d", i)
	}
}

func runTestSetWithDuplicates(t *testing.T) {
	t.Helper()
	s := []string{"a", "b", "a", "c", "b", "c"}
	got := slices.NewOrderedSet(s...)
	want := []string{"a", "b", "c"}
	core.AssertEqual(t, len(want), got.Len(), "set length")
	for i := range want {
		v, ok := got.GetByIndex(i)
		core.AssertEqual(t, true, ok, "GetByIndex should succeed")
		core.AssertEqual(t, want[i], v, "value at index %d", i)
	}
}

type sortedSetAddTestCase struct {
	name     string
	initial  []int
	add      []int
	result   []int
	expected int
}

func (tc sortedSetAddTestCase) Name() string {
	return tc.name
}

func (tc sortedSetAddTestCase) Test(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet(tc.initial...)
	n := set.Add(tc.add...)
	core.AssertEqual(t, tc.expected, n, "added count")

	values := set.Export()
	core.AssertSliceEqual(t, tc.result, values, "final set")
}

// newSortedSetAddTestCase creates a new sortedSetAddTestCase
func newSortedSetAddTestCase(name string, initial []int, add []int,
	expected int, result []int) sortedSetAddTestCase {
	return sortedSetAddTestCase{
		name:     name,
		initial:  initial,
		add:      add,
		expected: expected,
		result:   result,
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Add(t *testing.T) {
	tests := []sortedSetAddTestCase{
		newSortedSetAddTestCase(
			"add to empty set",
			nil,
			[]int{1, 2, 3},
			3,
			[]int{1, 2, 3},
		),
		newSortedSetAddTestCase(
			"add with duplicates",
			[]int{1, 2},
			[]int{2, 3, 3, 4},
			2,
			[]int{1, 2, 3, 4},
		),
		newSortedSetAddTestCase(
			"add nothing",
			[]int{1, 2, 3},
			[]int{},
			0,
			[]int{1, 2, 3},
		),
		newSortedSetAddTestCase(
			"add to nil set",
			nil,
			[]int{1, 2},
			2,
			[]int{1, 2},
		),
		newSortedSetAddTestCase(
			"add unsorted values",
			[]int{2, 4},
			[]int{5, 1, 3},
			3,
			[]int{1, 2, 3, 4, 5},
		),
	}

	core.RunTestCases(t, tests)
}

type sortedSetRemoveTestCase struct {
	name     string
	initial  []int
	remove   []int
	expected []int
	count    int
}

func (tc sortedSetRemoveTestCase) Name() string {
	return tc.name
}

func (tc sortedSetRemoveTestCase) Test(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet(tc.initial...)
	n := set.Remove(tc.remove...)
	core.AssertEqual(t, tc.count, n, "removed count")

	values := set.Export()
	core.AssertSliceEqual(t, tc.expected, values, "final set")
}

// newSortedSetRemoveTestCase creates a new sortedSetRemoveTestCase
func newSortedSetRemoveTestCase(name string, initial []int, remove []int,
	count int, expected []int) sortedSetRemoveTestCase {
	return sortedSetRemoveTestCase{
		name:     name,
		initial:  initial,
		remove:   remove,
		count:    count,
		expected: expected,
	}
}

// Compile-time verification
var _ core.TestCase = sortedSetRemoveTestCase{}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Remove(t *testing.T) {
	tests := []sortedSetRemoveTestCase{
		newSortedSetRemoveTestCase(
			"remove_from_empty",
			[]int{},
			[]int{1, 2, 3},
			0,
			[]int{},
		),
		newSortedSetRemoveTestCase(
			"remove_non_existing",
			[]int{1, 3, 5},
			[]int{2, 4},
			0,
			[]int{1, 3, 5},
		),
		newSortedSetRemoveTestCase(
			"remove_duplicates",
			[]int{1, 2, 3, 4, 5},
			[]int{2, 2, 4, 4},
			2,
			[]int{1, 3, 5},
		),
		newSortedSetRemoveTestCase(
			"remove_all",
			[]int{1, 2, 3},
			[]int{1, 2, 3},
			3,
			[]int{},
		),
		newSortedSetRemoveTestCase(
			"remove_partial",
			[]int{1, 2, 3, 4, 5},
			[]int{2, 4},
			2,
			[]int{1, 3, 5},
		),
	}

	core.RunTestCases(t, tests)
}

type sortedSetTrimNTestCase struct {
	set         *slices.CustomSet[int]
	name        string
	minCapacity int
	want        bool
}

func (tc sortedSetTrimNTestCase) Name() string {
	return tc.name
}

func (tc sortedSetTrimNTestCase) Test(t *testing.T) {
	t.Helper()
	got := tc.set.TrimN(tc.minCapacity)
	core.AssertEqual(t, tc.want, got, "TrimN result")
	if tc.set != nil {
		_, c := tc.set.Cap()
		core.AssertTrue(t, c >= tc.minCapacity, "capacity >= minCapacity")
	}
}

// newSortedSetTrimNTestCase creates a new sortedSetTrimNTestCase
//
//revive:disable-next-line:flag-parameter
func newSortedSetTrimNTestCase(name string, set *slices.CustomSet[int],
	minCapacity int, want bool) sortedSetTrimNTestCase {
	return sortedSetTrimNTestCase{
		name:        name,
		set:         set,
		minCapacity: minCapacity,
		want:        want,
	}
}

// Compile-time verification
var _ core.TestCase = sortedSetTrimNTestCase{}

func TestSortedSet_TrimN(t *testing.T) {
	tests := []sortedSetTrimNTestCase{
		newSortedSetTrimNTestCase(
			"nil set",
			nil,
			0,
			false,
		),
		newSortedSetTrimNTestCase(
			"empty set",
			slices.NewOrderedSet[int](),
			0,
			false,
		),
		newSortedSetTrimNTestCase(
			"set with elements above min capacity",
			slices.NewOrderedSet(1, 2, 3, 4, 5),
			3,
			false,
		),
		newSortedSetTrimNTestCase(
			"set with elements below min capacity",
			slices.NewOrderedSet(1, 2),
			5,
			true,
		),
		newSortedSetTrimNTestCase(
			"negative min capacity",
			slices.NewOrderedSet(1, 2, 3),
			-1,
			false,
		),
	}

	core.RunTestCases(t, tests)
}
