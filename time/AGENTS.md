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
- `num/format.go`: the shared side of `Format` and `GoString`, the verb
  tables, the sign, prefix and width padding, the base-10 chunking
  constants and the digit grouping; each type's `Format`, `GoString`
  and digit generation sit in its own file.
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
- `Format` owns every verb, since fmt consults nothing else once a type
  has it: `%#v` is routed to `GoString` by hand. `Uint128` generates
  the digits, base 10 by peeling 10^19 chunks and the power-of-two
  bases by shifting the words; `Int128` prints its sign and hands the
  magnitude over; `Int32` and `Int64` hand the native value to fmt with
  `fmt.FormatString`, under the verb they were given rather than a
  decimal rewrite of it, so their flags cannot drift from fmt's;
  `Decimal` prints its parts as magnitudes one at a time, since `Abs`
  on the backing's minimum wraps while the whole count never can. Fraction
  rounding is half away from zero; the precision of `%f` is fmt's, not
  the resolution's. Never reach for `math/big` for any of this.
- fmt's padding has corners worth knowing, since the 128-bit writer
  reimplements them: the `0` flag is a precision on the digits, so it
  leaves room for the sign but pushes the base prefix outside the
  width; a precision replaces that flag; the octal `0` prefix is not
  added when a zero already leads the digits, while `%#O` carries both
  prefixes; a zero value under precision zero prints as padding alone,
  the sign dropped with the digits; and the `+` of `%+v` reaches a
  `Formatter` as the `+` flag although fmt means the field names of a
  struct by it, so nothing signs under `v`.
- The two paths a `Formatter` takes over from fmt keep fmt's rules as
  well. `%#v` pads and truncates the `GoString` text as fmt pads and
  truncates any string, which is why it hands it back under `%s` rather
  than writing it plainly; and the `%!verb` form prints the value inside
  the brackets under the flags, width and precision the bad verb was
  given, which reach it unmunged, so `%+12t` signs and pads that value
  while nothing around it is padded.
- `TestFormatMatchesFmt` asserts the whole of this against fmt itself,
  over a matrix of formats and values: the integers against the native
  of the same value, the `s` verb against the `d` it prints as, `%#v`
  against a bare `GoStringer`, the `%!verb` form against the native it
  brackets, and a `Decimal` against the `float64` of a value exact in
  one. Extend that matrix rather than hand-writing an expectation; every
  defect in this surface so far has survived a careful reading and died
  on the first run of the table.
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
