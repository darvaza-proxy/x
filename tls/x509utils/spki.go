package x509utils

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"

	"darvaza.org/core"
)

// SubjectPublicKeySHA1 returns the SHA1 hash of the SubjectPublicKey
// of a [crypto.PublicKey]
func SubjectPublicKeySHA1(pub crypto.PublicKey) (hash [sha1.Size]byte, err error) {
	b, err := SubjectPublicKeyBytes(pub)
	if err != nil {
		return hash, err
	}

	return sha1.Sum(b), nil
}

// SubjectPublicKeySHA224 returns the SHA224 hash of the SubjectPublicKey
// of a [crypto.PublicKey]
func SubjectPublicKeySHA224(pub crypto.PublicKey) (hash [sha256.Size224]byte, err error) {
	b, err := SubjectPublicKeyBytes(pub)
	if err != nil {
		return hash, err
	}

	return sha256.Sum224(b), nil
}

// SubjectPublicKeySHA256 returns the SHA256 hash of the SubjectPublicKey
// of a [crypto.PublicKey]
func SubjectPublicKeySHA256(pub crypto.PublicKey) (hash [sha256.Size]byte, err error) {
	b, err := SubjectPublicKeyBytes(pub)
	if err != nil {
		return hash, err
	}

	return sha256.Sum256(b), nil
}

// SubjectPublicKeyBytes extracts the SubjectPublicKey bytes
// from a [crypto.PublicKey]
func SubjectPublicKeyBytes(pub crypto.PublicKey) ([]byte, error) {
	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		err = core.Wrap(err, "failed to encode public key")
		return nil, err
	}

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}

	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		err = core.Wrap(err, "failed to decode public key")
		return nil, err
	}

	return spki.SubjectPublicKey.Bytes, nil
}
