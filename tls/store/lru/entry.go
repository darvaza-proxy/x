package lru

import (
	"container/list"
	"crypto/x509"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

type Entry struct {
	lruList *list.List
	lruElem *list.Element

	hash     certpool.Hash
	size     int
	cert     *tls.Certificate
	names    []string
	suffixes []string
}

func (e *Entry) Export() (*tls.Certificate, bool) {
	if e != nil {
		return e.cert, e.cert != nil
	}
	return nil, false
}

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

	key, _ := cert.PrivateKey.(x509utils.PrivateKey)
	pk, _ := x509utils.EncodePKCS8PrivateKey(key)
	size += len(pk)

	return size
}
