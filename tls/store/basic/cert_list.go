package basic

import (
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/x/tls/x509utils/certpool"
)

// CertList is a double-linked list of [tls.Certificate]s.
type CertList certpool.List[*tls.Certificate]

// NewCertList creates a List optionally taking its init content as argument.
func NewCertList(certs ...*tls.Certificate) *CertList {
	l := new(CertList)
	for _, cert := range certs {
		if cert != nil {
			l.PushBack(cert)
		}
	}
	return l
}

// Sys returns the underlying [certpool.List].
func (l *CertList) Sys() *certpool.List[*tls.Certificate] {
	if l == nil {
		return nil
	}

	return (*certpool.List[*tls.Certificate])(l)
}

// PushBack appends a certificate at the end of the list.
// nil values are ignored.
func (l *CertList) PushBack(cert *tls.Certificate) {
	if cert != nil {
		l.Sys().PushBack(cert)
	}
}

// PushFront inserts a certificate at the beginning of the list.
// nil values are ignored.
func (l *CertList) PushFront(cert *tls.Certificate) {
	if cert != nil {
		l.Sys().PushFront(cert)
	}
}

// Contains tells if the [CertList] contains a [tls.Certificate]
// matching the given [x509.Certificate].
func (l *CertList) Contains(leaf *x509.Certificate) bool {
	_, found := l.Get(leaf)
	return found
}

// Get returns the [tls.Certificate] matching the given [x509.Certificate].
func (l *CertList) Get(leaf *x509.Certificate) (*tls.Certificate, bool) {
	var out *tls.Certificate

	if l == nil || leaf == nil {
		// no-op
		return nil, false
	}

	l.ForEach(func(c *tls.Certificate) bool {
		if leaf.Equal(c.Leaf) {
			out = c
		}
		return out == nil // continue until one is found
	})

	return out, out != nil
}

// ForEach calls a function for each certificate in the list until
// it returns false.
func (l *CertList) ForEach(fn func(*tls.Certificate) bool) {
	if l == nil || fn == nil {
		return
	}

	l.Sys().ForEach(fn)
}

// Len returns the number of entries in the [CertList].
func (l *CertList) Len() int {
	if l == nil {
		return 0
	}
	return l.Sys().Len()
}

// Export returns all certificates in the [CertList].
func (l *CertList) Export() []*tls.Certificate {
	if l == nil {
		return nil
	}
	return l.Sys().Values()
}
