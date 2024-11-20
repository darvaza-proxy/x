package certpool

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
)

// CertSet keeps a thread-safe set of unique [x509.Certificate]s.
type CertSet struct {
	set.Set[*x509.Certificate, Hash, *x509.Certificate]
}

// NewCertSet creates a [CertSet] optionally taking its initial content as argument.
func NewCertSet(certs ...*x509.Certificate) (*CertSet, error) {
	out := new(CertSet)
	if err := certSetConfig.Init(&out.Set, certs...); err != nil {
		return nil, err
	}
	return out, nil
}

// MustCertSet is like [NewCertSet] but panics on errors.
func MustCertSet(certs ...*x509.Certificate) *CertSet {
	out, err := NewCertSet(certs...)
	if err != nil {
		core.Panic(err)
	}
	return out
}

func certSetHash(leaf *x509.Certificate) (Hash, error) {
	hash, ok := HashCert(leaf)
	if !ok {
		return Hash{}, core.ErrInvalid
	}
	return hash, nil
}

func certSetItemKey(leaf *x509.Certificate) (*x509.Certificate, error) {
	if !validCert(leaf) {
		return nil, core.ErrInvalid
	}
	return leaf, nil
}

func certSetItemMatch(a, b *x509.Certificate) bool {
	return a.Equal(b)
}

// certSetConfig defines the behavior of CertSet's underlying Set implementation,
// including how certificates are hashed, validated, and compared for equality.
var certSetConfig = set.Config[*x509.Certificate, Hash, *x509.Certificate]{
	Hash:      certSetHash,
	ItemKey:   certSetItemKey,
	ItemMatch: certSetItemMatch,
}
