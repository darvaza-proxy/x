// Package lru implements a simple LRU cache.
package lru

import (
	"crypto/x509"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"darvaza.org/cache/x/simplelru"
	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls"
)

var _ tls.Store = (*LRU)(nil)

// LRU represents a Least Recently Used (LRU) cache for TLS certificates.
// It manages a cache of certificates with a maximum size, providing thread-safe
// operations and optional callbacks for certificate addition and eviction.
type LRU struct {
	mu sync.RWMutex

	lru *simplelru.LRU[string, *tls.Certificate]
	sf  singleflight.Group

	logger      slog.Logger
	upstream    tls.Store
	callOnAdd   OnAddFunc
	callOnEvict OnEvictFunc
}

func (s *LRU) readLock() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.upstream == nil:
		return ErrNoUpstream
	default:
		s.mu.RLock()
		return nil
	}
}

func (s *LRU) readUnlock() { s.mu.RUnlock() }

func (s *LRU) writeLock() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.upstream == nil:
		return ErrNoUpstream
	default:
		s.mu.Lock()
	}
}

func (s *LRU) writeUnlock() { s.mu.Unlock() }

// GetCAPool returns the CA certificate pool from the upstream TLS store.
// If the LRU cache or its upstream store is nil, it returns nil.
func (s *LRU) GetCAPool() *x509.CertPool {
	if err := s.readLock(); err != nil {
		return nil
	}
	defer s.mu.RUnlock()

	return s.upstream.GetCAPool()
}

// ForEach iterates over the LRU cache entries, calling the provided function for each entry.
func (s *LRU) ForEach(fn func(string, *tls.Certificate, int, time.Time) bool) error {
	if err := s.init(); err != nil {
		return err
	} else if fn == nil {
		return nil
	}

	return nil
}

// func (s *LRU) exportEntries() []*Entry {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	return nil
// }
