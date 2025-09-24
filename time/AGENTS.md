# Agent Documentation for x/time

## Overview

`darvaza.org/x/time` implements time-related primitives that don't belong in
the standard library.

## Subpackages

### `tai`

`darvaza.org/x/time/tai` implements TAI (International Atomic Time)
timestamps following the TAI64/TAI64N/TAI64NA specifications of DJB's
libtai, with an API modelled on Go's standard `time` package. TAI is a
continuous time scale without leap seconds, making it suitable for
applications that need monotonic timestamps without discontinuities.

The module depends only on `darvaza.org/core` and the Go standard
library.

For detailed API documentation and usage examples, see [README.md](README.md).

#### Types

- `TAI` — second precision (TAI64, 8 bytes).
- `TAIN` — nanosecond precision (TAI64N, 12 bytes).
- `TAINA` — attosecond precision (TAI64NA, 16 bytes).
- `AttosecondTimestamp` — split representation
  (`UnixNanoseconds int64` + `Attoseconds uint32`) that avoids `int64`
  overflow when expressing attoseconds since the Unix epoch.
- `DateConfig` — named-field alternative to `time.Date` arguments,
  used by the `DateTAIN` and `DateTAINA` constructors.

All three timestamp types provide:

- A uniform constructor family: `NowTAI`/`NowTAIN`/`NowTAINA`,
  `ParseTAI`/`ParseTAIN`/`ParseTAINA` (single string argument),
  `TAIFromTime`/`TAINFromTime`/`TAINAFromTime`, plus `UnixTAIN`,
  `UnixTAINA`, `DateTAIN`, and `DateTAINA`.
- Standard-library-style methods: `Add`, `Sub`, `Before`, `After`,
  `Equal`, `Compare`, `IsZero`, `Unix`, `UnixNano`, `Format`,
  `Truncate`, `Round`, `Since`, `Until`. All `Add` methods are
  overflow-checked, returning `ErrOverflow` or `ErrUnderflow`;
  `TAI.Add` truncates the duration to whole seconds.
- Conversion methods between all precisions (`TAI.TAIN`, `TAI.TAINA`,
  `TAIN.TAI`, `TAIN.TAINA`, `TAINA.TAI`, `TAINA.TAIN`) and to
  `time.Time` via `GoTime` (attoseconds are lost when converting to
  `time.Time`).
- `fmt.Stringer` using the external TAI64/TAI64N/TAI64NA hex format,
  plus JSON, binary, and text marshalling/unmarshalling.
- `TAINA` additionally offers attosecond arithmetic:
  `AddAttoseconds` with overflow protection, `SubAttoseconds`
  returning a split `AttosecondDuration` (a plain attosecond count
  would overflow `int64` beyond ±9.2 seconds), and
  `UnixAttosecondSplit` for the split timestamp representation.

#### Files

- `tai.go` — `TAI` type, constants (`TAICONST`, `TAILength`,
  `TAINLength`, `TAINALength`), and second-precision operations.
- `tain.go` — `TAIN` type, `DateConfig`, and nanosecond-precision
  operations.
- `taina.go` — `TAINA` and `AttosecondTimestamp` types with
  attosecond-precision operations.
- `leapsecs.go` — leap second table (as `int64` Unix timestamps) and
  the `lsoffset` UTC→TAI conversion helper.
- `*_test.go` — table-driven tests, including dedicated overflow
  coverage in `taina_overflow_test.go`.
- `math.go` — shared 128-bit arithmetic helpers (`toNanos128`,
  `sub128`, `durationFromSigned128`, `floorDivMod`, etc.) backing
  overflow-safe `Add`/`Sub`/`Truncate`/`Round`.
- `errors.go` — sentinel error definitions shared across `TAI`,
  `TAIN`, and `TAINA`.

## Development Notes

- Error handling follows `darvaza.org/core` conventions: sentinel
  errors wrap `core.ErrInvalid` via `core.QuietWrap`, context is
  added with `core.Wrap`/`core.Wrapf`, and constructor range
  violations panic via `core.PanicWrapf` with
  `ErrInvalidAttosecondRange`.
- UTC↔TAI conversion relies on the leap second table in
  `leapsecs.go`; the table must be updated whenever IERS announces a
  new leap second. `utcFromTAI` evaluates the offset at the UTC
  instant (with one refinement step); TAI labels inside an inserted
  leap second map to the first second after it.
- `Sub`/`Since`/`Until` are bounded by the `time.Duration` range
  (about ±292 years) and the `Unix*` accessors by `int64`; spans and
  dates beyond that overflow.
- The wire format is big-endian and byte-compatible with libtai;
  `MarshalBinary`/`UnmarshalBinary` must preserve this.
- Operations are designed to be zero-allocation where possible; avoid
  introducing allocations in hot paths such as `lsoffset`.
- Attosecond arithmetic must use the split representation to avoid
  `int64` overflow; see the "High-Precision Timestamp Handling"
  section of [README.md](README.md).

## Testing Patterns

Tests follow the conventions in [core's TESTING.md][core-testing]:

- `var _ core.TestCase = ...` declarations for every TestCase type.
- Factory functions decouple semantic argument order from
  memory-aligned struct field order.
- Table-driven suites use `core.RunTestCases`; scenario tests use
  `TestFoo() { t.Run("scenario", runTestFooScenario) }`.

[core-testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md
