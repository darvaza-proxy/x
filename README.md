# Darvaza Extra

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]

`darvaza.org/x` hosts mid complexity packages with
no big dependencies or assumptions.

[godoc]: https://pkg.go.dev/darvaza.org/x
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x.svg
[goreport]: https://goreportcard.com/report/darvaza.org/x
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/x

[darvaza-cache]: https://pkg.go.dev/darvaza.org/cache
[darvaza-core]: https://pkg.go.dev/darvaza.org/core
[darvaza-penne]: https://pkg.go.dev/darvaza.org/penne
[darvaza-resolver]: https://pkg.go.dev/darvaza.org/resolver
[darvaza-slog]: https://pkg.go.dev/darvaza.org/slog
[darvaza-sidecar]: https://pkg.go.dev/darvaza.org/sidecar
[darvaza-simplelru]: https://pkg.go.dev/darvaza.org/cache/x/simplelru
[darvaza-x-config]: https://pkg.go.dev/darvaza.org/x/config
[darvaza-x-tls]: https://pkg.go.dev/darvaza.org/x/tls
[darvaza-x-web]: https://pkg.go.dev/darvaza.org/x/web

## Dependencies

The _Darvaza Extra_ modules are built on top of a handful
low(ish) level packages in addition to the Go Standard Library.

* Our _core_ package, [`darvaza.org/core`][darvaza-core],
  dealing with network addresses, worker groups, errors and lists among
  other simple helpers.
* Our _structured logger_ interface, [`darvaza.org/slog`][darvaza-slog],
  allowing users to hook their favourite logger.
* And our thin and simple _LRU_ for local in-memory caching,
  [`darvaza.org/cache/x/simplelru`][darvaza-simplelru].

## Packages

### Config

[`darvaza.org/x/config`][darvaza-x-config] provides helpers
for dealing with config files.

### TLS

[`darvaza.org/x/tls`][darvaza-x-tls] provides helpers
to work with tls connections and certificates.

### Web

[`darvaza.org/x/web`][darvaza-x-web] provides helpers
for implementing http.Handlers.

## See also

* [JPI Technologies' Open Source Software](https://oss.jpi.io/)
* _darvaza libraries_
  * [`darvaza.org/cache`][darvaza-cache]
  * [`darvaza.org/core`][darvaza-core]
  * [`darvaza.org/resolver`][darvaza-resolver]
  * [`darvaza.org/slog`][darvaza-slog]
  * [`darvaza.org/x/config`][darvaza-x-config]
  * [`darvaza.org/x/tls`][darvaza-x-tls]
  * [`darvaza.org/x/web`][darvaza-x-web]
* _darvaza servers_
  * [`darvaza.org/penne`][darvaza-penne]
  * [`darvaza.org/sidecar`][darvaza-sidecar]
