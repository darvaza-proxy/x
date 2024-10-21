package lru

import (
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls"
)

const (
	// MinimumCacheSize defines the smallest acceptable size for the LRU cache.
	MinimumCacheSize = 8
	// DefaultCacheSize defines the LRU cache size used if a size smaller than the minimum is specified.
	DefaultCacheSize = 32
)

// OnAddFunc is a callback function type for handling certificate additions to the LRU cache.
type OnAddFunc func(string, *tls.Certificate, int, time.Time)

// OnEvictFunc is a callback function type for handling certificate evictions from the LRU cache.
type OnEvictFunc func(string, *tls.Certificate, int)

// Config defines the configuration parameters for an LRU-based TLS certificate store.
// It allows customization of the cache behavior, logging, and optional callback functions
// for monitoring certificate additions and evictions.
type Config struct {
	Logger slog.Logger

	Upstream tls.Store
	Workers  int
	MaxSize  int

	OnAdd   OnAddFunc
	OnEvict OnEvictFunc
}

// SetDefaults fills gaps in the [Config].
func (cfg *Config) SetDefaults() error {
	if cfg == nil {
		return core.ErrNilReceiver
	}

	if cfg.OnAdd == nil {
		cfg.OnAdd = func(_ string, _ *tls.Certificate, _ int, _ time.Time) {}
	}

	if cfg.OnEvict == nil {
		cfg.OnEvict = func(_ string, _ *tls.Certificate, _ int) {}
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
		return ErrNoUpstream
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
	for _, fn := range []func(*LRU, *Config) error{
		doInitUpstream,
	} {
		if err := fn(s, cfg); err != nil {
			return nil, err
	}

	return s, nil
}

func doInitUpstream(s *LRU, cfg *Config) error {
	s.upstream = cfg.Upstream
	s.logger = cfg.Logger
	s.callOnAdd = cfg.OnAdd
	s.callOnEvict = cfg.OnEvict
}
