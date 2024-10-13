// Package certpool provides an X.509 certificates store
package certpool

import (
	"crypto/x509"
	"sync"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ x509utils.CertPool = (*CertPool)(nil)
)

// CertPool is a collection of x.509 Certificates.
type CertPool struct {
	mu sync.RWMutex

	cache    *x509.CertPool
	hashed   map[Hash]*certPoolEntry
	names    map[string]*List[*certPoolEntry]
	patterns map[string]*List[*certPoolEntry]
}

// IsZero tells if the non-nil store is empty.
func (s *CertPool) IsZero() bool {
	if s == nil {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.hashed) == 0
}

// Count returns the number of certificates in the store.
func (s *CertPool) Count() int {
	var count int

	if s != nil {
		s.mu.RLock()
		defer s.mu.RUnlock()

		count = len(s.hashed)
	}

	return count
}

// IsCA tells if all the certificates in the store are CA.
func (s *CertPool) IsCA() bool {
	if s == nil {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ce := range s.hashed {
		if !ce.cert.IsCA {
			return false
		}
	}
	return true
}

func (s *CertPool) init() {
	if s != nil && s.hashed == nil {
		s.unsafeReset()
	}
}

// Reset removes all certificates from the store.
func (s *CertPool) Reset() {
	if s != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.unsafeReset()
	}
}

func (s *CertPool) unsafeReset() {
	s.cache = nil
	s.hashed = make(map[Hash]*certPoolEntry)
	s.names = make(map[string]*List[*certPoolEntry])
	s.patterns = make(map[string]*List[*certPoolEntry])
}

// New creates a blank [CertPool] store.
func New() *CertPool {
	return &CertPool{
		hashed:   make(map[Hash]*certPoolEntry),
		names:    make(map[string]*List[*certPoolEntry]),
		patterns: make(map[string]*List[*certPoolEntry]),
	}
}
