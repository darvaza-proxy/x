# `darvaza.org/x/time`

<!-- cspell:words Marshaler -->

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
the Go standard library. Subpackages are added as concrete needs surface.

## Packages

### `num`

Fixed-width numeric types for the time packages: the signed integers
`Int32`, `Int64` and `Int128`, and the unsigned `Uint128`, behind the
generic `Unsigned` and `Signed` constraints, plus the signed
fixed-point `Decimal`, whose `Milli32`, `Milli64` and `Atto128`
instantiations count sub-units at milli (10^-3) and atto (10^-18)
resolution.

Two constructor prefixes cover every type. `New` builds a value from
its parts, the two words of a 128-bit integer or the whole units and
sub-units of a `Decimal`; `As` takes a value that already is the count
and changes only its type, a native integer extended into a wider one
or a backing integer read as a `Decimal` at its resolution. So
`NewAtto128(1, 500e15)` and `AsAtto128(AsInt128(1500e15))` are both
1.5. `AsInt32` and `AsInt64` are the conversions of the native types
under the same name; `Int32` and `Int64` have no parts, so no `New`.

Every type converts to every other through a method named for the
target, `Int64` or `Atto128`, returning the value and whether it
fitted. The value is kept, not the count: an integer becomes whole
units and a `Decimal` is rescaled, so `AsInt32(5).Atto128()` is 5.0
and `NewMilli32(1, 500).Int64()` is 1, the fraction digits below the
target's resolution dropped towards zero. The flag is false only when
the whole units do not fit the target, the result then keeping the low
bits as a Go conversion does. A `Decimal` reads its count back through
`AsInt32`, `AsInt64` and `AsInt128`, the inverses of its `As`
constructor, each with a size check. `Number` names this whole
surface, the constraint for code generic over the family.

Under `%#v` every type prints as the call that rebuilds it, so a value
dumped from a failing test pastes back into a row: `num.AsInt32(-5)`,
`num.AsInt128(-42)`, `num.NewMilli32(1, 500)`. The `As` form gives way
to the `New` form over the two words in hex once a value no longer fits
the native word, and a `Decimal` whose whole count no longer fits an
`int64` prints as `As` over its backing integer.

Every other verb goes through `fmt.Formatter`, so the types print under
`fmt` the way its own numbers do. The integers take `d`, `v` and `s` for
decimal, `x` and `X` for hex, `o` and `O` for octal and `b` for binary,
with the `+`, space, `#`, `-` and `0` flags, width and precision as
`fmt` defines them for an integer, the `+` of `%+v` included, which
asks for the field names of a struct and so adds no sign. A `Decimal`
takes `v` and `s` at full resolution, `1.500` for a `Milli32`, and `f`,
or `F` under another name, with the fraction digits the precision asks
for, six without one, zero-filled past the resolution and rounded half
away from zero below it; `#` keeps the point a zero precision would
drop. A verb a type does not take prints in the `%!verb(type=value)`
form. `String` returns the text of `%v`, so every type is a
`fmt.Stringer` as well; `AppendText` writes it into a caller's buffer,
allocating nothing when the buffer has room, and `MarshalText` returns
it, so every type is an `encoding.TextAppender` and an
`encoding.TextMarshaler` too.

Arithmetic wraps on overflow, matching Go's built-in integer
operators, so `Add`, `Sub` and `Mul` never panic. Division by zero
panics with `ErrDivZero`, which wraps `core.ErrInvalid`; signed
division truncates towards zero, with the remainder taking the sign
of the dividend. `MulDivMod` fuses a multiply and a divide, forming
the product in an intermediate wide enough that it cannot overflow
before the division. `EuclideanDivMod` and `EuclideanMulDivMod`
instead keep the remainder non-negative.

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
