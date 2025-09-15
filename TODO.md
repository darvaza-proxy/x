# Migration from Testify to Core.Assert

<!-- cspell:ignore testutils -->

## Overview

This document outlines the plan to migrate the `darvaza.org/x` monorepo from
`github.com/stretchr/testify` to our internal `darvaza.org/core` assertion
functions, following the testing guidelines established in
`darvaza.org/core/TESTING.md`.

## Migration Progress Summary

| Package | Status | Coverage | Notes |
|---------|--------|----------|-------|
| `cmp/` | ✅ **COMPLETED** | 100.0% | Fully migrated to core assertions |
| `container/set/` | ✅ **COMPLETED** | 100.0% | Fully migrated with comprehensive tests |
| `container/list/` | 🎯 **NEXT** | 21.4% | Needs edge case tests |
| `container/slices/` | ⏳ Pending | 32.9% | Needs panic and edge case tests |
| `sync/` | ⏳ Pending | TBD | 464 assert calls to migrate from testify |
| `config/` | ⏳ Pending | TBD | Indirect dependency only |
| Other packages | ✅ Already using core/stdlib | - | No migration needed |

## Current State

### Packages Still Using Testify

- **`sync/`**: 9 test files across multiple sub-packages.
  - Heavy use of `assert.True`, `assert.False`, `assert.Nil`.
  - Extensive error checking with `assert.Error` and `assert.ErrorIs`.
  - Concurrent testing patterns.
- **`config/`**: Indirect dependency only.

### Packages Migrated to Core

- **`cmp/`**: ✅ Fully migrated with 100% coverage.
- **`container/set/`**: ✅ Fully migrated with 100% coverage.
- **`container/list/`**: 🎯 Next target, partial migration.
- **`container/slices/`**: Partial migration, needs completion.
- **`fs/`**: Already follows core testing patterns.
- **`net/`**: Uses standard library testing.
- **`web/`**: Uses standard library testing.

## Migration Strategy

### Phase 1: Preparation ✅ COMPLETED

1. ✅ Review and understand core testing patterns.
2. ✅ Identify testify-specific patterns requiring adaptation.
3. ✅ Create migration checklist for each package.

### Phase 2: Package-by-Package Migration

Migration order prioritised by complexity and dependencies:

1. ✅ `cmp` package (COMPLETED - simpler assertion patterns).
2. `container` package next (stdlib to core migration, no testify).
3. `sync` package third (complex concurrent tests, testify migration).
4. Remove testify from go.mod files after sync completion.

## Phase 1 Analysis Results

### Test Structure Issues

1. **Table-driven tests**: Currently using anonymous struct slices instead of
   TestCase.
2. **Complexity suppression**: `//revive:disable:cognitive-complexity` used.
3. **Message format**: Full sentences instead of short prefixes.
4. **No TestCase pattern**: None implement core.TestCase interface.

## Migration Patterns Learned from CMP

### 1. CMP Package ✅ COMPLETED (100% coverage achieved)

Successfully migrated from testify to core assertions with:

- TestCase pattern implementation for all table-driven tests
- Extraction of anonymous functions exceeding 3 lines
- Removal of all complexity suppressions
- Complete test coverage for all functions

### Common Patterns to Apply

#### Simple Assertions

```go
// Before (testify)
assert.Equal(t, expected, actual, "message")
assert.True(t, condition, "message")
assert.False(t, condition, "message")
assert.Nil(t, value, "message")
assert.NotNil(t, value, "message")
assert.Error(t, err, "message")
assert.ErrorIs(t, err, target, "message")
assert.Panics(t, func() { ... }, "message")

// After (core)
core.AssertEqual(t, expected, actual, "prefix")
core.AssertTrue(t, condition, "prefix")
core.AssertFalse(t, condition, "prefix")
core.AssertNil(t, value, "prefix")
core.AssertNotNil(t, value, "prefix")
core.AssertError(t, err, "prefix")
core.AssertErrorIs(t, err, target, "prefix")
core.AssertPanic(t, func() { ... }, "prefix")
```

#### Key Differences

- **Message format**: Testify uses complete messages; core uses prefixes.
- **Return values**: Core assertions return boolean results.
- **No require equivalent**: Use `AssertMust*` for fatal assertions.

## 2. CONTAINER Package ✅ COMPLETED (2025-01-11)

Successfully migrated from stdlib to core assertions with:

- 222 core assertions across all test files
- All stdlib t.Error/t.Fatal calls removed
- Proper use of core.AssertPanic for panic tests
- Example files preserved unchanged

### Files Migrated

- `list/list_test.go` - 17 assertions
- `list/list_more_test.go` - 53 assertions
- `set/set_test.go` - 76 assertions
- `set/set_simple_test.go` - 8 assertions
- `slices/set_ordered_test.go` - 13 assertions
- `slices/set_fn_test.go` - 55 assertions

### 3. SYNC Package Migration (DEFERRED)

**Current Status:**

- 9 test files across multiple subpackages
- 464 testify assert calls to migrate
- Still has `github.com/stretchr/testify v1.10.0` dependency

#### SYNC Test Files to Migrate

- [ ] `cond/barrier_test.go`
- [ ] `cond/count_test.go`
- [ ] `cond/utils_test.go`
- [ ] `mutex/mutex_test.go`
- [ ] `mutex/utils_test.go`
- [ ] `semaphore/semaphore_test.go`
- [ ] `spinlock/spinlock_test.go`
- [ ] `workgroup/workgroup_test.go`
- [ ] `once_test.go` (root sync package)

#### Special Considerations

- Concurrent testing patterns requiring careful migration
- Race condition tests that must be preserved
- Timeout and synchronisation assertions
- More complex than CMP due to concurrent test scenarios

### 3. Test Structure Refactoring

#### Implement TestCase Pattern Where Appropriate

For table-driven tests with multiple data variations:

1. Define TestCase types implementing `core.TestCase`.
2. Add interface validation: `var _ core.TestCase = myTestCase{}`.
3. Create factory functions: `newMyTestCase(...)`.
4. Use `core.RunTestCases(t, cases)`.

#### Extract Named Test Functions

For complex test scenarios:

```go
// Before
t.Run("complex test", func(t *testing.T) {
    // Long test logic
})

// After
func runTestComplexScenario(t *testing.T) {
    t.Helper()
    // Test logic
}

func TestFeature(t *testing.T) {
    t.Run("complex test", runTestComplexScenario)
}
```

## Automation Helpers

### Safe Regex Replacements

```bash
# Simple assertions
s/assert\.Equal\(t, ([^,]+), ([^,]+), "([^"]+)"\)/core.AssertEqual(t, \1,
\2, "\3")/g
s/assert\.True\(t, ([^,]+), "([^"]+)"\)/core.AssertTrue(t, \1, "\2")/g
s/assert\.False\(t, ([^,]+), "([^"]+)"\)/core.AssertFalse(t, \1, "\2")/g
s/assert\.Nil\(t, ([^,]+), "([^"]+)"\)/core.AssertNil(t, \1, "\2")/g
s/assert\.NotNil\(t, ([^,]+), "([^"]+)"\)/core.AssertNotNil(t, \1, "\2")/g
s/assert\.Panics\(t, ([^,]+), "([^"]+)"\)/core.AssertPanic(t, \1, "\2")/g
```

### Manual Work Required

- TestCase interface implementation.
- Factory function creation.
- Message shortening.
- Complex test refactoring.
- Removing complexity suppressions.

## Risk Assessment

### Low Risk

- Simple assertion replacements.
- Import changes.
- Basic message updates.

### Medium Risk

- Table-driven test restructuring.
- Factory function creation.
- TestCase implementation.

### High Risk

- Concurrent test modifications (sync package).
- Complex test splitting for compliance.
- Race condition test preservation.

## Implementation Steps

### For Each Package

1. **Analyse Current Tests**
   - [ ] Identify assertion patterns used.
   - [ ] Determine if TestCase pattern is applicable.
   - [ ] List special testing requirements.

2. **Create Branch**
   - [ ] Branch: `refactor-{package}-remove-testify`.

3. **Update Imports**
   - [ ] Remove: `"github.com/stretchr/testify/assert"`.
   - [ ] Add: `"darvaza.org/core"`.

4. **Replace Assertions**
   - [ ] Convert assert.*to core.Assert*.
   - [ ] Update message strings to prefixes.
   - [ ] Handle return values where needed.

5. **Refactor Test Structure**
   - [ ] Apply TestCase pattern for table-driven tests.
   - [ ] Extract named functions for complex tests.
   - [ ] Ensure compliance with TESTING.md guidelines.

6. **Validate**
   - [ ] Run tests: `go test ./...`.
   - [ ] Check coverage: `go test -cover ./...`.
   - [ ] Verify linting: `golangci-lint run`.

7. **Update Dependencies**
   - [ ] Remove testify from go.mod.
   - [ ] Run `go mod tidy`.

## Verification Checklist

### Per-File Checklist

- [ ] All testify imports removed.
- [ ] All assertions converted to core.
- [ ] Test messages converted to prefixes.
- [ ] TestCase pattern applied where appropriate.
- [ ] Factory functions created for TestCase types.
- [ ] Anonymous functions ≤3 lines.
- [ ] Complexity limits met (cognitive ≤7, cyclomatic ≤10).

### Per-Package Checklist

- [ ] All tests passing.
- [ ] Coverage maintained or improved.
- [ ] Linting passes without suppressions.
- [ ] go.mod updated and tidied.
- [ ] No remaining testify references.

## Benefits of Migration

1. **Consistency**: Uniform testing patterns across all darvaza.org projects.
2. **Compliance**: Tests meet strict linting requirements.
3. **Maintainability**: Simpler, more predictable test code.
4. **Performance**: Core assertions optimised for our use cases.
5. **Reduced Dependencies**: One less external dependency.

## Timeline

- **Phase 1 (Preparation)**: ✅ COMPLETED
- **Phase 2 (CMP Package)**: ✅ COMPLETED (2025-01-10)
- **Phase 3 (CONTAINER Package)**: 2-3 days (NEXT)
- **Phase 4 (SYNC Package)**: 3-4 days
- **Phase 5 (Cleanup & Review)**: 1 day

**Total Estimated Time**: 6-8 days remaining.

## Notes

- Migration should be done package by package to minimise disruption.
- Each package migration should be a separate PR for easier review.
- Tests should be run frequently during migration to catch issues early.
- Consider creating helper scripts to automate common replacements.
- Document any complex conversions or edge cases discovered.
- **Future**: Consider creating `x/testutils` module for shared testing
  utilities.

## References

- [Core Testing Guidelines](../core/TESTING.md)
- [TestCase Pattern Documentation](../core/TESTING.md#table-driven-test-
  structure-patterns-testcase)
- [Assertion Functions Reference](../core/TESTING.md#assertion-functions)
