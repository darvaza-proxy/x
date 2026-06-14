package certpool_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// testCert bundles a parsed certificate with the key that owns it, so it can
// sign children.
type testCert struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func baseTemplate(t *testing.T, cn string, dns []string) *x509.Certificate {
	t.Helper()
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	core.AssertMustNoError(t, err, "serial")

	return &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     dns,
	}
}

// signAndParse generates a key for tmpl and signs it with parent (self-signed
// when parent is nil), returning the parsed certificate and its key.
func signAndParse(t *testing.T, tmpl *x509.Certificate, parent *testCert) testCert {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "generate key")

	signerCert, signerKey := tmpl, crypto.Signer(key)
	if parent != nil {
		signerCert, signerKey = parent.cert, parent.key
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert,
		key.Public(), signerKey)
	core.AssertMustNoError(t, err, "create certificate")

	cert, err := x509.ParseCertificate(der)
	core.AssertMustNoError(t, err, "parse certificate")
	return testCert{cert: cert, key: key}
}

// mkTestCA builds a CA certificate signed by parent, or self-signed when parent
// is nil.
func mkTestCA(t *testing.T, cn string, parent *testCert) testCert {
	t.Helper()
	tmpl := baseTemplate(t, cn, nil)
	tmpl.IsCA = true
	tmpl.BasicConstraintsValid = true
	tmpl.KeyUsage |= x509.KeyUsageCertSign
	return signAndParse(t, tmpl, parent)
}

// mkTestLeaf builds a non-CA leaf certificate signed by parent.
func mkTestLeaf(t *testing.T, cn string, parent *testCert) testCert {
	t.Helper()
	return signAndParse(t, baseTemplate(t, cn, []string{cn}), parent)
}

// mustHashIssuer returns the subject-hash to look up a child's candidate
// issuers.
func mustHashIssuer(t *testing.T, child *x509.Certificate) certpool.Hash {
	t.Helper()
	h, ok := certpool.HashIssuer(child)
	core.AssertMustTrue(t, ok, "hash issuer")
	return h
}

// TestGetBySubjectHash confirms a candidate issuer is found by the child's
// issuer hash (and a cert by its own subject hash), and that the lookup is
// nil-safe and reports a miss for an empty pool.
func TestGetBySubjectHash(t *testing.T) {
	root := mkTestCA(t, "Lookup Root", nil)
	sub := mkTestCA(t, "Lookup Sub", &root)
	leaf := mkTestLeaf(t, "leaf.example.com", &sub)

	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(sub.cert), "add sub")

	// base primitive: by the child's issuer hash → its candidate issuer
	got := pool.GetBySubjectHash(mustHashIssuer(t, leaf.cert))
	core.AssertMustEqual(t, 1, len(got), "one candidate")
	core.AssertTrue(t, got[0].Equal(sub.cert), "candidate is sub")

	// GetByIssuer shortcut resolves to the same candidate
	core.AssertMustEqual(t, 1, len(pool.GetByIssuer(leaf.cert)), "by issuer")

	// GetBySubject shortcut finds the cert by its own Subject
	bySub := pool.GetBySubject(sub.cert)
	core.AssertMustEqual(t, 1, len(bySub), "by subject")
	core.AssertTrue(t, bySub[0].Equal(sub.cert), "subject match is sub")

	// misses / nil-safety
	core.AssertNil(t, certpool.New().GetByIssuer(leaf.cert), "empty pool")
	core.AssertNil(t, (*certpool.CertPool)(nil).GetByIssuer(leaf.cert), "nil pool")
	core.AssertNil(t, pool.GetBySubject(nil), "nil cert by subject")
	core.AssertNil(t, pool.GetByIssuer(nil), "nil cert by issuer")
}

// TestGetBySubjectHashReturnsTwins confirms the lookup is signature-agnostic:
// two certs sharing a Subject but holding different keys both land in the
// bucket and are both returned. Selecting the genuine signer via
// CheckSignatureFrom is the caller's job, not the index's.
func TestGetBySubjectHashReturnsTwins(t *testing.T) {
	root := mkTestCA(t, "Twin Root", nil)
	sub := mkTestCA(t, "Twin Sub", &root)
	twin := mkTestCA(t, "Twin Sub", &root) // same Subject, different key
	leaf := mkTestLeaf(t, "twin.example.com", &sub)

	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(sub.cert), "add sub")
	core.AssertMustTrue(t, pool.AddCert(twin.cert), "add twin")

	got := pool.GetByIssuer(leaf.cert)
	core.AssertMustEqual(t, 2, len(got), "both twins returned")

	var verified int
	for _, c := range got {
		if leaf.cert.CheckSignatureFrom(c) == nil {
			verified++
		}
	}
	core.AssertEqual(t, 1, verified, "exactly one genuine signer")
}

// TestGetBySubjectHashSurvivesClone confirms Clone rebuilds the subject index:
// the lookup works on the cloned pool even though bySubject is keyed by a
// derived hash and so is rebuilt from the cloned entries rather than copied.
func TestGetBySubjectHashSurvivesClone(t *testing.T) {
	root := mkTestCA(t, "Clone Root", nil)
	sub := mkTestCA(t, "Clone Sub", &root)
	leaf := mkTestLeaf(t, "clone.example.com", &sub)

	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(sub.cert), "add sub")

	clone, ok := pool.Clone().(*certpool.CertPool)
	core.AssertMustTrue(t, ok, "clone is *CertPool")

	got := clone.GetByIssuer(leaf.cert)
	core.AssertMustEqual(t, 1, len(got), "candidate found in clone")
	core.AssertTrue(t, got[0].Equal(sub.cert), "candidate is sub")
}

// TestGetBySubjectHashAfterDelete confirms the subject index is maintained on
// delete: once the cert is removed, the lookup misses.
func TestGetBySubjectHashAfterDelete(t *testing.T) {
	root := mkTestCA(t, "Delete Root", nil)
	sub := mkTestCA(t, "Delete Sub", &root)
	leaf := mkTestLeaf(t, "del.example.com", &sub)

	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(sub.cert), "add sub")
	core.AssertMustEqual(t, 1, len(pool.GetByIssuer(leaf.cert)), "present")

	core.AssertMustNoError(t, pool.DeleteCert(context.Background(), sub.cert), "delete")
	core.AssertNil(t, pool.GetByIssuer(leaf.cert), "gone")
}
