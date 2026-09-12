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
  intermediate, `int64` for `Int32` and `Int128` for `Int64`.
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
- **`ErrDivZero`**: the division-by-zero panic value, wrapping
  `core.ErrInvalid`.

Files:

- `num/atto128.go`: the `Atto128` instantiation and its scale.
- `num/const.go`: word primitives, the fixed-point scale factors and
  the sentinel bounds (`MaxUint128`, `MinInt128`, …).
- `num/decimal.go`: `Decimal` and the `DecimalScaler` interface.
- `num/doc.go`: package documentation.
- `num/errors.go`: `ErrDivZero`.
- `num/euclidean.go`: the `Euclidean` and `SignedEuclidean` constraints
  and the Euclidean division helpers.
- `num/format.go`: the digit grouping behind `GoString`; each type's
  `GoString` sits in its own file.
- `num/int128.go`: `Int128` and its operations.
- `num/int32.go`: `Int32` and its operations.
- `num/int64.go`: `Int64` and its operations.
- `num/milli.go`: the `Milli32` and `Milli64` instantiations and their
  scales.
- `num/num.go`: the `Unsigned` and `Signed` constraints.
- `num/u256.go`: the unexported 256-bit intermediate backing the wide
  multiply and 128-bit division.
- `num/uint128.go`: `Uint128` and its operations.

## Development Notes

- Constructors use two prefixes and every type follows them. `New` builds
  from parts: `NewUint128(hi, lo)` and `NewInt128(hi, lo)` take the two
  words, `NewMilli32`, `NewMilli64` and `NewAtto128` take whole units
  and sub-units. `As` changes only the type of a value that already is
  the count: `AsUint128` and `AsInt128` extend a native integer,
  `AsMilli32`, `AsMilli64` and `AsAtto128` read the backing integer at
  the resolution. `AsInt32` and `AsInt64` are the conversions of the
  native types; `Int32` and `Int64` have no parts, so no `New`. A new
  type gets both, or a comment saying why one is enough.
- `GoString` prints the constructor call that rebuilds the value, the
  `As` count form while the value fits the native word and the `New`
  words form in hex beyond it; a `Decimal` prints both parts with its
  sign, so the call holds under either sign rule of the constructor.
  `DecimalScaler` carries the instantiation's name and the int64 fit
  check through unexported methods, which closes the family to the
  package's scalers; the constraints stay free of formatting methods,
  and the `Decimal` fallback reaches its backing's form through `%#v`.
  A new type or instantiation adds a row to the `GoString` table.
- The package has one sentinel, `ErrDivZero`, a `core.QuietWrap` of
  `core.ErrInvalid`, and division by zero is the only failure: it
  panics with that value. Arithmetic wraps on overflow and the
  constructors never fail.
- Operations allocate nothing; keep it that way in the hot paths.

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
