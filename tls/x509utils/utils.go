package x509utils

import (
	"crypto"
	"crypto/x509"
)

// PublicKeyFromPrivateKey attempts to extract the [PublicKey] from the given
// [crypto.PrivateKey]. Returns nil if the private key or the public key are
// nil or they don't implement the required interfaces.
func PublicKeyFromPrivateKey(key crypto.PrivateKey) PublicKey {
	if k, ok := key.(PrivateKey); ok {
		if pub, ok := k.Public().(PublicKey); ok {
			return pub
		}
	}

	return nil
}

// PublicKeyFromCertificate attempts to extract the [PublicKey] from the given
// [x509.Certificate]. Returns nil if the certificate or public key are nil or
// they don't implement the required interfaces.
func PublicKeyFromCertificate(cert *x509.Certificate) PublicKey {
	if cert == nil || cert.PublicKey == nil {
		return nil
	}

	if pub, ok := cert.PublicKey.(PublicKey); ok {
		return pub
	}

	return nil
}
