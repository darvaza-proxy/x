// Package autocert implements a TLS store via a basic ACME client
package autocert

import (
	"context"
	"crypto/x509"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/basic"
)

var _ tls.Store = (*Store)(nil)

// Store ...
type Store struct {
	mu  sync.RWMutex
	cfg Config

	ss *basic.Store
}

func (s *Store) initStore() error {
	ss := new(basic.Store)

	sc := s.cfg.export()
	ctx := context.Background()
	if err := sc.Apply(ctx, ss); err != nil {
		return err
	}

	s.ss = ss
	return nil
}

// GetCAPool ...
func (s *Store) GetCAPool() *x509.CertPool {
	if s == nil {
		return x509.NewCertPool()
	}

	return s.ss.GetCAPool()
}

// GetCertificate ...
func (s *Store) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if s == nil {
		return nil, core.ErrNilReceiver
	}

	return nil, core.ErrTODO
}

// New ...
func New(cfg *Config) (*Store, error) {
	if cfg == nil {
		cfg = new(Config)
	}

	s := &Store{
		cfg: *cfg,
	}

	if err := s.unsafeInit(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) init() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	// RO
	s.mu.RLock()
	defer s.mu.Unlock()

	return s.unsafeInit()
}

func (s *Store) unsafeInit() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.ss == nil:
		// init
		for _, fn := range []func() error{
			s.cfg.SetDefaults,
			s.initStore,
		} {
			if err := fn(); err != nil {
				return err
			}
		}
	}

	return nil
}
