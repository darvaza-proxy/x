package slices

import "darvaza.org/core"

// NewOrderedSet creates a new CustomSet with a default comparison function for ordered types.
func NewOrderedSet[T core.Ordered](initial ...T) *CustomSet[T] {
	return MustCustomSet(cmpOrdered, initial...)
}

func cmpOrdered[T core.Ordered](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
