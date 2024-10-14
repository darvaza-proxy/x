package certpool

import (
	"context"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// Get attempts to find the certificate for a name.
func (s *CertPool) Get(ctx context.Context, name string) (*x509.Certificate, error) {
	if s == nil {
		// invalid
		return nil, ErrNilReceiver
	}

	sn, ok := x509utils.SanitizeName(name)
	if !ok {
		// bad name
		return nil, core.Wrap(core.ErrInvalid, name)
	}

	if err := ctx.Err(); err != nil {
		// cancelled
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if ce := s.getFirst(sn); ce != nil {
		return ce.cert, nil
	}

	return nil, core.Wrap(core.ErrNotExists, name)
}

// ForEach calls a function for each certificate in the store until the context is cancelled
// or the function returns false.
func (s *CertPool) ForEach(ctx context.Context, fn func(context.Context, *x509.Certificate) bool) {
	if s == nil || fn == nil || ctx.Err() != nil {
		return // NO-OP
	}

	// assemble iterable
	s.mu.RLock()
	certs := s.unsafeExportCerts()
	s.mu.RUnlock()

	sliceForEach(ctx, fn, certs)
}

func (s *CertPool) unsafeExportCerts() []*x509.Certificate {
	out := make([]*x509.Certificate, 0, len(s.hashed))
	for _, ce := range s.hashed {
		if ce.Valid() {
			out = append(out, ce.cert)
		}
	}
	return out
}
