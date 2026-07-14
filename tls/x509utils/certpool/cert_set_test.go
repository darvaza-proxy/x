package certpool_test

import (
	"crypto"
	"io"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// badSigner satisfies x509utils.PrivateKey but its Public returns a value
// without an Equal method, so GetByPrivateKey's non-PublicKey arm is reachable.
type badSigner struct{}

func (badSigner) Public() crypto.PublicKey { return struct{}{} }

func (badSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, nil
}

func (badSigner) Equal(crypto.PrivateKey) bool { return false }

// mkCertSet builds a CertSet holding two independently keyed leaves, returning
// the set and both leaves so a lookup can be checked against a known member and
// against a second, distinctly keyed one.
func mkCertSet(t *testing.T) (cs *certpool.CertSet, one, two testCert) {
	t.Helper()
	one = mkTestLeaf(t, "set-one.example.com", nil)
	two = mkTestLeaf(t, "set-two.example.com", nil)
	cs = certpool.MustCertSet(one.cert, two.cert)
	return cs, one, two
}

// TestCertSetGetByKey confirms GetByKey returns the certificate owning a public
// key and rejects a nil, non-key or nil-receiver lookup.
func TestCertSetGetByKey(t *testing.T) {
	cs, one, two := mkCertSet(t)

	got := cs.GetByKey(one.cert.PublicKey)
	core.AssertMustEqual(t, 1, len(got), "one match")
	core.AssertTrue(t, got[0].Equal(one.cert), "match is the owner")

	other := cs.GetByKey(two.cert.PublicKey)
	core.AssertMustEqual(t, 1, len(other), "one match")
	core.AssertTrue(t, other[0].Equal(two.cert), "second key finds its own cert")

	core.AssertNil(t, cs.GetByKey(nil), "nil key")
	core.AssertNil(t, cs.GetByKey("not a key"), "non-key value")
	core.AssertNil(t, (*certpool.CertSet)(nil).GetByKey(one.cert.PublicKey),
		"nil receiver")
}

// TestCertSetGetByPrivateKey confirms GetByPrivateKey resolves a certificate
// through its private key and rejects a nil, non-key, wrong-public-type or
// nil-receiver lookup.
func TestCertSetGetByPrivateKey(t *testing.T) {
	cs, one, two := mkCertSet(t)

	got := cs.GetByPrivateKey(one.key)
	core.AssertMustEqual(t, 1, len(got), "one match")
	core.AssertTrue(t, got[0].Equal(one.cert), "match is the owner")

	other := cs.GetByPrivateKey(two.key)
	core.AssertMustEqual(t, 1, len(other), "one match")
	core.AssertTrue(t, other[0].Equal(two.cert), "second key finds its own cert")

	core.AssertNil(t, cs.GetByPrivateKey(nil), "nil key")
	core.AssertNil(t, cs.GetByPrivateKey("not a key"), "non-key value")
	core.AssertNil(t, cs.GetByPrivateKey(badSigner{}), "public not a key")
	core.AssertNil(t, (*certpool.CertSet)(nil).GetByPrivateKey(one.key),
		"nil receiver")
}

// assertSetIndependent confirms mutating dup — a Copy or Clone of src — leaves
// src untouched, proving the duplicate does not share storage with its source.
// Equal length would pass even for a shared backing store; the mutation is the
// load-bearing proof.
func assertSetIndependent(t *testing.T, src, dup *certpool.CertSet, name string) {
	t.Helper()
	srcLen := src.Len()
	extra := mkTestLeaf(t, name+"-extra.example.com", nil).cert
	_, err := dup.Push(extra)
	core.AssertMustNoError(t, err, "%s: push extra", name)
	core.AssertEqual(t, srcLen+1, dup.Len(), "%s grew", name)
	core.AssertEqual(t, srcLen, src.Len(), "%s source unchanged", name)
}

// TestCertSetCopyAndClone confirms Copy fills a destination (creating one when
// absent), Clone duplicates, both produce a set independent of the source, and
// both handle a nil receiver.
func TestCertSetCopyAndClone(t *testing.T) {
	cs, _, _ := mkCertSet(t)
	before := cs.Len()

	fresh := cs.Copy(nil, nil)
	core.AssertMustNotNil(t, fresh, "copy into fresh")
	core.AssertNotSame(t, cs, fresh, "copy is a distinct set")
	core.AssertEqual(t, before, fresh.Len(), "copied all")
	assertSetIndependent(t, cs, fresh, "copy")

	clone := cs.Clone()
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertNotSame(t, cs, clone, "clone is a distinct set")
	core.AssertEqual(t, before, clone.Len(), "cloned all")
	assertSetIndependent(t, cs, clone, "clone")

	core.AssertNil(t, (*certpool.CertSet)(nil).Clone(), "nil clone")

	created := (*certpool.CertSet)(nil).Copy(nil, nil)
	core.AssertMustNotNil(t, created, "nil receiver, nil dst creates a set")
	core.AssertEqual(t, 0, created.Len(), "created set is empty")

	dst := certpool.MustCertSet()
	same := (*certpool.CertSet)(nil).Copy(dst, nil)
	core.AssertSame(t, dst, same, "nil receiver returns destination")
}
