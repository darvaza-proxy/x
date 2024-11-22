package certpool

import (
	"crypto"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
	"darvaza.org/x/tls/x509utils"
)

// CertSet keeps a thread-safe set of unique [x509.Certificate]s.
type CertSet struct {
	set.Set[*x509.Certificate, Hash, *x509.Certificate]
}

// GetByKey returns all certificates in the CertSet matching the given public key.
func (cs *CertSet) GetByKey(pub crypto.PublicKey) []*x509.Certificate {
	if cs == nil || pub == nil {
		return nil
	} else if pub1, ok := pub.(x509utils.PublicKey); ok {
		return cs.doGetByKey(pub1)
	}
	return nil
}

func (cs *CertSet) doGetByKey(pub x509utils.PublicKey) []*x509.Certificate {
	var out []*x509.Certificate

	cs.ForEach(func(cert *x509.Certificate) bool {
		if pub.Equal(cert.PublicKey) {
			out = append(out, cert)
		}

		return true
	})

	return out
}

// Copy copies all certificates satisfying the optional condition to the destination
// unless they are already there.
// If a destination isn't provided one will be created.
func (cs *CertSet) Copy(dst *CertSet, cond func(*x509.Certificate) bool) *CertSet {
	if cs == nil {
		if dst == nil {
			dst = MustCertSet()
		}
		return dst
	}

	if dst == nil {
		dst = new(CertSet)
	}
	cs.Set.Copy(&dst.Set, cond)
	return dst
}

// Clone makes a copy of the [CertSet].
func (cs *CertSet) Clone() *CertSet {
	if cs == nil {
		return nil
	}

	dst := new(CertSet)
	cs.Set.Copy(&dst.Set, nil)
	return dst
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
		core.Panic(core.Wrap(err, "failed to create CertSet"))
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
