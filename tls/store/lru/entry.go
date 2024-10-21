package lru

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// Entry represents a cached TLS certificate with associated metadata
// used in a least-recently-used (LRU) certificate store.
type Entry struct {
	// lruList *list.List
	// lruElem *list.Element

	hash     certpool.Hash
	size     int
	cert     *tls.Certificate
	names    []string
	suffixes []string
}

// Export returns the underlying TLS certificate and a boolean indicating whether the certificate is non-nil.
// If the Entry is nil, it returns nil and false.
func (e *Entry) Export() (*tls.Certificate, bool) {
	if e != nil {
		return e.cert, e.cert != nil
	}
	return nil, false
}

// NewEntry creates a new Entry from a TLS certificate after verifying it against the provided root certificate pool.
// It computes the certificate's hash, extracts its names and suffixes, calculates its size, and returns a new Entry.
// Returns an error if the certificate verification fails.
func NewEntry(cert *tls.Certificate, roots *x509.CertPool) (*Entry, error) {
	if err := tls.Verify(cert, roots); err != nil {
		return nil, err
	}

	hash, _ := certpool.HashCert(cert.Leaf)
	names, suffixes := x509utils.Names(cert.Leaf)
	size := getSize(cert)

	return &Entry{
		hash:     hash,
		size:     size,
		cert:     cert,
		names:    names,
		suffixes: suffixes,
	}, nil
}

func getSize(cert *tls.Certificate) int {
	if cert == nil {
		return 0
	}

	size := len(cert.Leaf.Raw)
	for _, p := range cert.Certificate {
		size += len(p)
	}

	key, ok := cert.PrivateKey.(x509utils.PrivateKey)
	if !ok {
		core.Panic(core.ErrUnreachable)
	}

	pk, _ := x509utils.EncodePKCS8PrivateKey(key)
	size += len(pk)

	return size
}
