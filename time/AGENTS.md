# Agent Documentation for x/time

## Overview

`darvaza.org/x/time` implements time-related primitives that don't belong in
the standard library.

## Development Notes

- Error handling follows `darvaza.org/core` conventions: sentinel
  errors wrap `core.ErrInvalid` via `core.QuietWrap`, context is
  added with `core.Wrap`/`core.Wrapf`, and constructor range
  violations panic via `core.PanicWrapf` with specific error types.
- Operations are designed to be zero-allocation where possible; avoid
  introducing allocations in hot paths.

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
