// Package x509utils provides utilities to aid working with x509 certificates
package x509utils

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
)

var (
	_ PrivateKey = (*rsa.PrivateKey)(nil)
	_ PrivateKey = (*ecdsa.PrivateKey)(nil)
	_ PrivateKey = (*ed25519.PrivateKey)(nil)

	_ PublicKey = (*rsa.PublicKey)(nil)
	_ PublicKey = (*ecdsa.PublicKey)(nil)
	_ PublicKey = (*ed25519.PublicKey)(nil)
)

// PrivateKey implements what crypto.PrivateKey should have
type PrivateKey interface {
	Public() crypto.PublicKey
	Equal(crypto.PrivateKey) bool
}

// PublicKey implements what crypto.PublicKey should have
type PublicKey interface {
	Equal(crypto.PublicKey) bool
}

// A CertPool contains x509 certificates and allows individual
// access to them which the standard [x509.CertPool] doesn't.
type CertPool interface {
	Get(ctx context.Context, name string) (*x509.Certificate, error)
	ForEach(ctx context.Context, fn func(context.Context, *x509.Certificate) bool)

	Clone() CertPool
	Export() *x509.CertPool
}

// A CertPoolWriter extends the CertPool interface with write capabilities.
type CertPoolWriter interface {
	CertPool

	Put(ctx context.Context, name string, cert *x509.Certificate) error
	Delete(ctx context.Context, name string) error
	DeleteCert(ctx context.Context, cert *x509.Certificate) error

	Import(ctx context.Context, src CertPool) (int, error)
	ImportPEM(ctx context.Context, b []byte) (int, error)
}
