# Agent Documentation for x/config

## Overview

The `config` package provides comprehensive utilities for handling
configuration files in Go applications. It includes support for multiple
file formats, environment variable expansion, validation, default values,
and platform-specific application directory handling.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Features

- **Multi-format Loading**: Support for JSON, YAML, TOML, and custom formats.
- **Environment Expansion**: Shell-style variable expansion in config files.
- **Validation**: Integration with go-playground/validator.
- **Defaults**: Automatic setting of default values.
- **AppDir**: Platform-specific application directory resolution.

### Main Files

- `loader.go`: Generic loader for configuration files with fallback support.
- `decoder.go`: Interface for format-specific decoders.
- `defaults.go`: Wrapper for github.com/amery/defaults.
- `validate.go`: Validation utilities using go-playground/validator.
- `prepare.go`: Combined defaults + validation workflow.
- `expand/file.go`: Environment variable expansion utilities.
- `appdir/`: Platform-specific directory helpers (XDG, FHS compliance).

### AppDir Subpackage

- `appdir.go`: Core directory resolution logic.
- `appdir_xdg.go`: XDG Base Directory Specification implementation.
- `appdir_fhs.go`: Filesystem Hierarchy Standard support.

## Architecture Notes

The package follows several design principles:

1. **Generic Support**: Uses Go generics for type-safe configuration loading.
2. **Platform Independence**: Handles Unix/Linux directory standards gracefully.
3. **Fail-Safe Loading**: Supports multiple fallback locations for config files.
4. **Extensibility**: Decoder interface allows custom format support.

Key patterns:

- Loader[T] provides a generic configuration loader with multiple fallbacks.
- Option[T] functions allow post-processing of loaded configurations.
- Platform-specific code is isolated in separate files (_unix.go suffixes).

## Development Commands

For common development commands and workflow, see the
[root AGENTS.md](../AGENTS.md).

## Testing Patterns

The package uses table-driven tests and focuses on:

- Platform-specific behaviour testing.
- Environment variable expansion edge cases.
- Validation error handling.
- File loading with various formats.

## Common Usage Patterns

### Loading Configuration

```go
type Config struct {
    Host string `default:"localhost" validate:"required,hostname"`
    Port int    `default:"8080" validate:"min=1,max=65535"`
}

loader := &config.Loader[Config]{
    NewDecoder: config.NewDecoderFunc,
}

cfg, err := loader.NewFromFile(os.DirFS("/"),
    "/etc/myapp/config.yaml",
    "$HOME/.config/myapp/config.yaml",
)
```

### Using AppDir

```go
// Get user-specific config directory
configDir, _ := appdir.UserConfigDir("myapp")

// Get system-wide config directory
sysConfigDir, _ := appdir.SysConfigDir("myapp")

// Get cache directory
cacheDir, _ := appdir.UserCacheDir("myapp")
```

### Environment Expansion

```go
// Expand variables in config file
expanded, err := expand.FromFile("config.template")
```

### Validation and Defaults

```go
// Prepare config: set defaults and validate
err := config.Prepare(&cfg)
```

## Dependencies

- `darvaza.org/core`: Core utilities.
- `github.com/amery/defaults`: Default value handling.
- `github.com/go-playground/validator/v10`: Struct validation.
- `mvdan.cc/sh/v3`: Shell-style expansion.
- Standard library (os, path, io/fs).

## Platform Support

- **Unix/Linux**: Full XDG Base Directory support.
- **Windows**: Uses standard OS paths (via os.UserCacheDir, etc.).
- **FHS Compliance**: System directories follow the Filesystem Hierarchy
  Standard.

## See Also

- [Package README](README.md) for API documentation.
- [Root AGENTS.md](../AGENTS.md) for mono-repo overview.
- [XDG Base Directory Specification][xdg-spec].

[xdg-spec]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
