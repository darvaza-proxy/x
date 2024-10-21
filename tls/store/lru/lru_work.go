package lru

import (
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

// Work Queue
func (s *LRU) emit(fn func()) {
	if fn != nil {
		defer func() {
			// ignore closed channel
			_ = recover()
		}()

		// block, don't drop.
		s.wq <- fn
	}
}

func (s *LRU) run(maxWorkers int) {
	// barrier
	workers := make(chan struct{}, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers <- struct{}{}
	}

	for fn := range s.wq {
		if fn != nil {
			<-workers
			go func(fn func()) {
				defer func() {
					workers <- struct{}{}
				}()

				s.tryRun(fn)
			}(fn)
		}
	}
}

func (s *LRU) tryRun(fn func()) {
	if fn != nil {
		defer func() {
			if err := core.AsRecovered(recover()); err != nil {
				s.logError(err, "panic in work queue")
			}
		}()

		fn()
	}
}

// LRU events
func (s *LRU) notifyEvicted(e *lruEntry) {
	if s == nil || e == nil {
		return
	}

	if fn := s.callOnEvict; fn != nil {
		name, cert, size := e.Name(), e.cert, e.size

		s.emit(func() { fn(name, cert, size) })
	}
}

func (s *LRU) onAdd(key string, cert *tls.Certificate, size int, expires time.Time) {
	if fn := s.callOnAdd; fn != nil {
		s.emit(func() { fn(key, cert, size, expires) })
	}
}
