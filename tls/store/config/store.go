// Package config provides helpers for working with darvaza.org/x/tls.Store objects.
package config

import (
	"context"
	"crypto/x509"
	"io/fs"
	"os"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// ReadStringPEM works over raw PEM data or a filename reading PEM blocks
// and invoking a callback for each.
//
// As opposed to x509utils.ReadStringPEM, this function doesn't support
// directories.
func ReadStringPEM(value string, fn x509utils.DecodePEMBlockFunc) error {
	if x509utils.ReadPEM([]byte(value), fn) == nil {
		// raw. done.
		return nil
	}

	b, err := os.ReadFile(value)
	if pe, ok := err.(*os.PathError); ok {
		// bad path
		if pe.Err == os.ErrInvalid {
			err = pe.Err
		}
	}

	if err != nil {
		return err
	}

	err = x509utils.ReadPEM(b, fn)
	if err != nil {
		return &fs.PathError{
			Op:   "pem.Decode",
			Path: value,
			Err:  err,
		}
	}

	return nil
}

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
