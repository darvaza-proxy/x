package tls

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// Verify checks if a [tls.Certificate] is good to use.
// If roots is provided, the chain will also be verified.
func Verify(cert *tls.Certificate, roots *x509.CertPool) error {
	switch {
	case cert == nil:
		return &x509utils.ErrInvalidCert{
			Reason: "none provided",
		}
	case cert.Leaf == nil || len(cert.Certificate) == 0:
		return &x509utils.ErrInvalidCert{
			Reason: "missing leaf certificate",
		}
	case len(cert.Leaf.Raw) == 0,
		cert.Leaf.NotAfter.Before(time.Now()),
		cert.Leaf.NotBefore.After(time.Now()),
		!bytes.Equal(cert.Leaf.Raw, cert.Certificate[0]):
		return &x509utils.ErrInvalidCert{
			Reason: "invalid leaf certificate",
		}
	case cert.PrivateKey == nil:
		return &x509utils.ErrInvalidCert{
			Reason: "missing private key",
		}
	case !x509utils.ValidCertKeyPair(cert.Leaf, cert.PrivateKey):
		return &x509utils.ErrInvalidCert{
			Reason: "invalid private key",
		}
	case roots != nil:
		return doVerifyRoots(cert, roots)
	default:
		return doVerifyNoRoots(cert)
	}
}

func doVerifyRoots(cert *tls.Certificate, roots *x509.CertPool) error {
	opt := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: x509.NewCertPool(),
	}
	for i, der := range cert.Certificate[1:] {
		c, err := doVerifyDERCert(i+1, der)
		if err != nil {
			return core.Wrap(err, "ReadCertificate")
		}

		opt.Intermediates.AddCert(c)
	}
	_, err := cert.Leaf.Verify(opt)
	return err
}

func doVerifyNoRoots(cert *tls.Certificate) error {
	for i, der := range cert.Certificate {
		_, err := doVerifyDERCert(i, der)
		if err != nil {
			return core.Wrap(err, "ReadCertificate")
		}
	}
	return nil
}

func doVerifyDERCert(slot int, der []byte) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, core.Wrapf(err, "bad certificate in slot %v", slot)
	}
	return cert, nil
}
