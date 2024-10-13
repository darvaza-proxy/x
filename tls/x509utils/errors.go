package x509utils

import (
	"crypto/x509"
	"strings"

	"darvaza.org/core"
)

var (
	_ error            = (*ErrInvalidCert)(nil)
	_ core.Unwrappable = (*ErrInvalidCert)(nil)
)

// ErrInvalidCert indicates the certificate wasn't acceptable.
type ErrInvalidCert struct {
	Cert   *x509.Certificate
	Err    error
	Reason string
}

func (err ErrInvalidCert) Error() string {
	s := make([]string, 0, 3)
	s = append(s, "invalid certificate")

	if err.Reason != "" {
		s = append(s, err.Reason)
	}

	if err.Err != nil {
		s = append(s, err.Err.Error())
	}

	return strings.Join(s, ": ")
}

func (err ErrInvalidCert) Unwrap() error {
	return err.Err
}
