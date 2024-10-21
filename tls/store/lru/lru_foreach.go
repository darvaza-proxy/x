package lru

import (
	"context"

	"darvaza.org/x/tls"
)

// ForEach iterates over all valid certificates in the LRU cache, calling the provided function
// for each certificate. The iteration stops if the context is canceled or if the
// callback function returns false.
func (s *LRU) ForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool) {
	if !s.checkForEach(ctx, fn) {
		return
	}

	certs, err := s.doExportWithContext(ctx)
	if err != nil {
		return
	}

	for _, cert := range certs {
		if ctx.Err() != nil || !fn(ctx, cert) {
			break
		}
	}
}

// ForEachMatch iterates over all valid certificates in the LRU cache matching the given server name,
// calling the provided function for each matching certificate. The iteration stops if the
// context is canceled or if the callback function returns false.
func (s *LRU) ForEachMatch(ctx context.Context, serverName string, fn func(context.Context, *tls.Certificate) bool) {
	serverName, ok := s.checkForEachMatch(ctx, serverName, fn)
	if !ok {
		return
	}

	certs, err := s.doExportMatchWithContext(ctx, serverName)
	if err != nil {
		return
	}

	for _, cert := range certs {
		if ctx.Err() != nil || !fn(ctx, cert) {
			break
		}
	}
}

func (s *LRU) checkForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool) bool {
	return fn != nil && s.checkInitWithContext(ctx) == nil
}

func (s *LRU) checkForEachMatch(ctx context.Context, serverName string,
	fn func(context.Context, *tls.Certificate) bool) (string, bool) {
	//
	if fn == nil {
		return "", false
	}

	name, err := s.checkExportMatchWithContext(ctx, serverName)
	return name, err == nil
}
