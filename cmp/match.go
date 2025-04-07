package cmp

// Matcher is a generic interface for filtering and combining predicates of type T.
// It allows logical AND and OR operations between conditions, and matching against a value.
//
// Matchers are composable, allowing complex matching logic to be built from simpler
// components. The zero value of any Matcher should either be safe to use or well-documented
// if it has special behavior.
type Matcher[T any] interface {
	// And combines this query with others using logical AND.
	// All conditions must match for the combined query to match.
	// Nil matchers in the provided list are ignored during matching.
	And(...Matcher[T]) Matcher[T]

	// Or combines this query with others using logical OR.
	// At least one condition must match for the combined query to match.
	// Nil matchers in the provided list are ignored during matching.
	Or(...Matcher[T]) Matcher[T]

	// Not negates this matcher's result
	Not() Matcher[T]

	// Match tests if the given value satisfies this query's conditions.
	Match(T) bool
}

// MatchAny returns a query that matches if any of the provided queries match.
// If no queries are provided, the result will match nothing (return false).
// Nil queries in the provided list are ignored during matching.
func MatchAny[T any](queries ...Matcher[T]) Matcher[T] {
	return ors[T](queries)
}

// MatchAll returns a query that matches if all of the provided queries match.
// If no queries are provided, the result will match everything (return true).
// Nil queries in the provided list are ignored during matching.
func MatchAll[T any](queries ...Matcher[T]) Matcher[T] {
	return ands[T](queries)
}

// ands is a slice of matchers that implements the Matcher interface with AND logic.
type ands[T any] []Matcher[T]

// Match returns true if all non-nil queries in the slice match the provided value.
// An empty slice matches everything (returns true), following the logical convention
// that an empty AND condition is always satisfied (similar to logical product of an empty set).
func (c ands[T]) Match(value T) bool {
	for _, q := range c {
		if q != nil && !q.Match(value) {
			return false
		}
	}
	return true
}

// And combines this AND query with others using logical AND.
func (c ands[T]) And(others ...Matcher[T]) Matcher[T] {
	return ands[T](qJoin(c, others))
}

// Or combines this AND query with others using logical OR.
func (c ands[T]) Or(others ...Matcher[T]) Matcher[T] {
	return ors[T](qJoin(c, others))
}

// Not returns a new Matcher that negates the current AND matcher's matching logic.
// It returns true for values that do not match the original AND matcher.
func (c ands[T]) Not() Matcher[T] {
	return AsMatcher(func(x T) bool {
		for _, q := range c {
			if q != nil && !q.Match(x) {
				return true
			}
		}
		return false
	})
}

// ors is a slice of matchers that implements the Matcher interface with OR logic.
type ors[T any] []Matcher[T]

// Match returns true if any non-nil query in the slice matches the provided value.
// An empty slice matches nothing (returns false), following the logical convention
// that an empty OR condition is never satisfied (similar to logical sum of an empty set).
func (c ors[T]) Match(value T) bool {
	for _, q := range c {
		if q != nil && q.Match(value) {
			return true
		}
	}
	return false
}

// And combines this OR query with others using logical AND.
func (c ors[T]) And(others ...Matcher[T]) Matcher[T] {
	return ands[T](qJoin(c, others))
}

// Or combines this OR query with others using logical OR.
func (c ors[T]) Or(others ...Matcher[T]) Matcher[T] {
	return ors[T](qJoin(c, others))
}

// Not returns a new Matcher that negates the current OR matcher's matching logic.
// It returns true for values that do not match the original OR matcher.
func (c ors[T]) Not() Matcher[T] {
	return AsMatcher(func(x T) bool {
		for _, q := range c {
			if q != nil && q.Match(x) {
				return false
			}
		}
		return true
	})
}

// qJoin combines a query with a slice of other queries into a single slice.
// If the first query is nil, it simply returns the others slice.
// This internal function is used by And/Or operations to combine matchers efficiently.
// qJoin combines a Matcher with a slice of Matchers into a single slice.
// It first cleans the 'others' slice by removing nil elements.
//
// The function handles several cases:
// - If fn is nil, returns the cleaned others slice
// - If others is empty, returns a single-element slice with fn
// - If others has enough capacity, reuses the slice by shifting elements
// - Otherwise, allocates a new slice and copies all elements
//
// The returned slice always has fn as its first element (if non-nil),
// followed by the elements from others.
func qJoin[T any](fn Matcher[T], others []Matcher[T]) []Matcher[T] {
	others = qClean(others)
	switch {
	case fn == nil:
		return others
	case len(others) == 0:
		return []Matcher[T]{fn}
	case cap(others) > len(others):
		result := others[:len(others)+1]
		copy(result[1:], result)
		result[0] = fn
		return result
	default:
		result := make([]Matcher[T], 1+len(others))
		result[0] = fn
		copy(result[1:], others)
		return result
	}
}

// qClean removes nil queries from the provided slice, operating in-place.
// The input slice is modified, and a slice containing only non-nil elements is returned.
// This helps to avoid nil checks during matching operations.
func qClean[T any](queries []Matcher[T]) []Matcher[T] {
	result := queries[:0]
	for _, q := range queries {
		if q != nil {
			result = append(result, q)
		}
	}
	return result
}
