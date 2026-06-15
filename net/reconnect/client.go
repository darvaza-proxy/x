// Package reconnect implements a generic retrying network client.
package reconnect

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/net"
	"darvaza.org/x/sync/workgroup"
)

var (
	_ WorkGroup = (*Client)(nil)
)

// Client is a reconnecting network client.
type Client struct {
	ctx    context.Context
	conn   net.Conn
	logger slog.Logger
	cfg    *Config

	waitReconnect func(context.Context) error
	onConnect     func(context.Context, net.Conn) error
	onSession     func(context.Context) error
	onDisconnect  func(context.Context, net.Conn) error
	onError       func(context.Context, net.Conn, error) error

	dialer  net.Dialer
	network string
	address string

	wg workgroup.Group
	mu sync.Mutex

	readTimeout  time.Duration
	writeTimeout time.Duration

	started atomic.Bool
}

// Config returns the [Config] object used when [Client.Reload] is called.
func (c *Client) Config() *Config {
	return c.cfg
}

// Reload attempts to apply changes done to the [Config] since the
// last time, or since created.
func (c *Client) Reload() error {
	if err := c.cfg.Valid(); err != nil {
		return err
	}

	return core.ErrTODO
}

// Connect launches the [Client], failing with [ErrRunning] when
// called more than once. A nil return means the reconnection loop
// has started, not that a connection is established — a failed
// first dial is retried in the background like any other
// disconnection. Calling it on a [Client] that has already been
// shut down fails with [ErrClosed].
func (c *Client) Connect() error {
	// once
	if !c.started.CompareAndSwap(false, true) {
		return ErrRunning
	}

	network, address := c.getRemote()
	conn, err := c.dial(network, address)
	if err = c.handleConnectError(conn, err); err != nil {
		return err
	}

	if err := c.wg.Go(func(context.Context) {
		c.run(conn)
	}); err != nil {
		// a concurrent Shutdown/terminate cancelled the group between
		// the dial and the spawn; don't leak the freshly dialled
		// connection, and report the client is shut down rather than
		// leaking the workgroup's internal error.
		unsafeClose(conn)
		return ErrClosed
	}

	return nil
}

// New creates a new [Client] using the given [Config] and options.
func New(cfg *Config, options ...OptionFunc) (*Client, error) {
	cfg, err := prepareNewConfig(cfg, options...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		wg: workgroup.Group{
			Parent: cfg.Context,
		},
		cfg:    cfg,
		dialer: cfg.ExportDialer(),
		logger: cfg.Logger,

		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,

		waitReconnect: cfg.WaitReconnect,
		onConnect:     cfg.OnConnect,
		onSession:     cfg.OnSession,
		onDisconnect:  cfg.OnDisconnect,
		onError:       cfg.OnError,
	}
	c.ctx = c.wg.Context()

	c.network, c.address, err = ParseRemote(cfg.Remote)
	if err != nil {
		return nil, err
	}

	cfg.unsafeBindClient(c)
	return c, nil
}

// Must is like [New] but it panics on errors.
func Must(cfg *Config, options ...OptionFunc) *Client {
	return core.Must(New(cfg, options...))
}

func (c *Client) getRemote() (network, addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.unsafeGetRemote()
}

func (c *Client) unsafeGetRemote() (network, addr string) {
	return c.network, c.address
}

func (c *Client) getWaitReconnect() func(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.waitReconnect
}

func (c *Client) getOnConnect() func(context.Context, net.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.onConnect
}

func (c *Client) getOnSession() func(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.onSession
}

func (c *Client) getOnDisconnect() func(context.Context, net.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.onDisconnect
}

func (c *Client) getOnError() func(context.Context, net.Conn, error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.onError
}
