package x509utils

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"io"

	"darvaza.org/core"
)

// WriteKey writes a PEM encoded private key
func WriteKey(w io.Writer, key PrivateKey) (int64, error) {
	var buf bytes.Buffer

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return 0, core.Wrap(err, "failed to encode key")
	}

	// pem.Encode to a bytes.Buffer cannot fail; only the writer can.
	core.MustNoError(pem.Encode(&buf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	}))

	return buf.WriteTo(w)
}

// WriteCert writes a PEM encoded certificate
func WriteCert(w io.Writer, cert *x509.Certificate) (int64, error) {
	var buf bytes.Buffer

	switch {
	case cert == nil:
		err := NewErrInvalidCert(nil, nil, "not provided")
		return 0, err
	case len(cert.Raw) == 0:
		err := NewErrInvalidCert(cert, nil, "missing Raw DER certificate")
		return 0, err
	}

	// cert.Raw is non-empty here and pem.Encode only fails when the writer
	// does; a bytes.Buffer never does, so a non-nil error is unreachable.
	core.MustNoError(pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}))

	return buf.WriteTo(w)
}
