# `darvaza.org/x/text`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/text.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/text
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/text
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/text
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
