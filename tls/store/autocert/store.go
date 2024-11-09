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
	wg  core.ErrGroup
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
	if acs.init() != nil {
		return x509.NewCertPool()
	}

	return acs.ss.GetCAPool()
}

// GetCertificate ...
func (acs *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if acs == nil {
		return nil, core.ErrNilReceiver
	} else if err := acs.init(); err != nil {
		return nil, err
	}

	return acs.ss.GetCertificate(chi)
}

// New ...
func New(ctx context.Context, cfg *Config, options ...StoreOption) (*Store, error) {
	if cfg == nil {
		cfg = new(Config)
	}

	acs := &Store{
		cfg: *cfg,
		wg: core.ErrGroup{
			Parent: ctx,
		},
	}

	if err := acs.unsafeInit(options...); err != nil {
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
	return acs != nil && acs.wg.Parent != nil && acs.ss != nil
}

func (acs *Store) unsafeInit(options ...StoreOption) error {
	switch {
	case acs == nil:
		return core.ErrNilReceiver
	case acs.ss != nil:
		return nil
	default:
		return acs.doUnsafeInit(options)
	}
}

func (acs *Store) doUnsafeInit(options []StoreOption) error {
	// apply options
	for _, opt := range options {
		if err := opt(acs); err != nil {
			return err
		}
	}

	// init
	for _, fn := range []func() error{
		acs.cfg.SetDefaults,
		acs.initDefaults,
		acs.initStore,
	} {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (acs *Store) initDefaults() error {
	if acs.wg.Parent == nil {
		acs.wg.Parent = context.Background()
	}

	return nil
}
