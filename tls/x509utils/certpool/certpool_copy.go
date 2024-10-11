package certpool

import (
	"crypto/x509"

	"darvaza.org/x/tls/x509utils"
)

// Export assembles a [x509.CertPool] with all the certificates
// contained in the store.
func (s *CertPool) Export() *x509.CertPool {
	if s == nil {
		return x509.NewCertPool()
	}

	// RO
	s.mu.RLock()
	out := s.cache
	s.mu.RUnlock()

	if out != nil {
		// cached
		return out
	}

	// RW
	s.mu.Lock()
	defer s.mu.Unlock()

	out = s.cache
	if out == nil {
		// assemble and store
		out = s.unsafeExport()
		s.cache = out
	}
	return out
}

func (s *CertPool) unsafeExport() *x509.CertPool {
	out := x509.NewCertPool()
	for _, ce := range s.hashed {
		out.AddCert(ce.cert)
	}
	return out
}

func (s *CertPool) unsafeInvalidateCache() {
	s.cache = nil
}

// Copy creates a copy of the [CertPool] store, optionally
// receiving the destination.
func (s *CertPool) Copy(out *CertPool, caOnly bool) *CertPool {
	switch {
	case s == nil:
		if out == nil {
			out = New()
		}
		return out
	case out == s:
		// avoid copying to itself
		return s
	default:
		cond := newCertPoolEntryCondFn(caOnly)

		if out == nil {
			return s.doClone(cond)
		}

		return s.doCopy(out, cond)
	}
}

func (s *CertPool) doCopy(out *CertPool, cond func(*certPoolEntry) bool) *CertPool {
	// extend condition to exclude those
	// already present
	cond2 := func(ce *certPoolEntry) bool {
		if !cond(ce) {
			// not wanted
			return false
		}
		_, found := out.hashed[ce.hash]
		return !found
	}

	// clone creates a copy of acceptable entries
	clone := newCertPoolEntryCopyFn(cond2)

	out.mu.Lock()
	defer out.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ce := range s.hashed {
		if ce2, ok := clone(ce); ok {
			out.unsafeAddCertEntry(ce2)
		}
	}

	return out
}

// Clone creates a copy of the [CertPool] store.
func (s *CertPool) Clone() x509utils.CertPool {
	if s == nil {
		return nil
	}

	return s.doClone(certPoolEntryValid)
}

func (s *CertPool) doClone(cond func(*certPoolEntry) bool) *CertPool {
	fn := newCertPoolEntryCopyFn(cond)

	s.mu.RLock()
	defer s.mu.RUnlock()

	return &CertPool{
		hashed:   copyMap(s.hashed, fn),
		names:    copyMapList(s.names, fn),
		patterns: copyMapList(s.patterns, fn),
	}
}
