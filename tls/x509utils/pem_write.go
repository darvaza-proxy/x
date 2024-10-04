package x509utils

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"darvaza.org/core"
)

// WriteKey writes a PEM encoded private key
func WriteKey(w io.Writer, key PrivateKey) (int64, error) {
	var buf bytes.Buffer

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err == nil {
		err = pem.Encode(&buf, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyDER,
		})
	}

	if err != nil {
		err = core.Wrap(err, "failed to encode key")
		return 0, err
	}

	return buf.WriteTo(w)
}

// WriteCert writes a PEM encoded certificate
func WriteCert(w io.Writer, cert *x509.Certificate) (int64, error) {
	var buf bytes.Buffer

	if len(cert.Raw) == 0 {
		err := errors.New("missing Raw DER certificate")
		return 0, err
	}

	err := pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	if err != nil {
		err = core.Wrap(err, "failed to encode certificate")
		return 0, err
	}

	return buf.WriteTo(w)
}