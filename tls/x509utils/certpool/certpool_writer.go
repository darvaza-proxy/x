package certpool

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

var (
	_ x509utils.CertPoolWriter = (*CertPool)(nil)
)

// Put inserts a certificate into the store, optionally including a reference name.
// The name will be appended to those included in the certificate.
func (s *CertPool) Put(ctx context.Context, name string, cert *x509.Certificate) error {
	sn, ok := x509utils.SanitizeName(name)
	if !ok {
		// invalid name
		return core.Wrapf(core.ErrInvalid, "invalid argument: %s: %q", "name", name)
	}

	hash, ok := HashCert(cert)
	if !ok {
		// invalid cert
		return core.Wrapf(core.ErrInvalid, "invalid argument: %s", "cert")
	}

	if err := ctx.Err(); err != nil {
		// cancelled
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.unsafeAddCert(hash, sn, cert) {
		return nil
	}
	return core.ErrExists
}

// Delete remove from the store all certificates associated to the given name
func (s *CertPool) Delete(ctx context.Context, name string) error {
	sn, ok := x509utils.SanitizeName(name)
	if !ok || s == nil {
		return core.ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// convert name to unique hashes
	hashes := core.SliceAsFn(func(ce *certPoolEntry) (Hash, bool) {
		return ce.Hash()
	}, s.getListForName(sn).Values())

	core.SliceUniquify(&hashes)

	if s.unsafeDeleteCerts(hashes...) > 0 {
		return nil
	}

	return core.Wrap(core.ErrNotExists, name)
}

// DeleteCert removes a certificate, by raw DER hash, from the store.
func (s *CertPool) DeleteCert(ctx context.Context, cert *x509.Certificate) error {
	hash, ok := HashCert(cert)
	if !ok || s == nil {
		return core.ErrInvalid
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.unsafeDeleteCerts(hash) > 0 {
		return nil
	}

	return core.ErrNotExists
}

func (s *CertPool) unsafeDeleteCerts(hashes ...Hash) int {
	var count int

	for _, hash := range hashes {
		if ce, ok := s.hashed[hash]; ok {
			s.unsafeDeleteCertEntry(ce)
			count++
		}
	}

	return count
}

func (s *CertPool) unsafeDeleteCertEntry(ce *certPoolEntry) {
	delete(s.hashed, ce.hash)

	eq := func(p *certPoolEntry) bool {
		return ce == p
	}

	deleteMapListMatchFn(s.names, ce.names, eq)
	deleteMapListMatchFn(s.patterns, ce.patterns, eq)
}

// Import certificates from another CertPool.
func (s *CertPool) Import(ctx context.Context, src x509utils.CertPool) (int, error) {
	if s == nil {
		return 0, ErrNilReceiver
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
		return 0, ErrNilReceiver
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
	if s != nil && cert != nil {
		hash, ok := HashCert(cert)
		if ok {
			s.mu.Lock()
			defer s.mu.Unlock()

			return s.unsafeAddCert(hash, "", cert)
		}
	}

	return false
}

func (s *CertPool) unsafeAddCert(hash Hash, name string, cert *x509.Certificate) bool {
	s.init()

	if ce, found := s.hashed[hash]; found && name != "" {
		return s.unsafeAddCertName(ce, name)
	}

	names, patterns := x509utils.Names(cert)
	if name != "" && !core.SliceContains(names, name) {
		names = append(names, name)
	}

	ce := &certPoolEntry{
		hash:     hash,
		cert:     cert,
		names:    names,
		patterns: patterns,
	}

	s.unsafeAddCertEntry(ce)
	return true
}

func (s *CertPool) unsafeAddCertName(ce *certPoolEntry, name string) bool {
	if name == "" || core.SliceContains(ce.names, name) {
		// nothing to add
		return false
	}

	ce.names = append(ce.names, name)
	appendMapList(s.names, name, ce)
	return true
}

func (s *CertPool) unsafeAddCertEntry(ce *certPoolEntry) {
	s.hashed[ce.hash] = ce
	for _, name := range ce.names {
		appendMapList(s.names, name, ce)
	}
	for _, pattern := range ce.patterns {
		appendMapList(s.patterns, pattern, ce)
	}
}
