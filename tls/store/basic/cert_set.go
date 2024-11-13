package basic

import (
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/x/tls/x509utils/certpool"
)

// CertSet keeps a unique set of [tls.Certificate]s.
type CertSet map[certpool.Hash]*CertList

// NewCertSet creates a [CertSet] optionally taking its initial content as argument.
func NewCertSet(certs ...*tls.Certificate) CertSet {
	set := make(CertSet)
	for _, cert := range certs {
		set.Push(cert)
	}

	return set
}

// Sys returns the underlying map.
func (set CertSet) Sys() map[certpool.Hash]*CertList {
	if set == nil {
		return nil
	}

	return (map[certpool.Hash]*CertList)(set)
}

// Hash tells if the certificate are usable, and returns the
// hash for the certificate.
func (CertSet) Hash(cert *tls.Certificate) (certpool.Hash, bool) {
	if cert == nil {
		return certpool.Hash{}, false
	}
	return certpool.HashCert(cert.Leaf)
}

// HashLeaf tells if the certificate are usable, and returns the
// hash for the certificate.
func (CertSet) HashLeaf(leaf *x509.Certificate) (certpool.Hash, bool) {
	return certpool.HashCert(leaf)
}

// Push attempts to add a [tls.Certificate] to the [CertSet], and
// returns true if succeeded, but also the unique certificate if found.
func (set CertSet) Push(cert *tls.Certificate) (*tls.Certificate, bool) {
	if set == nil || cert == nil {
		return nil, false
	}

	hash, ok := set.Hash(cert)
	if !ok {
		return nil, false
	}

	l, ok := set[hash]
	if !ok {
		set[hash] = NewCertList(cert)
		return cert, true
	}

	if c, _ := l.Get(cert.Leaf); c != nil {
		return c, false
	}

	l.PushFront(cert)
	return cert, true
}

// Pop removes from the [CertSet] the [tls.Certificate] matching the given
// [x509.Certificate].
func (set CertSet) Pop(leaf *x509.Certificate) (*tls.Certificate, bool) {
	if set == nil || leaf == nil {
		// nothing
		return nil, false
	}

	hash, ok := set.HashLeaf(leaf)
	if !ok {
		// bad
		return nil, false
	}

	l, ok := set[hash]
	if !ok {
		// not found
		return nil, false
	}

	return l.Sys().PopFirstMatchFn(func(c *tls.Certificate) bool {
		return leaf.Equal(c.Leaf)
	})
}

// Get attempts to find in the [CertSet] the [tls.Certificate] for the given leaf.
func (set CertSet) Get(cert *x509.Certificate) (*tls.Certificate, bool) {
	if set == nil || cert == nil {
		return nil, false
	}

	hash, ok := set.HashLeaf(cert)
	if !ok {
		return nil, false
	}

	l, ok := set[hash]
	if !ok {
		return nil, false
	}

	return l.Get(cert)
}

// ForEach calls a function for each certificate in the set until
// it returns false.
func (set CertSet) ForEach(fn func(*tls.Certificate) bool) {
	if set == nil || fn == nil {
		return
	}

	cont := true
	for _, l := range set {
		l.ForEach(func(c *tls.Certificate) bool {
			cont = fn(c)
			return cont
		})

		if !cont {
			break
		}
	}
}

// Len returns an estimate of the number of certificates in the [CertSet].
//
// Len returns the number of buckets in the hash table, which could be empty,
// or have more than one certificate if there is conflicting [certpool.Hash].
// Only returning 0 is certain, 1+ is an educated guess.
func (set CertSet) Len() int {
	if set == nil {
		return 0
	}
	return len(set)
}

// Export returns all certificates in the [CertSet].
func (set CertSet) Export() []*tls.Certificate {
	if set == nil {
		return nil
	}

	vv := make([][]*tls.Certificate, 0, len(set))
	for _, l := range set {
		vv = append(vv, l.Export())
	}

	return join(vv)
}
