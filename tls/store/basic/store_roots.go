package basic

import (
	"context"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// GetCAPool returns a snapshot of the CACert pool in the [Store].
func (ss *Store) GetCAPool() *x509.CertPool {
	if r, ok := ss.getRoots(); ok {
		return r.Export()
	}

	return x509.NewCertPool()
}

func (ss *Store) getRoots() (*certpool.CertPool, bool) {
	if ss == nil {
		return nil, false
	}

	// RO
	ss.mu.RLock()
	roots := ss.roots
	ss.mu.RUnlock()

	if roots != nil {
		// ready
		return roots, true
	}

	roots, err := certpool.SystemCertPool()
	if err != nil {
		// failed to get system certs
		return nil, false
	}

	// RW
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.roots == nil {
		if roots == nil {
			// fresh
			roots = certpool.New()
		}
		ss.roots = roots
	}

	return ss.roots, ss.roots != nil
}

// AddCACerts adds trusted root certificates to the store. If none has been added by the time
// GetCAPool() or Put()/AddCert()/AddCertPair() is called, system's roots will be loaded
// automatically.
// AddCACerts only accepts certificates marked as IsCA.
func (ss *Store) AddCACerts(ctx context.Context, roots ...*x509.Certificate) error {
	if err := ss.init(ctx); err != nil {
		return err
	}

	ch := make(chan *x509.Certificate, 1)
	defer close(ch)

	go ss.reportAddCACerts(ctx, ch)

	for _, root := range roots {
		if root != nil {
			if err := ss.doAddCACert(ctx, root); err != nil {
				return err
			}
			ch <- root
		}
	}

	return nil
}

func (ss *Store) reportAddCACerts(ctx context.Context, ch <-chan *x509.Certificate) {
	ss.mu.RLock()
	fn := ss.OnAddCACert
	ss.mu.RUnlock()

	for root := range ch {
		if fn != nil {
			fn(ctx, root)
		}
	}
}

func (ss *Store) reportAddCACert(ctx context.Context, root *x509.Certificate) {
	ss.mu.RLock()
	fn := ss.OnAddCACert
	ss.mu.RUnlock()

	if fn != nil {
		fn(ctx, root)
	}
}

func (ss *Store) doAddCACert(ctx context.Context, root *x509.Certificate) error {
	if err := ss.checkAddCACert(ctx, root); err != nil {
		return err
	}

	// RW
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.roots == nil {
		ss.roots = certpool.New()
	}

	ss.roots.AddCert(root)
	return nil
}

func (ss *Store) checkAddCACert(ctx context.Context, root *x509.Certificate) error {
	switch {
	case ss == nil:
		return core.ErrNilReceiver
	case root == nil:
		return ErrNoCert
	case !root.IsCA:
		return x509utils.ErrInvalidCert{
			Cert:   root,
			Reason: "not a CA",
		}
	default:
		return ss.init(ctx)
	}
}
