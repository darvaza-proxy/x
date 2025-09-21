# Generic Reconnecting Network Client

The `reconnect` package provides a robust, generic network client with automatic
reconnection capabilities for both TCP and Unix domain sockets. It implements
retry mechanisms, connection lifecycle management, and configurable callbacks
for handling connection events.

## Overview

The reconnecting network client automatically handles connection failures by
implementing configurable retry logic. It supports both TCP connections and
Unix domain sockets, provides hooks for custom connection handling, session
management, and error processing whilst maintaining thread safety and proper
resource clean-up.

## Key Features

- **Automatic Reconnection**: Transparent reconnection with configurable retry
  strategies.
- **Multiple Network Types**: Support for both TCP and Unix domain sockets.
- **Lifecycle Callbacks**: Customisable hooks for socket, connect, session,
  disconnect, and error events.
- **Thread Safety**: Built-in synchronisation for concurrent operations.
- **Timeout Management**: Separate dial, read, and write timeout configuration.
- **Structured Logging**: Integration with `darvaza.org/slog` for
  comprehensive logging.
- **Context Support**: Full `context.Context` integration for cancellation and
  deadlines.
- **Configuration Validation**: Built-in validation and default-setting
  mechanisms.

## Basic Usage

```go
import (
    "context"
    "net"
    "time"

    "darvaza.org/slog"
    "darvaza.org/x/net/reconnect"
)

// Get a logger instance (implementation-specific).
// This could be from any slog handler (filter, zap, etc.)
// or a custom implementation.
logger := getLogger()

// Create a configuration.
cfg := &reconnect.Config{
    Context: context.Background(),
    Logger:  logger,

    // Connection settings.
    // TCP: "host:port" or Unix: "/path/to/socket" or "unix:/path"
    Remote:       "example.com:8080",
    KeepAlive:    5 * time.Second,  // Default: 5s.
    DialTimeout:  2 * time.Second,  // Default: 2s.
    ReadTimeout:  2 * time.Second,  // Default: 2s.
    WriteTimeout: 2 * time.Second,  // Default: 2s.

    // Retry configuration.
    ReconnectDelay: time.Second,  // Delay between reconnection attempts.
}

// Create client.
client, err := reconnect.New(cfg)
if err != nil {
    logger.Fatal().Printf("Failed to create client: %v", err)
}

// Connect and start the client.
if err := client.Connect(); err != nil {
    logger.Fatal().Printf("Failed to connect: %v", err)
}

// Wait for completion.
defer func() {
    if err := client.Wait(); err != nil {
        logger.Error().Printf("Client error: %v", err)
    }
}()
```

### Unix Domain Socket Usage

The client automatically detects Unix domain sockets based on the Remote string:

```go
// Unix socket with explicit prefix
cfg := &reconnect.Config{
    Remote: "unix:/var/run/app.sock",
    // ... other configuration
}

// Unix socket with absolute path (auto-detected)
cfg := &reconnect.Config{
    Remote: "/var/run/app.sock",
    // ... other configuration
}

// Unix socket with .sock extension (auto-detected)
cfg := &reconnect.Config{
    Remote: "./app.sock",
    // ... other configuration
}
```

## Advanced Configuration

The client supports extensive customisation through callback functions and
options.

### Connection Lifecycle Callbacks

```go
import (
    "context"
    "net"
    "syscall"

    "darvaza.org/slog"
    "darvaza.org/x/net/reconnect"
)

// Assume logger is already obtained.
var logger slog.Logger

cfg := &reconnect.Config{
    // Called against the raw socket before connecting.
    OnSocket: func(ctx context.Context, conn syscall.RawConn) error {
        // Configure socket options.
        return nil
    },

    // Called when a connection is established.
    OnConnect: func(ctx context.Context, conn net.Conn) error {
        logger.Info().Printf("Connected to %s", conn.RemoteAddr())
        // Perform handshake or initial setup.
        return nil
    },

    // Called for each session after connection.
    OnSession: func(ctx context.Context) error {
        // Implement your protocol logic here.
        // This function should block until the session ends.
        return handleSession(ctx)
    },

    // Called when connection is about to close.
    OnDisconnect: func(ctx context.Context, conn net.Conn) error {
        logger.Info().Printf("Disconnecting from %s", conn.RemoteAddr())
        // Clean up connection-specific resources.
        return nil
    },

    // Called when errors occur.
    OnError: func(ctx context.Context, conn net.Conn, err error) error {
        logger.Error().
            WithField("error", err).
            Printf("Connection error")
        // Return nil to allow retry, non-nil to stop reconnection.
        return nil
    },
}
```

### Retry Configuration

```go
import (
    "context"
    "time"

    "darvaza.org/x/net/reconnect"
)

cfg := &reconnect.Config{
    // Simple constant delay between retries.
    ReconnectDelay: time.Second,

    // Custom wait function for retry logic.
    WaitReconnect: func(ctx context.Context) error {
        // Implement custom backoff logic.
        // Return error to stop reconnection.
        return customBackoff(ctx)
    },
}

// Use helper functions for common patterns.
cfg.WaitReconnect = reconnect.NewConstantWaiter(5 * time.Second)

// Immediate error return (no retry).
cfg.WaitReconnect = reconnect.NewImmediateErrorWaiter(errNoRetry)

// Prevent all reconnection attempts.
cfg.WaitReconnect = reconnect.NewDoNotReconnectWaiter(errStop)
```

## Configuration Options

The `Config` structure supports the following fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Context` | `context.Context` | `context.Background()` | Base context for the client. |
| `Logger` | `slog.Logger` | Default logger | Logger for structured logging. |
| `Remote` | `string` | Required | Target address. TCP: `host:port`, Unix: `/path/to/socket` or `unix:/path` |
| `KeepAlive` | `time.Duration` | `5s` | TCP keep-alive interval. |
| `DialTimeout` | `time.Duration` | `2s` | Connection establishment timeout. |
| `ReadTimeout` | `time.Duration` | `2s` | Read deadline for connections. |
| `WriteTimeout` | `time.Duration` | `2s` | Write deadline for connections. |
| `ReconnectDelay` | `time.Duration` | `0` | Delay between reconnection attempts. |
| `WaitReconnect` | `Waiter` | Constant waiter | Custom reconnection wait function. |
| `OnSocket` | `func` | `nil` | Raw socket configuration callback. |
| `OnConnect` | `func` | `nil` | Connection establishment callback. |
| `OnSession` | `func` | `nil` | Session handler (blocks until done). |
| `OnDisconnect` | `func` | `nil` | Disconnection callback. |
| `OnError` | `func` | `nil` | Error handler callback. |

## Client Methods

### Core Methods

```go
// New creates a new Client with options.
func New(cfg *Config, options ...OptionFunc) (*Client, error)

// Connect initiates the connection and starts the reconnection loop.
func (c *Client) Connect() error

// Config returns the configuration object.
func (c *Client) Config() *Config

// Reload attempts to apply configuration changes.
// Note: Currently returns ErrTODO.
func (c *Client) Reload() error

// Wait blocks until the client stops and returns the cancellation reason.
func (c *Client) Wait() error

// Err returns the cancellation reason.
func (c *Client) Err() error

// Done returns a channel that watches the client workers.
func (c *Client) Done() <-chan struct{}

// Shutdown initiates a shutdown and waits until done or context timeout.
func (c *Client) Shutdown(ctx context.Context) error
```

### Configuration Methods

```go
// SetDefaults fills gaps in the configuration.
func (cfg *Config) SetDefaults() error

// Valid checks if the configuration is usable.
func (cfg *Config) Valid() error

// ExportDialer creates a net.Dialer from the configuration.
func (cfg *Config) ExportDialer() net.Dialer
```

## Error Handling

The client distinguishes between recoverable and non-recoverable errors.

### Error Types

```go
var (
    // ErrConfigBusy indicates the Config is already in use.
    ErrConfigBusy = core.QuietWrap(fs.ErrPermission,
        "config already in use")

    // ErrRunning indicates the client is already running.
    ErrRunning = errors.New("already running")

    // Additional errors defined in the errors.go file.
)
```

### Error Classification

- **Recoverable**: Network timeouts, connection refused, temporary failures.
- **Non-recoverable**: Context cancellation, configuration errors, explicit
  stop requests.

The `OnError` callback can override the default retry behaviour by returning
a non-nil error to stop reconnection attempts.

## Thread Safety

All client operations are thread-safe. Multiple goroutines can safely:

- Call client methods concurrently.
- Access the client's context.
- Trigger cancellation.

The configuration becomes immutable after creating a client. Attempting to
reuse a configuration for another client returns `ErrConfigBusy`.

## Resource Management

The client properly manages resources:

- Connections are automatically closed on errors.
- Goroutines are cleaned up on shutdown.
- Context cancellation is propagated throughout.
- The `Wait()` method ensures proper shutdown sequencing.

## Helper Functions

### Waiter Functions

```go
// NewConstantWaiter creates a waiter with fixed delay.
func NewConstantWaiter(d time.Duration) func(context.Context) error

// NewImmediateErrorWaiter returns an error immediately.
func NewImmediateErrorWaiter(err error) func(context.Context) error

// NewDoNotReconnectWaiter prevents reconnection.
func NewDoNotReconnectWaiter(err error) func(context.Context) error
```

### Worker Functions

```go
// NewShutdownFunc creates a worker that shuts down gracefully.
func NewShutdownFunc(s Shutdowner, timeout time.Duration) WorkerFunc

// NewCatchFunc creates an error catcher with exceptions.
func NewCatchFunc(nonErrors ...error) CatcherFunc
```

## Implementation Notes

### Configuration Lifecycle

1. Create a `Config` with required fields.
2. Call `SetDefaults()` to fill optional fields (done automatically by `New`).
3. Call `Valid()` to verify the configuration (done automatically by `New`).
4. Pass to `New()` to create a client.
5. Configuration becomes immutable and bound to the client.

### Connection Flow

1. `Connect()` initiates the first connection attempt.
2. On success, `OnConnect` callback is invoked.
3. `OnSession` callback runs (blocks until session ends).
4. On disconnection, `OnDisconnect` callback is invoked.
5. `WaitReconnect` determines the retry delay.
6. Process repeats until context cancellation or fatal error.

### Session Handler Guidelines

The `OnSession` callback should:

- Block until the session is complete.
- Handle all protocol-specific logic.
- Return `nil` to trigger reconnection on completion.
- Return an error to stop reconnection.
- Respect context cancellation.

## Integration with darvaza.org/x/net

The reconnect client integrates seamlessly with other `darvaza.org/x/net`
components:

- Uses the same `net.Dialer` interface.
- Supports all standard network configurations.
- Compatible with the `bind` package for advanced binding.

## Example: Protocol Implementation

```go
import (
    "context"
    "net"
    "time"

    "darvaza.org/x/net/reconnect"
)

func createClient(addr string) (*reconnect.Client, error) {
    cfg := &reconnect.Config{
        Remote:         addr,
        ReconnectDelay: 5 * time.Second,

        OnConnect: func(ctx context.Context, conn net.Conn) error {
            // Send initial handshake.
            return sendHandshake(conn)
        },

        OnSession: func(ctx context.Context) error {
            // Main protocol loop.
            for {
                select {
                case <-ctx.Done():
                    return ctx.Err()
                case msg := <-messages:
                    if err := processMessage(msg); err != nil {
                        return err
                    }
                }
            }
        },

        OnError: func(ctx context.Context, conn net.Conn, err error) error {
            if isTemporary(err) {
                // Retry on temporary errors.
                return nil
            }
            // Stop on permanent errors.
            return err
        },
    }

    return reconnect.New(cfg)
}
```

### Example: Unix Domain Socket Connection

```go
import (
    "context"
    "net"
    "time"

    "darvaza.org/x/net/reconnect"
)

func createUnixClient() (*reconnect.Client, error) {
    cfg := &reconnect.Config{
        // Automatically detected as Unix socket
        Remote:         "/var/run/myapp.sock",
        ReconnectDelay: 5 * time.Second,

        OnConnect: func(ctx context.Context, conn net.Conn) error {
            // Unix socket connected
            logger.Info().Printf("Connected to Unix: %s", conn.RemoteAddr())
            return nil
        },

        OnSession: func(ctx context.Context) error {
            // Handle Unix socket communication
            return handleUnixProtocol(ctx)
        },
    }

    return reconnect.New(cfg)
}

// Alternative: explicit Unix socket prefix
func createExplicitUnixClient() (*reconnect.Client, error) {
    cfg := &reconnect.Config{
        Remote: "unix:/tmp/app.sock",
        // ... rest of configuration
    }
    return reconnect.New(cfg)
}
```

## Network Type Auto-Detection

The client automatically determines the network type based on the `Remote`
string:

| Pattern | Network Type | Example |
|---------|-------------|---------|
| `unix:` prefix | Unix socket | `unix:/var/run/app.sock` |
| Absolute path (`/`) | Unix socket | `/tmp/socket` |
| Contains `.sock` | Unix socket | `./app.sock`, `run/app.sock` |
| All others | TCP | `example.com:8080`, `192.168.1.1:443` |

This implementation provides enterprise-grade reliability for both TCP and Unix
domain socket connections whilst maintaining simplicity and flexibility for
various use cases.
