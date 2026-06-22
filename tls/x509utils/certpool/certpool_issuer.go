package certpool

import (
	"crypto/x509"
)

// GetBySubjectHash returns every certificate held by the pool whose RawSubject
// hashes to h. It is a pure index lookup over the pool's own certificates — no
// [x509.CertPool] export and no signature check. Callers resolving an issuer
// must verify each candidate themselves with
// [x509.Certificate.CheckSignatureFrom]; pair it with [HashIssuer] to find a
// child's candidate issuers natively.
func (s *CertPool) GetBySubjectHash(h Hash) []*x509.Certificate {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	l, ok := s.bySubject[h]
	if !ok || l.Len() == 0 {
		return nil
	}

	out := make([]*x509.Certificate, 0, l.Len())
	l.ForEach(func(ce *certPoolEntry) bool {
		if ce.Valid() {
			out = append(out, ce.cert)
		}
		return true
	})
	return out
}

// GetBySubject returns every certificate held by the pool that shares cert's
// Subject (cert itself included when held). Shortcut for
// GetBySubjectHash(HashSubject(cert)).
func (s *CertPool) GetBySubject(cert *x509.Certificate) []*x509.Certificate {
	if h, ok := HashSubject(cert); ok {
		return s.GetBySubjectHash(h)
	}
	return nil
}

// GetByIssuer returns cert's candidate issuers held by the pool: every
// certificate whose Subject matches cert's Issuer. The match is by name only —
// verifying which candidate actually signed cert (via CheckSignatureFrom) is
// the caller's job. Shortcut for GetBySubjectHash(HashIssuer(cert)).
func (s *CertPool) GetByIssuer(cert *x509.Certificate) []*x509.Certificate {
	if h, ok := HashIssuer(cert); ok {
		return s.GetBySubjectHash(h)
	}
	return nil
}

// unsafeIndexSubject adds ce to the bySubject index under the hash of its
// RawSubject. The caller must hold the write lock.
func (s *CertPool) unsafeIndexSubject(ce *certPoolEntry) {
	if h, ok := HashSubject(ce.cert); ok {
		appendMapList(s.bySubject, h, ce)
	}
}
