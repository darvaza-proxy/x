package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// A Store is used to set up a [tls.Config].
type Store interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetCAPool() *x509.CertPool
}

// WithStore binds a given [Store] to the [tls.Config]
func WithStore(cfg *tls.Config, store Store) error {
	if cfg == nil {
		return fmt.Errorf("missing argument: %s", "cfg")
	}

	if store == nil {
		return fmt.Errorf("missing argument: %s", "store")
	}

	pool := store.GetCAPool()
	if pool == nil {
		return fmt.Errorf("missing parameter: %s", "CAPool")
	}

	cfg.GetCertificate = store.GetCertificate
	cfg.RootCAs = pool
	cfg.ClientCAs = pool
	return nil
}

// NewConfig returns a basic [tls.Config] optionally configured to use the given Store.
func NewConfig(store Store) (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if store != nil {
		if err := WithStore(cfg, store); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
