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
that extend Go's standard crypto/tls package. It includes certificate stores,
SNI inspection, certificate bundling, and enhanced x509 utilities with a focus
on dynamic certificate management.

## Features

* **Dynamic Certificate Management**: Add and remove certificates at runtime
* **SNI Support**: Parse and route based on Server Name Indication
* **Certificate Bundling**: Automatic certificate chain construction
* **Enhanced X.509 Utilities**: Advanced certificate manipulation

## Components

### Certificate Store

Dynamic certificate storage with multiple backend implementations.

```go
// Create a store
store := &basic.Store{}
ctx := context.Background()

// Add certificates
err := store.AddCertPair(ctx, privateKey, cert, intermediates)

// Configure TLS
config := &tls.Config{
    GetCertificate: store.GetCertificate,
    RootCAs: store.GetCAPool(),
}
```

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

// SNI-based routing
dispatcher := sni.NewDispatcher()
dispatcher.Add("example.com", exampleHandler)
dispatcher.Add("*.api.com", apiHandler)
```

## Packages

### `sni`

Server Name Indication parsing and routing.

* ClientHello parsing without full handshake.
* SNI-based dispatching.
* Chi router integration.

### `store`

Certificate storage implementations.

* `basic`: Simple in-memory store.
* `buffer`: Buffered certificate operations.
* `config`: Configuration-based loading.

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
// Read PEM files
certs, err := x509utils.ReadCertificates(pemData)
key, err := x509utils.ReadPrivateKey(keyData)

// Write PEM
pemData := x509utils.EncodeCertificates(certs...)
keyData := x509utils.EncodePrivateKey(key)
```

### Custom Verification

```go
err := tls.Verify(cert, &tls.VerifyOptions{
    DNSName: "example.com",
    Roots: customRoots,
    Intermediates: customInter,
})
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
