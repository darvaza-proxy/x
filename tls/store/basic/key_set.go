package basic

import (
	"crypto"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// KeySet keeps a unique set of [x509utils.PrivateKey].
type KeySet map[certpool.Hash]*KeyList

// NewKeySet creates a KeySet optionally taking its initial content as argument.
func NewKeySet(keys ...crypto.Signer) KeySet {
	set := make(KeySet)
	for _, key := range keys {
		set.Push(key)
	}
	return set
}

// Sys returns the underlying map.
func (set KeySet) Sys() map[certpool.Hash]*KeyList {
	if set == nil {
		return nil
	}

	return (map[certpool.Hash]*KeyList)(set)
}

// Hash tells if the public key is usable, and returns the hash for it.
func (KeySet) Hash(pub crypto.PublicKey) (certpool.Hash, bool) {
	if pub == nil {
		// bad
		return certpool.Hash{}, false
	}

	b, err := x509utils.SubjectPublicKeyBytes(pub)
	if err != nil {
		// bad
		return certpool.Hash{}, false
	}

	hash := certpool.Sum(b)
	return hash, true
}

// Push adds a [crypto.Signer] to the set.
// It returns a singleton reference to the key in the set, and
// if the key is newly added or it existed before.
func (set KeySet) Push(key crypto.Signer) (crypto.Signer, bool) {
	if set == nil || key == nil {
		// bad
		return nil, false
	}

	hash, ok := set.Hash(key.Public())
	if !ok {
		// bad
		return nil, false
	}

	l, ok := set[hash]
	if !ok {
		// new
		set[hash] = NewKeyList(key)
		return key, true
	}

	if k, _ := l.Get(key.Public()); k != nil {
		// found
		return k, false
	}

	// new
	l.PushBack(key)
	return key, true
}

// Get returns the [crypto.Signer] matching the given [crypto.PublicKey].
func (set KeySet) Get(pub crypto.PublicKey) (crypto.Signer, error) {
	switch {
	case set == nil:
		return nil, core.ErrNilReceiver
	case pub == nil:
		// invalid public key
		return nil, ErrInvalid
	}

	hash, ok := set.Hash(pub)
	if !ok {
		// invalid public key
		return nil, ErrInvalid
	}

	l, ok := set[hash]
	if !ok {
		// miss
		return nil, ErrNotExists
	}

	if k, _ := l.Get(pub); k != nil {
		// found
		return k, nil
	}

	// miss
	return nil, ErrNotExists
}

// Copy copies all keys in the [KeySet] to another.
// If no destination is provided, a new one is created.
func (set KeySet) Copy(dst KeySet) KeySet {
	if dst == nil {
		dst = make(KeySet)
	}

	for hash, l1 := range set {
		l2, ok := dst[hash]
		if !ok {
			l2 = NewKeyList()
			dst[hash] = l2
		}

		set.doCopy(l1, l2)
	}

	return dst
}

func (KeySet) doCopy(l1, l2 *KeyList) {
	l1.ForEach(func(key crypto.Signer) bool {
		if !l2.Contains(key.Public()) {
			l2.PushBack(key)
		}

		return true
	})
}
