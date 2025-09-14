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
)

var (
	_ WorkGroup = (*Client)(nil)
)

// Client is a reconnecting network client.
type Client struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelCauseFunc
	started atomic.Bool
	err     error

	cfg     *Config
	dialer  net.Dialer
	network string
	address string
	logger  slog.Logger

	readTimeout  time.Duration
	writeTimeout time.Duration

	waitReconnect func(context.Context) error
	onConnect     func(context.Context, net.Conn) error
	onSession     func(context.Context) error
	onDisconnect  func(context.Context, net.Conn) error
	onError       func(context.Context, net.Conn, error) error

	conn net.Conn
}

// Config returns the [Config] object used when [Reload] is called.
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

// Connect launches the [Client].
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

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.run(conn)
	}()

	return nil
}

// New creates a new [Client] using the given [Config] and options.
func New(cfg *Config, options ...OptionFunc) (*Client, error) {
	cfg, err := prepareNewConfig(cfg, options...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(cfg.Context)

	c := &Client{
		ctx:    ctx,
		cancel: cancel,

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

	c.network, c.address = parseRemote(cfg.Remote)

	cfg.unsafeBindClient(c)
	return c, nil
}

// Must is like [New] but it panics on errors.
func Must(cfg *Config, options ...OptionFunc) *Client {
	c, err := New(cfg, options...)
	if err != nil {
		core.Panic(err)
	}
	return c
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
