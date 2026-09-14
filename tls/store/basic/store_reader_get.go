package basic

import (
	"context"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/container/list"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// GetCertificate returns the [tls.Certificate] to be used in response to
// a [tls.ClientHelloInfo].
func (ss *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	ctx, serverName, err := tls.SplitClientHelloInfo(chi)
	switch {
	case err != nil:
		return nil, err
	case ss == nil:
		return nil, core.ErrNilReceiver
	default:
		return ss.Get(ctx, serverName)
	}
}

// Get returns the certificate to use for a given serverName. If no match is found,
// the OnMissing callback will be used.
func (ss *Store) Get(ctx context.Context, serverName string) (*tls.Certificate, error) {
	if err := ss.init(ctx); err != nil {
		return nil, err
	}

	// fallback shortcut
	if serverName == "" || serverName == "." {
		return ss.doGetMissing(ctx, serverName)
	}

	name, ok := x509utils.SanitizeName(serverName)
	if !ok {
		return nil, core.Wrapf(ErrInvalid, "serverName:%q", serverName)
	}

	// get certificate for sanitized name.
	return ss.doGet(ctx, name)
}

func (ss *Store) doGetMissing(ctx context.Context, serverName string) (cert *tls.Certificate, err error) {
	ss.mu.RLock()
	fn := ss.OnMissing
	ss.mu.RUnlock()

	if fn != nil {
		cert, err = fn(ctx, serverName)
	}

	switch {
	case cert != nil:
		return cert, nil
	case err != nil:
		return nil, err
	default:
		return nil, ErrCertNotFound
	}
}

func (ss *Store) doGet(ctx context.Context, serverName string) (*tls.Certificate, error) {
	now := time.Now().UTC()
	cond := ss.newMetaNotExpired(ctx, now)

	ss.mu.RLock()
	cert := ss.unsafeGet(ctx, serverName, cond)
	ss.mu.RUnlock()

	if cert != nil {
		return cert, nil
	}

	return ss.doGetMissing(ctx, serverName)
}

func (ss *Store) unsafeGet(_ context.Context, serverName string, cond func(*storeCertMeta) bool) *tls.Certificate {
	if meta, ok := ss.unsafeGetNameList(serverName).FirstMatchFn(cond); ok {
		return meta.Cert
	}

	if serverName[0] != '[' {
		// unless it's an address, check for wildcard certificates.
		if meta, ok := ss.unsafeGetWildcardList(serverName).FirstMatchFn(cond); ok {
			return meta.Cert
		}
	}

	return nil
}

func (ss *Store) unsafeGetNameList(name string) *list.List[*storeCertMeta] {
	if l, ok := ss.names[name]; ok {
		return l
	}
	return nil
}

func (ss *Store) unsafeGetWildcardList(name string) *list.List[*storeCertMeta] {
	suffix, ok := x509utils.NameAsSuffix(name)
	if !ok {
		return nil
	}
	if l, ok := ss.patterns[suffix]; ok {
		return l
	}
	return nil
}

func (ss *Store) newMetaNotExpired(ctx context.Context, now time.Time) func(meta *storeCertMeta) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	isExpired := func(cert *tls.Certificate) bool {
		return cert.Leaf.NotAfter.Before(now) || cert.Leaf.NotBefore.After(now)
	}

	return func(meta *storeCertMeta) bool {
		if isExpired(meta.Cert) {
			go ss.deferredDelete(ctx, meta.Cert)

			return false
		}
		return true
	}
}
