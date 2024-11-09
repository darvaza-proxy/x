package basic

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/x/container/list"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// ForEach calls a function for each stored certificate.
func (ss *Store) ForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool) {
	if ss.checkForEach(ctx, fn) {
		doForEach(ctx, fn, ss.exportCerts())
	}
}

func (ss *Store) checkForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool) bool {
	if ss != nil && fn != nil && ctx.Err() == nil {
		ss.mu.RLock()
		defer ss.mu.RUnlock()

		if ctx.Err() == nil {
			// initialized only
			return ss.isInitialized()
		}
	}

	return false
}

func (ss *Store) exportCerts() []*tls.Certificate {
	ss.mu.RLock()
	out := ss.certs.Values()
	ss.mu.RUnlock()

	return out
}

// ForEachDuplicate calls a function for each name supported by more than one certificate.
func (ss *Store) ForEachDuplicate(ctx context.Context, fn func(context.Context, string, []*tls.Certificate) bool) {
	if ss.checkForEach2(ctx, fn) {
		doForEach2(ctx, fn, ss.exportDuplicateCerts())
	}
}

func (ss *Store) checkForEach2(ctx context.Context, fn func(context.Context, string, []*tls.Certificate) bool) bool {
	if ss != nil && fn != nil && ctx.Err() == nil {
		ss.mu.RLock()
		defer ss.mu.RUnlock()

		if ctx.Err() == nil {
			// initialized only
			return ss.isInitialized()
		}
	}

	return false
}

func (ss *Store) exportDuplicateCerts() map[string][]*tls.Certificate {
	out := make(map[string][]*tls.Certificate)

	ss.mu.RLock()
	for k, l := range ss.names {
		unsafeExportDuplicateCerts(out, k, l)
	}
	for k, l := range ss.patterns {
		unsafeExportDuplicateCerts(out, "*"+k, l)
	}
	ss.mu.RUnlock()

	return out
}

func unsafeExportDuplicateCerts(out map[string][]*tls.Certificate, name string, l *list.List[*storeCertMeta]) {
	if l.Len() > 1 {
		certs := l.Values()
		if len(certs) > 1 {
			out[name] = core.SliceAsFn(func(meta *storeCertMeta) (*tls.Certificate, bool) {
				return meta.Cert, true
			}, certs)
		}
	}
}

// ForEachMatch calls a function for each stored certificate matching the given
// serverName.
func (ss *Store) ForEachMatch(ctx context.Context, serverName string, fn func(context.Context, *tls.Certificate) bool) {
	if name, ok := ss.checkForEachMatch(ctx, serverName, fn); ok {
		doForEach(ctx, fn, ss.exportMatchingCerts(name))
	}
}

func (ss *Store) checkForEachMatch(ctx context.Context, serverName string, //
	fn func(context.Context, *tls.Certificate) bool) (string, bool) {
	//
	switch {
	case ss == nil, fn == nil:
		return "", false
	case serverName == "", serverName == ".":
		return "", false
	case ctx.Err() != nil:
		return "", false
	default:
		name, ok := x509utils.SanitizeName(serverName)
		if !ok {
			return "", false
		}

		ss.mu.RLock()
		defer ss.mu.RUnlock()

		if ctx.Err() == nil {
			// initialized only
			return name, ss.isInitialized()
		}

		return "", false
	}
}

func (ss *Store) exportMatchingCerts(name string) []*tls.Certificate {
	ss.mu.RLock()
	m := ss.unsafeExportMatchingCertsMap(name)
	ss.mu.RUnlock()

	return m.Export()
}

func (ss *Store) unsafeExportMatchingCertsMap(name string) storeCertMetaSet {
	var l1, l2 *list.List[*storeCertMeta]

	l1 = ss.unsafeGetNameList(name)
	if name[0] != '[' {
		l2 = ss.unsafeGetWildcardList(name)
	}

	count := l1.Len() + l2.Len()
	if count == 0 {
		return nil
	}

	m, fn := newUnsafeCertExporterSet(count)
	l1.ForEach(fn)
	l2.ForEach(fn)
	return m
}

func newUnsafeCertExporterSet(count int) (storeCertMetaSet, func(*storeCertMeta) bool) {
	m := make(storeCertMetaSet, count)
	fn := func(meta *storeCertMeta) bool {
		m[meta] = struct{}{}
		return true
	}

	return m, fn
}
