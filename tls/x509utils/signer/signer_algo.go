package signer

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"

	"darvaza.org/x/tls/x509utils"
)

// GenerateRSAKey generates a new RSA 3072 Private key.
func GenerateRSAKey() (x509utils.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 3072)
}

// GenerateECDSAKey generates a new ECDSA 256 Private Key.
func GenerateECDSAKey() (x509utils.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// GenerateED25519Key generates a new ED25519 Private Key
func GenerateED25519Key() (x509utils.PrivateKey, error) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	return key, err
}

// NewRSA creates a new [Signer] using a freshly generated
// RSA3072 key.
func NewRSA() (*Signer, error) {
	key, err := GenerateRSAKey()
	if err != nil {
		return nil, err
	}

	return makeNew(key, nil)
}

// NewED25519 creates a new [Signer] using a freshly generated
// ED25519 key.
func NewED25519() (*Signer, error) {
	key, err := GenerateED25519Key()
	if err != nil {
		return nil, err
	}

	return makeNew(key, nil)
}

// NewECDSA creates a new [Signer] using a freshly generated
// ECDSA 256 key.
func NewECDSA() (*Signer, error) {
	key, err := GenerateECDSAKey()
	if err != nil {
		return nil, err
	}

	return makeNew(key, nil)
}
