package basic

import (
	"context"
	"crypto"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls"
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

// AddCert ...
func (ss *Store) AddCert(ctx context.Context, leaf *x509.Certificate) error {
	if err := ss.checkAddCert(ctx, leaf); err != nil {
		return err
	}

	key, _ := ss.GetPrivateKey(ctx, leaf.PublicKey)
	if key != nil {
		return ss.doAddCertPair(ctx, key, leaf, nil)
	}

	return core.ErrTODO
}

func (ss *Store) checkAddCert(ctx context.Context, leaf *x509.Certificate) error {
	if leaf == nil {
		return ErrNoCert
	}

	return ss.init(ctx)
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
