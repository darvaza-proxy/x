# Agent Documentation for x/time

## Overview

`darvaza.org/x/time` will host clock-related primitives.
Subpackages land here as concrete users surface them; the module
starts as scaffolding and grows by demand.

For detailed API documentation and usage examples, see [README.md](README.md).

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

* Standard library only.
* `darvaza.org/core`: pulled in once tests land.

## See Also

* [Package README](README.md) for the public API tour.
* [Root AGENTS.md](../AGENTS.md) for mono-repo overview.

[core-testing]: https://github.com/darvaza-proxy/core/blob/main/TESTING.md
