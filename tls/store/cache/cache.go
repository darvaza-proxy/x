package cache

import (
	"context"
	"crypto/tls"

	"darvaza.org/cache"
	"darvaza.org/cache/x/memcache"
	"darvaza.org/core"
)

type gobCertSink = cache.GobSink[*tls.Certificate]

func (s *Store) doGetCertificate(ctx context.Context, serverName string) (*tls.Certificate, error) {
	out := new(gobCertSink)
	if err := s.sf.Get(ctx, serverName, out); err != nil {
		return nil, err
	}
	return out.Value(), nil
}

func (s *Store) doGet(ctx context.Context, _ string, dest cache.Sink) error {
	chi, ok := ClientHelloInfo(ctx)
	if !ok {
		return core.Wrap(core.ErrInvalid, "chi not provided")
	}

	cert, err := s.cfg.Store.GetCertificate(chi)
	if err != nil {
		return err
	}

	if sink, ok := dest.(*gobCertSink); ok {
		return sink.SetValue(cert, cert.Leaf.NotAfter)
	}

	return core.Wrap(core.ErrInvalid, "invalid sink")
}

func initCache(s *Store) error {
	getter := cache.GetterFunc(s.doGet)

	s.lru = memcache.NewLRU(int64(s.cfg.CacheSize), s.cfg.OnAdd, s.cfg.OnEvict)
	s.sf = memcache.NewSingleFlight("tlsStore", s.lru, getter)
	return nil
}
