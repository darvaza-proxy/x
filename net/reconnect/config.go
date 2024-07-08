package reconnect

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/config"
	"darvaza.org/x/net"
)

// A OptionFunc modifies a [Config] consistently before SetDefaults() and Validate().
type OptionFunc func(*Config) error

var (
	// ErrConfigBusy indicates the [Config] is in used and can't
	// be used to create another [Client].
	ErrConfigBusy = core.QuietWrap(fs.ErrPermission, "config already in use")
)

// Config describes the operation of the Client.
type Config struct {
	Context context.Context

	// Remote indicates the `host:port` address of the remote.
	Remote string

	// KeepAlive indicates the value to be set to TCP connections
	// for the low level keep alive messages.
	KeepAlive time.Duration `default:"5s"`
	// DialTimeout indicates how long are we willing to wait for new
	// connections getting established.
	DialTimeout time.Duration `default:"2s"`
	// ReadTimeout indicates the default what to use for the connection's
	// read deadline. zero or negative means the deadline should be disabled.
	ReadTimeout time.Duration `default:"2s"`
	// WriteTimeout indicates the default what to use for the connection's
	// write deadline. zero or negative means the deadline should be disabled.
	WriteTimeout time.Duration `default:"2s"`

	// ReconnectDelay specifies how long to wait between re-connections
	// unless [WaitReconnect] is specified. Negative implies reconnecting is disabled.
	ReconnectDelay time.Duration
	// WaitReconnect is a helper used to wait between re-connection attempts.
	WaitReconnect Waiter

	// OnSession is expected to block until it's done.
	OnSession func(context.Context) error
	// OnDisconnect is called after closing the connection and can be used to
	// prevent further connection retries.
	OnDisconnect func(context.Context, net.Conn) error
	// OnError is called after all errors and gives us the opportunity to
	// decide how the error should be treated by the reconnection logic.
	OnError func(context.Context, net.Conn, error) error

	// immutable data
	c   *Client
	ctx context.Context
}

func (cfg *Config) unsafeBindClient(c *Client) {
	// set immutables
	cfg.c = c
	cfg.ctx = cfg.Context
}

// busy indicates the immutables have already been set.
func (cfg *Config) busy() bool {
	return cfg.c != nil
}

// SetDefaults fills any gap in the config
func (cfg *Config) SetDefaults() error {
	if err := config.Set(cfg); err != nil {
		return err
	}

	if cfg.Context == nil {
		// either the immutable context or a fresh one
		cfg.Context = core.Coalesce(cfg.ctx, context.Background())
	}

	if cfg.WaitReconnect == nil {
		cfg.WaitReconnect = NewConstantWaiter(cfg.ReconnectDelay)
	}

	return nil
}

// Valid checks if the [Config] is fit to be used.
func (cfg *Config) Valid() error {
	// basic rules
	switch {
	case cfg.Context == nil:
		return errors.New("context missing")
	case cfg.WaitReconnect == nil:
		return errors.New("reconnect waiter missing")
	default:
		if err := cfg.validateRemote(cfg.Remote); err != nil {
			return core.Wrap(err, "invalid remote")
		}

		// TODO: more rules
	}

	if cfg.busy() {
		// bound rules
		return cfg.validateBusy()
	}

	return nil
}

func (*Config) validateRemote(remote string) error {
	_, port, err := core.SplitHostPort(remote)
	switch {
	case err != nil:
		return err
	case port == "":
		return fmt.Errorf("%q: port missing", remote)
	default:
		return nil
	}
}

func (cfg *Config) validateBusy() error {
	switch {
	case cfg.Context != cfg.ctx:
		return errors.New("invalid context")
	default:
		// TODO: more rules
	}

	return nil
}

// ExportDialer creates a [net.Dialer] from the
// [Config].
func (cfg *Config) ExportDialer() net.Dialer {
	return newDialer(cfg.KeepAlive, cfg.DialTimeout)
}

func prepareNewConfig(cfg *Config, options ...OptionFunc) (*Config, error) {
	switch {
	case cfg == nil:
		cfg = new(Config)
	case cfg.busy():
		return nil, ErrConfigBusy
	}

	for _, fn := range options {
		if err := fn(cfg); err != nil {
			return nil, err
		}
	}

	if err := cfg.SetDefaults(); err != nil {
		return nil, err
	}

	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	return cfg, nil
}
