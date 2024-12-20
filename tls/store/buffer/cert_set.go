package buffer

import (
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/container/set"

	"darvaza.org/x/tls/x509utils/certpool"
)

// CertSet keeps a thread-safe set of unique [tls.Certificate]s.
type CertSet struct {
	set.Set[*x509.Certificate, certpool.Hash, *tls.Certificate]
}

// Copy copies all certificates that satisfy the condition to the destination [CertSet]
// unless they are already there.
// If a destination isn't provided one will be created.
// If a condition function isn't provided, all certificates not present in the destination
// will be added.
func (cs *CertSet) Copy(dst *CertSet, cond func(*tls.Certificate) bool) *CertSet {
	switch {
	case cs != nil:
		if dst == nil {
			dst = new(CertSet)
		}

		cs.Set.Copy(&dst.Set, cond)
	case dst == nil:
		dst = MustCertSet()
	}

	return dst
}

// Clone creates a copy of the [CertSet].
func (cs *CertSet) Clone() *CertSet {
	if cs == nil {
		return nil
	}

	return cs.Copy(nil, nil)
}

// NewCertSet creates a [CertSet] optionally taking its initial content as argument.
func NewCertSet(certs ...*tls.Certificate) (*CertSet, error) {
	out := new(CertSet)
	if err := certSetConfig.Init(&out.Set, certs...); err != nil {
		return nil, err
	}
	return out, nil
}

// InitCertSet initializes a preallocated [CertSet].
func InitCertSet(out *CertSet, certs ...*tls.Certificate) error {
	if out == nil {
		return core.Wrap(core.ErrInvalid, "missing CertSet")
	}

	return certSetConfig.Init(&out.Set, certs...)
}

// MustCertSet is like [NewCertSet] but panics on errors.
func MustCertSet(certs ...*tls.Certificate) *CertSet {
	out, err := NewCertSet(certs...)
	if err != nil {
		core.Panic(core.Wrap(err, "failed to create CertSet"))
	}
	return out
}

// MustInitCertSet is like [InitCertSet] but panics on errors.
func MustInitCertSet(out *CertSet, certs ...*tls.Certificate) {
	err := InitCertSet(out, certs...)
	if err != nil {
		core.Panic(core.Wrap(err, "failed to initialize CertSet"))
	}
}

func certSetHash(leaf *x509.Certificate) (certpool.Hash, error) {
	hash, ok := certpool.HashCert(leaf)
	if !ok {
		return certpool.Hash{}, core.ErrInvalid
	}
	return hash, nil
}

func certSetItemKey(cert *tls.Certificate) (*x509.Certificate, error) {
	switch {
	case cert == nil, cert.Leaf == nil:
		return nil, core.ErrInvalid
	default:
		return cert.Leaf, nil
	}
}

func certSetItemMatch(leaf *x509.Certificate, cert *tls.Certificate) bool {
	return cert.Leaf.Equal(leaf)
}

// certSetConfig defines the behavior of CertSet's underlying Set implementation,
// including how certificates are hashed, validated, and compared for equality.
var certSetConfig = set.Config[*x509.Certificate, certpool.Hash, *tls.Certificate]{
	Hash:      certSetHash,
	ItemKey:   certSetItemKey,
	ItemMatch: certSetItemMatch,
}
