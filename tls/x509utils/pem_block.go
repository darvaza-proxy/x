package x509utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
)

var (
	// ErrIgnored is used when we ask the user to try a different function instead
	ErrIgnored = errors.New("type of value out of scope")

	// ErrNotSupported indicates the type of [PrivateKey] or [PublicKey] isn't
	// supported.
	ErrNotSupported = errors.New("key type not supported")
)

// BlockToPrivateKey parses a pem Block looking for rsa, ecdsa or ed25519 Private Keys
func BlockToPrivateKey(block *pem.Block) (PrivateKey, error) {
	if block.Type == "PRIVATE KEY" || strings.HasSuffix(block.Type, " PRIVATE KEY") {
		if pk, _ := x509.ParsePKCS1PrivateKey(block.Bytes); pk != nil {
			// *rsa.PrivateKey
			return pk, nil
		}

		pk, err := parsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		return pk, nil
	}

	return nil, ErrIgnored
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
	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		return cert, nil
	}
	return nil, ErrIgnored
}
