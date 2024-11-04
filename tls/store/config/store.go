// Package config provides helpers for working with darvaza.org/x/tls.Store objects.
package config

import (
	"context"
	"crypto/x509"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// AddCACerts adds all given certificates as trusted roots to the given [tls.Store].
// PEM content, a PEM fileName, or a directory containing PEM files.
func AddCACerts(ctx context.Context, s tls.StoreX509Writer, roots string) (int, error) {
	var cfg Config
	return cfg.AddCACerts(ctx, s, roots)
}

// AddCert adds all given certificates to the specified [tls.Store].
func AddCert(ctx context.Context, s tls.StoreX509Writer, cert string) error {
	var cfg Config
	return cfg.AddCert(ctx, s, cert)
}

// AddCertPair adds a cert-key pair to the specified [tls.Store].
func AddCertPair(ctx context.Context, s tls.StoreX509Writer, key, cert string) error {
	var cfg Config
	return cfg.AddCertPair(ctx, s, key, cert)
}

// AddPrivateKey adds a private key to the specified [tls.Store].
func AddPrivateKey(ctx context.Context, s tls.StoreX509Writer, key string) error {
	var cfg Config
	return cfg.AddPrivateKey(ctx, s, key)
}

// ImportCACerts adds all CACerts in a given [x509utils.CertPool] to the specified [tls.Store].
func ImportCACerts(ctx context.Context, s tls.StoreX509Writer, src x509utils.CertPool) (int, error) {
	var count int

	switch {
	case s == nil:
		return 0, tls.ErrNoStore
	case src == nil:
		return 0, nil
	default:
		src.ForEach(ctx, func(_ context.Context, cert *x509.Certificate) bool {
			if s.AddCACerts(ctx, cert) == nil {
				count++
			}
			// abort on time-out
			return ctx.Err() == nil
		})

		return count, ctx.Err()
	}
}

// ImportCerts adds all Certs in a given [x509utils.CertPool] to the specified [tls.Store].
func ImportCerts(ctx context.Context, s tls.StoreX509Writer, src x509utils.CertPool) (int, error) {
	var count int
	switch {
	case s == nil:
		return 0, tls.ErrNoStore
	case src == nil:
		return 0, nil
	default:
		src.ForEach(ctx, func(_ context.Context, cert *x509.Certificate) bool {
			if s.AddCert(ctx, cert) == nil {
				count++
			}
			// abort on time-out
			return ctx.Err() == nil
		})

		return count, ctx.Err()
	}
}
