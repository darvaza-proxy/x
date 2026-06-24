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

// NewErrInvalidCert builds an [ErrInvalidCert] reporting that cert was not
// acceptable for the stated reason, optionally wrapping the underlying err.
// Both cert and err may be nil; when err is nil the error defaults to wrapping
// [core.ErrInvalid], so every [ErrInvalidCert] from this constructor satisfies
// errors.Is(err, core.ErrInvalid).
func NewErrInvalidCert(cert *x509.Certificate, err error, reason string) *ErrInvalidCert {
	if err == nil {
		err = core.ErrInvalid
	}
	return &ErrInvalidCert{
		Cert:   cert,
		Err:    err,
		Reason: reason,
	}
}

func (err ErrInvalidCert) Error() string {
	s := make([]string, 0, 3)
	s = append(s, "invalid certificate")

	if err.Reason != "" {
		s = append(s, err.Reason)
	}

	// core.ErrInvalid is the default cause and adds nothing beyond the
	// "invalid certificate" prefix, so it is left out of the message; a
	// caller-supplied cause is always shown.
	if err.Err != nil && err.Err != core.ErrInvalid {
		s = append(s, err.Err.Error())
	}

	return strings.Join(s, ": ")
}

func (err ErrInvalidCert) Unwrap() error {
	return err.Err
}
