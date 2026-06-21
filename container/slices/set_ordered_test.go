package slices_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/slices"
)

//revive:disable-next-line:cognitive-complexity
func TestSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var s []string
		core.AssertEqual(t, 0, slices.NewOrderedSet(s...).Len(), "len")
	})

	t.Run("no duplicates", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		got := slices.NewOrderedSet(s...)
		core.AssertSliceEqual(t, s, got.Export(), "values")
	})

	t.Run("with duplicates", func(t *testing.T) {
		s := []string{"a", "b", "a", "c", "b", "c"}
		got := slices.NewOrderedSet(s...)
		want := []string{"a", "b", "c"}
		core.AssertSliceEqual(t, want, got.Export(), "values")
	})
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		add      []int
		result   []int
		expected int
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := slices.NewOrderedSet[int]()
			if tt.initial != nil {
				set.Add(tt.initial...)
			}

			core.AssertEqual(t, tt.expected, set.Add(tt.add...), "Add count")
			core.AssertSliceEqual(t, tt.result, set.Export(), "result")
		})
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Remove(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		remove   []int
		expected []int
		count    int
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := slices.NewOrderedSet[int]()
			set.Add(tt.initial...)

			core.AssertEqual(t, tt.count, set.Remove(tt.remove...), "Remove count")
			core.AssertSliceEqual(t, tt.expected, set.Export(), "result")
		})
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_TrimN(t *testing.T) {
	tests := []struct {
		set         *slices.CustomSet[int]
		name        string
		minCapacity int
		want        bool
	}{
		{
			name:        "nil set",
			set:         nil,
			minCapacity: 0,
			want:        false,
		},
		{
			name:        "empty set",
			set:         slices.NewOrderedSet[int](),
			minCapacity: 0,
			want:        false,
		},
		{
			name:        "set with elements above min capacity",
			set:         slices.NewOrderedSet(1, 2, 3, 4, 5),
			minCapacity: 3,
			want:        false,
		},
		{
			name:        "set with elements below min capacity",
			set:         slices.NewOrderedSet(1, 2),
			minCapacity: 5,
			want:        true,
		},
		{
			name:        "negative min capacity",
			set:         slices.NewOrderedSet(1, 2, 3),
			minCapacity: -1,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core.AssertEqual(t, tt.want, tt.set.TrimN(tt.minCapacity), "TrimN")
			if tt.set != nil {
				_, c := tt.set.Cap()
				core.AssertTrue(t, c >= tt.minCapacity, "capacity")
			}
		})
	}
}
