// Package cache implements a caching layer
// in front of another TLS store.
package cache

import (
	"crypto/x509"
	"os"

	"darvaza.org/cache"
	"darvaza.org/cache/x/memcache"
	"darvaza.org/core"
	"darvaza.org/x/tls"
)

var _ tls.Store = (*Store)(nil)

// Store ...
type Store struct {
	cfg Config

	lru *memcache.LRU[string]
	sf  *memcache.SingleFlight[string]
}

// GetCertificate ...
func (s *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if err := s.check(); err != nil {
		return nil, err
	} else if chi == nil {
		err = core.Wrap(core.ErrInvalid, "chi not provided")
		return nil, err
	}

	ctx := WithClientHelloInfo(chi.Context(), chi)
	serverName := chi.ServerName

	// TODO: check IP address support
	if serverName != "" {
		cert, err := s.doGetCertificate(ctx, serverName)
		switch {
		case cert != nil:
			return cert, nil
		case err != nil && !os.IsNotExist(err):
			return nil, err
		}
	}

	return s.doGetCertificate(ctx, s.cfg.Fallback)
}

// GetCAPool ...
func (s *Store) GetCAPool() *x509.CertPool {
	if err := s.check(); err != nil {
		core.Panic(err)
	}

	return s.cfg.Store.GetCAPool()
}

// check makes sure the Store is initialized.
func (s *Store) check() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.sf == nil:
		return core.Wrap(core.ErrInvalid, "not initialized")
	default:
		return nil
	}
}

// init is called by Config.NewStore() before making the object public.
func (s *Store) init() error {
	getter := cache.GetterFunc[string](s.doGet)

	s.lru = memcache.NewLRU(int64(s.cfg.CacheSize), s.cfg.OnAdd, s.cfg.OnEvict)
	s.sf = memcache.NewSingleFlight("tlsStore", s.lru, getter)
	return nil
}
