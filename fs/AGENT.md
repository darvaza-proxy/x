# Agent Documentation for x/fs

## Overview

The `fs` package provides enhanced tools for working with Go's `fs.FS`
interface. It extends the standard library with missing filesystem operations,
advanced globbing support, path utilities, and platform-specific file locking.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Features

- **Extended FS Interfaces**: Additional filesystem interfaces matching `os`
  package functionality.
- **Advanced Globbing**: Pattern matching with `**` support via gobwas/glob.
- **Path Utilities**: Enhanced path cleaning and splitting.
- **File Locking**: Cross-platform file locking support.
- **IO Utilities**: Helper functions for common file operations.

### Main Files

- `fs.go`: Extended filesystem interfaces (ChmodFS, MkdirFS, etc.).
- `glob.go`: Pattern matching and file walking.
- `clean.go`: Enhanced path cleaning with absolute path support.
- `split.go`: Path splitting utilities.
- `file.go`: File interface extensions.
- `io.go`: IO helper functions.
- `std.go`: Standard library aliases and proxies.
- `flock/flock.go`: File locking implementation.
- `fssyscall/`: Platform-specific syscall wrappers.

## Architecture Notes

The package follows several design principles:

1. **Interface Extension**: Adds missing FS operations while maintaining
   compatibility with standard fs.FS.
2. **Platform Abstraction**: Syscalls are isolated in fssyscall subpackage.
3. **Clean Import Space**: Uses type aliases to avoid import conflicts.
4. **Zero Allocation**: Path operations designed to minimize allocations.

Key patterns:
- Interfaces follow os package naming conventions (e.g., ChmodFS for Chmod).
- Platform-specific code uses build tags (_linux.go, _windows.go).
- Glob patterns support `**` for recursive matching.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:
- Path cleaning edge cases (absolute paths, `..` handling).
- Glob pattern matching accuracy.
- Cross-platform compatibility.
- File locking behavior.

## Common Usage Patterns

### Globbing Files

```go
// Find all Go files recursively
matches, err := fs.Glob(fsys, "**/*.go")

// Using compiled matchers
matchers, _ := fs.GlobCompile("*.txt", "data/**/*.json")
files, err := fs.Match(fsys, ".", matchers...)
```

### Path Operations

```go
// Clean paths with absolute support
clean, valid := fs.Clean("/path/../to/file.txt")

// Split without trailing slash
dir, file := fs.Split("path/to/file.txt")
// dir = "path/to", file = "file.txt"
```

### Extended FS Operations

```go
// Check if FS supports chmod
if chmodFS, ok := fsys.(fs.ChmodFS); ok {
    err := chmodFS.Chmod("file.txt", 0644)
}

// Create directories
if mkdirFS, ok := fsys.(fs.MkdirAllFS); ok {
    err := mkdirFS.MkdirAll("path/to/dir", 0755)
}
```

### File Locking

```go
// Exclusive file lock
handle, err := flock.LockEx("important.db")
if err != nil {
    // handle error
}
defer handle.Close()
```

### IO Helpers

```go
// Read file with FS fallback
data, err := fs.ReadFile(fsys, "config.json")

// Copy between filesystems
err := fs.Copy(destFS, "dest.txt", srcFS, "src.txt")
```

## Platform Support

- **Unix/Linux**: Full support including file locking.
- **Windows**: Adapted syscalls for compatibility.
- **Build Tags**: Platform-specific implementations isolated.

## Dependencies

- `darvaza.org/core`: Core utilities.
- `github.com/gobwas/glob`: Advanced glob pattern matching.
- Standard library (io/fs, os, path).

## See Also

- [Package README](README.md) for detailed API documentation.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
- [Go fs package documentation](https://pkg.go.dev/io/fs).
