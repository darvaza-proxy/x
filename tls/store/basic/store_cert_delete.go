package basic

import (
	"context"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/container/list"

	"darvaza.org/x/tls"
)

// Delete removes a [tls.Certificate] from the [Store].
func (ss *Store) Delete(ctx context.Context, cert *tls.Certificate) error {
	if err := ss.checkDelete(ctx, cert); err != nil {
		return err
	}

	return ss.doDelete(ctx, cert.Leaf)
}

func (ss *Store) checkDelete(ctx context.Context, cert *tls.Certificate) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if cert == nil {
		return ErrNoCert
	} else if cert.Leaf == nil {
		return ErrBadCert
	}

	return ctx.Err()
}

// deferredDelete is called in background when evicting an expired certificate
func (ss *Store) deferredDelete(ctx context.Context, cert *tls.Certificate) {
	_ = ss.doDelete(ctx, cert.Leaf)
}

func (ss *Store) doDelete(ctx context.Context, leaf *x509.Certificate) error {
	// RW
	ss.mu.Lock()
	ss.unsafeInit()
	cert := ss.unsafeDelete(leaf)
	ss.mu.Unlock()

	if cert == nil {
		return ErrCertNotFound
	}

	go ss.reportDelete(ctx, cert)
	return nil
}

func (ss *Store) unsafeDelete(leaf *x509.Certificate) *tls.Certificate {
	cert, _ := ss.certs.Pop(leaf)
	if cert == nil {
		return nil
	}

	meta := ss.meta[cert.Leaf]
	delete(ss.meta, cert.Leaf)

	// unlink
	ss.unsafeUnlinkNames(ss.names, meta.Names, meta)
	ss.unsafeUnlinkNames(ss.patterns, meta.Patterns, meta)
	return cert
}

func (*Store) unsafeUnlinkNames(m map[string]*list.List[*storeCertMeta], keys []string, meta *storeCertMeta) {
	for _, key := range keys {
		if l, ok := m[key]; ok {
			l.DeleteMatchFn(func(e2 *storeCertMeta) bool {
				return e2 == meta
			})
		}
	}
}

// DeleteCert removes a certificate from the [Store].
// This includes [tls.Certificates]s, but also the roots and intermediate [x509.Certificate]s.
func (ss *Store) DeleteCert(ctx context.Context, leaf *x509.Certificate) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if leaf == nil {
		return ErrNoCert
	} else if err := ctx.Err(); err != nil {
		return err
	}

	// RW
	ss.mu.Lock()
	ss.unsafeInit()
	cert, err := ss.unsafeDeleteCert(leaf)
	ss.mu.Unlock()

	switch {
	case err != nil:
		return err
	case cert == nil:
		go ss.reportDeleteCert(ctx, leaf)
	default:
		go ss.reportDelete(ctx, cert)
	}

	return nil
}

func (ss *Store) unsafeDeleteCert(leaf *x509.Certificate) (*tls.Certificate, error) {
	// no cancelling
	ctx := context.Background()

	cert := ss.unsafeDelete(leaf)
	ok1 := cert != nil
	ok2 := ss.roots.DeleteCert(ctx, leaf) == nil
	ok3 := ss.inter.DeleteCert(ctx, leaf) == nil

	if ok1 || ok2 || ok3 {
		// success
		return cert, nil
	}

	return nil, ErrCertNotFound
}

func (ss *Store) reportDelete(ctx context.Context, cert *tls.Certificate) {
	ss.mu.RLock()
	fn1 := ss.OnDelete
	fn2 := ss.OnDeleteCert
	ss.mu.RUnlock()

	switch {
	case fn1 != nil:
		fn1(ctx, cert)
	case fn2 != nil:
		fn2(ctx, cert.Leaf)
	}
}

func (ss *Store) reportDeleteCert(ctx context.Context, cert *x509.Certificate) {
	ss.mu.RLock()
	fn := ss.OnDeleteCert
	ss.mu.RUnlock()

	if fn != nil {
		fn(ctx, cert)
	}
}
