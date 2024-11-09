package basic

import (
	"context"
	"crypto/tls"

	"darvaza.org/x/container/list"
)

// Put adds a [*tls.Certificate] to the [Store] if it passes verification.
//
// If the certificate lacks private key or intermediate certificates, but they
// exist in the [Store] already, gaps will be filled.
func (ss *Store) Put(ctx context.Context, cert *tls.Certificate) error {
	if err := ss.checkPut(ctx, cert); err != nil {
		return err
	}

	return ss.doPut(ctx, cert)
}

func (ss *Store) checkPut(ctx context.Context, cert *tls.Certificate) error {
	if err := ss.Assemble(cert); err != nil {
		return err
	} else if err := ss.Verify(cert); err != nil {
		return err
	}
	return ctx.Err()
}

func (ss *Store) doPut(ctx context.Context, cert *tls.Certificate) error {
	// RW
	ss.mu.Lock()
	ss.unsafeInit()
	fn := ss.OnPut
	err := ss.unsafePut(cert)
	ss.mu.Unlock()

	switch {
	case err != nil:
		// failed
		return err
	case fn != nil:
		// report
		fn(ctx, cert)
		return nil
	default:
		// done
		return nil
	}
}

// unsafePut creates metadata.
func (ss *Store) unsafePut(cert *tls.Certificate) error {
	cert, err := ss.certs.Push(cert)
	if err != nil {
		return err
	}

	meta, err := newCertMeta(cert)
	if err != nil {
		return err
	}

	ss.unsafeLinkMeta(meta)
	return nil
}

// unsafePutMeta reuses the given metadata
func (ss *Store) unsafePutMeta(meta *storeCertMeta) bool {
	_, err := ss.certs.Push(meta.Cert)
	if err != nil {
		return false
	}

	ss.unsafeLinkMeta(meta)
	return true
}

func (ss *Store) unsafeLinkMeta(meta *storeCertMeta) {
	ss.meta[meta.Cert.Leaf] = meta

	ss.unsafeLinkNames(ss.names, meta.Names, meta)
	ss.unsafeLinkNames(ss.patterns, meta.Patterns, meta)
}

func (*Store) unsafeLinkNames(m map[string]*list.List[*storeCertMeta], keys []string, meta *storeCertMeta) {
	for _, key := range keys {
		l, ok := m[key]
		if !ok {
			l = new(list.List[*storeCertMeta])
			m[key] = l
		}
		l.PushBack(meta)
	}
}
