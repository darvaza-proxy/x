// Package certbotdir implements a read-only TLS store
// based on certbot's filesystem structure.
package certbotdir

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

var _ tls.Store = (*Store)(nil)

// Store ...
type Store struct {
	mu    sync.Mutex
	cfg   Config
	roots x509utils.CertPool
}

// GetCAPool ...
func (s *Store) GetCAPool() *x509.CertPool {
	if err := s.init(); err != nil {
		return nil
	}

	return s.roots.Export()
}

// GetCertificate ...
func (s *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	if chi != nil && chi.ServerName != SelfSignedID {
		// TODO: try wildcards
		cert, err := s.getbyServerName(chi.ServerName)
		switch {
		case cert != nil:
			return cert, nil
		case err != nil && !os.IsNotExist(err):
			return nil, err
		}
	}

	return s.getSelfSigned()
}

func (s *Store) getbyServerName(serverName string) (*tls.Certificate, error) {
	host, _, err := core.SplitHostPort(serverName)
	if err != nil {
		return nil, core.Wrapf(core.ErrInvalid, "invalid serverName: %q", serverName)
	}

	return s.getByName(host)
}

func (s *Store) getSelfSigned() (*tls.Certificate, error) {
	return s.getByName(SelfSignedID)
}

func (s *Store) init() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.roots == nil {
		// first
		if err := s.cfg.SetDefaults(); err != nil {
			return err
		}

		// sanitize
		s.cfg.BaseDirectory = filepath.Clean(s.cfg.BaseDirectory)
		s.cfg.LiveDirectory = filepath.Clean(s.cfg.LiveDirectory)

		s.roots = s.cfg.Roots
	}

	return nil
}

// New creates a [Store] using the given [Config].
func New(cfg *Config) (*Store, error) {
	if cfg == nil {
		cfg = new(Config)
	}

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
