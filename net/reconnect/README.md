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
    "time"

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

// Abstract Unix socket with @ prefix (Linux, auto-detected)
cfg := &reconnect.Config{
    Remote: "@app",
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

    // Called after the connection has been closed.
    OnDisconnect: func(ctx context.Context, conn net.Conn) error {
        logger.Info().Printf("Disconnected from %s", conn.RemoteAddr())
        // Clean up connection-specific resources.
        return nil
    },

    // Called when errors occur. The returned error replaces the
    // original for the reconnection logic.
    OnError: func(ctx context.Context, conn net.Conn, err error) error {
        logger.Error().
            WithField("error", err).
            Printf("Connection error")
        // Return nil to discard the error and keep retrying,
        // or reconnect.ErrDoNotReconnect to stop the client.
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
| --- | --- | --- | --- |
| `Context` | `context.Context` | `context.Background()` | Base context for the client. |
| `Logger` | `slog.Logger` | Default logger | Logger for structured logging. |
| `Remote` | `string` | Required | Target address. TCP: `host:port`, Unix: `/path/to/socket` or `unix:/path` |
| `KeepAlive` | `time.Duration` | `5s` | TCP keep-alive interval (ignored for Unix sockets). |
| `DialTimeout` | `time.Duration` | `2s` | Connection establishment timeout. |
| `ReadTimeout` | `time.Duration` | `2s` | Default read deadline, applied via the `Reset*Deadline` helpers, not automatically. |
| `WriteTimeout` | `time.Duration` | `2s` | Default write deadline, applied via the `Reset*Deadline` helpers, not automatically. |
| `ReconnectDelay` | `time.Duration` | `0` | Delay between reconnection attempts. Zero means 5s (`DefaultWaitReconnect`); negative disables reconnection. |
| `WaitReconnect` | `Waiter` | `NewConstantWaiter(ReconnectDelay)` | Custom reconnection wait function. |
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

// Connect starts the reconnection loop. A nil return doesn't mean
// a connection is established, only that the loop has started.
func (c *Client) Connect() error

// Config returns the configuration object.
func (c *Client) Config() *Config

// Reload attempts to apply configuration changes.
// Note: Currently returns ErrTODO.
func (c *Client) Reload() error

// Wait blocks until the client stops and returns the cancellation
// reason, nil when the shutdown was user-initiated.
func (c *Client) Wait() error

// Err returns the cancellation reason.
func (c *Client) Err() error

// Done returns a channel that closes once the client has stopped.
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

### Utility Functions

```go
// ParseRemote determines network type and address from a remote string.
// Returns (network, address, error) where network is reconnect.NetworkTCP or
// reconnect.NetworkUnix.
func ParseRemote(remote string) (network, address string, err error)

// ValidateRemote validates a remote address for TCP or Unix socket connection.
// Returns nil if the address is valid for either protocol.
func ValidateRemote(remote string) error

// TimeoutToAbsoluteTime adds duration to base time.
// Returns zero time if duration is negative.
func TimeoutToAbsoluteTime(base time.Time, d time.Duration) time.Time
```

#### Using Utility Functions

```go
import "darvaza.org/x/net/reconnect"

// Parse and validate remote addresses
network, address, err := reconnect.ParseRemote("unix:/var/run/app.sock")
if err != nil {
    // Handle parsing error
}
// network = "unix", address = "/var/run/app.sock"

// Validate before using in configuration
if err := reconnect.ValidateRemote("example.com:8080"); err != nil {
    // Handle invalid address
}

// Convert relative timeout to absolute time
deadline := reconnect.TimeoutToAbsoluteTime(time.Now(), 30*time.Second)
```

## Error Handling

The client distinguishes between recoverable and non-recoverable errors.

### Error Types

```go
var (
    // ErrConfigBusy indicates the Config is already in use.
    ErrConfigBusy = core.QuietWrap(fs.ErrPermission,
        "config already in use")

    // ErrRunning indicates the client has already been started.
    ErrRunning = core.QuietWrap(syscall.EBUSY,
        "client already running")

    // ErrDoNotReconnect tells the client to stop reconnecting.
    ErrDoNotReconnect = errors.New("don't reconnect")

    // ErrNotConnected indicates the client isn't currently connected.
    ErrNotConnected = core.QuietWrap(fs.ErrClosed, "not connected")

    // Additional errors defined in the errors.go file.
)
```

### Error Classification

- **Fatal**: `ErrDoNotReconnect`, possibly wrapped, terminates the client.
- **Context termination**: the client stops retrying once its context is
  cancelled or its deadline expires.
- **Recoverable**: everything else — connection refused/reset/aborted,
  timeouts, and other session errors — leads to a reconnection attempt.

The `OnError` callback observes every error, and its return value replaces
it: return `nil` to discard the error, or `ErrDoNotReconnect` to stop the
client.

## Thread Safety

All client operations are thread-safe. Multiple goroutines can safely:

- Call client methods concurrently.
- Access the client's context.
- Trigger cancellation.

The configuration becomes immutable after creating a client. Attempting to
reuse a configuration for another client returns `ErrConfigBusy`.

## Resource Management

The client properly manages resources:

- Connections are closed at the end of each session.
- Goroutines are cleaned up on shutdown.
- Context cancellation is propagated to callbacks and workers.
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
- Return `nil` or a non-fatal error to trigger reconnection.
- Return `ErrDoNotReconnect`, possibly wrapped, to stop the client.
- Respect context cancellation for a graceful wind-down; a session left
  blocked on a read is unblocked by shutdown closing the connection.

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
            return reconnect.ErrDoNotReconnect
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
| --- | --- | --- |
| `unix:` prefix | Unix socket | `unix:/var/run/app.sock` |
| `@` prefix | Unix socket (abstract namespace, Linux) | `@app` |
| Absolute path (`/`) | Unix socket | `/tmp/socket` |
| Ends with `.sock` | Unix socket | `./app.sock`, `run/app.sock` |
| All others | TCP | `example.com:8080`, `192.168.1.1:443` |
