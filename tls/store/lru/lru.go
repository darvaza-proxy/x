// Package lru implements a simple LRU cache.
package lru

import (
	"context"
	"crypto/x509"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/container/list"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils/certpool"
)

// LRU represents a Least Recently Used (LRU) cache for TLS certificates.
// It manages a cache of certificates with a maximum size, providing thread-safe
// operations and optional callbacks for certificate addition and eviction.
type LRU struct {
	// sync
	mu sync.Mutex
	wq chan func()

	// LRU
	size     int
	count    int
	entries  map[certpool.Hash]*lruEntry
	names    map[string]*list.List[*lruEntry]
	suffixes map[string]*list.List[*lruEntry]
	eviction *list.List[*lruEntry]

	// params
	maxSize     int
	logger      slog.Logger
	upstream    tls.Store
	callOnAdd   OnAddFunc
	callOnEvict OnEvictFunc
}

// Interface assertions
var (
	_ tls.Store       = (*LRU)(nil)
	_ tls.StoreReader = (*LRU)(nil)
	_ tls.StoreWriter = (*LRU)(nil)
)

// GetCAPool returns the CA certificate pool from the upstream TLS store.
// If the LRU cache or its upstream store is nil, it returns nil.
func (s *LRU) GetCAPool() *x509.CertPool {
	if err := s.tryLock(); err != nil {
		return nil
	}
	defer s.unlock()

	return s.upstream.GetCAPool()
}

// Init initializes
func (s *LRU) Init(cfg *Config) error {
	if s == nil {
		return core.ErrNilReceiver
	} else if err := cfg.Validate(); err != nil {
		return err
	} else if s == cfg.Upstream {
		return core.Wrap(core.ErrInvalid, "self reference")
	}

	// RW
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.entries != nil {
		return ErrAlreadyInitialized
	}

	s.unsafeInit(cfg)
	return nil
}

// unsafeInit is called by Config.New() or LRU.Init() after validating the [Config]
func (s *LRU) unsafeInit(cfg *Config) {
	// sync
	s.wq = make(chan func())

	// LRU
	s.entries = make(map[certpool.Hash]*lruEntry)
	s.names = make(map[string]*list.List[*lruEntry])
	s.suffixes = make(map[string]*list.List[*lruEntry])
	s.eviction = list.New[*lruEntry]()

	// params
	s.maxSize = cfg.MaxSize
	s.upstream = cfg.Upstream
	s.logger = cfg.Logger
	s.callOnAdd = cfg.OnAdd
	s.callOnEvict = cfg.OnEvict

	// work
	go s.run(cfg.MaxWorkers)
}

// checkInit checks if the LRU cache has been initialized.
func (s *LRU) checkInit() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.entries == nil:
		return ErrNotInitialized
	default:
		return nil
	}
}

func (s *LRU) checkInitWithContext(ctx context.Context) error {
	if err := s.checkInit(); err != nil {
		return err
	}
	return ctx.Err()
}

// tryLock locks the LRU if it's been initialized.
func (s *LRU) tryLock() error {
	if err := s.checkInit(); err != nil {
		return err
	}

	s.mu.Lock()
	return nil
}

func (s *LRU) unlock() {
	s.mu.Unlock()
}
