package basic

import (
	"context"
	"crypto"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// Verify checks if the certificate is valid and ready to use
// considering the root certificates in the [Store].
//
// If no CACerts have been added, system's root certificates will be loaded.
func (ss *Store) Verify(cert *tls.Certificate) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if cert == nil || cert.Leaf == nil {
		return ErrNoCert
	}

	return tls.Verify(cert, ss.GetCAPool())
}

// Assemble attempts to fill gaps in the [tls.Certificate] in preparation
// for a call to [Store#Verify].
func (ss *Store) Assemble(cert *tls.Certificate) error {
	if ss == nil {
		return core.ErrNilReceiver
	} else if cert == nil || cert.Leaf == nil {
		return ErrNoCert
	}

	return core.ErrTODO
}

// AddCert adds a certificate to the [Store].
// Self signed certificates will automatically become trusted.
// CA certificates are added as intermediate.
// And for non-CA certificates added this way, the corresponding
// private key should have been added before.
func (ss *Store) AddCert(ctx context.Context, leaf *x509.Certificate) error {
	if err := ss.checkAddCert(ctx, leaf); err != nil {
		return err
	} else if x509utils.IsSelfSigned(leaf) {
		// trust self signed
		ss.doAddCertRoots(ctx, leaf)
	} else if leaf.IsCA {
		ss.doAddCertInter(ctx, leaf)
	}

	return ss.doAddCert(ctx, leaf)
}

func (ss *Store) checkAddCert(ctx context.Context, leaf *x509.Certificate) error {
	if leaf == nil {
		return ErrNoCert
	}

	return ss.init(ctx)
}

func (ss *Store) doAddCertRoots(ctx context.Context, leaf *x509.Certificate) {
	// RW
	ss.mu.Lock()
	if ss.roots == nil {
		ss.roots = certpool.New()
	}
	added := ss.roots.AddCert(leaf)
	ss.mu.Unlock()

	if added {
		ss.reportAddCACert(ctx, leaf)
	}
}

func (ss *Store) doAddCertInter(_ context.Context, leaf *x509.Certificate) {
	ss.inter.AddCert(leaf)
}

func (ss *Store) doAddCert(ctx context.Context, leaf *x509.Certificate) error {
	key, _ := ss.GetPrivateKey(ctx, leaf.PublicKey)
	switch {
	case key != nil:
		return ss.doAddCertPair(ctx, key, leaf, nil)
	case !leaf.IsCA:
		return &x509utils.ErrInvalidCert{
			Cert:   leaf,
			Reason: "private key not found",
		}
	default:
		return nil
	}
}

// AddCertPair ...
func (ss *Store) AddCertPair(ctx context.Context, //
	key crypto.Signer, leaf *x509.Certificate, inter []*x509.Certificate) error {
	//
	if err := ss.checkAddCert(ctx, leaf); err != nil {
		return err
	}

	return ss.doAddCertPair(ctx, key, leaf, inter)
}

func (*Store) doAddCertPair(_ context.Context, //
	_ crypto.Signer, _ *x509.Certificate, _ []*x509.Certificate) error {
	return core.ErrTODO
}
