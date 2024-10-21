package lru

import (
	"container/list"
	"crypto/x509"
	"sync"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

const (
	MinimumCacheSize = 8
	DefaultCacheSize = 32
)

type OnAddFunc func(string, *tls.Certificate, int, time.Time)
type OnEvictFunc func(string, *tls.Certificate, int)

var _ tls.Store = (*LRU)(nil)

type LRU struct {
	mu sync.Mutex

	lru     *list.List
	maxSize int

	upstream    tls.Store
	callOnAdd   OnAddFunc
	callOnEvict OnEvictFunc
}

func New(upstream tls.Store, size int, onAdd OnAddFunc, onEvict OnEvictFunc) (*LRU, error) {
	if size < MinimumCacheSize {
		size = DefaultCacheSize
	}

	s := &LRU{
		lru:     new(list.List),
		maxSize: size,

		upstream:    upstream,
		callOnAdd:   onAdd,
		callOnEvict: onEvict,
	}

	if err := s.unsafeInit(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *LRU) unsafeInit() error {
	if s.upstream == nil {
		return ErrNoUpstream
	}

	if s.maxSize == 0 {
		s.maxSize = DefaultCacheSize
	}

	if s.lru == nil {
		s.lru = new(list.List)
	}

	return nil
}

func (s *LRU) init() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.unsafeInit()
}

func (s *LRU) GetCAPool() *x509.CertPool {
	if s != nil && s.upstream != nil {
		return s.upstream.GetCAPool()
	}

	return nil
}

func (s *LRU) onAdd(string, *tls.Certificate, int, time.Time)
func (s *LRU) onEvict(string, *tls.Certificate, int)

func (s *LRU) ForEach(fn func(string, *tls.Certificate, int, time.Time) bool) error {
	if err := s.init(); err != nil {
		return err
	} else if fn == nil {
		return nil
	}

	return nil
}

func (s *LRU) exportEntries() []*Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
}
