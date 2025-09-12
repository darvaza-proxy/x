# `darvaza.org/x/net`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/net.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/net
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/net
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/net
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=net
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/net
[socket-link]: https://socket.dev/go/package/darvaza.org/x/net

## Overview

The `net` package provides advanced networking utilities that extend Go's
standard net package. It includes sophisticated port binding mechanisms,
automatic reconnection clients, and platform-specific socket control.

## Packages

### `bind`

Advanced TCP/UDP port binding with retry logic, multi-interface support, and
socket option control.

```go
cfg := &bind.Config{
    Interfaces: []string{"lo", "eth0"},
    Port: 8080,
    PortAttempts: 4,
    ReusePort: true,
    KeepAlive: 30 * time.Second,
}

listeners, err := bind.Bind(cfg)
```

Features:

* Multi-interface and multi-address binding.
* Automatic port retry with configurable attempts.
* Socket option control (SO_REUSEADDR, SO_REUSEPORT).
* Buffer size configuration for UDP.
* Connection upgrading capabilities.

### `reconnect`

Generic reconnecting TCP client with automatic retry and lifecycle management.

```go
cfg := &reconnect.Config{
    Address: "server:9000",
    DialTimeout: 5 * time.Second,
    RetryWait: 1 * time.Second,
    RetryBackoff: true,
}

client := reconnect.NewClient(cfg,
    reconnect.WithLogger(logger),
    reconnect.WithOnConnect(onConnect),
    reconnect.WithOnError(onError),
)

// Start with automatic reconnection
err := client.Spawn(ctx)
```

Features:

* Automatic connection retry with backoff.
* Session management with lifecycle callbacks.
* Context-based cancellation.
* Thread-safe connection handling.
* Configurable timeouts and retry strategies.

## Features

* **Platform Abstraction**: Socket control operations work across Unix/Linux
  and Windows.
* **Graceful Degradation**: Features degrade safely on unsupported platforms.
* **Context Integration**: Full context.Context support throughout.
* **Thread Safety**: Concurrent-safe operations in reconnect client.

## Installation

```bash
go get darvaza.org/x/net
```

## Examples

### Advanced Port Binding

```go
// Custom listener with socket options
control := bind.ControlFunc(func(fd uintptr) error {
    // Set custom socket options
    return bind.SetReusePort(fd, true)
})

ln, err := bind.ListenTCP("tcp", addr, control)
```

### Connection Lifecycle Management

```go
client := reconnect.NewClient(cfg,
    reconnect.WithOnConnect(func(ctx context.Context, conn net.Conn) error {
        // Initialize connection
        return nil
    }),
    reconnect.WithOnSession(func(ctx context.Context) error {
        // Handle active session
        return nil
    }),
    reconnect.WithOnDisconnect(func(ctx context.Context, conn net.Conn) error {
        // Cleanup on disconnect
        return nil
    }),
)
```

## Dependencies

* [`darvaza.org/core`][core-link]: Core utilities.
* [`darvaza.org/slog`][slog-link]: Structured logging.
* Standard library (net, syscall, context).

[core-link]: https://pkg.go.dev/darvaza.org/core
[slog-link]: https://pkg.go.dev/darvaza.org/slog

## Platform Support

* **Unix/Linux**: Full socket control support.
* **Windows**: Adapted control operations.
* Platform-specific implementations use build tags.

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).

## Licence

This project is licensed under the MIT Licence. See [LICENCE.txt](LICENCE.txt)
for details.
