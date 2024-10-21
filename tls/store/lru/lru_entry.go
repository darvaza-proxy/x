package lru

import (
	"crypto/x509"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

type lruEntry struct {
	hash certpool.Hash
	size int

	cert     *tls.Certificate
	names    []string
	suffixes []string
}

// Exists tells if the entry is initialized.
func (e *lruEntry) Exists() bool {
	return e != nil && e.cert != nil
}

// Export returns the certificate, its hash, size, and a boolean indicating whether the entry exists.
// If the entry is not initialized, it returns nil, an empty hash, zero size, and false.
func (e *lruEntry) Export() (*tls.Certificate, certpool.Hash, int, bool) {
	if e.Exists() {
		return e.cert, e.hash, e.size, true
	}
	return nil, certpool.Hash{}, 0, false
}

// Valid tells if the entry can remain in the cache.
func (e *lruEntry) Valid() bool {
	if cert, _, _, ok := e.Export(); ok {
		return !cert.Leaf.NotAfter.Before(time.Now())
	}
	return false
}

// Name returns the first name associated with the certificate entry, or an empty string if no names exist.
func (e *lruEntry) Name() string {
	var s string
	if e.Exists() && len(e.names) > 0 {
		s = e.names[0]
	}
	return s
}

func (e *lruEntry) Names() ([]string, []string, bool) {
	if e.Exists() {
		return e.names, e.suffixes, true
	}
	return nil, nil, false
}

func (e *lruEntry) Hash() certpool.Hash {
	if e.Exists() {
		return e.hash
	}
	return certpool.Hash{}
}

func (e *lruEntry) Size() int {
	if e.Exists() {
		return e.size
	}
	return 0
}

// newEntry creates a new [lruEntry] from a [tls.Certificate] optionally
// verifying it against a root certificate pool.
func newEntry(cert *tls.Certificate, roots *x509.CertPool) (*lruEntry, error) {
	if err := tls.Verify(cert, roots); err != nil {
		return nil, err
	}

	hash, _ := certpool.HashCert(cert.Leaf)
	names, suffixes := x509utils.Names(cert.Leaf)
	size := getSize(cert)

	e := &lruEntry{
		hash:     hash,
		size:     size,
		cert:     cert,
		names:    names,
		suffixes: suffixes,
	}

	return e, nil
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
