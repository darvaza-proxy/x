// Package set implements a generic programmatic Set data container.
package set

import (
	"sync"

	"darvaza.org/core"

	"darvaza.org/x/container/list"
)

var (
	// ErrExist is returned by Set.Push when a matching entry is
	// already present.
	ErrExist = core.ErrExists
	// ErrNotExist is returned by Set.Pop and Set.Get when a matching entry
	// is not present.
	ErrNotExist = core.ErrNotExists
)

// Set implements a simple set using generics.
type Set[K, H comparable, T any] struct {
	cfg     Config[K, H, T]
	mu      sync.RWMutex
	buckets map[H]*list.List[T]
}

func (set *Set[K, H, T]) init(cfg Config[K, H, T], items ...T) error {
	var err error

	if set == nil {
		return core.ErrNilReceiver
	}

	// RW
	set.mu.Lock()
	if !set.unsafeIsReady() {
		// first
		set.cfg = cfg
		set.unsafeReset()
	} else {
		err = core.Wrap(core.ErrInvalid, "Set already initialized")
	}
	set.mu.Unlock()

	if err == nil {
		err = set.unsafeInitPush(items)
	}
	return err
}

func (set *Set[K, H, T]) unsafeInitPush(items []T) error {
	for _, item := range items {
		_, err := set.Push(item)
		if err != nil && err != ErrExist {
			return err
		}
	}

	return nil
}

func (set *Set[K, H, T]) isReady() bool {
	var ready bool

	if set != nil {
		set.mu.RLock()
		ready = set.unsafeIsReady()
		set.mu.RUnlock()
	}

	return ready
}

func (set *Set[K, H, T]) unsafeIsReady() bool {
	return set.buckets != nil
}

// Reset removes all entires from the [Set].
func (set *Set[K, H, T]) Reset() error {
	if set == nil {
		return core.ErrNilReceiver
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	if !set.unsafeIsReady() {
		return core.ErrInvalid
	}

	set.unsafeReset()
	return nil
}

func (set *Set[K, H, T]) unsafeReset() {
	set.buckets = make(map[H]*list.List[T])
}

// Push adds entries to the set unless it already exist.
// It returns the value with matching key stored in the Set so it
// can be treated as a global reference.
func (set *Set[K, H, T]) Push(value T) (T, error) {
	var zero T

	key, hash, err := set.checkWithValue(value)
	if err != nil {
		return zero, err
	}

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	l, ok := set.buckets[hash]
	if !ok {
		// new
		set.buckets[hash] = set.newList(value)
		return value, nil
	}

	if v, err := set.unsafeGet(key, l); err == nil {
		// found
		return v, ErrExist
	}

	// new
	l.PushBack(value)
	return value, nil
}

// Get returns the item matching the key
func (set *Set[K, H, T]) Get(key K) (T, error) {
	var zero T

	hash, err := set.checkWithKey(key)
	if err != nil {
		return zero, err
	}

	// RO
	set.mu.RLock()
	defer set.mu.RUnlock()

	l, ok := set.buckets[hash]
	if !ok {
		return zero, ErrNotExist
	}

	return set.unsafeGet(key, l)
}

func (set *Set[K, H, T]) unsafeGet(key K, l *list.List[T]) (T, error) {
	var zero T

	match := set.newMatchKey(key)
	value, found := l.FirstMatchFn(match)
	if !found {
		return zero, ErrNotExist
	}

	return value, nil
}

// Pop removes and return the item matching the given key from the
// Set.
func (set *Set[K, H, T]) Pop(key K) (T, error) {
	var zero T

	hash, err := set.checkWithKey(key)
	if err != nil {
		return zero, err
	}

	match := set.newMatchKey(key)

	// RW
	set.mu.Lock()
	defer set.mu.Unlock()

	l, ok := set.buckets[hash]
	if !ok {
		return zero, ErrNotExist
	}

	value, ok := l.PopFirstMatchFn(match)
	if !ok {
		return zero, ErrNotExist
	}
	return value, nil
}

func (set *Set[K, H, T]) checkWithValue(value T) (key K, hash H, err error) {
	switch {
	case set == nil:
		return key, hash, core.ErrNilReceiver
	case !set.isReady():
		return key, hash, core.ErrNotImplemented
	default:
		// T -> K
		key, err = set.cfg.ItemKey(value)
		if err != nil {
			return key, hash, err
		}

		// K -> H
		hash, err = set.cfg.Hash(key)
		return key, hash, err
	}
}

func (set *Set[K, H, T]) checkWithKey(key K) (hash H, err error) {
	switch {
	case set == nil:
		return hash, core.ErrNilReceiver
	case !set.isReady():
		return hash, core.ErrNotImplemented
	default:
		return set.cfg.Hash(key)
	}
}

func (*Set[K, H, T]) newList(values ...T) *list.List[T] {
	l := new(list.List[T])
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

func (set *Set[K, H, T]) newMatchKey(key K) func(T) bool {
	fn1 := set.cfg.ItemMatch
	return func(v T) bool {
		return fn1(key, v)
	}
}
