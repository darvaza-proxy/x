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

`darvaza.org/x/text` hosts shared text-processing primitives. The first
inhabitant is the [`lexer`](lexer) subpackage; further building blocks land
here as concrete users surface them.

## Subpackages

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
