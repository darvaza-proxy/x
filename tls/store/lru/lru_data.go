package lru

import (
	"darvaza.org/x/container/list"
)

// Len returns the number of certificates in the cache.
// It is thread-safe.
func (s *LRU) Len() int {
	if err := s.tryLock(); err != nil {
		return 0
	}
	defer s.unlock()

	return s.count
}

// Size returns the added size of all certificates in the cache
// It is thread-safe.
func (s *LRU) Size() int {
	if err := s.tryLock(); err != nil {
		return 0
	}
	defer s.unlock()

	return s.size
}

// Available tells how much space is available without evictions.
// It is thread-safe.
func (s *LRU) Available() int {
	if err := s.tryLock(); err != nil {
		return 0
	}
	defer s.unlock()

	return s.maxSize - s.size
}

// PruneExpired removes all expired entries from the cache.
// It is thread-safe.
func (s *LRU) PruneExpired() int {
	if err := s.tryLock(); err != nil {
		return 0
	}
	defer s.unlock()

	// all expired
	return s.pruneFn(s.evictExpired, nil)
}

// Prune removes expired and least recently used entries from the cache
// until size is within acceptable limits.
// It is thread-safe.
func (s *LRU) Prune() int {
	if err := s.tryLock(); err != nil {
		return 0
	}
	defer s.unlock()

	return s.doPrune()
}

// doPrune performs internal pruning of the LRU cache by first removing expired entries
// and then removing least recently used entries until the cache size is within
// the maximum allowed size. It is not thread-safe and should only be called
// from methods that have already acquired a lock.
func (s *LRU) doPrune() int {
	isDone := func() bool {
		return s.size <= s.maxSize
	}

	return s.pruneFn(s.evictExpired, isDone) + s.pruneFn(s.evict, isDone)
}

// pruneFn iterates through the eviction queue and removes entries based on the provided eviction function.
// It stops iterating when the optional isDone function returns true or all entries have been processed.
// Returns the number of entries successfully evicted. It is not thread-safe.
func (s *LRU) pruneFn(evictEntry func(*lruEntry) bool, isDone func() bool) int {
	var evicted int

	if isDone == nil || !isDone() {
		s.eviction.ForEach(func(e *lruEntry) bool {
			if evictEntry(e) {
				evicted++
			}
			return isDone != nil && isDone()
		})
	}

	return evicted
}

// evict removes an entry from the cache unconditionally.
// It is not thread-safe.
func (s *LRU) evict(e *lruEntry) bool {
	return s.evictFn(e, nil)
}

// evictExpired removes an entry from the cache if it is no longer valid.
// It is not thread-safe.
func (s *LRU) evictExpired(e *lruEntry) bool {
	cond := func(e *lruEntry) bool {
		return !e.Valid()
	}

	return s.evictFn(e, cond)
}

// evictFn attempts to remove an entry from the cache based on optional conditions.
// It returns true if the entry was successfully evicted, false otherwise.
// It is not thread-safe and should only be called from methods that have already acquired a lock.
func (s *LRU) evictFn(e *lruEntry, cond func(*lruEntry) bool) bool {
	switch {
	case e == nil, cond != nil && !cond(e):
		return false
	case !s.unlink(e):
		return false
	default:
		s.notifyEvicted(e)
		return true
	}
}

// link adds an entry to the LRU cache, tracking its hash, size, names, and suffixes.
// It returns true if the entry was successfully linked, false if the entry is invalid
// or already exists in the cache. It is not thread-safe.
func (s *LRU) link(e *lruEntry) bool {
	_, hash, size, ok := e.Export()
	if !ok || s == nil {
		return false // invalid
	} else if _, found := s.entries[hash]; found {
		return false // already linked
	}

	names, suffixes, _ := e.Names()
	s.size += size
	s.count++

	s.entries[hash] = e
	s.linkNames(s.names, names, e)
	s.linkNames(s.suffixes, suffixes, e)
	s.eviction.PushBack(e) // end of the queue

	return true
}

// linkNames adds an entry to the front of the specified map of lists
// It is not thread-safe and is an internal method for managing name-based references in the LRU cache.
func (s *LRU) linkNames(m map[string]*list.List[*lruEntry], keys []string, e *lruEntry) {
	for _, key := range keys {
		l, ok := m[key]
		if !ok {
			m[key] = list.New[*lruEntry](e)
		} else {
			l.PushFront(e)
		}
	}
}

// unlink removes an entry from the LRU cache, decrementing its size and count.
// It returns true if the entry was successfully unlinked, false if the entry is invalid
// or not currently in the cache. It is not thread-safe and should only be called
// from methods that have already acquired a lock.
func (s *LRU) unlink(e *lruEntry) bool {
	_, hash, size, ok := e.Export()
	if !ok || s == nil {
		return false // invalid
	} else if _, found := s.entries[hash]; !found {
		return false // not linked
	}

	names, suffixes, _ := e.Names()
	s.size -= size
	s.count--

	eq := func(ep *lruEntry) bool {
		return e == ep
	}

	s.unlinkNamesFn(s.names, names, eq)
	s.unlinkNamesFn(s.suffixes, suffixes, eq)
	s.eviction.DeleteMatchFn(eq)
	delete(s.entries, hash)
	return true
}

// unlinkNamesFn removes entries from name-based lists that match the given equality function.
// It is an internal method used for managing name references during cache entry removal.
// It is not thread-safe and should only be called from methods that have already acquired a lock.
func (s *LRU) unlinkNamesFn(m map[string]*list.List[*lruEntry], keys []string, eq func(*lruEntry) bool) {
	if eq == nil || len(keys) == 0 {
		return
	}

	for _, key := range keys {
		if l, ok := m[key]; ok {
			l.DeleteMatchFn(eq)
		}
	}
}

// postponeEviction moves the specified entry to the back of the eviction list,
// effectively delaying its potential removal from the cache.
// It is not thread-safe.
func (s *LRU) postponeEviction(e *lruEntry) {
	s.eviction.MoveToBackFirstMatchFn(func(ep *lruEntry) bool {
		return e == ep
	})
}
