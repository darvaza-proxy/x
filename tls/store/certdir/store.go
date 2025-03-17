// Package certdir uses an on-disk certificate store.
package certdir

import (
	"context"
	"crypto/x509"
	"sync"

	"darvaza.org/cache/x/simplelru"
	"darvaza.org/core"
	"darvaza.org/x/tls"
)

var _ tls.Store = (*Store)(nil)

type StoreOption func(*Store) error

type Store struct {
	mu  sync.RWMutex
	wg  core.ErrGroup
	cfg Config

	roots *x509.CertPool
	cache *simplelru.LRU[string, *meta]
}

type meta struct {
	cert *tls.Certificate
}

func (s *Store) init() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	// RO
	s.mu.RLock()
	ready := s.unsafeIsReady()
	s.mu.RUnlock()

	if ready {
		return nil
	}

	// RW
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.unsafeInit()
}

func (s *Store) unsafeIsReady() bool {
	return s.roots != nil && s.cache != nil
}

func (s *Store) unsafeInit(options ...StoreOption) error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.unsafeIsReady():
		return nil
	default:
		return s.doUnsafeInit(options...)
	}
}

func (s *Store) doUnsafeInit(options ...StoreOption) error {
	for _, opt := range options {
		if err := opt(s); err != nil {
			return err
		}
	}

	for _, fn := range []func() error{
		s.initDefaults,
		s.initRoots,
		s.initCache,
		s.initFSNotify,
	} {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) initDefaults() error {
	if err := s.cfg.SetDefaults(); err != nil {
		return err
	}

	if s.wg.Parent == nil {
		s.wg.Parent = context.Background()
	}

	return nil
}

func (s *Store) initRoots() error
func (s *Store) initCache() error
func (s *Store) initFSNotify() error

func (s *Store) GetCAPool() *x509.CertPool {
	if s.init() != nil {
		return x509.NewCertPool()
	}

	return s.roots
}

func (s *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	switch {
	case s == nil:
		return nil, core.ErrNilReceiver
	case chi == nil:
		return nil, core.Wrap(core.ErrInvalid, "CHI not provided")
	default:
		if err := s.init(); err != nil {
			return nil, err
		}

		ctx, serverName, err := tls.SplitClientHelloInfo(chi)
		if err != nil {
			return nil, core.Wrap(err, "bad CHI")
		}

		m, _, ok := s.cache.Get(serverName)
		if ok {
			// hit
			return m.cert, nil
		}

		// miss
		return s.doGetCertificate(ctx, serverName, chi)
	}
}

func (s *Store) doGetCertificate(ctx context.Context, serverName string, chi *tls.ClientHelloInfo) (*tls.Certificate, error)
