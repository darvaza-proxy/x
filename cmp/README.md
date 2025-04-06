# cmp

[![Go Reference][godoc_badge]][godoc_link]
[![Go Report Card][goreportcard_badge]][goreportcard_link]

[godoc_badge]: https://pkg.go.dev/badge/darvaza.org/x/cmp.svg
[godoc_link]: https://pkg.go.dev/darvaza.org/x/cmp
[goreportcard_badge]: https://goreportcard.com/badge/darvaza.org/x/cmp
[goreportcard_link]: https://goreportcard.com/report/darvaza.org/x/cmp

## Overview

The `cmp` package provides generic helpers for comparing and matching values in Go. It defines type-safe comparison functions and utilities to adapt them for various use cases, leveraging Go's generics.

## Interfaces

### `CompFunc[T any]`

`CompFunc` is a generic comparison function type that follows the standard comparison convention:
- Returns a negative value if `a < b`
- Returns zero if `a == b`
- Returns a positive value if `a > b`

```go
type CompFunc[T any] func(a, b T) int
```

Utility functions for working with `CompFunc`:

- `Reverse[T any](cmp CompFunc[T]) CompFunc[T]`: Creates a new comparison function that inverts the order of the original comparison.

### `CondFunc[T any]`

`CondFunc` is a generic condition function type that evaluates two values and returns a boolean result:

```go
type CondFunc[T any] func(a, b T) bool
```

Conversion functions:

- `AsLess[T any](cmp CompFunc[T]) CondFunc[T]`: Converts a comparison function to a "less than" condition function.
- `AsEqual[T any](cmp CompFunc[T]) CondFunc[T]`: Converts a comparison function to an equality condition function.

## `Matcher`

The `Matcher` interface provides a powerful abstraction for building composable filtering and matching logic with a fluent API. It allows for combining predicates using logical operators (AND, OR, NOT) and testing values against these conditions.

```go
type Matcher[T any] interface {
    And(...Matcher[T]) Matcher[T]
    Or(...Matcher[T]) Matcher[T]
    Not() Matcher[T]
    Match(T) bool
}
```

### `MatchFunc`

`MatchFunc` is a function type that implements the `Matcher` interface, allowing simple predicate functions to be used as matchers:

```go
type MatchFunc[T any] func(T) bool
```

The `MatchFunc` type makes it easy to create matchers from simple functions. If a `MatchFunc` is nil, its `Match` method returns true (matches everything), making it act as a logical identity element.

### Utility Functions

- `AsMatcher[T any](fn MatchFunc[T]) Matcher[T]`: Converts a function to a `Matcher`, allowing simple functions to be used with the matcher API.
- `MatchAny[T any](queries ...Matcher[T]) Matcher[T]`: Creates a matcher that returns true if any of the provided matchers match (logical OR).
- `MatchAll[T any](queries ...Matcher[T]) Matcher[T]`: Creates a matcher that returns true if all of the provided matchers match (logical AND).

### Examples

```go
// Create matchers from predicate functions
isEven := cmp.AsMatcher(func(n int) bool { return n%2 == 0 })
isPositive := cmp.AsMatcher(func(n int) bool { return n > 0 })

// Combine matchers using logical operators
isEvenAndPositive := isEven.And(isPositive)
isEvenOrPositive := isEven.Or(isPositive)
isOdd := isEven.Not()

// Test values against matchers
fmt.Println(isEvenAndPositive.Match(4))  // true
fmt.Println(isEvenAndPositive.Match(-2)) // false
fmt.Println(isEvenOrPositive.Match(3))   // true
fmt.Println(isOdd.Match(5))              // true
```

### Implementation Details

The package provides two concrete implementations of the `Matcher` interface:

1. `ands[T]`: A slice-backed matcher that implements logical AND semantics
   - Returns true if all contained matchers match the value
   - An empty ands matcher matches everything (true)

2. `ors[T]`: A slice-backed matcher that implements logical OR semantics
   - Returns true if any contained matcher matches the value
   - An empty ors matcher matches nothing (false)

Both implementations handle nil matchers gracefully by ignoring them during matching.

## Comparison functions

The package provides several comparison functions that can be used directly or as building blocks for more complex comparisons:

### Equality
- `Eq[T comparable](a, b T) bool`: Returns true if `a` equals `b` using Go's equality operator.
- `EqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` equals `b` using a custom comparison function.
- `EqFn2[T any](a, b T, eq CondFunc[T]) bool`: Returns true if `a` equals `b` using a custom equality condition function.

### Inequality
- `NotEq[T comparable](a, b T) bool`: Returns true if `a` is not equal to `b` using Go's inequality operator.
- `NotEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is not equal to `b` using a custom comparison function.
- `NotEqFn2[T any](a, b T, eq CondFunc[T]) bool`: Returns true if `a` is not equal to `b` using a custom equality condition function.

### Less Than
- `Lt[T core.Ordered](a, b T) bool`: Returns true if `a` is less than `b` using Go's < operator.
- `LtFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is less than `b` using a custom comparison function.
- `LtFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is less than `b` using a custom less-than condition function.

### Less Than or Equal
- `LtEq[T core.Ordered](a, b T) bool`: Returns true if `a` is less than or equal to `b` using Go's <= operator.
- `LtEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is less than or equal to `b` using a custom comparison function.
- `LtEqFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is less than or equal to `b` using a custom less-than condition function.

### Greater Than
- `Gt[T core.Ordered](a, b T) bool`: Returns true if `a` is greater than `b` using Go's > operator.
- `GtFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is greater than `b` using a custom comparison function.

### Greater Than or Equal
- `GtEq[T core.Ordered](a, b T) bool`: Returns true if `a` is greater than or equal to `b` using Go's >= operator.
- `GtEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is greater than or equal to `b` using a custom comparison function.
- `GtEqFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is greater than or equal to `b` using a custom less-than condition function.

All functions that accept custom comparison or condition functions will panic if a nil function is provided.

## License

This project is licensed under the MIT License.
