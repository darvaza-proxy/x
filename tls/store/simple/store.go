// Package simple implements a generic programmable TLS store
package simple

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

var _ tls.Store = (*Store)(nil)

// Store ...
type Store struct{}

// GetCAPool ...
func (*Store) GetCAPool() *x509.CertPool { return nil }

// GetCertificate ...
func (*Store) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return nil, core.ErrTODO
}

// New ...
func New() (*Store, error) {
	return &Store{}, nil
}
