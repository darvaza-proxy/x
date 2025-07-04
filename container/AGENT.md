# Agent Documentation for x/container

## Overview

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Subpackages

- **`list`**: Type-safe wrapper around container/list for doubly linked lists.
- **`set`**: Map-based set implementation with thread-safety.
- **`slices`**: Slice-based set implementations and utilities.

### Main Types

#### list Package

- **`List[T]`**: Generic doubly linked list wrapper providing type safety.

#### set Package

- **`Set[K, H comparable, T any]`**: Thread-safe set with configurable hashing.
- **`Config[K, H, T]`**: Configuration for Set behavior (key extraction, hashing).

#### slices Package

- **`Set[T]`**: Interface for slice-based sets.
- **`CustomSet[T]`**: Sorted slice-based set with custom comparison.
- **`OrderedSet[T]`**: Convenience type for ordered types.

## Architecture Notes

The package follows several design principles:

1. **Type Safety**: All containers use Go generics for compile-time type safety.
2. **Thread Safety**: The map-based set includes built-in mutex protection.
3. **Performance**: Slice-based sets maintain sorted order for efficient operations.
4. **Flexibility**: Custom comparison and hashing functions supported.

Key patterns:
- Generic type parameters enable reuse across different types.
- Interface-based design allows multiple implementations.
- Error handling uses wrapped errors from darvaza.org/core.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:
- Type safety verification.
- Thread safety for concurrent operations.
- Performance benchmarks for different set sizes.
- Edge cases (nil receivers, empty sets).

## Common Usage Patterns

### Using List

```go
// Create a typed list
l := list.New[string]("first", "second")

// Add elements
l.PushBack("third")
l.PushFront("zero")

// Iterate
for e := l.First(); e != nil; e = e.Next() {
    fmt.Println(e.Value)
}
```

### Using Map-based Set

```go
// Configure a set
cfg := set.Config[string, string, Person]{
    Key: func(p Person) string { return p.ID },
    Hash: func(k string) string { return k },
}

// Create and initialize
s := &set.Set[string, string, Person]{}
s.Init(cfg, person1, person2)

// Operations
s.Push(person3)
p, err := s.Get("id123")
s.ForEach(func(p Person) error {
    fmt.Println(p.Name)
    return nil
})
```

### Using Slice-based Set

```go
// Create ordered set for comparable types
s := slices.NewOrderedSet[int](3, 1, 4, 1, 5)

// Set operations
s.Add(9)
s.Delete(1)
exists := s.Contains(4)

// Convert to slice
values := s.Values()
```

### Custom Comparison Set

```go
// Create set with custom comparison
cmp := func(a, b Person) int {
    return strings.Compare(a.Name, b.Name)
}
s := slices.NewCustomSet(cmp, person1, person2)

// Operations maintain sorted order
s.Add(person3)
```

## Performance Characteristics

- **List**: O(1) push/pop at ends, O(n) search.
- **Map Set**: O(1) average case for all operations.
- **Slice Set**: O(log n) search, O(n) insert/delete.

## Dependencies

- `darvaza.org/core`: Core utilities and error handling.
- Standard library (container/list, sync).

## See Also

- [slices README](slices/README.md) for slice-based set implementation details.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
- Go standard library container documentation.
