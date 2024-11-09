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
	OnPut           func(context.Context, *tls.Certificate)

	// OnDelete is called when a certificate is deleted
	OnDelete func(context.Context, *tls.Certificate)
	// OnDeleteCert is called when a certificate is deleted and OnDelete
	// isn't defined.
	OnDeleteCert func(context.Context, *x509.Certificate)

	// OnMissing is called when no matching certificate is found. If a certificate
	// is returned, it will be used. if an error is returned, the response won't
	// be remembered.
	OnMissing func(context.Context, string) (*tls.Certificate, error)
}

// Clone returns a copy of the [Store].
func (ss *Store) Clone() *Store {
	out := New()
	if ss != nil {
		ss.mu.RLock()
		defer ss.mu.RUnlock()

		ss.unsafeInit()
		ss.unsafeCopy(out)
	}
	return out
}

// unsafeCopy copies entries in this [Store] onto the provided [Store].
// assuming both are locked already.
func (ss *Store) unsafeCopy(out *Store) {
	// CAs
	out.roots = ss.roots.Copy(out.roots, nil)
	ss.inter.Copy(&out.inter, nil)

	// keys
	ss.keys.Copy(&out.keys, nil)

	// certs
	for _, v := range ss.meta {
		out.unsafePutMeta(v)
	}

	// callbacks. don't override if already set.
	out.OnAddPrivateKey = core.Coalesce(out.OnAddPrivateKey, ss.OnAddPrivateKey)
	out.OnAddCACert = core.Coalesce(out.OnAddCACert, ss.OnAddCACert)
	out.OnAddCert = core.Coalesce(out.OnAddCert, ss.OnAddCert)
	out.OnDeleteCert = core.Coalesce(out.OnDeleteCert, ss.OnDeleteCert)
	out.OnDelete = core.Coalesce(out.OnDelete, ss.OnDelete)
	out.OnPut = core.Coalesce(out.OnPut, ss.OnPut)
	out.OnMissing = core.Coalesce(out.OnMissing, ss.OnMissing)
}

// New creates a blank [Store].
func New() *Store {
	ss := &Store{}
	ss.unsafeInit()

	return ss
}

// init attempts to initialize the [Store] using locks and taking
// a context for cancellation.
func (ss *Store) init(ctx context.Context) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if err := ctx.Err(); err != nil {
		return err
	}

	// RO
	ss.mu.RLock()
	ready := ss.isInitialized()
	ss.mu.RUnlock()

	if ready {
		return ctx.Err()
	}

	// RW
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.unsafeInit()
	return ctx.Err()
}

// unsafeInit initializes the [Store] if it isn't initialized already.
// lock before use.
func (ss *Store) unsafeInit() {
	if !ss.isInitialized() {
		buffer.MustInitCertSet(&ss.certs)
		buffer.MustInitKeySet(&ss.keys)

		ss.meta = make(map[*x509.Certificate]*storeCertMeta)
		ss.names = make(map[string]*list.List[*storeCertMeta])
		ss.patterns = make(map[string]*list.List[*storeCertMeta])
	}
}

// isInitialized tells if the [Store] is ready to be used.
// lock before use.
func (ss *Store) isInitialized() bool {
	return ss != nil && ss.meta != nil
}
