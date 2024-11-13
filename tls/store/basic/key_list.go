package basic

import (
	"crypto"

	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// KeyList is a double-linked list of private keys.
type KeyList certpool.List[crypto.Signer]

// NewKeyList creates a KeyList optionally taking its initial content as argument.
func NewKeyList(keys ...crypto.Signer) *KeyList {
	l := new(KeyList)
	for _, key := range keys {
		if key != nil {
			l.PushBack(key)
		}
	}
	return l
}

// Sys returns the underlying [certpool.List].
func (l *KeyList) Sys() *certpool.List[crypto.Signer] {
	if l == nil {
		return nil
	}

	return (*certpool.List[crypto.Signer])(l)
}

// PushBack appends a key at the end of the list.
// nil values are ignored.
func (l *KeyList) PushBack(key crypto.Signer) {
	if key != nil {
		l.Sys().PushBack(key)
	}
}

// PushFront inserts a key at the beginning of the list.
// nil values are ignored.
func (l *KeyList) PushFront(key crypto.Signer) {
	if key != nil {
		l.Sys().PushFront(key)
	}
}

// Contains tells if the [KeyList] contains a [crypto.Signer]
// matching the given [crypto.PublicKey].
func (l *KeyList) Contains(pub crypto.PublicKey) bool {
	_, found := l.Get(pub)
	return found
}

// Get returns the [crypto.Signer] matching the given
// [crypto.PublicKey].
func (l *KeyList) Get(pub crypto.PublicKey) (crypto.Signer, bool) {
	var out crypto.Signer

	if l == nil || pub == nil {
		// no-op
		return nil, false
	}

	pk, ok := pub.(x509utils.PublicKey)
	if !ok {
		// invalid key
		return nil, false
	}

	l.ForEach(func(key crypto.Signer) bool {
		if pk.Equal(key.Public()) {
			out = key
		}
		return out == nil // continue until one is found.
	})

	return out, out != nil
}

// ForEach calls a function for each private key in the list until
// it returns false.
func (l *KeyList) ForEach(fn func(crypto.Signer) bool) {
	if l == nil || fn == nil {
		return
	}

	l.Sys().ForEach(fn)
}
