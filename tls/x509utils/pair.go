package x509utils

import "crypto/x509"

// ValidCertKeyPair confirms the given key can use the given certificate.
func ValidCertKeyPair(cert *x509.Certificate, key PrivateKey) bool {
	if cert == nil || key == nil {
		// missing
		return false
	}
	pub, ok := cert.PublicKey.(PublicKey)
	if !ok {
		// invalid - unreachable
		return false
	}

	return pub.Equal(key.Public())
}
