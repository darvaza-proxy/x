package cache

import (
	"crypto/tls"
	"time"

	"darvaza.org/cache"
)

// Sink represents a [cache.Sink] that handles [tls.Certificate] values
type Sink = cache.TSink[tls.Certificate]

// NewGobSink creates a new Sink using Gob encoding.
func NewGobSink() (cache.TSink[tls.Certificate], error) {
	return new(cache.GobSink[tls.Certificate]), nil
}

// DecodeCached uses decodes an encoded certificate using the
// encoding linked to the [Store]'s cache sink type.
func (s *Store) DecodeCached(b []byte) (*tls.Certificate, bool) {
	if s != nil && s.cfg.NewSink != nil {
		sink, err := s.cfg.NewSink()
		if err == nil {
			err = sink.SetBytes(b, time.Time{})
		}
		if err != nil {
			return nil, false
		}
		return sink.Value()
	}
	return nil, false
}
