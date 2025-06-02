# Agent Documentation for x/cmp

## Overview

The `cmp` package provides generic comparison and matching utilities for Go.
It offers type-safe comparison functions, condition predicates, and a powerful
`Matcher` interface for building composable filtering logic.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Types

- **`CompFunc[T any]`**: Generic comparison function following standard
  conventions (negative for less than, zero for equal, positive for greater).
- **`CondFunc[T any]`**: Generic condition function returning boolean results.
- **`Matcher[T any]`**: Interface for composable filtering with And/Or/Not
  operations.
- **`MatchFunc[T any]`**: Function type implementing the Matcher interface.

### Main Files

- `types.go`: Core type definitions (CompFunc, CondFunc).
- `match.go`: Matcher interface and implementations (ands, ors).
- `matchfunc.go`: MatchFunc implementation.
- `match_cmp.go`: Matcher creation functions for comparison operations.
- `cmp.go`: Direct comparison functions (Eq, Lt, Gt, etc.).
- `compose.go`: Matcher composition for cross-type operations.

## Architecture Notes

The package uses Go generics extensively to provide type-safe operations while
maintaining flexibility. Key design patterns:

1. **Nil Safety**: All functions that accept comparison/condition functions
   panic if a nil is provided, preventing runtime errors.
2. **Logical Identity**: Empty AND matchers match everything (true), empty OR
   matchers match nothing (false).
3. **Composability**: Matchers can be combined using logical operators and
   composed across types using accessor functions.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests are organized alongside their implementation files:
- `*_test.go` files contain unit tests for corresponding source files.
- Tests use table-driven patterns for comprehensive coverage.
- Benchmarks are included for performance-critical operations.

Example test pattern:
```go
func TestMatchEq(t *testing.T) {
    tests := []struct {
        name   string
        value  int
        target int
        want   bool
    }{
        {"equal", 5, 5, true},
        {"not equal", 5, 3, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := MatchEq(tt.target)
            if got := m.Match(tt.value); got != tt.want {
                t.Errorf("Match() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Common Usage Patterns

### Creating Matchers

```go
// From comparison functions
isPositive := AsMatcher(func(n int) bool { return n > 0 })

// Using built-in creators
equals5 := MatchEq(5)
greaterThan10 := MatchGt(10)
```

### Combining Matchers

```go
// Logical AND
between5And10 := MatchGt(5).And(MatchLt(10))

// Logical OR
zeroOrNegative := MatchEq(0).Or(MatchLt(0))

// Negation
notZero := MatchEq(0).Not()
```

### Cross-Type Matching

```go
// Match Person objects by age
isAdult := Compose(
    func(p Person) (int, bool) { return p.Age, true },
    MatchGtEq(18),
)
```

## Dependencies

- `darvaza.org/core`: For core utilities and panic handling.
- Standard library only.

## See Also

- [Package README](README.md) for detailed API documentation.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
