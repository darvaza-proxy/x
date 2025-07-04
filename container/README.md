# `darvaza.org/x/container`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/container.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/container
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/container
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/container

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
```

### `slices`

Provides slice-based set implementations and utilities. See the
[slices README](slices/README.md) for detailed documentation.

```go
// Create an ordered set
s := slices.NewOrderedSet[int](3, 1, 4, 1, 5)
s.Add(9)
exists := s.Contains(4)
```

## Features

* **Type Safety**: All containers use Go generics for compile-time type safety
* **Thread Safety**: Map-based set includes built-in synchronization
* **Performance**: Optimized implementations for common operations
* **Flexibility**: Configurable comparison and hashing functions

## Installation

```bash
go get darvaza.org/x/container
```

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## Development

For development guidelines, architecture notes, and AI agent instructions, see [AGENT.md](AGENT.md).

## License

See [LICENCE.txt](LICENCE.txt) for details.
