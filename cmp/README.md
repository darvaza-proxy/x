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

## License

This project is licensed under the MIT License.
