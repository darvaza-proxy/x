package buffer

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/container/set"

	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// KeySet keeps a thread-safe set of unique [x509utils.PrivateKey]s.
type KeySet struct {
	set.Set[x509utils.PublicKey, certpool.Hash, x509utils.PrivateKey]
}

// Public handles type-casting of public keys.
func (*KeySet) Public(key x509utils.PrivateKey) x509utils.PublicKey {
	if key != nil {
		pub, ok := key.Public().(x509utils.PublicKey)
		if ok {
			return pub
		}
	}

	return nil
}

// GetFromCertificate is like Get but uses the public key associated with the
// given certificate.
func (ks *KeySet) GetFromCertificate(cert *x509.Certificate) (x509utils.PrivateKey, error) {
	if ks == nil {
		return nil, core.ErrNilReceiver
	}

	if pub := x509utils.PublicKeyFromCertificate(cert); pub != nil {
		return ks.Get(pub)
	}

	return nil, core.Wrap(core.ErrInvalid, "invalid certificate provided")
}

// Copy copies all keys that satisfy the condition to the destination [KeySet]
// unless they are already there.
// If a destination isn't provided one will be created.
// If a condition function isn't provided, all keys not present in the destination
// will be added.
func (ks *KeySet) Copy(dst *KeySet, cond func(x509utils.PrivateKey) bool) *KeySet {
	if ks == nil {
		if dst == nil {
			dst = MustKeySet()
		}
		return dst
	}

	if dst == nil {
		dst = new(KeySet)
	}

	ks.Set.Copy(&dst.Set, cond)
	return dst
}

// Clone creates a copy of the [KeySet].
func (ks *KeySet) Clone() *KeySet {
	if ks == nil {
		return nil
	}

	return ks.Copy(nil, nil)
}

// NewKeySet creates a KeySet optionally taking its initial content as argument.
func NewKeySet(keys ...x509utils.PrivateKey) (*KeySet, error) {
	out := new(KeySet)
	if err := keySetConfig.Init(&out.Set, keys...); err != nil {
		return nil, err
	}
	return out, nil
}

// InitKeySet initializes a preallocated [KeySet].
func InitKeySet(out *KeySet, keys ...x509utils.PrivateKey) error {
	if out == nil {
		return core.Wrap(core.ErrInvalid, "missing KeySet")
	}

	return keySetConfig.Init(&out.Set, keys...)
}

// MustKeySet is like [NewKeySet] but panics on errors.
func MustKeySet(keys ...x509utils.PrivateKey) *KeySet {
	out, err := NewKeySet(keys...)
	if err != nil {
		core.Panic(core.Wrap(err, "failed to create KeySet"))
	}
	return out
}

// MustInitKeySet is like [InitKeySet] but panics on errors.
func MustInitKeySet(out *KeySet, keys ...x509utils.PrivateKey) {
	err := InitKeySet(out, keys...)
	if err != nil {
		core.Panic(core.Wrap(err, "failed to initialize KeySet"))
	}
}

func keySetHash(pub x509utils.PublicKey) (certpool.Hash, error) {
	if pub == nil {
		// bad
		return certpool.Hash{}, core.ErrInvalid
	}

	b, err := x509utils.SubjectPublicKeyBytes(pub)
	if err != nil {
		return certpool.Hash{}, err
	}

	return certpool.Sum(b), nil
}

func keySetItemKey(key x509utils.PrivateKey) (x509utils.PublicKey, error) {
	if key != nil {
		if pub, ok := key.Public().(x509utils.PublicKey); ok {
			return pub, nil
		}
	}
	return nil, core.ErrInvalid
}

func keySetItemMatch(pub x509utils.PublicKey, key x509utils.PrivateKey) bool {
	if pub == nil || key == nil {
		return false
	}

	return pub.Equal(key.Public())
}

// keySetConfig defines the behaviour of KeySet's underlying Set implementation,
// including how keys are hashed, validated, and compared for equality.
var keySetConfig = set.Config[x509utils.PublicKey, certpool.Hash, x509utils.PrivateKey]{
	Hash:      keySetHash,
	ItemKey:   keySetItemKey,
	ItemMatch: keySetItemMatch,
}
