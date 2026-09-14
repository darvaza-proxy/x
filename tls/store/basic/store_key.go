package basic

import (
	"context"
	"crypto"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
	"darvaza.org/x/tls/x509utils"
)

// AddPrivateKey adds a private key to the [Store], to be used later
// when bundling certificates.
func (ss *Store) AddPrivateKey(ctx context.Context, key crypto.Signer) error {
	pk, err := ss.checkAddPrivateKey(ctx, key)
	if err != nil {
		return err
	}

	// RW
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.unsafeInit()
	return ss.unsafeAddPrivateKey(ctx, pk)
}

func (ss *Store) checkAddPrivateKey(ctx context.Context, key crypto.Signer) (x509utils.PrivateKey, error) {
	if ss == nil {
		return nil, core.ErrNilReceiver
	} else if key == nil {
		return nil, ErrNoKey
	} else if key.Public() == nil {
		return nil, ErrBadKey
	}

	pk, ok := key.(x509utils.PrivateKey)
	if !ok {
		return nil, ErrBadKey
	}

	return pk, ctx.Err()
}

func (ss *Store) unsafeAddPrivateKey(ctx context.Context, key x509utils.PrivateKey) error {
	key, err := ss.keys.Push(key)
	switch {
	case err == nil:
		go ss.reportAddPrivateKey(ctx, key)
		return nil
	case key != nil:
		return set.ErrExist
	default:
		return ErrBadKey
	}
}

func (ss *Store) reportAddPrivateKey(ctx context.Context, key crypto.Signer) {
	ss.mu.RLock()
	fn := ss.OnAddPrivateKey
	ss.mu.RUnlock()

	if fn != nil {
		fn(ctx, key)
	}
}

// GetPrivateKey attempts to get a private key matching the given public key from the [Store].
func (ss *Store) GetPrivateKey(ctx context.Context, pub crypto.PublicKey) (crypto.Signer, error) {
	pub2, err := ss.checkGetPrivateKey(ctx, pub)
	if err != nil {
		return nil, err
	}

	// RO
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if !ss.isInitialized() {
		return nil, ErrNotExist
	}

	return ss.keys.Get(pub2)
}

func (ss *Store) checkGetPrivateKey(ctx context.Context, pub crypto.PublicKey) (x509utils.PublicKey, error) {
	if ss == nil {
		return nil, core.ErrNilReceiver
	} else if pub == nil {
		return nil, ErrNoKey
	}

	pub2, ok := pub.(x509utils.PublicKey)
	if !ok {
		return nil, ErrBadKey
	}

	return pub2, ctx.Err()
}
