package certpool

import (
	"crypto/x509"

	"github.com/zeebo/blake3"

	"darvaza.org/x/tls/x509utils"
)

const (
	// HashSize is the number of bytes of HashCert's output
	HashSize = 32
)

// Hash is a blake3.Sum256 representation of a DER encoded certificate
type Hash [HashSize]byte

// Equal says if a hash is identical to this one.
func (hash Hash) Equal(other Hash) bool {
	return hash == other
}

// EqualCert says if the certificate matches the hash.
func (hash Hash) EqualCert(cert *x509.Certificate) bool {
	other, ok := HashCert(cert)
	if !ok {
		return false
	}

	return hash.Equal(other)
}

// IsZero tells if the hash is at its zero value.
func (hash Hash) IsZero() bool {
	return hash == Hash{}
}

// HashCert produces a blake3 digest of the DER representation of a Certificate
func HashCert(cert *x509.Certificate) (Hash, bool) {
	if cert == nil || len(cert.Raw) == 0 {
		return Hash{}, false
	}
	return Sum(cert.Raw), true
}

// HashSubject produces a blake3 digest of the raw subject of the Certificate
func HashSubject(cert *x509.Certificate) (Hash, bool) {
	if cert == nil || len(cert.RawSubject) == 0 {
		return Hash{}, false
	}
	return Sum(cert.RawSubject), true
}

// HashSubjectPublicKey produces a blake3 digest of the PublicKey of the Certificate
func HashSubjectPublicKey(cert *x509.Certificate) (Hash, bool) {
	if cert != nil && cert.PublicKey != nil {
		if b, err := x509utils.SubjectPublicKeyBytes(cert.PublicKey); err == nil {
			return Sum(b), true
		}
	}
	return Hash{}, false
}

// Sum is a shortcut to our preferred hash function, blake3.Sum256()
func Sum(data []byte) Hash {
	return blake3.Sum256(data)
}
