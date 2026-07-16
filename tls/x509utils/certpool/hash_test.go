package certpool_test

import (
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// TestHashProducers confirms the four hash producers digest a valid
// certificate and reject an invalid one, and that each keys on a distinct
// field: subject, issuer and DER all differ for a leaf signed by a CA.
func TestHashProducers(t *testing.T) {
	root := mkTestCA(t, "Hash Root", nil)
	leaf := mkTestLeaf(t, "hash.example.com", &root)
	invalid := &x509.Certificate{}

	cert, ok := certpool.HashCert(leaf.cert)
	core.AssertMustTrue(t, ok, "hash cert")
	subject, ok := certpool.HashSubject(leaf.cert)
	core.AssertMustTrue(t, ok, "hash subject")
	issuer, ok := certpool.HashIssuer(leaf.cert)
	core.AssertMustTrue(t, ok, "hash issuer")
	spki, ok := certpool.HashSubjectPublicKey(leaf.cert)
	core.AssertMustTrue(t, ok, "hash spki")

	core.AssertFalse(t, cert.Equal(subject), "cert vs subject differ")
	core.AssertFalse(t, subject.Equal(issuer), "subject vs issuer differ")
	core.AssertFalse(t, cert.Equal(spki), "cert vs spki differ")
	core.AssertEqual(t, subject, certpool.Sum(leaf.cert.RawSubject), "subject")

	assertHashMiss(t, invalid)
}

// assertHashMiss confirms every producer reports failure for a certificate that
// fails validCert.
func assertHashMiss(t *testing.T, cert *x509.Certificate) {
	t.Helper()
	_, ok := certpool.HashCert(cert)
	core.AssertFalse(t, ok, "hash cert miss")
	_, ok = certpool.HashSubject(cert)
	core.AssertFalse(t, ok, "hash subject miss")
	_, ok = certpool.HashIssuer(cert)
	core.AssertFalse(t, ok, "hash issuer miss")
	_, ok = certpool.HashSubjectPublicKey(cert)
	core.AssertFalse(t, ok, "hash spki miss")
}

// TestHashSubjectPublicKeyUnsupportedKey confirms a certificate that passes
// validCert but carries a public key that cannot be encoded is a miss:
// validCert only checks the field is non-nil, so the SubjectPublicKeyBytes
// error arm is reachable.
func TestHashSubjectPublicKeyUnsupportedKey(t *testing.T) {
	cert := &x509.Certificate{
		Raw:        []byte{0x01},
		RawSubject: []byte{0x02},
		RawIssuer:  []byte{0x03},
		PublicKey:  struct{}{},
	}
	_, ok := certpool.HashSubjectPublicKey(cert)
	core.AssertFalse(t, ok, "unsupported key miss")
}

// TestHashEqualAndZero confirms Hash.Equal, IsZero and EqualCert: a digest
// equals only itself, the zero value reports IsZero, and EqualCert matches the
// certificate that produced the hash but not another or an invalid one.
func TestHashEqualAndZero(t *testing.T) {
	leaf := mkTestLeaf(t, "eq.example.com", nil)
	other := mkTestLeaf(t, "other.example.com", nil)

	h, ok := certpool.HashCert(leaf.cert)
	core.AssertMustTrue(t, ok, "hash")

	core.AssertTrue(t, h.Equal(h), "equal self")
	core.AssertFalse(t, h.IsZero(), "non-zero")
	core.AssertTrue(t, certpool.Hash{}.IsZero(), "zero value")

	core.AssertTrue(t, h.EqualCert(leaf.cert), "matches own cert")
	core.AssertFalse(t, h.EqualCert(other.cert), "rejects other cert")
	core.AssertFalse(t, h.EqualCert(&x509.Certificate{}), "rejects invalid")
}
