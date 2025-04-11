// Package cmp provides generic helpers to compare and match values.
// It includes functions for equality, ordering and custom comparisons,
// along with utilities to convert between different comparison function types.
package cmp

import "darvaza.org/core"

// Eq returns true if a equals b for comparable types.
// Uses the standard Go equality operator.
func Eq[T comparable](a, b T) bool {
	return a == b
}

// EqFn returns true if a equals b using a custom comparison function.
// Returns true when the comparison function returns zero.
// Panics if the provided comparison function is nil.
func EqFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) == 0
}

// EqFn2 returns true if a equals b using a custom equality condition.
// Returns the direct result of the equality condition function.
// Panics if the provided equality function is nil.
func EqFn2[T any](a, b T, eq CondFunc[T]) bool {
	if eq == nil {
		core.Panic(newNilCondFuncErr())
	}
	return eq(a, b)
}

// NotEq returns true if a does not equal b for comparable types.
// Uses the standard Go inequality operator.
func NotEq[T comparable](a, b T) bool {
	return a != b
}

// NotEqFn returns true if a does not equal b using a custom comparison function.
// Returns true when the comparison function does not return zero.
// Panics if the provided comparison function is nil.
func NotEqFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) != 0
}

// NotEqFn2 returns true if a does not equal b using an equality condition.
// Returns the negation of the equality condition function result.
// Panics if the provided equality function is nil.
func NotEqFn2[T any](a, b T, eq CondFunc[T]) bool {
	if eq == nil {
		core.Panic(newNilCondFuncErr())
	}
	return !eq(a, b)
}

// Lt returns true if a is less than b for ordered types.
// Uses the standard Go less-than operator.
func Lt[T core.Ordered](a, b T) bool {
	return a < b
}

// LtFn returns true if a is less than b using a custom comparison function.
// Returns true when the comparison function returns a negative value.
// Panics if the provided comparison function is nil.
func LtFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) < 0
}

// LtFn2 returns true if a is less than b using a less-than condition.
// Returns the direct result of the less-than condition function.
// Panics if the provided less-than condition function is nil.
func LtFn2[T any](a, b T, less CondFunc[T]) bool {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return less(a, b)
}

// LtEq returns true if a is less than or equal to b for ordered types.
// Uses the standard Go less-than-or-equal operator.
func LtEq[T core.Ordered](a, b T) bool {
	return a <= b
}

// LtEqFn returns true if a is less than or equal to b using a comparison function.
// Returns true when comparison result is less than or equal to zero.
// Panics if the provided comparison function is nil.
func LtEqFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) <= 0
}

// LtEqFn2 returns true if a is less than or equal to b using a condition.
// Returns true when b is not less than a using the provided function.
// Panics if the provided less-than condition function is nil.
func LtEqFn2[T any](a, b T, less CondFunc[T]) bool {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return !less(b, a)
}

// Gt returns true if a is greater than b for ordered types.
// Uses the standard Go greater-than operator.
func Gt[T core.Ordered](a, b T) bool {
	return a > b
}

// GtFn returns true if a is greater than b using a custom comparison function.
// Returns true when the comparison function returns a positive value.
// Panics if the provided comparison function is nil.
func GtFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) > 0
}

// GtEq returns true if a is greater than or equal to b for ordered types.
// Uses the standard Go greater-than-or-equal operator.
func GtEq[T core.Ordered](a, b T) bool {
	return a >= b
}

// GtEqFn returns true if a is greater than or equal to b using a comparison.
// Returns true when comparison result is greater than or equal to zero.
// Panics if the provided comparison function is nil.
func GtEqFn[T any](a, b T, cmp CompFunc[T]) bool {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}
	return cmp(a, b) >= 0
}

// GtEqFn2 returns true if a is greater than or equal to b using a condition.
// Returns true when a is not less than b using the provided function.
// Panics if the provided less-than condition function is nil.
func GtEqFn2[T any](a, b T, less CondFunc[T]) bool {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}
	return !less(a, b)
}
