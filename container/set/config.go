package set

import (
	"reflect"

	"darvaza.org/core"
)

// Config defines the callbacks used by a [Set].
type Config[K, H comparable, T any] struct {
	// Hash computes the bucket identifier for the given K type
	Hash func(k K) (H, error)
	// ItemKey computes the K value from the T instance
	ItemKey func(v T) (K, error)
	// ItemMatch confirms the T instance matches the K value
	ItemMatch func(k K, v T) bool
}

// Validate confirms the [Config] is good for use.
func (cfg Config[K, H, T]) Validate() error {
	var errs core.CompoundError
	if cfg.Hash == nil {
		_ = errs.Append(core.ErrInvalid, "missing callback: %s", "Hash")
	}
	if cfg.ItemKey == nil {
		_ = errs.Append(core.ErrInvalid, "missing callback: %s", "ItemKey")
	}
	if cfg.ItemMatch == nil {
		_ = errs.Append(core.ErrInvalid, "missing callback: %s", "ItemMatch")
	}
	return errs.AsError()
}

// New creates a [Set] based on the [Config] optionally receiving items.
// duplicates won't cause errors, just make sure to [Set.Get] the
// actually stored instances before assuming uniqueness.
func (cfg Config[K, H, T]) New(items ...T) (*Set[K, H, T], error) {
	set := new(Set[K, H, T])
	if err := cfg.Init(set, items...); err != nil {
		return nil, err
	}
	return set, nil
}

// Init initializes a [Set] that wasn't created using [Config.New].
func (cfg Config[K, H, T]) Init(set *Set[K, H, T], items ...T) error {
	var err error
	if err = cfg.Validate(); err != nil {
		return err
	} else if set == nil {
		return core.Wrap(core.ErrInvalid, "set")
	}

	return set.init(cfg, items...)
}

// Must is equivalent to [New] but it panics on error.
func (cfg Config[K, H, T]) Must(items ...T) *Set[K, H, T] {
	set, err := cfg.New(items...)
	if err != nil {
		core.Panic(err)
	}
	return set
}

// Equal determines if two Config instances use exactly the same callback functions
// by comparing their memory addresses using reflection. This is used to optimize
// Set operations like Copy() when configs are identical.
//
// Equal isn't cheap but it saves Copy() from performing unnecessary rehashing.
func (cfg Config[K, H, T]) Equal(other Config[K, H, T]) bool {
	switch {
	case !equalFuncPointer(cfg.Hash, other.Hash):
		return false
	case !equalFuncPointer(cfg.ItemKey, other.ItemKey):
		return false
	case !equalFuncPointer(cfg.ItemMatch, other.ItemMatch):
		return false
	default:
		return true
	}
}

func equalFuncPointer[T any](f1, f2 T) bool {
	v1 := reflect.ValueOf(f1)
	v2 := reflect.ValueOf(f2)
	nil1 := v1.IsNil()
	nil2 := v2.IsNil()

	switch {
	case nil1 && nil2:
		return true
	case nil1, nil2:
		return false
	default:
		return v1.Pointer() == v2.Pointer()
	}
}
