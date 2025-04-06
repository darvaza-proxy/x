package cmp

import "darvaza.org/core"

// CompFunc is a generic comparison function that takes two values of type T and returns an integer.
// The return value follows the standard comparison convention:
// - Negative value if a < b
// - Zero if a == b
// - Positive value if a > b
type CompFunc[T any] func(a, b T) int

// Reverse returns a new CompFunc that inverts the comparison result of the given CompFunc.
// It returns a function that negates the original comparison, effectively reversing the order.
// It panics if the provided comparison function is nil.
func Reverse[T any](cmp CompFunc[T]) CompFunc[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}

	return func(a, b T) int {
		return -cmp(a, b)
	}
}

// CondFunc is a generic condition function that takes two values of type T and returns a boolean.
// The return value indicates whether the condition is true or false for the given pair of values.
type CondFunc[T any] func(a, b T) bool

// AsLess converts a CompFunc into a less-than condition function.
// It returns a function that returns true if the first argument is less than the second argument.
// It panics if the provided comparison function is nil.
func AsLess[T any](cmp CompFunc[T]) CondFunc[T] {
	if cmp == nil {
		core.Panic(newNilCompFuncErr())
	}

	return func(a, b T) bool {
		return cmp(a, b) < 0
	}
}

// AsEqual converts a CompFunc into an equality condition function.
// It returns a function that returns true if the first argument is equal to the second argument.
// It panics if the provided comparison function is nil.
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
