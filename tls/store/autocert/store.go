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

func (acs *Store) initStore() error {
	ss := new(basic.Store)

	sc := acs.cfg.export()
	ctx := context.Background()
	if err := sc.Apply(ctx, ss); err != nil {
		return err
	}
	acs.ss = ss
	return nil
}

// GetCAPool ...
func (acs *Store) GetCAPool() *x509.CertPool {
	if acs == nil {
		return x509.NewCertPool()
	}

	return acs.ss.GetCAPool()
}

// GetCertificate ...
func (acs *Store) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if acs == nil {
		return nil, core.ErrNilReceiver
	}

	return nil, core.ErrTODO
}

// New ...
func New(cfg *Config) (*Store, error) {
	if cfg == nil {
		cfg = new(Config)
	}

	acs := &Store{
		cfg: *cfg,
	}

	if err := acs.unsafeInit(); err != nil {
		return nil, err
	}

	return acs, nil
}

func (acs *Store) init() error {
	if acs == nil {
		return core.ErrNilReceiver
	}

	// RO
	acs.mu.RLock()
	ok := acs.isInitialized()
	acs.mu.Unlock()

	if ok {
		return nil
	}

	// RW
	acs.mu.Lock()
	defer acs.mu.Unlock()

	if acs.isInitialized() {
		return nil
	}

	return acs.unsafeInit()
}

func (acs *Store) isInitialized() bool {
	switch {
	case acs == nil, acs.ss == nil:
		return false
	default:
		return true
	}
}

func (acs *Store) unsafeInit() error {
	switch {
	case acs == nil:
		return core.ErrNilReceiver
	case !acs.isInitialized():
		// init
		for _, fn := range []func() error{
			acs.cfg.SetDefaults,
			acs.initStore,
		} {
			if err := fn(); err != nil {
				return err
			}
		}
	}

	return nil
}
