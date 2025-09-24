# `darvaza.org/x/time`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/time.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/time
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/time
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/time
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=time
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/time
[socket-link]: https://socket.dev/go/package/darvaza.org/x/time

## Overview

`darvaza.org/x/time` hosts time-related primitives that don't belong in
the Go standard library. Subpackages land here as concrete needs surface.

## Subpackages

### `tai`

TAI (International Atomic Time) timestamp implementations with varying
precision levels that follow the Go time package API conventions.

#### Types

- `TAI` - TAI timestamp with second precision (TAI64)
- `TAIN` - TAI timestamp with nanosecond precision (TAI64N)
- `TAINA` - TAI timestamp with attosecond precision (TAI64NA)
- `AttosecondTimestamp` - Split representation of high-precision timestamps

#### Key Features

- Compatible with Go's `time` package API
- Handles leap seconds correctly
- JSON, binary, and text marshaling/unmarshaling
- Full arithmetic operations (Add, Sub, etc.) with overflow detection
- High-precision attosecond arithmetic (AddAttoseconds, SubAttoseconds)
- Split timestamp representation to avoid overflow issues
- Comparison operations (Before, After, Equal, Compare)
- Conversion to/from Go's `time.Time`
- String formatting using TAI64/TAI64N/TAI64NA format

#### Usage Examples

```go
package main

import (
 "fmt"
 "time"

 "darvaza.org/x/time/tai"
)

func main() {
 // Create TAI timestamps
 nowTAIN := tai.NowTAIN()   // Current time with nanoseconds (TAIN)
 nowTAINA := tai.NowTAINA() // Current time with attoseconds (TAINA)
 fmt.Println("now (TAIN):", nowTAIN)
 fmt.Println("now (TAINA):", nowTAINA)

 // From Go time
 goTime := time.Now()
 taiTime := tai.TAIFromTime(goTime)     // TAI type (seconds)
 tainTime := tai.TAINFromTime(goTime)   // TAIN type (nanoseconds)
 tainaTime := tai.TAINAFromTime(goTime) // TAINA type (attoseconds)
 fmt.Println("TAI:", taiTime, "TAIN:", tainTime, "TAINA:", tainaTime)

 // Constructors
 timestamp := tai.UnixTAIN(1234567890, 123456789)
 date := tai.DateTAIN(tai.DateConfig{
  Year: 2023, Month: time.May, Day: 15,
  Hour: 10, Min: 30, Sec: 45, Nsec: 0, Loc: time.UTC,
 })
 fmt.Println("from Unix:", timestamp, "from date:", date)

 // TAINA with attosecond precision
 tainaTimestamp := tai.UnixTAINA(1234567890, 123456789, 987654321)
 tainaDate := tai.DateTAINA(tai.DateConfig{
  Year: 2023, Month: time.May, Day: 15,
  Hour: 10, Min: 30, Sec: 45, Nsec: 123456789, Loc: time.UTC,
 }, 987654321)
 fmt.Println("TAINA from Unix:", tainaTimestamp, "TAINA from date:", tainaDate)

 // Conversion between types
 tai64 := nowTAIN.TAI()   // Convert TAIN to TAI (truncate)
 tain := taiTime.TAIN()   // Convert TAI to TAIN (zero nanos)
 taina := nowTAIN.TAINA() // Convert TAIN to TAINA (zero attos)
 fmt.Println("converted:", tai64, tain, taina)

 // Arithmetic (Add returns an error on overflow/underflow)
 later, err := nowTAIN.Add(time.Hour)
 if err != nil {
  fmt.Println("add failed:", err)
 }
 duration := later.Sub(nowTAIN)
 fmt.Println("one hour later:", later, "duration:", duration)

 // Attosecond arithmetic
 // Add 1 nanosecond worth of attoseconds
 laterAtto, err := nowTAINA.AddAttoseconds(1000000000)
 if err != nil {
  fmt.Println("add attoseconds failed:", err)
 }
 // Get the difference as split nanoseconds + attoseconds
 attoDiff := laterAtto.SubAttoseconds(nowTAINA)
 fmt.Printf("%d ns + %d as\n", attoDiff.Nanoseconds, attoDiff.Attoseconds)

 // Formatting
 fmt.Println(nowTAIN.String())             // TAI64N format
 fmt.Println(nowTAINA.String())             // TAI64NA format
 fmt.Println(nowTAIN.Format(time.RFC3339)) // Standard time format

 // Precision accessors
 fmt.Printf("Nanoseconds: %d\n", nowTAINA.Nanosecond())
 fmt.Printf("Attoseconds: %d\n", nowTAINA.Attosecond())

 // High precision timestamp representation
 splitTime := nowTAINA.UnixAttosecondSplit()
 fmt.Printf("Unix nanoseconds: %d\n", splitTime.UnixNanoseconds)
 fmt.Printf("Additional attoseconds: %d\n", splitTime.Attoseconds)

 // Conversion back to Go time (attoseconds lost in conversion)
 goTimeBack := nowTAINA.GoTime()
 fmt.Println("back to time.Time:", goTimeBack)
}
```

#### TAI vs UTC

TAI (International Atomic Time) is a continuous time scale that does not include
leap seconds, unlike UTC. This makes it useful for precise timing applications
where you need monotonic time that doesn't have discontinuities.

The package automatically handles the conversion between UTC and TAI using the
current leap second table.

#### Precision Levels

- **TAI64**: Second precision (8 bytes) - suitable for most applications
- **TAI64N**: Nanosecond precision (12 bytes) - compatible with Go's time.Time
- **TAI64NA**: Attosecond precision (16 bytes) - for ultra-high precision timing

Note: Go's `time.Time` only supports nanosecond precision, so attosecond
precision is lost when converting TAINA to `time.Time`.

#### Limitations

- `Sub`, `Since`, and `Until` return `time.Duration`, which overflows for
  spans over roughly ±292 years.
- `UnixNano` overflows `int64` for dates before the year 1678 or after
  2262; `UnixMicro` and `UnixMilli` have proportionally wider but still
  bounded ranges.
- Go's `time.Time` cannot represent an inserted leap second (23:59:60
  UTC); converting such a TAI instant with `GoTime` returns the first
  second after it.

#### High-Precision Timestamp Handling

For modern timestamps, calculating attoseconds as a single int64 value from the
Unix epoch would cause integer overflow. To solve this, the package provides
`AttosecondTimestamp`:

```go
type AttosecondTimestamp struct {
 UnixNanoseconds int64  // Nanoseconds since Unix epoch (from UnixNano())
 Attoseconds     uint32 // Additional attoseconds within that nanosecond
                        // [0, 999999999]
}
```

##### Type Choice Rationale

The `AttosecondTimestamp` uses different types for its fields for specific
technical reasons:

**UnixNanoseconds (int64):**

- Compatible with Go's standard `time` package (`time.Time.UnixNano()` returns
  `int64`)
- Can represent the full range of timestamps that Go's `time` package supports
- Matches the return type of `TAINA.UnixNano()` method
- Signed to handle dates before Unix epoch (negative values)

**Attoseconds (uint32):**

- Represents sub-nanosecond precision: 0 to 999999999 attoseconds
- A nanosecond contains exactly 1000000000 attoseconds (1e9)
- `uint32` can hold values up to 4294967295, which is sufficient for the
  range [0, 999999999]
- Matches the internal `atto` field type in the `TAINA` struct
- Unsigned because attoseconds are always a positive offset within a nanosecond
- Uses minimal memory (4 bytes) for the constrained range

**Why not both int64?** Using `int64` for attoseconds would waste memory since
the valid range [0, 999999999] fits comfortably in `uint32`, and consistency
with the internal `TAINA.atto` field (also `uint32`) simplifies the
implementation.

This split representation allows handling attosecond precision without overflow
issues:

```go
// Get split representation
timestamp := tai.NowTAINA()
split := timestamp.UnixAttosecondSplit()

fmt.Printf("Nanoseconds since epoch: %d\n", split.UnixNanoseconds)
fmt.Printf("Additional attoseconds: %d\n", split.Attoseconds)
```

#### Comparison with Original libtai

This Go implementation follows the same TAI64/TAI64N/TAI64NA specifications as
DJB's original libtai C library but provides several enhancements:

##### Similarities

| Feature | libtai (C) | darvaza.org/x/time/tai |
| --------- | ------------ | --------------------------- |
| **TAI64 Format** | 8 bytes, big-endian | ✅ Same format |
| **TAI64N Format** | 12 bytes (8+4), big-endian | ✅ Same format |
| **TAI64NA Format** | 16 bytes (12+4), big-endian | ✅ Same format |
| **Time Range** | Hundreds of billions of years | ✅ Same range |
| **Leap Second Handling** | Automatic UTC-TAI conversion | ✅ Same functionality |
| **High Precision** | 1-attosecond precision | ✅ Same precision |

##### Go Package Enhancements

| Feature | libtai (C) | darvaza.org/x/time/tai |
| --------- | ------------ | --------------------------- |
| **Type Safety** | Manual struct management | ✅ Strong typing with Go structs |
| **Memory Safety** | Manual memory management | ✅ Garbage collected |
| **API Design** | C-style functions | ✅ Go idiomatic methods |
| **Standard Library Integration** | Custom time functions | ✅ Compatible with Go's `time` package |
| **Serialization** | Manual byte manipulation | ✅ JSON/Binary/Text marshal/unmarshal |
| **Overflow Protection** | Limited int64 range | ✅ Split timestamp representation |
| **Error Handling** | Return codes | ✅ Go error interface |
| **String Formatting** | Custom sprintf usage | ✅ Implements `fmt.Stringer` |
| **Arithmetic Safety** | Manual overflow checks | ✅ Built-in overflow detection |

##### Performance Improvements

- **Zero Allocations**: Most operations avoid heap allocations
- **Efficient Conversions**: Optimised TAI-UTC conversion algorithms
- **Modern Architecture**: Designed for 64-bit systems and modern Go runtime

##### API Differences

**libtai approach (C):**

```c
struct tai t;
tai_now(&t);
tai_uint64(&t, seconds);
```

**Go package approach:**

```go
t := tai.NowTAI()
seconds := t.Unix()
```

The Go implementation prioritises type safety, idiomatic Go patterns, and
integration with the standard library while maintaining full compatibility with
the original TAI64/TAI64N/TAI64NA specifications.

#### Missing Features from Original libtai

While this Go package provides full TAI64/TAI64N/TAI64NA compatibility and
integrates well with Go's standard library, some specialised features from the
original libtai C library are not currently implemented. A future
`darvaza.org/x/time/calendar` subpackage may cover the calendar-date
functionality below; see [Subpackages](#subpackages) once it lands.

##### Calendar Date Utilities

**Missing Julian Day Functions:**

- `caldate_frommjd()` - Convert modified Julian day number to calendar date
- `caldate_mjd()` - Convert calendar date to modified Julian day number
- `caldate_normalize()` - Normalize out-of-range dates
- Weekday and yearday calculations

**Missing Calendar Formatting:**

- `caldate_fmt()` - Format dates in ISO style (YYYY-MM-DD)
- `caldate_scan()` - Parse dates from ISO style strings
- `caltime_fmt()` - Format calendar time with UTC offset
  (YYYY-MM-DD HH:MM:SS +OOOO)
- `caltime_utc()` - Convert TAI64 to calendar time with UTC offset

##### Data Structures

**Missing Calendar Types:**

- `struct caldate` - Calendar date representation (year, month, day)
- `struct caltime` - Calendar time with UTC offset

##### Design Philosophy

The Go package takes a different approach from libtai:

| Feature | libtai (C) | darvaza.org/x/time/tai |
| --------- | ------------ | --------------------------- |
| **Calendar Functions** | Separate caldate/caltime utilities | ✅ Uses Go's standard `time` package |
| **Date Formatting** | Custom ISO formatters | ✅ Uses Go's time formatting |
| **Julian Day Support** | Built-in MJD functions | ❌ Not implemented (see workarounds below) |
| **Date Normalization** | Built-in normalize function | ✅ Go's `time` package handles this |
| **UTC Offset Handling** | Custom caltime struct | ✅ Go's `time.Location` handles this |

##### Workarounds for Missing Features

For applications that need Julian day functionality, you can use Go's standard
library or third-party packages:

```go
// Example: Calculate day of year using Go's standard library
t := time.Now()
dayOfYear := t.YearDay()  // 1-365 or 1-366 for leap years
weekday := t.Weekday()    // Sunday = 0, Monday = 1, etc.

// Example: Date formatting using Go's time package
formatted := t.Format("2006-01-02 15:04:05 -0700")  // ISO style with offset
```

**For precise Julian day calculations**, consider:

- Third-party astronomy packages
- Manual calculation using the standard Gregorian calendar algorithms
- Integration with existing astronomical libraries

##### Future Considerations

If there's demand for these features, they could be added in future versions
while maintaining the Go-idiomatic API design. However, most use cases are
well-served by Go's standard `time` package combined with this TAI
implementation.

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).

## Dependencies

This module depends only on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## Licence

This project is licensed under the MIT Licence. See [LICENCE.txt](LICENCE.txt)
for details.
