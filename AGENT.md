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

1. Run `make tidy` for Go code formatting.
2. Remove trailing whitespace: `sed -i 's/[ \t]*$//' *.md`.
3. Check markdown files with `pnpx markdownlint-cli *.md **/*.md`.
4. Check markdown files with LanguageTool for grammar and style issues.
5. Verify all tests pass with `make test`.
6. Ensure no linting violations remain.
7. Update `AGENT.md` to reflect any changes in development workflow or
   standards.
8. Update `README.md` to reflect significant changes in functionality or API.
9. Ensure all markdown files follow the 80-character line length rule.
