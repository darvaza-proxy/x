# Agent Documentation for x/text

## Overview

`darvaza.org/x/text` hosts shared text-processing primitives. Subpackages
are added as concrete users surface them; the module starts small and
grows by demand.

For detailed API documentation and usage examples, see [README.md](README.md).

## Subpackages

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

## Architecture Notes

* **Encoding is hidden.** Callers see runes and strings, never bytes.
  Adding byte-level escape hatches is deferred until a real caller
  needs one.
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
