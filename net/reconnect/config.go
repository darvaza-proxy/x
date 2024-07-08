package reconnect

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/config"
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

	// ReadTimeout indicates the default what to use for the connection's
	// read deadline. zero or negative means the deadline should be disabled.
	ReadTimeout time.Duration `default:"2s"`
	// WriteTimeout indicates the default what to use for the connection's
	// write deadline. zero or negative means the deadline should be disabled.
	WriteTimeout time.Duration `default:"2s"`

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
	return nil
}

// Valid checks if the [Config] is fit to be used.
func (cfg *Config) Valid() error {
	// basic rules
	switch {
	case cfg.Context == nil:
		return errors.New("context missing")
	default:
		// TODO: more rules
	}

	if cfg.busy() {
		// bound rules
		return cfg.validateBusy()
	}

	return nil
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
