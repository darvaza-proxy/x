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
- **`Decimal[T, S]`**: signed fixed-point number backed by a signed
  integer, with the exported `DecimalScaler` supplying the resolution
  and the qualified name; `Milli32`, `Milli64` and `Atto128` are its
  instantiations, built through `NewDecimal` and `AsDecimal`, which
  another package calls over a backing and a scaler of its own.
- **`Unsigned[T]`**, **`Signed[T]`**: generic constraints naming the
  method surface the family shares, including the fused `MulDivMod`
  wide multiply-divide.
- **`Number[T]`**: the constraint naming the whole surface, `Unsigned`
  with `IsNegative` and `Abs`, the fmt, encoding and JSON forms, and
  the conversion to every type of the family, each `(T, bool)`.
- **`NewFromInt128`**, **`NewDecimalFromInt128`**: the generic
  factories, whole units as a type named by its type arguments, the
  first over a closed switch on the seven types of the package and the
  second through the scaler, so an instantiation of another package
  takes it; the nearest bound with `ErrRange` where the units do not
  fit, `core.ErrUnsupported` for a type the switch does not know. The
  second takes any backing `Decimal` takes.
- **`EuclideanDivMod`**, **`EuclideanMulDivMod`**: division helpers
  correcting the remainder into `[0, |divisor|)`, constrained on
  `Euclidean`; `SignedEuclidean` combines it with `Signed`. A `Decimal`
  backing needs only `Signed`: `Decimal` finds the backing's unit and
  its `Int128` widening through a type switch over the package's
  integers. Any other backing takes the scale divided by itself for its
  unit, and its widening falls back to a read one bit at a time, by
  division by a two built from the value divided by itself, so neither
  boxes the backing.
- **`Pow10`**: the powers of ten a `Uint128` holds, 10^0 to 10^38, read
  by exponent from a table.
- **`ErrDivZero`**: the division-by-zero panic value, wrapping
  `core.ErrInvalid`.
- **`ErrPow10Range`**: the panic value of `Pow10` for an exponent
  outside the table.
- **`ErrSyntax`**, **`ErrRange`**: the text-parsing errors, each a
  `core.QuietWrap` over a `core.CompoundError` holding the `strconv`
  sentinel of the same name and `core.ErrInvalid`, so `errors.Is`
  matches either.
- **`ParseError`**: the parse report, a defined type over
  `strconv.NumError` with the package's own text; `Unwrap` returns the
  cause, `ErrSyntax` or `ErrRange` from the package's parsers and one
  matching `core.ErrInvalid` from `AsParseError`. `errors.As` matches
  `*ParseError`, not the strconv type.
- **`AsParseError`**: the report builder, for the package's parsers and
  for one outside it: a function name, the text and the failure, which
  may be strconv's own `NumError` or a `ParseError`, opened: its cause
  is reported, `core.ErrInvalid` when it carries none, and its function
  and text stand in for an empty `fn` or `s`. The cause comes out as the
  package's own, strconv's sentinels translated, an error already
  matching `core.ErrInvalid` kept and anything else compounded with
  it. Nil in, a typed nil included, is nil out.
- **`ParseInt32`**, **`ParseInt64`**: the text parsers of the native
  integers, `strconv.ParseInt` at the type's width with the failure
  translated into a `ParseError`; `UnmarshalText` on the pointer
  stores their value.
- **`ParseUint128`**, **`ParseInt128`**: the text parsers of the
  128-bit integers, the digits over `strconv.ParseUint` a group at a
  time, with the optional sign on `Int128` alone and `UnmarshalText`
  on the pointer likewise.

Files:

- `num/atto128.go`: the `Atto128` instantiation and its scale.
- `num/const.go`: word primitives, the fixed-point scale factors, the
  table of powers of ten behind `Pow10` and the sentinel bounds
  (`MaxUint128`, `MinInt128`, …).
- `num/convert.go`: `NewFromInt128` and `NewDecimalFromInt128`, with
  the range rule under them, a conversion's value when it fitted and
  the bound on the side of the input with `ErrRange` when it did not;
  then the `wide` intermediate the conversions share, which rescales a
  count between resolutions and narrows it into the target. Each
  type's `wide` and conversion methods sit in its own file.
- `num/decimal.go`: `Decimal`, its constructors and its methods.
- `num/doc.go`: package documentation.
- `num/errors.go`: `ErrDivZero`, `ErrPow10Range`, `ErrSyntax`,
  `ErrRange` and `ParseError`.
- `num/euclidean.go`: the `Euclidean` and `SignedEuclidean` constraints
  and the Euclidean division helpers.
- `num/format.go`: the shared side of `Format` and `GoString`, the verb
  tables, the sign, prefix and width padding, the base-10 digit group
  constants and the thousands grouping; each type's `Format`,
  `GoString`, `String` and digit generation sit in its own file.
- `num/int128.go`: `Int128` and its operations.
- `num/int32.go`: `Int32` and its operations.
- `num/int64.go`: `Int64` and its operations.
- `num/json.go`: the quoted form and the two bounds under which
  `MarshalJSON` emits a number; each type's `MarshalJSON` sits in its
  own file.
- `num/milli.go`: the `Milli32` and `Milli64` instantiations and their
  scales.
- `num/num.go`: the `Unsigned`, `Signed` and `Number` constraints and
  the `DecimalScaler` interface.
- `num/parse.go`: the shared side of the parsers, the sign split, the
  digit groups behind `doParseUint128`, `AsParseError` with the cause
  translation under it and the store step of the unmarshalers; each
  type's parser and `UnmarshalText` sit in its own file.
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
- The conversions between the types are methods named for the target,
  `Int32()` through `Atto128()` on every type, each `(T, bool)`, and
  they keep the value where `As` keeps the count: `AsMilli32(1500)` is
  1.5 and `AsInt32(1500).Milli32()` is 1500.0. Fraction digits below
  the target's resolution drop towards zero with the flag true; the
  flag is false only when the whole units do not fit, the result then
  keeping the low bits, so a negative into `Uint128` is its bit
  pattern. `Decimal` alone has `AsInt32`, `AsInt64` and `AsInt128`,
  the count with a size check, the inverses of its constructor; they
  stay off `Number`, as does `sys`. Every conversion takes one path:
  the receiver widens into a `wide`, an `Int128` count and the
  exponent of its scale, `at` rescales it by the power of ten between
  the two exponents, multiplying towards a finer resolution and
  detecting the wrap by comparing the magnitude against `pow10Bound`,
  or dividing towards a coarser one, and the target narrows it with a
  fit check. `Uint128` widens with the flag already clear when its top
  bit is set, since the bits read as a negative `Int128` from then on.
  `TestConvert` pins every cell of the matrix by hand, the flag declared
  per row, and asserts the round trip where nothing truncates;
  `TestCount` pins the count accessors. A new type adds its `wide`, its
  seven methods, a `wide` method named for it, and a row per cell, then
  its arm in the switch of `NewFromInt128` with its rows in
  `TestNewFromInt128`.
- `GoString` prints the constructor call that rebuilds the value, the
  `As` count form while the value fits the native word and the `New`
  words form in hex beyond it; a `Decimal` prints both parts with its
  sign, so the call holds under either sign rule of the constructor.
  `DecimalScaler`'s `Name` carries the qualified name, `num.Milli32`,
  from which `GoString` derives the `New` and `As` constructors after
  the qualifier and `Format` its bad-verb label; the constraints stay
  free of formatting methods, and the `Decimal` fallback reaches its
  backing's form through `%#v`.
  A new type or instantiation adds a row to the `GoString` table.
- `Format` owns every verb, since fmt consults nothing else once a type
  has it: `%#v` is routed to `GoString` by hand. `Uint128` generates
  the digits, base 10 in groups of 19 divided out by 10^19 and the
  power-of-two bases by shifting the words; `Int128` prints its sign
  and hands the magnitude over; `Int32` and `Int64` hand the native
  value to fmt with `fmt.FormatString`, under the verb they were given
  rather than a decimal rewrite of it, so their flags cannot drift from
  fmt's; `Decimal` prints its parts as unsigned 128-bit magnitudes one
  at a time, as `Int128` takes its own, which keeps the backing's
  minimum where `Abs` wraps and holds a fraction past an int64.
  Fraction rounding is half away from zero; the precision of `%f` is
  fmt's, not the resolution's. Never reach for `math/big` for any of
  this.
- `String` returns the `%v` text over the same digit generation, for
  the callers that ask for it by name; fmt never does, since a
  `Formatter` takes precedence over a `Stringer`. The primitive under
  both is the unexported `doAppendText` of each type; `AppendText` is
  its exported form behind an always-nil error, and `MarshalText` is
  `AppendText(nil)`. A `Decimal` writes its whole count through the
  `doAppendText` of `Uint128`, so the count never goes through fmt.
  `TestText` checks the four agree on every row and that `AppendText`
  allocates nothing into a buffer with room.
- `MarshalJSON` is the `MarshalText` text, bare while a `float64`
  consumer reads the value back safely and quoted beyond that: an
  integer at a magnitude of at most 2^53 through `Int64()`, a
  `Decimal` at a count and a scale both below 10^15, each read through
  its `Int128` widening, which makes a `Milli32` always a number and an
  `Atto128` never one, so its field type stays stable. Exactness in a
  `float64` is the wrong test, since 2^60 is exact and 2^60+1 is not.
  `TestJSON` reads each result back through the standard decoder to
  check the token kind.
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
- Every sentinel is a `core.QuietWrap` carrying a `num:`-prefixed text
  and matching `core.ErrInvalid`; the parsing pair also matches its
  `strconv` counterpart through a `core.CompoundError`. Division by
  zero panics with `ErrDivZero`. `Pow10` takes an exponent outside its
  table as the caller's mistake and panics with `ErrPow10Range` through
  `core.PanicFrom`, so the stack starts at the caller. Parsing returns
  a `ParseError` carrying `ErrSyntax` or `ErrRange`. Arithmetic wraps
  on overflow and the constructors never fail.
- The parsers behave as strconv does wherever strconv fits: the whole
  input is the number, base 10 only, no underscores, prefixes or
  spaces, a syntax failure returns zero and a range failure the
  nearest bound. The native integers hand the text to
  `strconv.ParseInt` at their width and pass its `NumError` through
  `AsParseError` under the bare parser name, `ParseInt32`, which opens
  it and reports its sentinel as the package's own. The 128-bit
  integers keep to strconv's shape, with the unexported
  `doParseUint128` as the primitive under both, speaking strconv's own
  errors: `ParseUint128` wraps its answer under its name and takes no
  sign; `ParseInt128` splits the sign, reads the magnitude at the full
  unsigned width through it and then holds it to the sign's bound, the
  `Int128()` conversion for a positive value and one further, 2^127,
  for a negative one so `MinInt128` parses back, as `strconv.ParseInt`
  does over `ParseUint`. The magnitude is read in groups of 19 digits
  cut from the right, each read by `strconv.ParseUint`, which is the
  digit check, and added to the magnitude read so far, multiplied by
  10^19, with the overflow caught in a 256-bit product. A text both
  malformed and too long reports the failure strconv meets first,
  reading one digit at a time: the range once the digits before the
  bad byte pass 128 bits, the syntax otherwise. `UnmarshalText` parses
  first and stores second, so a bad text reports its own failure
  before a nil receiver reports `core.ErrNilReceiver`; both come back
  through `core.Wrap` with the method's name in front, so `errors.As`
  still finds the one `ParseError`, and a failed call leaves the
  receiver as it was. `TestParseMatchesStrconv` runs every text of
  `parseCorpus` against strconv: the natives against
  `strconv.ParseInt` at their width, comparing the value and the class
  of failure, and the 128-bit integers against the 64-bit parser of
  their kind, `strconv.ParseUint` or `strconv.ParseInt`, comparing the
  syntax verdict, the value while strconv has one, and on its range
  failure a value past the bound strconv clamps to, on the same side;
  extend the corpus rather than hand-write an expectation.
  `TestParseRoundTrip` parses the text tables of `TestText` back, so a
  new type that prints also reads.
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
