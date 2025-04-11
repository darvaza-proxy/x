package cmp

import "darvaza.org/core"

// Compose creates a new Matcher by applying an accessor function to transform
// input values before matching against an existing matcher. It enables
// composition of matchers across different types, allowing you to build
// complex matching logic by combining simple matchers.
//
// The accessor function (fn) extracts or transforms a value of type T into a
// value of type V and indicates whether the extraction was successful. The
// returned matcher evaluates to false if the accessor function returns
// ok=false.
//
// Common uses include:
// - Extracting and matching on struct fields
// - Transforming values before matching
// - Creating conditional matching chains
//
// Example:
//
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//
//	// Create a matcher that checks if a person is an adult
//	isAdult := Compose(
//	    func(p Person) (int, bool) { return p.Age, true },
//	    GtEq(18),
//	)
//
// Panics if the accessor function or the base matcher is nil.
func Compose[T any, V any](fn func(T) (V, bool), match Matcher[V]) Matcher[T] {
	if fn == nil {
		core.Panic(core.NewPanicError(1, "nil accessor function"))
	}

	if match == nil {
		core.Panic(core.NewPanicError(1, "no match condition"))
	}

	return MatchFunc[T](func(x T) bool {
		if v, ok := fn(x); ok {
			return match.Match(v)
		}
		return false
	})
}
