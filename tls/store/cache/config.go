package cache

import (
	"time"

	"darvaza.org/cache/x/memcache"
	"darvaza.org/core"
	"darvaza.org/x/tls"
)

// DefaultCacheSize ...
const DefaultCacheSize = 2 * memcache.MiB

// Config ...
type Config struct {
	CacheSize int
	Store     tls.Store
	Fallback  string

	OnAdd   func(serverName string, b []byte, size int64, expiration *time.Time)
	OnEvict func(serverName string, b []byte, size int64)

	NewSink func() (Sink, error)
}

// SetDefaults ...
func (cfg *Config) SetDefaults() error {
	switch {
	case cfg == nil:
		return core.ErrNilReceiver
	case cfg.Store == nil:
		return core.Wrap(core.ErrInvalid, "upstream store not specified")
	}

	if cfg.CacheSize == 0 {
		cfg.CacheSize = DefaultCacheSize
	}

	if cfg.NewSink == nil {
		cfg.NewSink = NewGobSink
	}
	return nil
}

// NewStore creates [tls.Store] middleware for caching [tls.Certificate]s
func (cfg *Config) NewStore() (*Store, error) {
	if err := cfg.SetDefaults(); err != nil {
		return nil, err
	}

	s := &Store{
		cfg: *cfg,
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
