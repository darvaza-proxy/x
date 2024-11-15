package certpool

import (
	"darvaza.org/x/tls/x509utils"
)

func (s *CertPool) getListForName(name string) *List[*certPoolEntry] {
	if l, ok := s.names[name]; ok {
		return l
	}

	if ip, ok := x509utils.NameAsIP(name); ok {
		if l, ok := s.names[ip]; ok {
			return l
		}
	}
	return nil
}

func (s *CertPool) getListForSuffix(name string) *List[*certPoolEntry] {
	if name[0] == '[' {
		// skip IP Addresses
		return nil
	}

	if suffix, ok := x509utils.NameAsSuffix(name); ok {
		if l, ok := s.patterns[suffix]; ok {
			return l
		}
	}

	return nil
}

func (s *CertPool) getFirst(name string) *certPoolEntry {
	// exact
	if l := s.getListForName(name); l != nil {
		return l.Front()
	}

	// wildcard
	if l := s.getListForSuffix(name); l != nil {
		return l.Front()
	}

	return nil
}
