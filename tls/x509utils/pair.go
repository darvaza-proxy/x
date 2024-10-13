package x509utils

import (
	"crypto"
	"crypto/x509"
)

// PrivateKeyEqual tells if two private keys are the same.
// nil keys aren't considered comparable.
func PrivateKeyEqual(a, b crypto.PrivateKey) bool {
	if a2, ok := a.(PrivateKey); ok {
		return a2.Equal(b)
	}
	return false
}

// PublicKeyEqual tells if two public keys are the same.
// nil keys aren't considered comparable.
func PublicKeyEqual(a, b crypto.PublicKey) bool {
	if a2, ok := a.(PublicKey); ok {
		return a2.Equal(b)
	}
	return false
}

// ValidCertKeyPair confirms the given key can use the given certificate.
// nil keys aren't considered comparable.
func ValidCertKeyPair(cert *x509.Certificate, key crypto.PrivateKey) bool {
	if cert == nil || key == nil {
		return false
	}

	if pk, ok := key.(PrivateKey); ok {
		return PublicKeyEqual(cert.PublicKey, pk.Public())
	}

	return false
}

// ValidKeyPair confirms the public key matches the private one.
// nil keys aren't considered comparable.
func ValidKeyPair(pub crypto.PublicKey, key crypto.PrivateKey) bool {
	if pub == nil || key == nil {
		return false
	}
	if priv, ok := key.(PrivateKey); ok {
		return PublicKeyEqual(pub, priv.Public())
	}
	return false
}
