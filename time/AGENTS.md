# Agent Documentation for x/time

## Overview

`darvaza.org/x/time` implements time-related primitives that don't belong in
the standard library.

## Key Components

### Subpackages

- **`num`**: fixed-width numeric types for the time packages.

### num Package

- **`Uint128`**, **`Int128`**: unsigned and signed 128-bit integers on a
  two-word layout, wrapping on overflow like Go's built-in operators.
- **`Int32`**, **`Int64`**: the native integers wrapped into the same
  method surface, forming the `MulDivMod` product in a wider
  intermediate — `int64` for `Int32`, `Int128` for `Int64`.
- **`Decimal[T, S]`**: signed fixed-point number backed by one of the
  signed integers, with the exported `DecimalScaler` supplying the
  resolution; `Milli32`, `Milli64` and `Atto128` are its
  instantiations.
- **`Unsigned[T]`**, **`Signed[T]`**: generic constraints naming the
  method surface the family shares, including the fused `MulDivMod`
  wide multiply-divide.
- **`EuclideanDivMod`**, **`EuclideanMulDivMod`**: division helpers
  correcting the remainder into `[0, |divisor|)`, constrained on
  `Euclidean`; `SignedEuclidean` combines it with `Signed` and is the
  `Decimal` backing constraint.
- **Base-10 text**: the integer types render through `String` and
  `MarshalText` and parse through `UnmarshalText`, all in base 10 and
  without a big-number dependency.
- **`ErrDivZero`**: the division-by-zero panic value, wrapping
  `core.ErrInvalid`. **`ErrSyntax`** and **`ErrRange`** are the
  text-parse failures: `strconv.ErrSyntax` and `strconv.ErrRange`
  re-exported, and additionally wrapping `core.ErrInvalid`.

Files:

- `num/atto128.go`: the `Atto128` instantiation and its scale.
- `num/const.go`: word primitives, the fixed-point scale factors and
  the sentinel bounds (`MaxUint128`, `MinInt128`, …).
- `num/decimal.go`: `Decimal` and the `DecimalScaler` interface.
- `num/doc.go`: package documentation.
- `num/errors.go`: `ErrDivZero`.
- `num/euclidean.go`: the `Euclidean` and `SignedEuclidean` constraints
  and the Euclidean division helpers.
- `num/int128.go`: `Int128` and its operations.
- `num/int32.go`: `Int32` and its operations.
- `num/int64.go`: `Int64` and its operations.
- `num/milli.go`: the `Milli32` and `Milli64` instantiations and their
  scales.
- `num/num.go`: the `Unsigned` and `Signed` constraints.
- `num/text.go`: the shared base-10 parsing and formatting helpers
  behind each type's `String`, `MarshalText` and `UnmarshalText`
  methods, reporting `ErrSyntax`/`ErrRange` on bad input.
- `num/u256.go`: the unexported 256-bit intermediate backing the wide
  multiply and 128-bit division.
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
