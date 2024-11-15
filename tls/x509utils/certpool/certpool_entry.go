package certpool

import (
	"crypto/x509"

	"darvaza.org/core"
)

type certPoolEntry struct {
	cert *x509.Certificate

	hash     Hash
	names    []string
	patterns []string
}

func (ce *certPoolEntry) Clone() *certPoolEntry {
	if !ce.Valid() {
		return nil
	}

	return &certPoolEntry{
		cert:     ce.cert,
		hash:     ce.hash,
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

func (ce *certPoolEntry) Hash() (Hash, bool) {
	if ce.Valid() {
		return ce.hash, true
	}

	return Hash{}, false
}

func (ce *certPoolEntry) IsCA() bool {
	if ce.Valid() {
		return ce.cert.IsCA
	}
	return false
}

//revive:disable:flag-parameter
func newCertPoolEntryCondFn(caOnly bool) func(*certPoolEntry) bool {
	//revive:enable:flag-parameter

	if caOnly {
		return certPoolEntryIsCA
	}

	return certPoolEntryValid
}

func newCertPoolEntryCopyFn(cond func(*certPoolEntry) bool) func(*certPoolEntry) (*certPoolEntry, bool) {
	if cond == nil {
		cond = certPoolEntryValid
	}

	return func(ce *certPoolEntry) (ce2 *certPoolEntry, ok bool) {
		if cond(ce) {
			ce2 = ce.Clone()
		}
		return ce2, ce2 != nil
	}
}
func certPoolEntryIsCA(ce *certPoolEntry) bool {
	return ce.IsCA()
}

func certPoolEntryValid(ce *certPoolEntry) bool {
	return ce.Valid()
}
