package slices

import (
	"testing"

	"darvaza.org/core"
)

//revive:disable-next-line:cognitive-complexity
func TestSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var s []string
		got := NewOrderedSet(s...)
		core.AssertEqual(t, 0, got.Len(), "empty set length")
	})

	t.Run("no duplicates", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		got := NewOrderedSet(s...)
		want := []string{"a", "b", "c"}
		core.AssertEqual(t, len(want), got.Len(), "set length")
		for i := range want {
			v, ok := got.GetByIndex(i)
			core.AssertEqual(t, true, ok, "GetByIndex should succeed")
			core.AssertEqual(t, want[i], v, "value at index %d", i)
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		s := []string{"a", "b", "a", "c", "b", "c"}
		got := NewOrderedSet(s...)
		want := []string{"a", "b", "c"}
		core.AssertEqual(t, len(want), got.Len(), "set length")
		for i := range want {
			v, ok := got.GetByIndex(i)
			core.AssertEqual(t, true, ok, "GetByIndex should succeed")
			core.AssertEqual(t, want[i], v, "value at index %d", i)
		}
	})
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
	set := NewOrderedSet(tc.initial...)
	n := set.Add(tc.add...)
	core.AssertEqual(t, tc.expected, n, "added count")

	values := set.Export()
	core.AssertSliceEqual(t, tc.result, values, "final set")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = sortedSetAddTestCase{}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Add(t *testing.T) {
	tests := []sortedSetAddTestCase{
		{
			name:     "add to empty set",
			initial:  nil,
			add:      []int{1, 2, 3},
			expected: 3,
			result:   []int{1, 2, 3},
		},
		{
			name:     "add with duplicates",
			initial:  []int{1, 2},
			add:      []int{2, 3, 3, 4},
			expected: 2,
			result:   []int{1, 2, 3, 4},
		},
		{
			name:     "add nothing",
			initial:  []int{1, 2, 3},
			add:      []int{},
			expected: 0,
			result:   []int{1, 2, 3},
		},
		{
			name:     "add to nil set",
			initial:  nil,
			add:      []int{1, 2},
			expected: 2,
			result:   []int{1, 2},
		},
		{
			name:     "add unsorted values",
			initial:  []int{2, 4},
			add:      []int{5, 1, 3},
			expected: 3,
			result:   []int{1, 2, 3, 4, 5},
		},
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
	set := NewOrderedSet(tc.initial...)
	n := set.Remove(tc.remove...)
	core.AssertEqual(t, tc.count, n, "removed count")

	values := set.Export()
	core.AssertSliceEqual(t, tc.expected, values, "final set")
}

// Compile-time verification
var _ core.TestCase = sortedSetRemoveTestCase{}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Remove(t *testing.T) {
	tests := []sortedSetRemoveTestCase{
		{
			name:     "remove_from_empty",
			initial:  []int{},
			remove:   []int{1, 2, 3},
			expected: []int{},
			count:    0,
		},
		{
			name:     "remove_non_existing",
			initial:  []int{1, 3, 5},
			remove:   []int{2, 4},
			expected: []int{1, 3, 5},
			count:    0,
		},
		{
			name:     "remove_duplicates",
			initial:  []int{1, 2, 3, 4, 5},
			remove:   []int{2, 2, 4, 4},
			expected: []int{1, 3, 5},
			count:    2,
		},
		{
			name:     "remove_all",
			initial:  []int{1, 2, 3},
			remove:   []int{1, 2, 3},
			expected: []int{},
			count:    3,
		},
		{
			name:     "remove_partial",
			initial:  []int{1, 2, 3, 4, 5},
			remove:   []int{2, 4},
			expected: []int{1, 3, 5},
			count:    2,
		},
	}

	core.RunTestCases(t, tests)
}

type sortedSetTrimNTestCase struct {
	set         *CustomSet[int]
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

// Compile-time verification
var _ core.TestCase = sortedSetTrimNTestCase{}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_TrimN(t *testing.T) {
	tests := []sortedSetTrimNTestCase{
		{
			name:        "nil set",
			set:         nil,
			minCapacity: 0,
			want:        false,
		},
		{
			name:        "empty set",
			set:         NewOrderedSet[int](),
			minCapacity: 0,
			want:        false,
		},
		{
			name:        "set with elements above min capacity",
			set:         NewOrderedSet(1, 2, 3, 4, 5),
			minCapacity: 3,
			want:        false,
		},
		{
			name:        "set with elements below min capacity",
			set:         NewOrderedSet(1, 2),
			minCapacity: 5,
			want:        true,
		},
		{
			name:        "negative min capacity",
			set:         NewOrderedSet(1, 2, 3),
			minCapacity: -1,
			want:        false,
		},
	}

	core.RunTestCases(t, tests)
}
