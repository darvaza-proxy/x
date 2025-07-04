# Agent Documentation for x/net

## Overview

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Subpackages

- **`bind`**: Advanced TCP/UDP port binding with retry logic and control.
- **`reconnect`**: Generic reconnecting TCP client with automatic retry.

### Main Features

#### bind Package
- Multi-interface/address binding support.
- Port retry logic with configurable attempts.
- Socket option control (SO_REUSEADDR, SO_REUSEPORT).
- Buffer size configuration for UDP.
- Connection upgrading capabilities.

#### reconnect Package
- Automatic connection retry with backoff.
- Session management with lifecycle callbacks.
- Context-based cancellation.
- Thread-safe connection handling.
- Configurable timeouts and retry strategies.

### Main Files

- `dialer.go`: Enhanced dialer implementations.
- `addrs.go`: Address manipulation utilities.
- `std.go`: Standard library extensions.
- `bind/bind.go`: Core binding configuration and logic.
- `bind/control_*.go`: Platform-specific socket control.
- `bind/listen.go`: TCP/UDP listener creation.
- `reconnect/client.go`: Reconnecting client implementation.
- `reconnect/config.go`: Client configuration.
- `reconnect/worker.go`: Background worker management.

## Architecture Notes

The package follows several design principles:

1. **Platform Abstraction**: Socket control operations isolated by platform.
2. **Graceful Degradation**: Features degrade safely on unsupported platforms.
3. **Context Integration**: Full context.Context support throughout.
4. **Thread Safety**: Concurrent-safe operations in reconnect client.

Key patterns:
- Config structs for complex initialization.
- Callback-based lifecycle management.
- Interface-based abstraction for extensibility.
- Atomic operations for state management.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:
- Port binding edge cases (conflicts, retries).
- Reconnection logic and timing.
- Platform-specific behavior.
- Concurrent access patterns.

## Common Usage Patterns

### Advanced Port Binding
```go
cfg := &bind.Config{
    Interfaces: []string{"lo", "eth0"},
    Port: 8080,
    PortAttempts: 4,
    ReusePort: true,
    KeepAlive: 30 * time.Second,
}

listeners, err := bind.Bind(cfg)
for _, l := range listeners {
    defer l.Close()
    go handleListener(l)
}
```

### Reconnecting Client
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

// Start client with automatic reconnection
err := client.Spawn(ctx)
```

### Socket Control
```go
// Custom listener with socket options
control := bind.ControlFunc(func(fd uintptr) error {
    // Set custom socket options
    return bind.SetReusePort(fd, true)
})

ln, err := bind.ListenTCP("tcp", addr, control)
```

### Connection Lifecycle
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

## Performance Characteristics

- **Bind**: O(n) for n interfaces/addresses.
- **Reconnect**: Exponential backoff available for retry delays.
- **Socket Control**: Minimal overhead on socket creation.

## Dependencies

- `darvaza.org/core`: Core utilities.
- `darvaza.org/slog`: Structured logging.
- Standard library (net, syscall, context).

## Platform Support

- **Unix/Linux**: Full socket control support.
- **Windows**: Adapted control operations.
- **Build Tags**: Platform-specific implementations.

## See Also

- [reconnect README](reconnect/README.md) for automatic reconnection client details.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
