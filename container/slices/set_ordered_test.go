package slices

import (
	"slices"
	"testing"
)

//revive:disable-next-line:cognitive-complexity
func TestSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var s []string
		if got := NewOrderedSet(s...); got.Len() != 0 {
			t.Errorf("Set() = %v, want empty slice", got)
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		got := NewOrderedSet(s...)
		want := []string{"a", "b", "c"}
		if got.Len() != len(want) {
			t.Errorf("Set() = %v, want %v", got, want)
		}
		for i := range want {
			if v, ok := got.GetByIndex(i); !ok || v != want[i] {
				t.Errorf("Set()[%d] = %v, want %v", i, v, want[i])
			}
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		s := []string{"a", "b", "a", "c", "b", "c"}
		got := NewOrderedSet(s...)
		want := []string{"a", "b", "c"}
		if got.Len() != len(want) {
			t.Errorf("Set() = %v, want %v", got, want)
		}
		for i := range want {
			if v, ok := got.GetByIndex(i); !ok || v != want[i] {
				t.Errorf("Set()[%d] = %v, want %v", i, v, want[i])
			}
		}
	})
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		add      []int
		expected int
		result   []int
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
			set := NewOrderedSet[int]()
			if tt.initial != nil {
				set.Add(tt.initial...)
			}

			count := set.Add(tt.add...)
			if count != tt.expected {
				t.Errorf("Add() returned %v, want %v", count, tt.expected)
			}

			if got := set.Export(); !slices.Equal(got, tt.result) {
				t.Errorf("After Add() got %v, want %v", got, tt.result)
			}
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
			set := NewOrderedSet[int]()
			set.Add(tt.initial...)

			count := set.Remove(tt.remove...)

			if count != tt.count {
				t.Errorf("Remove() count = %v, want %v", count, tt.count)
			}

			if got := set.Export(); !slices.Equal(got, tt.expected) {
				t.Errorf("After Remove() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

//revive:disable-next-line:cognitive-complexity
func TestSortedSet_TrimN(t *testing.T) {
	tests := []struct {
		name        string
		set         *CustomSet[int]
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.TrimN(tt.minCapacity); got != tt.want {
				t.Errorf("SortedSet.TrimN() = %v, want %v", got, tt.want)
			}
			if tt.set != nil {
				if _, c := tt.set.Cap(); c < tt.minCapacity {
					t.Errorf("SortedSet.Cap() = %v, want >= %v", c, tt.minCapacity)
				}
			}
		})
	}
}
