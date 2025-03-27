package slices

import (
	"sort"
	"sync"

	"darvaza.org/core"
)

// Compile-time check to ensure CustomSet implements the Set interface
var _ Set[struct{}] = (*CustomSet[struct{}])(nil)

// CustomSet is a generic set implementation with custom comparison and concurrency-safe operations.
// The set maintains a sorted slice of unique elements using a provided comparison function.
type CustomSet[T any] struct {
	mu  sync.RWMutex
	s   []T
	cmp func(T, T) int
}

// New creates a new CustomSet with the same comparison function as the current set.
// Returns nil if the current set is nil or has no comparison function.
func (set *CustomSet[T]) New() Set[T] {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	if set.cmp == nil {
		return nil
	}

	return &CustomSet[T]{
		cmp: set.cmp,
	}
}

// NewCustomSet creates a new CustomSet with the provided comparison function and initial elements.
// If the comparison function is nil, it returns an error.
func NewCustomSet[T any](cmp func(T, T) int, initial ...T) (*CustomSet[T], error) {
	if cmp == nil {
		return nil, core.Wrap(core.ErrInvalid, "comparison function is required")
	}

	return &CustomSet[T]{
		s:   dedupeFn(cmp, initial),
		cmp: cmp,
	}, nil
}

// MustCustomSet creates a new CustomSet with the provided comparison function and initial elements.
// It panics if the comparison function is nil or if an error occurs during set creation.
// This is a convenience function that simplifies set initialization when panicking on error is acceptable.
func MustCustomSet[T any](cmd func(T, T) int, initial ...T) *CustomSet[T] {
	set, err := NewCustomSet(cmd, initial...)
	if err != nil {
		core.Panic(err)
	}
	return set
}

// InitCustomSet initializes a pre-allocated CustomSet with a comparison function and optional initial elements.
// It performs thread-safe initialization, ensuring the set can only be initialized once.
// Returns an error if the set is nil, the comparison function is nil, or the set has already been initialized.
func InitCustomSet[T any](set *CustomSet[T], cmp func(T, T) int, initial ...T) error {
	switch {
	case set == nil:
		return core.ErrNilReceiver
	case cmp == nil:
		return core.Wrap(core.ErrInvalid, "comparison function is required")
	default:
		set.mu.Lock()
		defer set.mu.Unlock()

		if set.cmp != nil || set.s != nil {
			return core.Wrap(core.ErrInvalid, "set already initialized")
		}

		set.s = dedupeFn(cmp, initial)
		set.cmp = cmp
		return nil
	}
}

// dedupeFn removes duplicate elements from a slice using the provided comparison function.
// It first sorts the slice and then eliminates consecutive duplicate elements in-place.
// The function modifies the input slice and returns a slice with unique elements.
func dedupeFn[T any](cmp func(T, T) int, s []T) []T {
	var zero T

	if len(s) < 2 {
		return s
	}

	// sort all elements at once
	sort.Slice(s, func(i, j int) bool {
		return cmp(s[i], s[j]) < 0
	})

	// remove duplicates
	j := 0
	for i := 1; i < len(s); i++ {
		if cmp(s[j], s[i]) == 0 {
			continue
		}
		j++
		s[j] = s[i]
	}

	// zero out remaining elements
	for i := j + 1; i < len(s); i++ {
		s[i] = zero
	}

	// and resize the slice to remove the zeroed elements
	return s[:j+1]
}

// Contains checks if the given value is present in the CustomSet.
// It returns true if the value is found, false otherwise.
// If the set is nil or its comparison function is not set, it returns false.
// The method is concurrency-safe, using a read lock to protect access to the underlying slice.
func (set *CustomSet[T]) Contains(v T) bool {
	if set == nil || set.cmp == nil {
		return false
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	_, ok := set.search(0, v)
	return ok
}

func (set *CustomSet[T]) search(start int, v T) (int, bool) {
	view := set.s[start:]

	if l := len(view); l > 0 {
		i := sort.Search(l, func(i int) bool {
			return set.cmp(view[i], v) >= 0
		})
		return start + i, i < l && set.cmp(view[i], v) == 0
	}

	return start, false
}

// Clone creates a copy of the CustomSet.
// If the set is nil or has no comparison function, it returns nil.
// The method is concurrency-safe, using a read lock to protect access to the underlying slice.
// The returned set has a new slice with the same elements and comparison function as the original set.
func (set *CustomSet[T]) Clone() Set[T] {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	if set.cmp == nil {
		return nil
	}

	s := make([]T, len(set.s))
	copy(s, set.s)

	return &CustomSet[T]{
		s:   s,
		cmp: set.cmp,
	}
}

// Len returns the number of elements in the CustomSet.
// If the set is nil, it returns 0.
// The method is concurrency-safe, using a read lock to protect access to the underlying slice.
func (set *CustomSet[T]) Len() int {
	if set == nil {
		return 0
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	return len(set.s)
}

// Cap returns the available and total capacity of the CustomSet.
// If the set is nil, it returns (0, 0).
// The method is concurrency-safe, using a read lock to protect access to the underlying slice.
// The first return value is the number of additional elements that can be added without reallocation.
// The second return value is the total capacity of the underlying slice.
func (set *CustomSet[T]) Cap() (available, total int) {
	if set == nil {
		return 0, 0
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	return cap(set.s) - len(set.s), cap(set.s)
}

// Add adds the given values to the CustomSet, returning the number of unique values added.
// If the set is nil, has no comparison function, or no values are provided, it returns 0.
// The method is concurrency-safe, using a lock to protect modifications to the underlying slice.
func (set *CustomSet[T]) Add(values ...T) int {
	if set == nil || set.cmp == nil || len(values) == 0 {
		return 0
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doAdd(values)
}

func (set *CustomSet[T]) doAdd(values []T) int {
	var count int

	start := 0
	for _, v := range dedupeFn(set.cmp, values) {
		var added bool
		start, added = set.doAddOne(start, v)
		if added {
			count++
		}
	}

	return count
}

func (set *CustomSet[T]) doAddOne(start int, v T) (int, bool) {
	switch {
	case len(set.s) == 0 && start == 0:
		// first ever
		set.s = []T{v}
		return 0, true
	default:
		var zero T

		i, exists := set.search(start, v)
		if !exists {
			// insert
			set.s = append(set.s, zero)
			copy(set.s[i+1:], set.s[i:])
			set.s[i] = v
		}
		return i, !exists
	}
}

// Remove removes the given values from the CustomSet, returning the number of unique values removed.
// If the set is nil, has no comparison function, no values are provided, or the set is empty, it returns 0.
// The method is concurrency-safe, using a lock to protect modifications to the underlying slice.
func (set *CustomSet[T]) Remove(values ...T) int {
	if set == nil || set.cmp == nil || len(values) == 0 {
		return 0
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if len(set.s) == 0 {
		return 0
	}

	return set.doRemove(values)
}

func (set *CustomSet[T]) doRemove(values []T) int {
	var count int

	start := 0
	for _, v := range dedupeFn(set.cmp, values) {
		var removed bool
		start, removed = set.doRemoveOne(start, v)
		if removed {
			count++
		}
	}

	return count
}

func (set *CustomSet[T]) doRemoveOne(start int, v T) (int, bool) {
	switch {
	case len(set.s) == 0:
		// empty, nothing to remove.
		return 0, false
	default:
		var zero T

		i, exists := set.search(start, v)
		if exists {
			last := len(set.s) - 1
			// remove
			copy(set.s[i:], set.s[i+1:])
			set.s[last] = zero
			set.s = set.s[:last]
		}

		return i, exists
	}
}

// Clear removes all elements from the set, resetting it to an empty state.
// It zeroes out the underlying slice and truncates it to zero length.
// If the set is nil, the method does nothing.
func (set *CustomSet[T]) Clear() {
	if set == nil {
		return
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	for i := range set.s {
		var zero T
		set.s[i] = zero
	}
	set.s = set.s[:0]
}

// Purge removes and returns all elements from the set, resetting it to an empty state.
// If the set is nil, it returns nil. The method is thread-safe and efficiently
// transfers ownership of the underlying slice to the caller.
func (set *CustomSet[T]) Purge() []T {
	if set == nil {
		return nil
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	out := set.s
	set.s = []T{}
	return out
}

// Export returns a copy of the set's underlying slice, ensuring thread-safe access.
// If the set is nil, it returns nil. The returned slice is a new slice with the same
// elements as the set, preventing direct modification of the set's internal state.
func (set *CustomSet[T]) Export() []T {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	out := make([]T, len(set.s))
	copy(out, set.s)
	return out
}

// ForEach iterates over each element in the set, applying the provided function.
// It stops iteration if the function returns false. The method is thread-safe
// and handles nil sets or nil functions by returning immediately.
func (set *CustomSet[T]) ForEach(fn func(T) bool) {
	if set == nil || fn == nil {
		return
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	for _, v := range set.s {
		if !fn(v) {
			break
		}
	}
}

// Reserve increases the capacity of the set to at least the specified capacity.
// If the set is nil, it returns false. The method is thread-safe and ensures
// that the underlying slice can accommodate the requested capacity without
// reallocation.
func (set *CustomSet[T]) Reserve(capacity int) bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doReserve(capacity)
}

func (set *CustomSet[T]) doReserve(capacity int) bool {
	if capacity > cap(set.s) {
		s2 := make([]T, len(set.s), capacity)
		copy(s2, set.s)
		set.s = s2
		return true
	}
	return false
}

// Grow increases the capacity of the set by the specified additional amount.
// If the set is nil, it returns false. The method is concurrency-safe and ensures
// that the underlying slice can accommodate the increased capacity without
// reallocation.
func (set *CustomSet[T]) Grow(additional int) bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doReserve(cap(set.s) + additional)
}

// Trim reduces the capacity of the set to match its length.
// If the set is nil, it returns false. The method is thread-safe and ensures
// that the underlying slice's capacity is minimized without changing its contents.
func (set *CustomSet[T]) Trim() bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doTrim(0)
}

func (set *CustomSet[T]) doTrim(minCapacity int) bool {
	l := len(set.s)
	capacity := max(minCapacity, l)

	switch {
	case capacity == cap(set.s):
		return false
	case capacity == 0:
		set.s = []T{}
		return true
	default:
		s2 := make([]T, l, capacity)
		if l > 0 {
			copy(s2, set.s)
		}
		set.s = s2
		return true
	}
}

// TrimN reduces the capacity of the set to at least the specified minimum capacity.
// If the set is nil, it returns false. The method is concurrency-safe and ensures
// that the underlying slice's capacity is minimized while maintaining at least
// the specified minimum capacity.
func (set *CustomSet[T]) TrimN(minCapacity int) bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doTrim(minCapacity)
}

// GetByIndex returns the element at the specified index in the set.
// If the index is out of bounds, it returns the zero value and false.
// This is concurrency-safe but it's only intended to be used when testing,
// always prefer ForEach() or Export() instead.
func (set *CustomSet[T]) GetByIndex(i int) (T, bool) {
	var zero T

	if set == nil || i < 0 {
		// invalid
		return zero, false
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	if i >= len(set.s) {
		// out of range
		return zero, false
	}

	return set.s[i], true
}
