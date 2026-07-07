package x509utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"darvaza.org/core"
)

var (
	// ErrEmpty indicates the file, string or bytes slice is empty.
	ErrEmpty = errors.New("empty")

	// ErrIgnored is used when we ask the user to try a different function instead.
	ErrIgnored = errors.New("type of value out of scope")

	// ErrNotSupported indicates the type of [PrivateKey] or [PublicKey] isn't
	// supported.
	ErrNotSupported = errors.New("key type not supported")
)

// BlockToPrivateKey parses a PEM block into an rsa, ecdsa or ed25519 private
// key, accepting the PKCS1, SEC1 and PKCS8 encodings.
func BlockToPrivateKey(block *pem.Block) (PrivateKey, error) {
	if block == nil {
		return nil, core.ErrInvalid
	}

	if block.Type == "PRIVATE KEY" || strings.HasSuffix(block.Type, " PRIVATE KEY") {
		return parsePrivateKeyDER(block.Bytes)
	}

	return nil, ErrIgnored
}

// parsePrivateKeyDER decodes a DER private key, trying the PKCS1, SEC1 and
// PKCS8 encodings in turn.
func parsePrivateKeyDER(b []byte) (PrivateKey, error) {
	if pk, _ := x509.ParsePKCS1PrivateKey(b); pk != nil {
		// *rsa.PrivateKey (PKCS1)
		return pk, nil
	}

	if pk, _ := x509.ParseECPrivateKey(b); pk != nil {
		// *ecdsa.PrivateKey (SEC1)
		return pk, nil
	}

	return parsePKCS8PrivateKey(b)
}

func parsePKCS8PrivateKey(b []byte) (PrivateKey, error) {
	pk, err := x509.ParsePKCS8PrivateKey(b)
	if err != nil {
		return nil, err
	}

	if key, ok := pk.(PrivateKey); ok {
		return key, nil
	}
	return nil, ErrNotSupported
}

// BlockToRSAPrivateKey attempts to parse a pem.Block to extract an rsa.PrivateKey
func BlockToRSAPrivateKey(block *pem.Block) (*rsa.PrivateKey, error) {
	pk, err := BlockToPrivateKey(block)
	if err != nil {
		return nil, err
	}

	if key, ok := pk.(*rsa.PrivateKey); ok {
		return key, nil
	}

	return nil, ErrIgnored
}

// BlockToCertificate attempts to parse a pem.Block to extract a x509.Certificate
func BlockToCertificate(block *pem.Block) (*x509.Certificate, error) {
	if block == nil {
		return nil, core.ErrInvalid
	}

	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		return cert, nil
	}
	return nil, ErrIgnored
}

// EncodeBytes produces a PEM encoded block
func EncodeBytes(label string, body []byte, headers map[string]string) []byte {
	var b bytes.Buffer
	_ = pem.Encode(&b, &pem.Block{
		Type:    label,
		Bytes:   body,
		Headers: headers,
	})
	return b.Bytes()
}

// EncodePKCS1PrivateKey produces a PEM encoded RSA Private Key
func EncodePKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	var out []byte
	if key != nil {
		body := x509.MarshalPKCS1PrivateKey(key)
		out = EncodeBytes("RSA PRIVATE KEY", body, nil)
	}
	return out
}

// EncodePKCS8PrivateKey produces a PEM encoded Private Key
func EncodePKCS8PrivateKey(key PrivateKey) ([]byte, error) {
	var out []byte
	if key != nil {
		body, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}
		out = EncodeBytes("PRIVATE KEY", body, nil)
	}
	return out, nil
}

// EncodeCertificate produces a PEM encoded x509 Certificate
// without optional headers
func EncodeCertificate(der []byte) []byte {
	return EncodeBytes("CERTIFICATE", der, nil)
}
