# `darvaza.org/x/cmp`

[![Go Reference][godoc_badge]][godoc_link]
[![Go Report Card][goreportcard_badge]][goreportcard_link]

[godoc_badge]: https://pkg.go.dev/badge/darvaza.org/x/cmp.svg
[godoc_link]: https://pkg.go.dev/darvaza.org/x/cmp
[goreportcard_badge]: https://goreportcard.com/badge/darvaza.org/x/cmp
[goreportcard_link]: https://goreportcard.com/report/darvaza.org/x/cmp

## Overview

The `cmp` package provides generic helpers for comparing and matching values
in Go. It defines type-safe comparison functions and utilities to adapt them
for various use cases, leveraging Go's generics.

## Interfaces

### `CompFunc[T any]`

`CompFunc` is a generic comparison function type that follows the standard
comparison convention:

* Returns a negative value if `a < b`
* Returns zero if `a == b`
* Returns a positive value if `a > b`

```go
type CompFunc[T any] func(a, b T) int
```

Utility functions for working with `CompFunc`:

* `Reverse[T any](cmp CompFunc[T]) CompFunc[T]`: Creates a new comparison
  function that inverts the order of the original comparison.

### `CondFunc[T any]`

`CondFunc` is a generic condition function type that evaluates two values and
returns a boolean result:

```go
type CondFunc[T any] func(a, b T) bool
```

Conversion functions:

* `AsLess[T any](cmp CompFunc[T]) CondFunc[T]`: Converts a comparison function
  to a "less than" condition function.
* `AsEqual[T any](cmp CompFunc[T]) CondFunc[T]`: Converts a comparison
  function to an equality condition function.
* `AsCmp[T any](less CondFunc[T]) CompFunc[T]`: Converts a "less than" condition
  function to a comparison function. Returns -1 if less(a,b) is true, 1 if less(b,a)
  is true, and 0 otherwise.

## `Matcher`

The `Matcher` interface provides a powerful abstraction for building
composable filtering and matching logic with a fluent API. It allows for
combining predicates using logical operators (AND, OR, NOT) and testing values
against these conditions.

```go
type Matcher[T any] interface {
    And(...Matcher[T]) Matcher[T]
    Or(...Matcher[T]) Matcher[T]
    Not() Matcher[T]
    Match(T) bool
}
```

### `MatchFunc`

`MatchFunc` is a function type that implements the `Matcher` interface,
allowing simple predicate functions to be used as matchers:

```go
type MatchFunc[T any] func(T) bool
```

The `MatchFunc` type makes it easy to create matchers from simple functions.
If a `MatchFunc` is nil, its `Match` method returns true (matches everything),
making it act as a logical identity element.

### Implementation Details

The package provides two concrete implementations of the `Matcher` interface:

1. `ands[T]`: A slice-backed matcher that implements logical AND semantics
     * Returns true if all contained matchers match the value.
     * An empty ands matcher matches everything (true).

2. `ors[T]`: A slice-backed matcher that implements logical OR semantics
     * Returns true if any contained matcher matches the value.
     * An empty ors matcher matches nothing (false).

Both implementations handle nil matchers gracefully by ignoring them during
matching.

### Utility Functions

* `AsMatcher[T any](fn MatchFunc[T]) Matcher[T]`: Converts a function to a
  `Matcher`, allowing simple functions to be used with the matcher API.
* `M[T any](m Matcher[T]) func(T) bool`: Converts a Matcher to a simple
  boolean function. If the Matcher is nil, it returns a function that always
  returns true.
* `MatchAny[T any](queries ...Matcher[T]) Matcher[T]`: Creates a matcher that
  returns true if any of the provided matchers match (logical OR).
* `MatchAll[T any](queries ...Matcher[T]) Matcher[T]`: Creates a matcher that
  returns true if all of the provided matchers match (logical AND).

### Matcher Creation Functions

The package provides a comprehensive set of functions to create matchers for
common comparison operations:

#### Equality Matchers

* `MatchEq[T comparable](v T) Matcher[T]`: Creates a matcher that checks for
  equality with the given value.
* `MatchEqFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher that
  checks for equality using a custom comparison function.
* `MatchEqFn2[T any](v T, eq CondFunc[T]) Matcher[T]`: Creates a matcher that
  checks for equality using a custom equality function.

#### Inequality Matchers

* `MatchNotEq[T comparable](v T) Matcher[T]`: Creates a matcher that checks
  for inequality with the given value.
* `MatchNotEqFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher
  that checks for inequality using a custom comparison function.
* `MatchNotEqFn2[T any](v T, eq CondFunc[T]) Matcher[T]`: Creates a matcher
  that checks for inequality using a custom equality function.

#### Greater Than Matchers

* `MatchGt[T core.Ordered](v T) Matcher[T]`: Creates a matcher that checks if
  a value is strictly greater than the given value.
* `MatchGtFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher that
  checks if a value is strictly greater than the given value using a custom
  comparison function.

#### Greater Than or Equal Matchers

* `MatchGtEq[T core.Ordered](v T) Matcher[T]`: Creates a matcher that checks
  if a value is greater than or equal to the given value.
* `MatchGtEqFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher
  using a custom comparison function for greater than or equal checks.
* `MatchGtEqFn2[T any](v T, less CondFunc[T]) Matcher[T]`: Creates a matcher
  using a custom condition function for greater than or equal checks.

#### Less Than Matchers

* `MatchLt[T core.Ordered](v T) Matcher[T]`: Creates a matcher that checks if
  a value is strictly less than the given value.
* `MatchLtFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher that
  checks if a value is strictly less than the given value using a custom
  comparison function.
* `MatchLtFn2[T any](v T, less CondFunc[T]) Matcher[T]`: Creates a matcher
  that checks if a value is strictly less than the given value using a custom
  condition function.

#### Less Than or Equal Matchers

* `MatchLtEq[T core.Ordered](v T) Matcher[T]`: Creates a matcher that checks
  if a value is less than or equal to the given value.
* `MatchLtEqFn[T any](v T, cmp CompFunc[T]) Matcher[T]`: Creates a matcher
  using a custom comparison function for less than or equal checks.
* `MatchLtEqFn2[T any](v T, less CondFunc[T]) Matcher[T]`: Creates a matcher
  using a custom condition function for less than or equal checks.

### Matcher Composition

The package provides a powerful composition mechanism to create matchers that
operate across different types:

* `Compose[T any, V any](fn func(T) (V, bool), match Matcher[V]) Matcher[T]`:
  Creates a new Matcher by applying an accessor function to transform input
  values before matching against an existing matcher.

The `Compose` function enables building complex matching logic by combining
simpler matchers with accessor functions. The accessor function extracts or
transforms a value of type T into a value of type V and indicates whether the
extraction was successful. The resulting matcher returns false if the accessor
function returns `ok=false`.

#### Example of Matcher Composition

```go
type Person struct {
    Name string
    Age  int
}

// Create a matcher that checks if a person is an adult
isAdult := cmp.Compose(
    func(p Person) (int, bool) { return p.Age, true },
    cmp.MatchGtEq(18),
)

alice := Person{Name: "Alice", Age: 25}
bob := Person{Name: "Bob", Age: 16}

fmt.Println(isAdult.Match(alice))  // true
fmt.Println(isAdult.Match(bob))    // false

// Composition with conditional extraction
hasValidEmail := cmp.Compose(
    func(p Person) (string, bool) {
        // Only extract email if it contains "@"
        if strings.Contains(p.Name, "@") {
            return p.Name, true
        }
        return "", false
    },
    cmp.MatchNotEq(""),
)
```

### More Examples

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

## Comparison Functions

The package provides several comparison functions that can be used directly or
as building blocks for more complex comparisons:

### Equality Functions

* `Eq[T comparable](a, b T) bool`: Returns true if `a` equals `b` using Go's
  equality operator.
* `EqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` equals `b`
  using a custom comparison function.
* `EqFn2[T any](a, b T, eq CondFunc[T]) bool`: Returns true if `a` equals `b`
  using a custom equality condition function.

### Inequality Functions

* `NotEq[T comparable](a, b T) bool`: Returns true if `a` is not equal to `b`
  using Go's inequality operator.
* `NotEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is not
  equal to `b` using a custom comparison function.
* `NotEqFn2[T any](a, b T, eq CondFunc[T]) bool`: Returns true if `a` is not
  equal to `b` using a custom equality condition function.

### Ordering Functions

* `Lt[T core.Ordered](a, b T) bool`: Returns true if `a` is less than `b`.
* `LtFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is less
  than `b` using a custom comparison function.
* `LtFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is less
  than `b` using a custom less-than condition function.

* `LtEq[T core.Ordered](a, b T) bool`: Returns true if `a` is less than or
  equal to `b`.
* `LtEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is less
  than or equal to `b` using a custom comparison function.
* `LtEqFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is less
  than or equal to `b` using a custom condition function.

* `Gt[T core.Ordered](a, b T) bool`: Returns true if `a` is greater than `b`.
* `GtFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is greater
  than `b` using a custom comparison function.

* `GtEq[T core.Ordered](a, b T) bool`: Returns true if `a` is greater than or
  equal to `b`.
* `GtEqFn[T any](a, b T, cmp CompFunc[T]) bool`: Returns true if `a` is
  greater than or equal to `b` using a custom comparison function.
* `GtEqFn2[T any](a, b T, less CondFunc[T]) bool`: Returns true if `a` is
  greater than or equal to `b` using a custom condition function.

All functions that accept custom comparison or condition functions will panic
if a nil function is provided.

## Development

For development guidelines, architecture notes, and AI agent instructions, see [AGENT.md](AGENT.md).

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## License

This project is licensed under the MIT License.
