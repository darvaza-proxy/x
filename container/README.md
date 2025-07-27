# `darvaza.org/x/container`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/container.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/container
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/container
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/container
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=container
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x

## Overview

The `container` package provides generic data structure implementations that
extend Go's standard library containers. It includes type-safe wrappers and
advanced collection types with a focus on performance and ergonomics.

## Packages

### `list`

Type-safe wrapper around the standard `container/list`, providing a
generic doubly linked list implementation.

```go
// Create a typed list
l := list.New[string]("first", "second")
l.PushBack("third")

// Iterate over elements
for e := l.First(); e != nil; e = e.Next() {
    fmt.Println(e.Value)
}

// Use ForEach for iteration
l.ForEach(func(value string) bool {
    fmt.Println(value)
    return true // true continues, false stops
})
```

### `set`

Thread-safe set implementation using a map internally with configurable key
extraction and hashing functions.

```go
// Configure a set
cfg := set.Config[string, string, Person]{
    Key: func(p Person) string { return p.ID },
    Hash: func(k string) string { return k },
}

// Create and use
s := &set.Set[string, string, Person]{}
s.Init(cfg, person1, person2)
s.Push(person3)

// Check existence and iterate
if s.Has(person1) {
    s.ForEach(func(p Person) bool {
        fmt.Printf("Person: %s\n", p.ID)
        return false // continue iteration
    })
}
```

### `slices`

Provides slice-based set implementations and utilities. See the
[slices README](slices/README.md) for detailed documentation.

```go
// Create an ordered set (duplicates are automatically removed)
s := slices.NewOrderedSet[int](3, 1, 4, 1, 5)
s.Add(9)
exists := s.Contains(4)

// Get sorted values
values := s.Values() // [1, 3, 4, 5, 9]

// Use custom comparison function
custom := slices.NewCustomSet(func(a, b string) int {
    return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}, "Hello", "world", "HELLO") // case-insensitive set
```

## Features

* **Type Safety**: All containers use Go generics for compile-time type safety
* **Thread Safety**: Map-based set includes built-in synchronization
* **Performance**: Optimized implementations for common operations
* **Flexibility**: Configurable comparison and hashing functions
* **Comprehensive Testing**: All packages now have extensive test coverage
  with examples

## Installation

```bash
go get darvaza.org/x/container
```

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENT.md](AGENT.md).

## License

See [LICENCE.txt](LICENCE.txt) for details.
