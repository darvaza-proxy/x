// Package reconnect implement a generic retrying TCP client
package reconnect

import (
	"bufio"
	"context"
	"net"
	"sync"
	"time"

	"darvaza.org/core"
)

// Client is a reconnecting TCP Client
type Client struct {
	mu     sync.Mutex
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelCauseFunc
	err    error

	cfg *Config

	readTimeout  time.Duration
	writeTimeout time.Duration

	conn net.Conn
	in   *bufio.Reader
	out  *bufio.Writer
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

// New creates a new [Client] using the given [Config] and options
func New(cfg *Config, options ...OptionFunc) (*Client, error) {
	cfg, err := prepareNewConfig(cfg, options...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(cfg.Context)

	c := &Client{
		ctx:    ctx,
		cancel: cancel,

		cfg: cfg,

		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,
	}

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
