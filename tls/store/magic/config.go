package magic

import (
	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/tls/store/config"
)

// Config ...
type Config struct {
	Logger slog.Logger
}

func (cfg *Config) export() *config.Config {
	return &config.Config{
		Logger: cfg.Logger,
	}
}

// SetDefaults fills any gap in the [Config].
func (cfg *Config) SetDefaults() error {
	if cfg == nil {
		return core.ErrNilReceiver
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	return nil
}

// New creates a new [Store] from the config.
func (cfg *Config) New() (*Store, error) {
	return New(cfg)
}
