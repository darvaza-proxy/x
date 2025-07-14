# Darvaza Config

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]
[![codecov][codecov-badge]][codecov]

`darvaza.org/x/config` provides helpers
for dealing with config files.

[godoc]: https://pkg.go.dev/darvaza.org/x/config
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/config.svg
[goreport]: https://goreportcard.com/report/darvaza.org/x/config
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/x/config
[codecov]: https://codecov.io/gh/darvaza-proxy/x
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=config

[darvaza-core]: https://pkg.go.dev/darvaza.org/core
[darvaza-penne]: https://pkg.go.dev/darvaza.org/penne
[darvaza-sidecar]: https://pkg.go.dev/darvaza.org/sidecar
[darvaza-x]: https://github.com/darvaza-proxy/x

[amery-defaults]: https://pkg.go.dev/github.com/amery/defaults
[go-playground-validator]: https://pkg.go.dev/github.com/go-playground/validator/v10

## AppDir

`appdir` contains helpers to determine the location of application
specific files.

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

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENT.md](AGENT.md).

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
