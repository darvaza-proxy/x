package slices

// This file implements a thread-safe generic set data structure with custom comparison support.

import (
	"sort"
	"sync"

	"darvaza.org/core"
)

// Compile-time check to ensure CustomSet implements the Set interface
var _ Set[struct{}] = (*CustomSet[struct{}])(nil)

// CustomSet is a generic thread-safe set implementation with custom element comparison.
// It maintains a sorted slice of unique elements using a provided comparison function.
type CustomSet[T any] struct {
	mu  sync.RWMutex
	s   []T
	cmp func(T, T) int // comparison function: negative if a<b, zero if a==b, positive if a>b
}

// New creates a new empty CustomSet with the same comparison function as the current set.
// Returns nil if the current set is nil, and panics if the set is not initialized.
func (set *CustomSet[T]) New() Set[T] {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	if set.cmp == nil {
		core.Panic(set.newNotInitialized(1))
	}

	return &CustomSet[T]{
		cmp: set.cmp,
	}
}

// NewCustomSet creates a new CustomSet with the provided comparison function and initial elements.
// The comparison function should return negative if a<b, zero if a==b, and positive if a>b.
// Returns an error if the comparison function is nil.
func NewCustomSet[T any](cmp func(T, T) int, initial ...T) (*CustomSet[T], error) {
	if cmp == nil {
		return nil, core.Wrap(core.ErrInvalid, "comparison function is required")
	}

	return &CustomSet[T]{
		s:   dedupeFn(cmp, initial),
		cmp: cmp,
	}, nil
}

// MustCustomSet creates a new CustomSet, panicking if initialization fails.
// This is a convenience function for when error handling is not needed.
func MustCustomSet[T any](cmd func(T, T) int, initial ...T) *CustomSet[T] {
	set, err := NewCustomSet(cmd, initial...)
	if err != nil {
		core.Panic(err)
	}
	return set
}

// InitCustomSet initializes a pre-allocated CustomSet with thread-safe semantics.
// Returns an error if the set is nil, the comparison function is nil, or the set is already initialized.
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
// It sorts the slice and then eliminates consecutive duplicate elements in-place.
// Returns the deduplicated slice.
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
// Returns false if the set is nil or not initialized.
func (set *CustomSet[T]) Contains(v T) bool {
	if set == nil || set.cmp == nil {
		return false
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	_, ok := set.search(0, v)
	return ok
}

// search is an internal helper that finds a value in the set starting from the given index.
// Returns the found index and true if found, or insertion point and false if not found.
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
// Returns nil if the current set is nil, and panics if not initialized.
func (set *CustomSet[T]) Clone() Set[T] {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	if set.cmp == nil {
		core.Panic(set.newNotInitialized(1))
	}

	s := make([]T, len(set.s))
	copy(s, set.s)

	return &CustomSet[T]{
		s:   s,
		cmp: set.cmp,
	}
}

// Len returns the number of elements in the CustomSet.
// Returns 0 if the set is nil.
func (set *CustomSet[T]) Len() int {
	if set == nil {
		return 0
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	return len(set.s)
}

// Cap returns the available and total capacity of the CustomSet.
// The first return value is available capacity, and the second is total capacity.
// Returns (0, 0) if the set is nil.
func (set *CustomSet[T]) Cap() (available, total int) {
	if set == nil {
		return 0, 0
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	return cap(set.s) - len(set.s), cap(set.s)
}

// Add adds the given values to the CustomSet.
// Returns the number of unique values that were added.
// Panics if the set is not properly initialized.
func (set *CustomSet[T]) Add(values ...T) int {
	if set == nil || len(values) == 0 {
		return 0
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.cmp == nil {
		core.Panic(set.newNotInitialized(1))
	}

	return set.doAdd(values)
}

// doAdd implements the non-locking portion of Add.
// Returns the count of added elements.
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

// doAddOne adds a single value to the set starting search from the given index.
// Returns the insertion position and true if added, or position and false if already exists.
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

// Remove removes the given values from the CustomSet.
// Returns the number of values that were actually removed.
// Panics if the set is not properly initialized.
func (set *CustomSet[T]) Remove(values ...T) int {
	if set == nil || len(values) == 0 {
		return 0
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.cmp == nil {
		core.Panic(set.newNotInitialized(1))
	} else if len(set.s) == 0 {
		return 0
	}

	return set.doRemove(values)
}

// doRemove implements the non-locking portion of Remove.
// Returns the count of removed elements.
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

// doRemoveOne removes a single value from the set starting search from the given index.
// Returns the position and true if removed, or position and false if not found.
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
// Does nothing if the set is nil.
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

// Purge removes and returns all elements from the set.
// Unlike Clear(), this transfers ownership of the underlying slice to the caller.
// Returns nil if the set is nil.
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

// Export returns a copy of the set's elements as a new slice.
// Returns nil if the set is nil.
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
// Stops iteration if the function returns false.
// Does nothing if either the set or function is nil.
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
// Returns true if the capacity was increased, false if the set is nil or already has sufficient capacity.
func (set *CustomSet[T]) Reserve(capacity int) bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doReserve(capacity)
}

// doReserve implements the non-locking portion of Reserve.
// Returns true if capacity was changed, false otherwise.
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
// Returns true if the capacity was increased, false otherwise.
func (set *CustomSet[T]) Grow(additional int) bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doReserve(cap(set.s) + additional)
}

// Trim reduces the capacity of the set to match its length.
// Returns true if the capacity was reduced, false otherwise.
func (set *CustomSet[T]) Trim() bool {
	if set == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	return set.doTrim(0)
}

// doTrim implements the non-locking portion of Trim and TrimN.
// Returns true if capacity was changed, false otherwise.
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

// TrimN reduces the capacity of the set while maintaining at least the specified minimum capacity.
// Returns true if the capacity was reduced, false otherwise.
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

func (*CustomSet[T]) newNotInitialized(skip int) error {
	return core.NewPanicError(skip+1, "CustomSet not initialized")
}
