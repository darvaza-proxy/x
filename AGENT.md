# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Repository Overview

`darvaza.org/x` is a Go mono-repo hosting mid-complexity packages with
minimal dependencies. Each package is a separate Go module providing
utility libraries that extend Go's standard library capabilities.

## Prerequisites

Before starting development, ensure you have:

- Go 1.23 or later installed (check with `go version`).
- `make` command available (usually pre-installed on Unix systems).
- `$GOPATH` configured correctly (typically `~/go`).
- Git configured for proper line endings.

## Packages

Each package has its own AGENT.md file with detailed documentation:

- **cmp**: See [cmp/AGENT.md](cmp/AGENT.md).
- **config**: See [config/AGENT.md](config/AGENT.md).
- **container**: See [container/AGENT.md](container/AGENT.md).
- **fs**: See [fs/AGENT.md](fs/AGENT.md).
- **net**: See [net/AGENT.md](net/AGENT.md).
- **sync**: See [sync/AGENT.md](sync/AGENT.md).
- **tls**: See [tls/AGENT.md](tls/AGENT.md).
- **web**: See [web/AGENT.md](web/AGENT.md).

## Common Development Commands

```bash
# Full build cycle (get deps, generate, tidy, build)
make all

# Run tests for all packages
make test

# Format code and tidy dependencies (run before committing)
make tidy

# Clean build artifacts
make clean

# Update dependencies
make up

# Run go:generate directives for all packages
make generate
```

## Working with Individual Packages

Each package has its own directory and can be worked on independently:

```bash
# Navigate to a specific package
cd cmp/

# Run package-specific commands
go test ./...
go mod tidy
```

## Architecture Principles

### Key Design Principles

- **Minimal dependencies**: Primarily the Go standard library and minimal
  golang.org/x packages.
- **Generic programming**: Extensive use of Go 1.23+ generics for
  type-safe utilities.
- **Interface-driven design**: For extensibility and testability.
- **Common foundation**: Each package depends on `darvaza.org/core` for
  basic utilities.

### Module Structure

- Each package is a separate Go module with its own `go.mod`.
- No circular dependencies between modules.
- Clear separation of concerns between packages.

### Code Quality Standards

The project enforces strict linting rules via revive (configuration in each
package's `internal/build/revive.toml`):

- Max function length: 40 lines.
- Max function results: 3.
- Max arguments: 5.
- Cognitive complexity: 7.
- Cyclomatic complexity: 10.

Always run `make tidy` before committing to ensure proper formatting.

### Testing Patterns

- Table-driven tests are preferred.
- Helper functions like `S[T]()` create test slices.
- Comprehensive coverage for generic functions is expected.

### Build System

- Dynamic Makefile generation via shell scripts.
- Automatic detection of sub-modules.
- Integrated linting with golangci-lint and revive.
- Version-aware tool selection based on Go version.

## Build System Features

### Whitespace and EOF Handling

The `internal/build/fix_whitespace.sh` script automatically:

- Removes trailing whitespace from all text files.
- Ensures files end with a newline.
- Excludes binary files and version control directories.
- Integrates with `make fmt` for non-Go files.
- Supports both directory scanning and explicit file arguments.

### Markdownlint Integration

The build system includes automatic Markdown linting:

- Detects markdownlint-cli via pnpx.
- Configuration in `internal/build/markdownlint.json`.
- 80-character line limits and strict formatting rules.
- Selective HTML allowlist (comments, br, kbd, etc.).
- Runs automatically with `make fmt` when available.

### CSpell Integration

Spell checking for both Markdown and Go source files:

- Detects cspell via pnpx.
- British English configuration in `internal/build/cspell.json`.
- New `check-spelling` target.
- Integrated into `make tidy`.
- Custom word list for project-specific terminology.
- Checks both documentation and code comments.

### LanguageTool Integration

Grammar and style checking for Markdown files:

- Detects LanguageTool via pnpx.
- British English configuration in `internal/build/languagetool.cfg`.
- New `check-grammar` target.
- Checks for missing articles, punctuation, and proper hyphenation.

### ShellCheck Integration

Shell script analysis for all `.sh` files:

- Detects shellcheck via pnpx.
- New `check-shell` target.
- Integrated into `make tidy`.
- Uses inline disable directives for SC1007 (empty assignments) and SC3043
  (`local` usage).
- Checks for common shell scripting issues and best practices.

### Test Coverage Collection

Automated coverage reporting across all modules:

- New `coverage` target runs tests with coverage profiling.
- Uses `internal/build/make_coverage.sh` to orchestrate testing.
- Tests each module independently via generated `test-*` targets.
- Merges coverage profiles automatically.
- Stores results in `.tmp/coverage/` directory.
- Displays coverage summary after test runs.
- Optional HTML report generation with `COVERAGE_HTML=true`.

### CI/CD Workflow Separation

GitHub Actions workflows split for better performance:

- **Build workflow** (`.github/workflows/build.yml`): Focuses on compilation
  only.
- **Test workflow** (`.github/workflows/test.yml`): Dedicated testing
  pipeline.
  - Race condition detection job with Go 1.23.
  - Multi-version testing matrix (Go 1.23 and 1.24).
  - Conditional execution to avoid duplicate runs on PRs.
- Workflows skip branches ending in `-wip`.
- Improves parallelism and reduces redundant work.

### Codecov Integration

Automated coverage reporting with monorepo support:

- **Codecov workflow** (`.github/workflows/codecov.yml`): Coverage collection
  and upload.
- Enhanced `make_coverage.sh` generates:
  - `codecov.yml`: Dynamic configuration with per-module flags.
  - Module-specific coverage targets (80% default).
  - Path mappings for accurate coverage attribution.
  - `codecov.sh`: Upload script for bulk submission.
- Supports both GitHub Actions and local coverage uploads.
- PR comments show coverage changes per module.

## Testing with GOTEST_FLAGS

The `GOTEST_FLAGS` environment variable allows flexible test execution by
passing additional flags to `go test`. This variable is defined in the
Makefile with an empty default value and is used when running tests through
the generated rules.

### Common Usage Examples

```bash
# Run tests with race detection
make test GOTEST_FLAGS="-race"

# Run specific tests by pattern
make test GOTEST_FLAGS="-run TestSpecific"

# Generate coverage profile (alternative to 'make coverage')
make test GOTEST_FLAGS="-coverprofile=coverage.out"

# Run tests with timeout
make test GOTEST_FLAGS="-timeout 30s"

# Combine multiple flags
make test GOTEST_FLAGS="-v -race -coverprofile=coverage.out"

# Run benchmarks
make test GOTEST_FLAGS="-bench=. -benchmem"

# Skip long-running tests
make test GOTEST_FLAGS="-short"

# Test with specific CPU count
make test GOTEST_FLAGS="-cpu=1,2,4"
```

### Integration with Coverage

While `make coverage` provides automated coverage collection across all
modules, you can use `GOTEST_FLAGS` for more targeted coverage analysis:

```bash
# Coverage for specific package with detailed output
make test GOTEST_FLAGS="-v -coverprofile=coverage.out -covermode=atomic"

# Coverage with HTML output
make test GOTEST_FLAGS="-coverprofile=coverage.out"
go tool cover -html=coverage.out
```

### How It Works

1. The Makefile defines `GOTEST_FLAGS ?=` (empty by default).
2. The generated rules use it in the test target:
   `$(GO) test $(GOTEST_FLAGS) ./...`.
3. Any flags passed via `GOTEST_FLAGS` are forwarded directly to `go test`.

This provides a clean interface for passing arbitrary test flags without
modifying the Makefile, making it easy to run tests with different
configurations for debugging, coverage analysis, or CI/CD pipelines.

## Important Notes

- Go 1.23 is the minimum required version.
- Each package maintains its own README.md and AGENT.md.
- The Makefile dynamically generates rules for subprojects.
- Tool versions (golangci-lint, revive) are selected based on Go version.
- These are utility libraries - no business logic, only reusable helpers.
- Always use `pnpm` instead of `npm` for any JavaScript/TypeScript tooling.
- Follow existing patterns when adding new functionality.

## Linting and Code Quality

### Documentation Standards

When editing markdown files, ensure compliance with:

- **LanguageTool**: Check for missing articles ("a", "an", "the"), punctuation,
  and proper hyphenation of compound modifiers.
- **Markdownlint**: Follow standard Markdown formatting rules.
- **Line Length**: Keep lines at 80 characters or less (enforced by
  .editorconfig).

### Common Documentation Issues to Check

1. **Missing Articles**: Ensure proper use of "a", "an", and "the".
   - ❌ "converts value using provided function"
   - ✅ "converts value using a provided function"

2. **Missing Punctuation**: End all list items consistently.
   - ❌ "Comprehensive coverage for generic functions is expected"
   - ✅ "Comprehensive coverage for generic functions is expected."

3. **Compound Modifiers**: Hyphenate when used as modifiers.
   - ❌ "capture specific stack frame"
   - ✅ "capture-specific stack frame"
   - ❌ "Open Source Software"
   - ✅ "Open-Source Software"

4. **Formal Language**: Use professional, precise wording.
   - ❌ "no big dependencies"
   - ✅ "no significant dependencies"

5. **Verb Forms**: Ensure verbs match their context and subjects.
   - ❌ "package dealing with network addresses"
   - ✅ "package which handles network addresses"

6. **List Structure**: Avoid starting list items with conjunctions.
   - ❌ "And our thin and simple LRU"
   - ✅ "Our thin and simple LRU"

### Writing Documentation Guidelines

When creating or editing documentation files:

1. **File Structure**:
   - Always include a link to related documentation (e.g., AGENT.md should
     link to README.md).
   - Add prerequisites or setup instructions before diving into commands.
   - Include paths to configuration files when mentioning tools
     (e.g., revive.toml).

2. **Formatting Consistency**:
   - **Line Length**: Wrap lines at 80 characters maximum.
   - End all bullet points with periods for consistency.
   - Capitalize proper nouns correctly (JavaScript, TypeScript, Markdown).
   - Use consistent punctuation in examples and lists.
   - In "See also" sections, add colons after section headers and periods
     after all items.

3. **Markdownlint Compliance**:
   - Add blank lines before and after lists, code blocks, and headings.
   - End files with exactly one newline character.
   - Avoid spaces inside emphasis markers (use `_text_` not `_ text _`).
   - Follow all markdownlint rules (run `pnpx markdownlint-cli *.md`).

4. **Clarity and Context**:
   - Provide context for AI agents and developers alike.
   - Include "why" explanations, not just "what" descriptions.
   - Add examples for complex concepts or common pitfalls.

5. **Maintenance**:
   - Update documentation when adding new tools or changing workflows.
   - Keep the pre-commit checklist current with project practices.
   - Review documentation changes for the issues listed above.

### Pre-commit Checklist

1. **ALWAYS run `make tidy` first** - Fix ALL issues before committing:
   - Go code formatting and whitespace clean-up.
   - Markdown files checked with CSpell and markdownlint.
   - Shell scripts checked with ShellCheck.
   - If `make tidy` fails, fix the issues and run it again until it passes.
2. Verify all tests pass with `make test`.
3. Ensure no linting violations remain.
4. Update `AGENT.md` to reflect any changes in development workflow or
   standards.
5. Update `README.md` to reflect significant changes in functionality or API.
6. Update package documentation if modifying package behaviour.
7. Verify package examples still compile and run correctly.

### Grammar and Style Checking

The project now includes integrated grammar checking via LanguageTool:

```bash
# Run formatting and spell/shell checks
make tidy

# Run only grammar checks (Markdown and Go files)
make check-grammar
```

LanguageTool is automatically installed via npm (using pnpx) when available.
It checks both Markdown documentation and Go source files (comments and
strings). The following rules are disabled for technical documentation
compatibility:

- COMMA_PARENTHESIS_WHITESPACE (conflicts with Markdown links).
- ARROWS (used in code examples).
- EN_QUOTES (technical docs use straight quotes).
- MORFOLOGIK_RULE_EN_GB (flags technical terms).
- UPPERCASE_SENTENCE_START (conflicts with inline code).

Configuration files are located in `internal/build/`:

- `markdownlint.json` - Markdown formatting rules.
- `languagetool.cfg` - Grammar checking rules for British English.

## Release Process

For information about releasing packages and managing dependencies, see
[RELEASE.md](RELEASE.md). The mono-repo structure requires careful coordination
of releases due to internal dependencies between packages.
