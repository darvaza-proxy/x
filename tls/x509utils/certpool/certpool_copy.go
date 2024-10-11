package certpool

import (
	"crypto/x509"

	"darvaza.org/x/tls/x509utils"
)

// Export assembles a [x509.CertPool] with all the certificates
// contained in the store.
func (s *CertPool) Export() *x509.CertPool {
	out := x509.NewCertPool()
	if s != nil {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for _, ce := range s.hashed {
			out.AddCert(ce.cert)
		}
	}
	return out
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
	case out == nil:
		return s.doClone(caOnly)
	case out == s:
		// avoid copying to itself
		return s
	default:
		return s.doCopy(out, caOnly)
	}
}

//revive:disable:flag-parameter
func (s *CertPool) doCopy(out *CertPool, caOnly bool) *CertPool {
	//revive:enable:flag-parameter
	out.mu.Lock()
	defer out.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for hash, ce := range s.hashed {
		if !caOnly || ce.IsCA() {
			if _, found := out.hashed[hash]; !found {
				// new
				out.unsafeAddCertEntry(ce)
			}
		}
	}

	return out
}

// Clone creates a copy of the [CertPool] store.
func (s *CertPool) Clone() x509utils.CertPool {
	if s == nil {
		return nil
	}

	return s.doClone(false)
}

func (s *CertPool) doClone(caOnly bool) *CertPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fn := func(ce *certPoolEntry) (*certPoolEntry, bool) {
		return ce, !caOnly || ce.IsCA()
	}

	return &CertPool{
		hashed:   copyMap(s.hashed, fn),
		names:    copyMapList(s.names, fn),
		patterns: copyMapList(s.patterns, fn),
	}
}
