# `darvaza.org/x/text`

[![Go Reference][godoc-badge]][godoc-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/text.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/text
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=text
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/text
[socket-link]: https://socket.dev/go/package/darvaza.org/x/text

## Overview

`darvaza.org/x/text` hosts shared text-processing primitives. Subpackages
land here as concrete users surface them.

## Subpackages

### `buffer`

A chainable wrapper around `strings.Builder` for the "write many,
export once" pattern used by lexers, parsers, and similar text
emitters.

* `Buffer` — wrapper type. Chainable helpers (`WriteStrings`,
  `WriteRunes`, `WriteBytes`, `Print`, `Println`, `Printf`, `Grow`,
  `Reset`) discard errors and return `*Buffer` for fluent
  composition.
* Standard interfaces (`io.Writer`, `io.WriterTo`, `io.StringWriter`,
  `io.ByteWriter`) keep their stdlib signatures, so the buffer plugs
  directly into `fmt.Fprintf`, `io.Copy`, and similar.
* `Bytes()` exposes a non-nil byte view aliasing the internal
  storage. Read-only — mutating it also mutates every previously
  returned `String()`. Copy out (`append([]byte(nil), buf.Bytes()...)`)
  before handing to anything that may mutate.
* `WriteTo(w)` drains the buffer on success (storage is discarded,
  ready for reuse). On error it leaves the buffer intact so the
  caller can retry or inspect the unsent payload.
* `New(capacity int) *Buffer` — factory with pre-allocated storage.

```go
var buf buffer.Buffer
buf.WriteStrings(`{"id":`).Printf("%d", 42).WriteRunes('}')
fmt.Println(buf.String())
// {"id":42}
```

Single-use semantics: build, export, discard. Inherits
`strings.Builder`'s no-copy contract.

### `lexer`

A small toolkit for hand-written state-function parsers:

* `Cursor` — a UTF-8-aware read cursor over a string source with an emit
  buffer. Encoding is hidden: the API speaks runes and strings.
* `StateFn[P]` and `Run[P]` — a generic state-function machine that
  threads a caller-defined parser state through every transition.

Typical use embeds `*lexer.Cursor` in the caller's parser state and lets
the state functions share the scanning primitives:

```go
type parser struct {
    *lexer.Cursor
    // ...additional state
}

func stateStart(p *parser) (lexer.StateFn[*parser], error) {
    r, ok := p.Peek()
    if !ok {
        return nil, nil
    }
    // ...
}

err := lexer.Run(&parser{Cursor: lexer.New(line)}, stateStart)
```

### `versionsort`

Orders strings by version: every run of decimal digits compares by the
number it writes, so `ttyS10` follows `ttyS9`, and the text between
compares a character at a time.

* `Compare[S core.String](a, b S) int` — the comparison, ready for
  `slices.SortFunc` and usable on any type over `string`. Strings that
  tie on every rule, such as `ttyS1` and `ttyS01`, or `tty` and `tty0`,
  fall back to their bytes, so only equal strings compare as 0.

```go
names := []string{"eth10", "eth2", "enp10s0", "enp2s0"}
slices.SortFunc(names, versionsort.Compare)
// [enp2s0 enp10s0 eth2 eth10]
```

Text ranks, from first to last:

* A tilde, ahead of even the end of the string, so `1.0~rc1` sorts
  before `1.0`.
* The end of the string, and a digit, so `1.0` sorts before `1.0a` and
  `tty1` before `ttyS0`.
* Letters, by code point. Letters are Unicode letters, so `é` sorts
  before `-`.
* Every other character, by code point, so `dma_heap` sorts before
  `dm-0`.
* A byte that is not valid UTF-8, by value.

Numbers compare by value, however long, and a string that ends where the
other goes on with a number counts as having a zero there. Digits are
ASCII only.

The empty string sorts first, ahead of a tilde. Strings are not
normalised: a letter followed by a combining accent sorts apart from the
same accented letter written as one character. Dots and file suffixes
are ordinary characters.

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
