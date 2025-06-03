# Agent Documentation for x/tls

## Overview

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Features

- **Certificate Store**: Dynamic certificate storage and retrieval.
- **SNI Handling**: Server Name Indication parsing and routing.
- **Certificate Bundling**: Automatic certificate chain building.
- **X509 Utilities**: Enhanced certificate and key manipulation.

### Main Types

- **`Store`**: Interface for certificate storage and retrieval.
- **`StoreReader/Writer`**: Extended store interfaces for CRUD operations.
- **`Bundler`**: Certificate chain builder with quality metrics.
- **`ClientHelloInfo`**: SNI packet parsing and inspection.

### Subpackages

#### sni Package

- ClientHello parsing without full TLS handshake.
- SNI-based routing and dispatching.
- Integration with chi router for HTTP multiplexing.

#### store Package

- **`basic`**: Simple in-memory certificate store.
- **`buffer`**: Buffered certificate operations.
- **`config`**: Configuration-based certificate loading.

#### x509utils Package

- **`certpool`**: Enhanced certificate pool management.
- PEM encoding/decoding utilities.
- Certificate name matching and validation.
- System certificate pool integration.

## Architecture Notes

The package follows several design principles:

1. **Dynamic Management**: Certificates can be added/removed at runtime.
2. **Thread Safety**: All stores implement concurrent-safe operations.
3. **Chain Building**: Automatic certificate chain construction.
4. **System Integration**: Seamless integration with system cert stores.

Key patterns:

- Store interface allows multiple backend implementations.
- SNI parsing enables early routing decisions.
- Bundler optimizes certificate chains for TLS handshakes.
- Context-aware operations throughout.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:

- Certificate chain validation.
- SNI parsing accuracy.
- Thread-safe store operations.
- System certificate pool integration.
- PEM encoding/decoding edge cases.

## Common Usage Patterns

### Basic Certificate Store

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

### SNI Inspection

```go
// Parse ClientHello without full handshake
info := sni.GetInfo(clientHelloBytes)
if info != nil {
    fmt.Printf("SNI: %s\n", info.ServerName)
    // Route based on SNI
}
```

### SNI-based Routing

```go
// Create SNI dispatcher
dispatcher := sni.NewDispatcher()

// Add handlers for different domains
dispatcher.Add("example.com", exampleHandler)
dispatcher.Add("*.api.com", apiHandler)

// Use in TLS config
config := &tls.Config{
    GetCertificate: dispatcher.GetCertificate,
}
```

### Working with Certificate Pools

```go
// Create custom cert pool
pool := certpool.New()

// Add certificates
pool.AddCert(rootCA)
pool.AddCert(intermediateCA)

// Clone system pool and extend
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

### Certificate Verification

```go
// Custom verification
err := tls.Verify(cert, &tls.VerifyOptions{
    DNSName: "example.com",
    Roots: customRoots,
    Intermediates: customInter,
})
```

## Performance Characteristics

- **Store**: O(1) certificate lookup by name (map-based).
- **SNI Parsing**: Minimal overhead, early termination.
- **Bundling**: O(n*m) for n certs and m intermediates.
- **CertPool**: Optimized for fast verification.

## Dependencies

- `darvaza.org/core`: Core utilities.
- `golang.org/x/crypto/cryptobyte`: Low-level crypto parsing.
- Standard library (crypto/tls, crypto/x509).

## Security Considerations

- Private keys are stored in memory (consider HSM for production).
- Certificate validation follows standard x509 rules.
- SNI parsing is resistant to malformed packets.
- System cert pool access may require elevated privileges.

## See Also

- [sni Package](sni/) for SNI handling details.
- [store Package](store/) for storage implementations.
- [x509utils Package](x509utils/) for certificate utilities.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
