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
the Go standard library. Subpackages are added as concrete needs surface.

## Packages

### `num`

Fixed-width integer types for the time packages: the signed integers
`Int32`, `Int64` and `Int128`, and the unsigned `Uint128`, behind the
generic `Unsigned` and `Signed` constraints.

Arithmetic wraps on overflow, matching Go's built-in integer
operators, so `Add`, `Sub` and `Mul` never panic. Division by zero
panics with `ErrDivZero`, which wraps `core.ErrInvalid`; signed
division truncates towards zero, with the remainder taking the sign
of the dividend. `MulDivMod` fuses a multiply and a divide, forming
the product in an intermediate wide enough that it cannot overflow
before the division.

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
