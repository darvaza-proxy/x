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

// InitOrderedSet initializes a CustomSet with a default comparison function for ordered types.
// It sets the comparison function and populates the set with the initial elements.
// Returns an error if the initialization fails.
func InitOrderedSet[T core.Ordered](set *CustomSet[T], initial ...T) error {
	return InitCustomSet(set, cmpOrdered, initial...)
}
