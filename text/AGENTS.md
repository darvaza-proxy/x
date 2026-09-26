# Agent Documentation for x/text

## Overview

`darvaza.org/x/text` hosts shared text-processing primitives. Subpackages
are added as concrete users surface them; the module starts small and
grows by demand.

For detailed API documentation and usage examples, see [README.md](README.md).

## Subpackages

### `buffer`

A chainable wrapper around `strings.Builder` for write-many-then-
export-once text assembly.

* `Buffer` — wrapper type. Chainable helpers (`WriteStrings`,
  `WriteRunes`, `WriteBytes`, `Print`/`Println`/`Printf`, `Grow`,
  `Reset`) discard errors and return `*Buffer` for fluent
  composition. Standard interfaces (`io.Writer`, `io.WriterTo`,
  `io.StringWriter`, `io.ByteWriter`) keep their stdlib signatures.
* `New(capacity int) *Buffer` — factory with pre-allocated storage.

Files:

* `buffer/doc.go` — package overview.
* `buffer/buffer.go` — type, methods, factory.
* `buffer/buffer_test.go` — external API tests (`package buffer_test`).
* `buffer/buffer_internal_test.go` — covers the unexported `sys()`
  accessor (`package buffer`).

### `lexer`

A small toolkit for hand-written state-function parsers.

* `Cursor` — UTF-8-aware read cursor over a string source with an emit
  buffer. The API is rune- and text-shaped; encoding is an implementation
  detail.
* `StateFn[P]` / `Run[P]` — generic state-function machine driver. The
  parser-state type `P` is threaded through every transition, allowing
  callers to embed `*Cursor` in their own struct and share the scanning
  primitives.

Files:

* `lexer/doc.go` — package overview.
* `lexer/cursor.go` — the cursor and emit-buffer primitives.
* `lexer/run.go` — the state-function machine driver.
* `lexer/cursor_test.go` and `lexer/run_test.go` — table-driven tests
  via `core.TestCase` plus scenario tests via `t.Run`.

### `versionsort`

The version string comparison, and sorts built on it.

* `Compare[S core.String](a, b S) int` — digit runs compare by the number
  they write, however long, and a string that ends where the other goes
  on with a number counts as having a zero there. Text compares a
  character at a time: a tilde first, ahead of the end of the string;
  then the end and a digit; then Unicode letters and then every other
  character, each by code point; invalid UTF-8 bytes last. Digits are
  ASCII only, strings are not normalised, the empty string sorts first,
  and strings that tie on every rule fall back to their bytes.
* `Sort[S core.String](a []S)` — sorts a slice of strings in place with
  `Compare`.
* `SortBy[S core.String, T any](a []T, fn func(T) S)` — sorts any slice
  in place by the string `fn` gives each element; a stable sort, calling
  `fn` once per element. A nil `fn` leaves the slice as it is.

Files:

* `versionsort/doc.go` — package overview.
* `versionsort/versionsort.go` — the comparison and the sorts.
* `versionsort/versionsort_test.go` — rows per rule for `Compare`, then
  rows for `Sort` and `SortBy`: nil and empty slices, one element,
  duplicates, empty strings, a mixed list of device and interface names,
  a nil `fn`, and a list long enough to tell `SortBy`'s stable sort from
  an unstable one, each checked for the order it leaves and for not
  panicking. Further tests use a named type over `string` with each
  function, check that an unstable sort does break that list's order,
  and count `SortBy`'s calls to `fn`.

## Architecture Notes

* **Single-use buffer.** `buffer.Buffer` exists for the write-many-
  then-export-once pattern. Lexers, parsers, and similar emitters
  embed it (or use `buffer.New`) rather than rolling local
  `strings.Builder` usage. Inherits `strings.Builder`'s no-copy
  contract.
* **Typed-nil defence.** `buffer.Buffer.sys()` keeps an explicit
  nil-receiver check so a `(*Buffer)(nil)` does not silently convert
  to `(*strings.Builder)(nil)` and leak across `interface{}`
  boundaries. The check is load-bearing — do not strip as
  "redundant".
* **Encoding is hidden.** Callers of `buffer` and `lexer` see runes and
  strings, never bytes. Adding byte-level escape hatches is deferred
  until a real caller needs one.
* **Generic state-function machine.** `StateFn[P]` keeps the caller's
  parser state type opaque to the package while still allowing free
  functions (rather than methods) to serve as states.

## Development Commands

For common development commands and workflow, see the
[root AGENTS.md](../AGENTS.md).

## Testing Patterns

Tests follow the conventions in [core's TESTING.md][core-testing]:

* `var _ core.TestCase = ...` declarations for every TestCase type.
* Factory functions decouple semantic argument order from
  memory-aligned struct field order.
* Table-driven suites use `core.RunTestCases`; scenario tests use
  `TestFoo() { t.Run("scenario", runTestFooScenario) }`.

## Dependencies

* `darvaza.org/core`: For test utilities and common helpers.
* Standard library only otherwise.

## See Also

* [Package README](README.md) for the public API tour.
* [Root AGENTS.md](../AGENTS.md) for mono-repo overview.

[core-testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md
