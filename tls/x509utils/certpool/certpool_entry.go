package certpool

import "crypto/x509"

type certPoolEntry struct {
	cert *x509.Certificate

	hash     Hash
	names    []string
	patterns []string
}

func (ce *certPoolEntry) Equal(other *certPoolEntry) bool {
	switch {
	case ce == other:
		return true
	case ce == nil, other == nil:
		return false
	default:
		return ce.hash.Equal(other.hash)
	}
}

func (ce *certPoolEntry) Hash() (Hash, bool) {
	var zero Hash
	switch {
	case ce == nil, ce.cert == nil:
		return zero, false
	default:
		return ce.hash, true
	}
}

func (ce *certPoolEntry) IsCA() bool {
	switch {
	case ce == nil, ce.cert == nil:
		return false
	default:
		return ce.cert.IsCA
	}
}
