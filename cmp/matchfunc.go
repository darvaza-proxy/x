package cmp

// MatchFunc is a function type that implements the Matcher interface,
// allowing simple functions to be used as matchers.
type MatchFunc[T any] func(T) bool

// And combines this matcher function with others using logical AND.
func (fn MatchFunc[T]) And(others ...Matcher[T]) Matcher[T] {
	return ands[T](qJoin(fn, others))
}

// Or combines this matcher function with others using logical OR.
func (fn MatchFunc[T]) Or(others ...Matcher[T]) Matcher[T] {
	return ors[T](qJoin(fn, others))
}

// Not negates the result of this matcher function, returning a new Matcher
// that produces the opposite of the original match result.
func (fn MatchFunc[T]) Not() Matcher[T] {
	return MatchFunc[T](func(x T) bool {
		return !fn(x)
	})
}

// Match calls the matcher function with the provided value.
// If the function is nil, it returns true (matches everything).
// This default behaviour makes nil MatchFunc act as a logical identity
// element, simplifying composition when conditions are optional.
func (fn MatchFunc[T]) Match(value T) bool {
	if fn == nil {
		return true
	}
	return fn(value)
}

// AsMatcher converts a MatchFunc to a Matcher. Returns nil if the
// provided function is nil.
func AsMatcher[T any](fn MatchFunc[T]) Matcher[T] {
	if fn == nil {
		return nil
	}
	return fn
}

// M converts a Matcher to a simple function that can be used for matching.
// If the Matcher is nil, it returns a function that always returns true,
// allowing seamless usage of Matcher types as boolean functions.
func M[T any](m Matcher[T]) func(T) bool {
	return func(value T) bool {
		if m == nil {
			return true
		}
		return m.Match(value)
	}
}
