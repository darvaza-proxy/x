package certpool

import (
	"crypto/x509"

	"darvaza.org/core"
)

type certPoolEntry struct {
	cert *x509.Certificate

	names    []string
	patterns []string
}

func (ce *certPoolEntry) Clone() *certPoolEntry {
	if !ce.Valid() {
		return nil
	}

	return &certPoolEntry{
		cert:     ce.cert,
		names:    core.SliceCopy(ce.names),
		patterns: core.SliceCopy(ce.patterns),
	}
}

func (ce *certPoolEntry) Equal(other *certPoolEntry) bool {
	switch {
	case ce == other:
		return true
	case ce == nil, other == nil:
		return false
	default:
		return ce.cert.Equal(other.cert)
	}
}

func (ce *certPoolEntry) Valid() bool {
	return ce != nil && ce.cert != nil
}

func (ce *certPoolEntry) IsCA() bool {
	if ce.Valid() {
		return ce.cert.IsCA
	}
	return false
}

func newCertPoolEntryCondFn(fn func(*x509.Certificate) bool) func(*certPoolEntry) bool {
	return func(ce *certPoolEntry) bool {
		if !ce.Valid() {
			return false
		}

		return fn == nil || fn(ce.cert)
	}
}

func newCertPoolEntryCopyFn(cond func(*certPoolEntry) bool) func(*certPoolEntry) (*certPoolEntry, bool) {
	if cond == nil {
		cond = func(ce *certPoolEntry) bool {
			return ce.Valid()
		}
	}

	return func(ce *certPoolEntry) (ce2 *certPoolEntry, ok bool) {
		if cond(ce) {
			ce2 = ce.Clone()
		}
		return ce2, ce2 != nil
	}
}
