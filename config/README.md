# Darvaza Config

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

`darvaza.org/x/config` provides helpers
for dealing with config files.

[godoc-link]: https://pkg.go.dev/darvaza.org/x/config
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/config.svg
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/config
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/config
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=config
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/config
[socket-link]: https://socket.dev/go/package/darvaza.org/x/config

[darvaza-core]: https://pkg.go.dev/darvaza.org/core
[darvaza-penne]: https://pkg.go.dev/darvaza.org/penne
[darvaza-sidecar]: https://pkg.go.dev/darvaza.org/sidecar
[darvaza-x]: https://github.com/darvaza-proxy/x

[amery-defaults]: https://pkg.go.dev/github.com/amery/defaults
[go-playground-validator]: https://pkg.go.dev/github.com/go-playground/validator/v10

## AppDir

`appdir` determines where an application should keep its files —
cache, configuration, persistent data, and run-time data — following
the XDG Base Directory Specification in user mode and the Filesystem
Hierarchy Standard in system mode, with native equivalents on
Windows. On macOS the XDG variables still take precedence, with
Apple-native locations as the user-mode fallbacks.

The `Prefix` type selects the system-mode filesystem root:
`PrefixSystem` (`/`), `PrefixLocal` (`/usr/local`) and
`PrefixOptional` (`/opt`) on unix-like systems, or `PrefixSystem`
alone — mapping to `%ProgramData%` — on Windows, or a custom
directory validated via `NewPrefix()`. The special `PrefixUser`
value makes the methods return the user-mode locations instead.
Any value other than the well-known constants must be an absolute
path to an existing directory, as checked by `Validate()`; the
methods reject malformed prefixes — including the zero value —
with an error.

* `UserCacheDir()`, `UserConfigDir()`, `UserDataDir()`,
  `UserRuntimeDir()` — user mode (XDG).
* `Prefix.CacheDir()`, `Prefix.ConfigDir()`, `Prefix.DataDir()`,
  `Prefix.RuntimeDir()` — system mode under the `Prefix`.
* `SysCacheDir()`, `SysConfigDir()`, `SysDataDir()`,
  `SysRuntimeDir()` — shortcuts using the package default set via
  `SetSysPrefix()` and read via `SysPrefix()`.
* `AllConfigDir()` — configuration search path: working directory,
  user mode, then system mode.

## Default values

Wrappers for [`github.com/amery/defaults`][amery-defaults]:

* `SetDefaults()`
* `Set()`
* `CanUpdate()`

## Environment

Expand shell-style variables:

* `FromString()`
* `FromBytes()`
* `FromReader()`
* `FromFile()`

## Loader

Attempts to decode an object from one of a list of filenames.

## Validations

Wrappers for
[`github.com/go-playground/validator/v10`][go-playground-validator]:

* `Validate()`
* `AsValidationError()`
* and `Prepare()`, calling `SetDefaults()` and `Validate()`.

## Installation

```bash
go get darvaza.org/x/config
```

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).

## See also

* [Apptly Software's Open Source Projects](https://oss.apptly.co/)
* _darvaza libraries_
  * [darvaza.org/core][darvaza-core]
  * [darvaza.org/x][darvaza-x]
* _darvaza servers_
  * [darvaza.org/penne][darvaza-penne]
  * [darvaza.org/sidecar][darvaza-sidecar]
* _third party libraries_
  * [github.com/amery/defaults][amery-defaults]
  * [github.com/go-playground/validator][go-playground-validator]
  * [mvdan.cc/sh](https://pkg.go.dev/mvdan.cc/sh/v3)
