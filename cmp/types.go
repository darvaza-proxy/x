package cmp

import "darvaza.org/core"

// CompFunc is a generic comparison function for type T that returns an integer.
// The return value follows the standard comparison convention:
// - Negative value if a < b
// - Zero if a == b
// - Positive value if a > b
type CompFunc[T any] func(a, b T) int

// Reverse returns a new CompFunc that inverts the result of the given CompFunc.
// It negates the original comparison, effectively reversing the order.
// Panics if the provided comparison function is nil.
func Reverse[T any](cmp CompFunc[T]) CompFunc[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}

	return func(a, b T) int {
		return -cmp(a, b)
	}
}

// CondFunc is a generic condition function for type T that returns a boolean.
// The return value indicates whether the condition is true or false for the
// given pair of values.
type CondFunc[T any] func(a, b T) bool

// AsLess converts a CompFunc into a less-than condition function.
// Returns a function that evaluates to true if the first argument is less
// than the second. Panics if the provided comparison function is nil.
func AsLess[T any](cmp CompFunc[T]) CondFunc[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}

	return func(a, b T) bool {
		return cmp(a, b) < 0
	}
}

// AsCmp converts a less-than condition function into a CompFunc.
// Panics if the provided condition function is nil.
func AsCmp[T any](less CondFunc[T]) CompFunc[T] {
	if less == nil {
		core.Panic(newNilCondFuncErr())
	}

	return func(a, b T) int {
		switch {
		case less(a, b):
			return -1
		case less(b, a):
			return 1
		default:
			return 0
		}
	}
}

// AsEqual converts a CompFunc into an equality condition function.
// Returns a function that evaluates to true if the first argument equals
// the second. Panics if the provided comparison function is nil.
func AsEqual[T any](cmp CompFunc[T]) CondFunc[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}

	return func(a, b T) bool {
		return cmp(a, b) == 0
	}
}

func newNilCompFuncErr() error {
	return core.NewPanicError(2, "nil comparison function")
}

func newNilCondFuncErr() error {
	return core.NewPanicError(2, "nil condition function")
}
