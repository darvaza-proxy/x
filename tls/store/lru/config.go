package lru

import (
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/tls"
)

const (
	// MiB is a megabyte.
	MiB = 1024 * 1024

	// MinimumCacheSize defines the smallest acceptable size for the LRU cache.
	MinimumCacheSize = 1 * MiB
	// DefaultCacheSize defines the amount of certificates the [LRU] will store if no value is
	// specified in the [Config]
	DefaultCacheSize = 8 * MiB

	// MinimumWorkers specifies the minimum acceptable number of workers we can have.
	MinimumWorkers = 1
	// DefaultWorkers specifies the default number of workers we will have if none is
	// specified in the [Config]
	DefaultWorkers = 4
)

// OnAddFunc is a callback function type for handling certificate additions to the LRU cache.
type OnAddFunc func(string, *tls.Certificate, int, time.Time)

// OnEvictFunc is a callback function type for handling certificate evictions from the LRU cache.
type OnEvictFunc func(string, *tls.Certificate, int)

// Config defines the configuration parameters for an LRU-based TLS certificate store.
// It allows customization of the cache behavior, logging, and optional callback functions
// for monitoring certificate additions and evictions.
type Config struct {
	Logger     slog.Logger
	Upstream   tls.Store
	MaxWorkers int
	MaxSize    int

	OnAdd   OnAddFunc
	OnEvict OnEvictFunc
}

// SetDefaults fills gaps in the [Config].
func (cfg *Config) SetDefaults() error {
	if cfg == nil {
		return core.ErrNilReceiver
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	if cfg.MaxWorkers == 0 {
		cfg.MaxWorkers = DefaultWorkers
	}

	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultCacheSize
	}

	return nil
}

// Validate checks the configuration for correctness.
func (cfg *Config) Validate() error {
	switch {
	case cfg == nil:
		return core.ErrNilReceiver
	case cfg.Upstream == nil:
		return core.Wrap(core.ErrInvalid, "missing upstream store")
	case cfg.MaxWorkers < MinimumWorkers:
		return core.Wrapf(core.ErrInvalid, "minimum workers %d", MinimumWorkers)
	case cfg.MaxSize < MinimumCacheSize:
		return core.Wrapf(core.ErrInvalid, "minimum cache size %d", MinimumCacheSize)
	default:
		return nil
	}
}

// New creates a new [LRU] from the [Config]. Validate() is called
// internally but SetDefaults() is not.
func (cfg *Config) New() (*LRU, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := new(LRU)
	s.unsafeInit(cfg)
	return s, nil
}
