package cache

import (
	"context"
	"crypto/tls"
	"time"

	"darvaza.org/cache"
	"darvaza.org/core"
)

func (s *Store) doGetCertificate(ctx context.Context, serverName string) (*tls.Certificate, error) {
	out, err := s.cfg.NewSink()
	if err != nil {
		return nil, err
	}
	if err := s.sf.Get(ctx, serverName, out); err != nil {
		return nil, err
	}
	cert, _ := out.Value()
	return cert, nil
}

// doGet is called by SingleFlight to get a [tls.Certificate]
func (s *Store) doGet(ctx context.Context, serverName string, dest cache.Sink) error {
	chi, ok := ClientHelloInfo(ctx)
	if !ok {
		return core.Wrap(core.ErrInvalid, "chi not provided")
	}

	chi2 := *chi
	chi2.ServerName = serverName
	cert, err := s.cfg.Store.GetCertificate(&chi2)
	if err != nil {
		return err
	}

	if sink, ok := dest.(Sink); ok {
		return sink.SetValue(cert, cert.Leaf.NotAfter)
	}

	return cache.ErrInvalidSink
}

// Get attempts to get an certificate from the cache by name,
// without triggering an upstream request.
func (s *Store) Get(serverName string) ([]byte, *time.Time, error) {
	switch {
	case s == nil:
		return nil, nil, core.ErrNilReceiver
	case s.lru == nil:
		return nil, nil, core.Wrap(core.ErrInvalid, "store not initialized")
	default:
		b, exp, ok := s.lru.Get(serverName)
		if !ok {
			return nil, nil, core.Wrap(core.ErrNotExists, serverName)
		}
		return b, exp, nil
	}
}
