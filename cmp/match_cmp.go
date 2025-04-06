package cmp

import "darvaza.org/core"

// MatchEq creates a Matcher that checks for equality with the given value.
//
// It returns a function that returns true if the input is equal to the specified value.
// This function is useful for creating filters or validators that need to match exact values.
//
// Example:
//
//	isApple := MatchEq("apple")
//	isApple("apple")  // returns true
//	isApple("banana") // returns false
func MatchEq[T comparable](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return Eq(v0, v)
	})
}

// MatchEqFn creates a Matcher that checks for equality using a custom comparison function.
// It returns a function that returns true if the input is equal to the specified value
// according to the provided comparison function. Panics if the comparison function is nil.
func MatchEqFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return EqFn(v0, v, cmp)
	})
}

// MatchEqFn2 creates a Matcher that checks for equality using a custom equality function.
// It returns a function that returns true if the input is equal to the specified value
// according to the provided equality function. Panics if the equality function is nil.
func MatchEqFn2[T any](v T, eq CondFunc[T]) Matcher[T] {
	if eq == nil {
		core.Panic(newNilCondFuncErr())
	}

	return MatchFunc[T](func(v0 T) bool {
		return EqFn2(v0, v, eq)
	})
}

// MatchNotEq creates a Matcher that checks for inequality with the given value.
// It returns a function that returns true if the input is not equal to the specified value.
func MatchNotEq[T comparable](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return NotEq(v0, v)
	})
}

// MatchNotEqFn creates a Matcher that checks for inequality using a custom comparison function.
// It returns a function that returns true if the input is not equal to the specified value
// according to the provided comparison function. Panics if the comparison function is nil.
func MatchNotEqFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return NotEqFn(v0, v, cmp)
	})
}

// MatchNotEqFn2 creates a Matcher that checks for inequality using a custom equality function.
// It returns a function that returns true if the input is not equal to the specified value
// according to the provided equality function. Panics if the equality function is nil.
func MatchNotEqFn2[T any](v T, eq CondFunc[T]) Matcher[T] {
	if eq == nil {
		core.Panic(newNilCondFuncErr())
	}

	return MatchFunc[T](func(v0 T) bool {
		return NotEqFn2(v0, v, eq)
	})
}

// MatchGt creates a Matcher that checks if a value is strictly greater than the given value.
// It returns a function that returns true if the input is greater than the specified value.
func MatchGt[T core.Ordered](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return Gt(v0, v)
	})
}

// MatchGtFn creates a Matcher that checks if a value is strictly greater than the given value
// using a custom comparison function. It returns a function that returns true if the input
// is greater than the specified value according to the provided comparison function.
// Panics if the comparison function is nil.
func MatchGtFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return GtFn(v0, v, cmp)
	})
}

// MatchGtEq creates a Matcher that checks if a value is greater than or equal to the given value.
// It returns a function that returns true if the input is greater than or equal to the specified value.
func MatchGtEq[T core.Ordered](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return GtEq(v0, v)
	})
}

// MatchGtEqFn creates a Matcher that checks if a value is greater than or equal to the given value
// using a custom comparison function. It returns a function that returns true if the input
// is greater than or equal to the specified value according to the provided comparison function.
// Panics if the comparison function is nil.
func MatchGtEqFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return GtEqFn(v0, v, cmp)
	})
}

// MatchGtEqFn2 creates a Matcher that checks if a value is greater than or equal to the given value
// using a custom condition function. It returns a function that returns true if the input
// is greater than or equal to the specified value according to the provided condition function.
// Panics if the condition function is nil.
func MatchGtEqFn2[T any](v T, less CondFunc[T]) Matcher[T] {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return GtEqFn2(v0, v, less)
	})
}

// MatchLt creates a Matcher that checks if a value is strictly less than the given value.
// It returns a function that returns true if the input is less than the specified value.
func MatchLt[T core.Ordered](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return Lt(v0, v)
	})
}

// MatchLtFn creates a Matcher that checks if a value is strictly less than the given value
// using a custom comparison function. It returns a function that returns true if the input
// is less than the specified value according to the provided comparison function.
// Panics if the comparison function is nil.
func MatchLtFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return LtFn(v0, v, cmp)
	})
}

// MatchLtFn2 creates a Matcher that checks if a value is strictly less than the given value
// using a custom condition function. It returns a function that returns true if the input
// is less than the specified value according to the provided condition function.
// Panics if the condition function is nil.
func MatchLtFn2[T any](v T, less CondFunc[T]) Matcher[T] {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return LtFn2(v0, v, less)
	})
}

// MatchLtEq creates a Matcher that checks if a value is less than or equal to the given value.
// It returns a function that returns true if the input is less than or equal to the specified value.
func MatchLtEq[T core.Ordered](v T) Matcher[T] {
	return MatchFunc[T](func(v0 T) bool {
		return LtEq(v0, v)
	})
}

// MatchLtEqFn creates a Matcher that checks if a value is less than or equal to the given value
// using a custom comparison function. It returns a function that returns true if the input
// is less than or equal to the specified value according to the provided comparison function.
// Panics if the comparison function is nil.
func MatchLtEqFn[T any](v T, cmp CompFunc[T]) Matcher[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return LtEqFn(v0, v, cmp)
	})
}

// MatchLtEqFn2 creates a Matcher that checks if a value is less than or equal to the given value
// using a custom condition function. It returns a function that returns true if the input
// is less than or equal to the specified value according to the provided condition function.
// Panics if the condition function is nil.
func MatchLtEqFn2[T any](v T, less CondFunc[T]) Matcher[T] {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return MatchFunc[T](func(v0 T) bool {
		return LtEqFn2(v0, v, less)
	})
}
