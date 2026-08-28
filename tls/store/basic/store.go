// Package basic implements a generic programmable TLS store
package basic

import (
	"context"
	"crypto"
	"crypto/x509"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/container/list"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils/certpool"
)

var (
	_ tls.Store           = (*Store)(nil)
	_ tls.StoreReader     = (*Store)(nil)
	_ tls.StoreReadWriter = (*Store)(nil)
	_ tls.StoreWriter     = (*Store)(nil)
	_ tls.StoreX509Writer = (*Store)(nil)
)

// Store is a programmable implementation of the [tls.Store].
type Store struct {
	mu sync.RWMutex

	// roots is initially nil to handle the case where no roots
	// have been explicitly added, in which case SystemCertPool()
	// is used.
	roots *certpool.CertPool

	inter    certpool.CertPool
	keys     buffer.KeySet
	certs    buffer.CertSet
	meta     map[*x509.Certificate]*storeCertMeta
	names    map[string]*list.List[*storeCertMeta]
	patterns map[string]*list.List[*storeCertMeta]

	// callbacks
	OnAddPrivateKey func(context.Context, crypto.Signer)
	OnAddCACert     func(context.Context, *x509.Certificate)
	OnAddCert       func(context.Context, *x509.Certificate) // TODO: use
	OnDeleteCert    func(context.Context, *x509.Certificate)
	OnPut           func(context.Context, *tls.Certificate)
	OnDelete        func(context.Context, *tls.Certificate)
	OnMissing       func(context.Context, string) (*tls.Certificate, error)
}

// Clone returns a copy of the [Store].
func (ss *Store) Clone() *Store {
	out := new(Store)
	if ss != nil {
		ss.mu.RLock()
		defer ss.mu.RUnlock()

		ss.unsafeCopy(out)
	}
	return out
}

func (ss *Store) unsafeCopy(out *Store) {
	if !ss.isInitialized() {
		out.unsafeInit()
	}

	// CAs
	out.roots = ss.roots.Copy(out.roots, nil)
	ss.inter.Copy(&out.inter, nil)

	// keys
	ss.keys.Copy(&out.keys, nil)

	// certs
	for _, v := range ss.meta {
		out.unsafePutMeta(v)
	}

	// callbacks
	out.OnAddPrivateKey = core.Coalesce(out.OnAddPrivateKey, ss.OnAddPrivateKey)
	out.OnAddCACert = core.Coalesce(out.OnAddCACert, ss.OnAddCACert)
	out.OnAddCert = core.Coalesce(out.OnAddCert, ss.OnAddCert)
	out.OnDeleteCert = core.Coalesce(out.OnDeleteCert, ss.OnDeleteCert)
	out.OnPut = core.Coalesce(out.OnPut, ss.OnPut)
	out.OnDelete = core.Coalesce(out.OnDelete, ss.OnDelete)
	out.OnMissing = core.Coalesce(out.OnMissing, ss.OnMissing)
}

// New creates a blank [Store].
func New() *Store {
	ss := &Store{}
	ss.unsafeInit()

	return ss
}

func (ss *Store) init(ctx context.Context) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if err := ctx.Err(); err != nil {
		return err
	}

	// RO
	ss.mu.RLock()
	if ss.meta != nil {
		ss.mu.RUnlock()
		return ctx.Err()
	}
	ss.mu.RUnlock()

	// RW
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.unsafeInit()
	return ctx.Err()
}

func (ss *Store) unsafeInit() {
	if ss.meta == nil {
		buffer.MustInitCertSet(&ss.certs)
		buffer.MustInitKeySet(&ss.keys)

		ss.meta = make(map[*x509.Certificate]*storeCertMeta)
		ss.names = make(map[string]*list.List[*storeCertMeta])
		ss.patterns = make(map[string]*list.List[*storeCertMeta])
	}
}

func (ss *Store) isInitialized() bool {
	switch {
	case ss == nil || ss.meta == nil:
		return false
	default:
		return true
	}
}
