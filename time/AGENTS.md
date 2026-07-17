# Agent Documentation for x/time

## Overview

`darvaza.org/x/time` implements time-related primitives that don't belong in
the standard library.

## Key Components

### Subpackages

- **`num`**: fixed-width integer types for the time packages.

### num Package

- **`Uint128`**, **`Int128`**: unsigned and signed 128-bit integers on a
  two-word layout, wrapping on overflow like Go's built-in operators.
- **`Unsigned[T]`**, **`Signed[T]`**: generic constraints naming the
  method surface the family shares.
- **`ErrDivZero`**: the division-by-zero panic value, wrapping
  `core.ErrInvalid`.

Files:

- `num/const.go`: word primitives and the sentinel bounds
  (`MaxUint128`, `MinInt128`, …).
- `num/doc.go`: package documentation.
- `num/errors.go`: `ErrDivZero`.
- `num/int128.go`: `Int128` and its operations.
- `num/num.go`: the `Unsigned` and `Signed` constraints.
- `num/u256.go`: the unexported 256-bit intermediate backing 128-bit
  division.
- `num/uint128.go`: `Uint128` and its operations.

## Development Notes

- Error handling follows `darvaza.org/core` conventions: sentinel
  errors wrap `core.ErrInvalid` via `core.QuietWrap`, context is
  added with `core.Wrap`/`core.Wrapf`, and constructor range
  violations panic via `core.PanicWrapf` with specific error types.
- Operations are designed to be zero-allocation where possible; avoid
  introducing allocations in hot paths.

## Testing Patterns

Tests follow the conventions in [core's TESTING.md][core-testing]:

- `var _ core.TestCase = ...` declarations for every TestCase type.
- Factory functions decouple semantic argument order from
  memory-aligned struct field order.
- Table-driven suites use `core.RunTestCases`; scenario tests use
  `TestFoo() { t.Run("scenario", runTestFooScenario) }`.

## See Also

- [Package README](README.md) for API documentation.
- [Root AGENTS.md](../AGENTS.md) for mono-repo overview.

[core-testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md
