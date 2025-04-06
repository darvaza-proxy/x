package cmp

// MatchFunc is a function type that implements the Matcher interface.
// It allows simple functions to be used as matchers.
type MatchFunc[T any] func(T) bool

// And combines this query function with others using logical AND.
func (fn MatchFunc[T]) And(others ...Matcher[T]) Matcher[T] {
	return ands[T](qJoin(fn, others))
}

// Or combines this query function with others using logical OR.
func (fn MatchFunc[T]) Or(others ...Matcher[T]) Matcher[T] {
	return ors[T](qJoin(fn, others))
}

// Not negates the result of the current MatchFunc, returning a new
// Matcher that returns the opposite of the original match condition.
func (fn MatchFunc[T]) Not() Matcher[T] {
	return MatchFunc[T](func(x T) bool {
		return !fn(x)
	})
}

// Match calls the query function with the provided value.
// If the function is nil, it returns true (matches everything).
// This "match everything" default behavior makes nil MatchFunc act as a logical identity
// element, simplifying composition when conditions are conditionally included.
func (fn MatchFunc[T]) Match(value T) bool {
	if fn == nil {
		return true
	}
	return fn(value)
}

// AsMatcher converts a MatchFunc to a Matcher, allowing simple functions to be used as matchers.
func AsMatcher[T any](fn MatchFunc[T]) Matcher[T] {
	if fn == nil {
		return nil
	}
	return fn
}
