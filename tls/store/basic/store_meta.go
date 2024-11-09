package basic

import (
	"crypto/tls"

	"darvaza.org/x/tls/x509utils"
)

type storeCertMeta struct {
	Cert     *tls.Certificate
	Names    []string
	Patterns []string
}

func newCertMeta(cert *tls.Certificate) (*storeCertMeta, error) {
	switch {
	case cert == nil:
		return nil, ErrNoCert
	case cert.Leaf == nil || len(cert.Leaf.Raw) == 0:
		return nil, ErrBadCert
	default:
		names, patterns := x509utils.Names(cert.Leaf)
		out := &storeCertMeta{
			Cert:     cert,
			Names:    names,
			Patterns: patterns,
		}
		return out, nil
	}
}

type storeCertMetaSet map[*storeCertMeta]struct{}

func (set storeCertMetaSet) ForEach(fn func(*tls.Certificate) bool) {
	if fn != nil {
		for meta := range set {
			if !set.doForEach(fn, meta) {
				return
			}
		}
	}
}

func (storeCertMetaSet) doForEach(fn func(*tls.Certificate) bool, meta *storeCertMeta) bool {
	if meta != nil && meta.Cert != nil {
		return fn(meta.Cert)
	}
	return true
}

func (set storeCertMetaSet) Len() int {
	return len(set)
}

func (set storeCertMetaSet) Export() []*tls.Certificate {
	var out []*tls.Certificate

	if n := len(set); n > 0 {
		out = make([]*tls.Certificate, 0, n)

		set.ForEach(func(cert *tls.Certificate) bool {
			out = append(out, cert)
			return true
		})
	}

	return out
}
