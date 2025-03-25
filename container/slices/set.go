package slices

// Set is a generic interface for a mutable set of elements with dynamic capacity management.
// It provides methods for adding, removing, checking, and manipulating a collection of unique elements.
// The type parameter T allows the set to work with any type that supports comparison.
type Set[T any] interface {
	// Contains reports whether the set contains the given value v.
	Contains(v T) bool
	// Len returns the number of elements currently in the set.
	Len() int
	// Cap returns the available and current capacity of the underlying slice.
	Cap() (available, total int)

	// Add adds one or more elements to the set and returns the number of elements successfully added.
	// Elements that already exist in the set are not added and do not contribute to the return value.
	Add(...T) int
	// Remove removes one or more elements from the set and returns the number of elements successfully removed.
	// Elements that do not exist in the set are ignored and do not contribute to the return value.
	Remove(...T) int

	// Clear removes all elements from the set, leaving it empty.
	Clear()
	// Export returns a copy of the elements currently in the set as a new slice.
	// The order of elements in the returned slice is not necessarily the same as the order in which
	// they were added.
	Export() []T
	// ForEach applies the provided function to each element in the set, stopping if the function returns false.
	// The function is passed the element as an argument and should return true to continue iterating, or false to stop.
	ForEach(func(T) bool)

	// Reserve attempts to increase the underlying slice's capacity to accommodate the
	// specified number of additional elements.
	// It returns true if the capacity was increased, or otherwise false.
	Reserve(capacity int) bool
	// Grow increases the capacity of the underlying slice to accommodate the specified number of additional elements.
	// It returns true if the capacity was increased, or otherwise false.
	Grow(additional int) bool
	// Trim reduces the capacity of the underlying slice to match its length, potentially freeing unused memory.
	// It returns true if the capacity was effectively reduced, or false otherwise.
	Trim() bool
	// TrimN reduces the capacity of the underlying slice to match its length, optionally ensuring a minimum capacity.
	// It returns true if the capacity was effectively reduced, or false otherwise.
	TrimN(minimumCapacity int) bool
}
