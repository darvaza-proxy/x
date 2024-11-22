package certpool

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/x/container/set"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ x509utils.CertPoolWriter = (*CertPool)(nil)
)

// Put inserts a certificate into the store, optionally including a reference name.
// The name will be appended to those included in the certificate.
func (s *CertPool) Put(ctx context.Context, name string, cert *x509.Certificate) error {
	name, err := s.checkPut(ctx, name, cert)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.unsafeAddCert(name, cert)
}

func (s *CertPool) checkPut(ctx context.Context, name string, cert *x509.Certificate) (string, error) {
	sn, ok := x509utils.SanitizeName(name)
	switch {
	case !ok:
		return "", core.Wrapf(core.ErrInvalid, "%s: %q", "name", name)
	case !validCert(cert):
		return "", core.Wrap(core.ErrInvalid, "cert")
	case s == nil:
		return "", core.ErrNilReceiver
	default:
		return sn, ctx.Err()
	}
}

// Delete remove from the store all certificates associated to the given name
func (s *CertPool) Delete(ctx context.Context, name string) error {
	sn, err := s.checkDelete(ctx, name)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// convert name to unique certificates
	certs := core.SliceAsFn(func(ce *certPoolEntry) (*x509.Certificate, bool) {
		return ce.cert, ce.cert != nil
	}, s.getListForName(sn).Values())

	if s.unsafeDeleteCerts(certs...) > 0 {
		return nil
	}

	return core.Wrap(core.ErrNotExists, name)
}

func (s *CertPool) checkDelete(ctx context.Context, name string) (string, error) {
	sn, ok := x509utils.SanitizeName(name)
	switch {
	case !ok:
		return "", core.Wrapf(core.ErrInvalid, "%s: %q", "name", name)
	case s == nil:
		return "", core.ErrNilReceiver
	default:
		return sn, ctx.Err()
	}
}

// DeleteCert removes a certificate, by raw DER hash, from the store.
func (s *CertPool) DeleteCert(ctx context.Context, cert *x509.Certificate) error {
	if !validCert(cert) {
		return core.Wrap(core.ErrInvalid, "cert")
	} else if s == nil {
		return core.ErrNilReceiver
	} else if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.unsafeDeleteCerts(cert) > 0 {
		return nil
	}

	return core.ErrNotExists
}

func (s *CertPool) unsafeDeleteCerts(certs ...*x509.Certificate) int {
	var count int

	for _, cert := range certs {
		if ce, ok := s.entries[cert]; ok {
			s.unsafeDeleteCertEntry(ce)
			count++
		}
	}

	return count
}

func (s *CertPool) unsafeDeleteCertEntry(ce *certPoolEntry) {
	s.unsafeInvalidateCache()
	cert := ce.cert
	if c, _ := s.certs.Pop(cert); c != nil {
		cert = c
	}

	delete(s.entries, cert)

	eq := func(p *certPoolEntry) bool {
		return ce == p
	}

	deleteMapListMatchFn(s.names, ce.names, eq)
	deleteMapListMatchFn(s.patterns, ce.patterns, eq)
}

// Import certificates from another CertPool.
func (s *CertPool) Import(ctx context.Context, src x509utils.CertPool) (int, error) {
	if s == nil {
		return 0, core.ErrNilReceiver
	} else if err := ctx.Err(); err != nil {
		return 0, err
	} else if src == nil || src == s {
		return 0, nil
	}

	return s.doImport(ctx, src)
}

func (s *CertPool) doImport(ctx context.Context, src x509utils.CertPool) (int, error) {
	var count int
	var err error

	fn := func(ctx context.Context, cert *x509.Certificate) bool {
		if e := ctx.Err(); e != nil {
			err = e
			return false
		}

		if s.AddCert(cert) {
			count++
		}

		return true
	}

	src.ForEach(ctx, fn)
	return count, err
}

// ImportPEM adds x509 certificates contained in the PEM encoded data.
func (s *CertPool) ImportPEM(ctx context.Context, b []byte) (int, error) {
	if s == nil {
		return 0, core.ErrNilReceiver
	} else if err := ctx.Err(); err != nil {
		return 0, err
	} else if len(b) == 0 {
		return 0, nil
	}

	return s.doImportPEM(ctx, b)
}

func (s *CertPool) doImportPEM(ctx context.Context, b []byte) (int, error) {
	var count int
	var err error

	fn := func(_ fs.FS, _ string, block *pem.Block) bool {
		cert, e1 := x509utils.BlockToCertificate(block)
		if cert != nil && s.AddCert(cert) {
			count++
		}

		if e2 := s.checkImportError(ctx, e1); e2 != nil {
			err = e2
			return false
		}

		return true
	}

	if e := x509utils.ReadPEM(b, fn); e != nil {
		return 0, e
	}

	return count, err
}

func (*CertPool) checkImportError(ctx context.Context, err error) error {
	if err != nil && err != x509utils.ErrIgnored {
		return err
	}
	return ctx.Err()
}

// AddCert adds a certificate to the store if it wasn't known
// already.
func (s *CertPool) AddCert(cert *x509.Certificate) bool {
	if s == nil || !validCert(cert) {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.unsafeAddCert("", cert)
	return err == nil
}

func (s *CertPool) unsafeAddCert(name string, cert *x509.Certificate) error {
	if err := s.unsafeInit(); err != nil {
		return err
	}

	cert, err := s.certs.Push(cert)
	switch {
	case err == set.ErrExist:
		return s.unsafeAddCertName(cert, name)
	case err != nil:
		return err
	}

	names, patterns := x509utils.Names(cert)
	if name != "" && !core.SliceContains(names, name) {
		names = append(names, name)
	}

	ce := &certPoolEntry{
		cert:     cert,
		names:    names,
		patterns: patterns,
	}

	s.unsafeAddCertEntry(ce)
	return nil
}

func (s *CertPool) unsafeAddCertName(cert *x509.Certificate, name string) error {
	ce := s.entries[cert]

	if name == "" || core.SliceContains(ce.names, name) {
		// nothing to add
		return core.ErrExists
	}

	ce.names = append(ce.names, name)
	s.unsafeInvalidateCache()
	appendMapList(s.names, name, ce)
	return nil
}

func (s *CertPool) unsafeAddCertEntry(ce *certPoolEntry) {
	s.unsafeInvalidateCache()
	s.entries[ce.cert] = ce
	for _, name := range ce.names {
		appendMapList(s.names, name, ce)
	}
	for _, pattern := range ce.patterns {
		appendMapList(s.patterns, pattern, ce)
	}
}
