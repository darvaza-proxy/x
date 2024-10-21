package lru

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

// Export returns all valid certificates in the LRU cache.
func (s *LRU) Export() ([]*tls.Certificate, error) {
	return s.ExportWithContext(context.Background())
}

// ExportMatch returns all valid certificates matching the given server name.
func (s *LRU) ExportMatch(serverName string) ([]*tls.Certificate, error) {
	return s.ExportMatchWithContext(context.Background(), serverName)
}

// ExportWithContext returns all valid certificates in the LRU cache
// allowing cancellations via the context.
func (s *LRU) ExportWithContext(ctx context.Context) ([]*tls.Certificate, error) {
	if err := s.checkInitWithContext(ctx); err != nil {
		return nil, err
	}

	return s.doExportWithContext(ctx)
}

// ExportMatchWithContext returns all valid certificates matching the given server name,
// allowing cancellations via the context.
func (s *LRU) ExportMatchWithContext(ctx context.Context, serverName string) ([]*tls.Certificate, error) {
	_, err := s.checkExportMatchWithContext(ctx, serverName)
	if err != nil {
		return nil, err
	}

	return s.doExportMatchWithContext(ctx, serverName)
}

func (s *LRU) doExportWithContext(ctx context.Context) ([]*tls.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]*tls.Certificate, 0, s.count)
	for _, e := range s.entries {
		if e.Valid() {
			out = append(out, e.cert)
		} else {
			s.evict(e)
		}

		if ctx.Err() != nil {
			break
		}
	}

	if len(out) == 0 {
		return nil, core.CoalesceError(ctx.Err(), ErrNotFound)
	}

	return out, ctx.Err()
}

func (s *LRU) doExportMatchWithContext(ctx context.Context, name string) ([]*tls.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// exact match only
	if l, ok := s.names[name]; ok {
		out := make([]*tls.Certificate, 0, l.Len())

		l.ForEach(func(e *lruEntry) bool {
			if e.Valid() {
				out = append(out, e.cert)
			} else {
				s.evict(e)
			}
			return ctx.Err() == nil // continue until cancelled
		})

		if len(out) > 0 {
			return out, ctx.Err()
		}
	}

	return nil, core.CoalesceError(ctx.Err(), ErrNotFound)
}

func (s *LRU) checkExportMatchWithContext(ctx context.Context, serverName string) (string, error) {
	if err := s.checkInitWithContext(ctx); err != nil {
		return "", err
	}
	return sanitizeName(serverName)
}
