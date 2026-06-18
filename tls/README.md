# `darvaza.org/x/tls`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/tls.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/tls
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/tls
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/tls
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=tls
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/tls
[socket-link]: https://socket.dev/go/package/darvaza.org/x/tls

## Overview

The `tls` package provides advanced TLS/SSL certificate handling utilities
that extend Go's standard crypto/tls package: the `Store` abstraction, SNI
inspection, certificate bundling, and enhanced x509 utilities.

Unlike the standard library's static configuration, a `Store` is a living
source of certificates and trust — material is added, renewed and removed while
in use. The `Store` interfaces are designed to be layered, so higher-level
systems can compose behaviour (on-demand issuance, replication) over a simple
local store while keeping it both flexible and consistent.

## Features

* **Dynamic Certificate Management**: Add and remove certificates at runtime
* **SNI Support**: Parse and route based on Server Name Indication
* **Certificate Bundling**: Automatic certificate chain construction
* **Enhanced X.509 Utilities**: Advanced certificate manipulation

## Components

### Certificate Store

The `Store` interfaces describe a dynamic source of certificates and trust. A
read/write store adds lookup, iteration and the building-block writers
(`AddCACerts`, `AddPrivateKey`, `AddCert`, `AddCertPair`) on top of the base
`Store`. Any implementation wires straight into a `tls.Config`:

```go
// store is any tls.Store implementation.
config, err := tls.NewConfig(store)

// or bind one to an existing config:
err = tls.WithStore(config, store)
```

`WithStore` sets `GetCertificate`, `RootCAs` and `ClientCAs` from the store.

### Certificate Bundling

Automatic certificate chain building with quality optimization.

```go
bundler := &tls.Bundler{
    Roots: systemRoots,
    Inter: intermediateCerts,
    Less: func(a, b []*x509.Certificate) bool {
        // Prefer shorter chains
        return len(a) < len(b)
    },
}

// Bundle certificate with optimal chain
tlsCert, err := bundler.Bundle(cert, privateKey)
```

### SNI Handling

Parse ClientHello packets without full TLS handshake.

```go
// Parse SNI from ClientHello
info := sni.GetInfo(clientHelloBytes)
if info != nil {
    fmt.Printf("SNI: %s\n", info.ServerName)
}

// SNI-based routing: GetHandler claims a connection for a dedicated
// Handler, or returns nil to let it fall through to the TLS listener.
dispatcher := &sni.Dispatcher{
    GetHandler: func(chi *tls.ClientHelloInfo) sni.Handler {
        if chi.ServerName == "example.com" {
            return exampleHandler
        }
        return nil
    },
}

go dispatcher.Serve(rawListener)  // feed it the raw net.Listener
tlsListener := tls.NewListener(dispatcher, cfg)  // unclaimed ones land here
```

## Packages

### `sni`

Server Name Indication parsing and routing.

* ClientHello parsing without full handshake.
* SNI-based dispatching.
* Chi router integration.

### `store`

Loading and population for a `Store`.

* `buffer`: collect certificates and keys, then flush them into a `Store`.
* `config`: load certificates and keys from configuration and PEM sources.

### `x509utils`

Enhanced X.509 certificate utilities.

* `certpool`: Advanced certificate pool management.
* PEM encoding/decoding.
* Certificate validation.
* System certificate integration.

## Examples

### Working with Certificate Pools

```go
// Create custom cert pool
pool := certpool.New()
pool.AddCert(rootCA)

// Clone and extend system pool
sysPool, _ := certpool.SystemCertPool()
customPool := sysPool.Clone()
customPool.AddCert(internalCA)
```

### PEM Operations

```go
// Decode certificates from PEM via a per-block callback
err := x509utils.ReadPEM(pemData, func(_ fs.FS, _ string, block *pem.Block) bool {
    if cert, err := x509utils.BlockToCertificate(block); err == nil {
        certs = append(certs, cert)
    }
    return true // keep reading
})

// Encode a certificate (DER) back to PEM
pemBytes := x509utils.EncodeCertificate(cert.Raw)
```

### Certificate Verification

```go
// Verify a tls.Certificate. Pass a roots pool to also verify the chain;
// nil checks only the certificate's own validity.
err := tls.Verify(cert, customRoots)
```

## Installation

```bash
go get darvaza.org/x/tls
```

## Dependencies

* [`darvaza.org/core`][core-link]: Core utilities.
* [`golang.org/x/crypto`][xcrypto-link]: Low-level crypto parsing.
* Standard library (crypto/tls, crypto/x509).

[core-link]: https://pkg.go.dev/darvaza.org/core
[xcrypto-link]: https://pkg.go.dev/golang.org/x/crypto

## Security Considerations

* Private keys are stored in memory (consider HSM for production).
* Certificate validation follows standard x509 rules.
* SNI parsing is resistant to malformed packets.
* System cert pool access may require elevated privileges.

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).

## Licence

This project is licensed under the MIT Licence. See [LICENCE.txt](LICENCE.txt)
for details.
