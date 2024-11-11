package tls

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"darvaza.org/core"
)

// ErrNoStore is an error indicating the [Store] wasn't provided.
var ErrNoStore = core.Wrap(core.ErrInvalid, "store not provided")

// A Store is used to set up a [tls.Config].
type Store interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetCAPool() *x509.CertPool
}

// StoreReader adds read methods to the [Store].
type StoreReader interface {
	Store

	Get(ctx context.Context, name string) (*tls.Certificate, error)

	ForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool)
	ForEachMatch(ctx context.Context, name string, fn func(context.Context, *tls.Certificate) bool)
}

// StoreWriter adds [tls.Certificate] write methods to the [Store].
type StoreWriter interface {
	Store

	Put(ctx context.Context, cert *tls.Certificate) error
	Delete(ctx context.Context, cert *tls.Certificate) error
}

// StoreX509Writer adds [x509.Certificate] write methods to the [Store].
type StoreX509Writer interface {
	Store

	AddCACerts(ctx context.Context, roots ...*x509.Certificate) error

	AddPrivateKey(ctx context.Context, key crypto.Signer) error
	AddCert(ctx context.Context, cert *x509.Certificate) error
	AddCertPair(ctx context.Context, key crypto.Signer, cert *x509.Certificate, intermediates []*x509.Certificate) error

	DeleteCert(ctx context.Context, cert *x509.Certificate) error
}

// StoreReadWriter includes read and write methods for the [Store]
type StoreReadWriter interface {
	StoreReader
	StoreWriter
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
