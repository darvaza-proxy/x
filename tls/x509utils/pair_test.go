package x509utils_test

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = certKeyPairTestCase{}
	_ core.TestCase = isSelfSignedTestCase{}
	_ core.TestCase = keyEqualTestCase{}
	_ core.TestCase = privateKeyEqualTestCase{}
	_ core.TestCase = validKeyPairTestCase{}
)

// isSelfSignedTestCase exercises IsSelfSigned: only a self-signed CA whose
// signature verifies returns true.
type isSelfSignedTestCase struct {
	cert *x509.Certificate
	name string
	want bool
}

func (tc isSelfSignedTestCase) Name() string { return tc.name }

func (tc isSelfSignedTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, x509utils.IsSelfSigned(tc.cert), "self-signed")
}

func newIsSelfSignedTestCase(name string, cert *x509.Certificate,
	want bool) isSelfSignedTestCase {
	return isSelfSignedTestCase{cert: cert, name: name, want: want}
}

func isSelfSignedTestCases(t *testing.T) []isSelfSignedTestCase {
	t.Helper()

	root := certSpec{cn: "Root", isCA: true}.build(t)
	sub := certSpec{cn: "Sub", isCA: true, parent: &root}.build(t)
	leaf := certSpec{cn: "leaf.example.com",
		dns: core.S("leaf.example.com"), parent: &sub}.build(t)
	selfLeaf := certSpec{cn: "self.example.com"}.build(t)
	expiredRoot := certSpec{cn: "Expired Root", isCA: true,
		notBefore: time.Now().Add(-2 * time.Hour),
		notAfter:  time.Now().Add(-time.Hour)}.build(t)
	// self-issued but not self-signed (RFC 5280): a root re-key — same
	// subject and issuer name as root, signed by root's old key. The
	// premise assertion keeps the row honest: subject and issuer must
	// be byte-identical or the row stops reaching the signature check.
	rekeyedRoot := certSpec{cn: "Root", isCA: true, parent: &root}.build(t)
	core.AssertMustTrue(t, bytes.Equal(rekeyedRoot.cert.RawSubject,
		rekeyedRoot.cert.RawIssuer), "self-issued premise")

	return core.S(
		newIsSelfSignedTestCase("nil", nil, false),
		newIsSelfSignedTestCase("self-signed CA", root.cert, true),
		newIsSelfSignedTestCase("intermediate CA", sub.cert, false),
		newIsSelfSignedTestCase("end-entity leaf", leaf.cert, false),
		// self-signed but not a CA: subject==issuer yet IsCA is false.
		newIsSelfSignedTestCase("self-signed non-CA", selfLeaf.cert, false),
		// IsSelfSigned is a structural check, not a validity check: an expired
		// self-signed root is still recognised so the basic store classifies
		// it as a root rather than mis-filing it as an intermediate.
		newIsSelfSignedTestCase("expired self-signed CA", expiredRoot.cert,
			true),
		// subject and issuer bytes match and IsCA holds, so only the
		// CheckSignatureFrom clause rejects it — the row that keeps the
		// signature check load-bearing.
		newIsSelfSignedTestCase("self-issued not self-signed",
			rekeyedRoot.cert, false),
	)
}

func TestIsSelfSigned(t *testing.T) {
	core.RunTestCases(t, isSelfSignedTestCases(t))
}

// certKeyPairTestCase exercises ValidCertKeyPair: the key must own the cert's
// public key; nil operands are never comparable.
type certKeyPairTestCase struct {
	cert *x509.Certificate
	key  crypto.PrivateKey
	name string
	want bool
}

func (tc certKeyPairTestCase) Name() string { return tc.name }

func (tc certKeyPairTestCase) Test(t *testing.T) {
	t.Helper()
	got := x509utils.ValidCertKeyPair(tc.cert, tc.key)
	core.AssertEqual(t, tc.want, got, "valid pair")
}

func newCertKeyPairTestCase(name string, cert *x509.Certificate,
	key crypto.PrivateKey, want bool) certKeyPairTestCase {
	return certKeyPairTestCase{cert: cert, key: key, name: name, want: want}
}

func certKeyPairTestCases(t *testing.T) []certKeyPairTestCase {
	t.Helper()

	leaf := certSpec{cn: "pair.example.com",
		dns: core.S("pair.example.com")}.build(t)
	stray := mkKey(t)

	return core.S(
		newCertKeyPairTestCase("matching", leaf.cert, leaf.key, true),
		newCertKeyPairTestCase("mismatched key", leaf.cert, stray, false),
		newCertKeyPairTestCase("nil cert", nil, leaf.key, false),
		newCertKeyPairTestCase("nil key", leaf.cert, nil, false),
		// a value that is not an x509utils.PrivateKey never matches.
		newCertKeyPairTestCase("non-key value", leaf.cert, "not a key", false),
	)
}

func TestValidCertKeyPair(t *testing.T) {
	core.RunTestCases(t, certKeyPairTestCases(t))
}

// keyEqualTestCase exercises PublicKeyEqual: the first operand must satisfy the
// PublicKey interface and Equal the second.
type keyEqualTestCase struct {
	a    crypto.PublicKey
	b    crypto.PublicKey
	name string
	want bool
}

func (tc keyEqualTestCase) Name() string { return tc.name }

func (tc keyEqualTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, x509utils.PublicKeyEqual(tc.a, tc.b), "equal")
}

func newKeyEqualTestCase(name string, a, b crypto.PublicKey,
	want bool) keyEqualTestCase {
	return keyEqualTestCase{a: a, b: b, name: name, want: want}
}

func keyEqualTestCases(t *testing.T) []keyEqualTestCase {
	t.Helper()

	k1, k2 := mkKey(t), mkKey(t)
	pub1, pub2 := k1.Public(), k2.Public()

	return core.S(
		newKeyEqualTestCase("same key", pub1, pub1, true),
		newKeyEqualTestCase("different keys", pub1, pub2, false),
		newKeyEqualTestCase("nil first", nil, pub1, false),
		newKeyEqualTestCase("nil second", pub1, nil, false),
		// the first operand must implement PublicKey (Equal); a bare value
		// fails the type assertion and is treated as incomparable.
		newKeyEqualTestCase("non-key first", "not a key", pub1, false),
	)
}

func TestPublicKeyEqual(t *testing.T) {
	core.RunTestCases(t, keyEqualTestCases(t))
}

// privateKeyEqualTestCase exercises PrivateKeyEqual: the first operand must
// satisfy the PrivateKey interface and Equal the second.
type privateKeyEqualTestCase struct {
	a    crypto.PrivateKey
	b    crypto.PrivateKey
	name string
	want bool
}

func (tc privateKeyEqualTestCase) Name() string { return tc.name }

func (tc privateKeyEqualTestCase) Test(t *testing.T) {
	t.Helper()
	got := x509utils.PrivateKeyEqual(tc.a, tc.b)
	core.AssertEqual(t, tc.want, got, "equal")
}

func newPrivateKeyEqualTestCase(name string, a, b crypto.PrivateKey,
	want bool) privateKeyEqualTestCase {
	return privateKeyEqualTestCase{a: a, b: b, name: name, want: want}
}

func privateKeyEqualTestCases(t *testing.T) []privateKeyEqualTestCase {
	t.Helper()

	k1, k2 := mkKey(t), mkKey(t)

	return core.S(
		newPrivateKeyEqualTestCase("same key", k1, k1, true),
		newPrivateKeyEqualTestCase("different keys", k1, k2, false),
		newPrivateKeyEqualTestCase("nil first", nil, k1, false),
		newPrivateKeyEqualTestCase("nil second", k1, nil, false),
		// the first operand must implement PrivateKey (Equal); a bare value
		// fails the type assertion and is treated as incomparable.
		newPrivateKeyEqualTestCase("non-key first", "not a key", k1, false),
	)
}

func TestPrivateKeyEqual(t *testing.T) {
	core.RunTestCases(t, privateKeyEqualTestCases(t))
}

// validKeyPairTestCase exercises ValidKeyPair: the public key must match the
// private key that owns it; nil operands are never comparable.
type validKeyPairTestCase struct {
	pub  crypto.PublicKey
	key  crypto.PrivateKey
	name string
	want bool
}

func (tc validKeyPairTestCase) Name() string { return tc.name }

func (tc validKeyPairTestCase) Test(t *testing.T) {
	t.Helper()
	got := x509utils.ValidKeyPair(tc.pub, tc.key)
	core.AssertEqual(t, tc.want, got, "valid pair")
}

func newValidKeyPairTestCase(name string, pub crypto.PublicKey,
	key crypto.PrivateKey, want bool) validKeyPairTestCase {
	return validKeyPairTestCase{pub: pub, key: key, name: name, want: want}
}

func validKeyPairTestCases(t *testing.T) []validKeyPairTestCase {
	t.Helper()

	k1, k2 := mkKey(t), mkKey(t)

	return core.S(
		newValidKeyPairTestCase("matching", k1.Public(), k1, true),
		newValidKeyPairTestCase("mismatched", k2.Public(), k1, false),
		newValidKeyPairTestCase("nil pub", nil, k1, false),
		newValidKeyPairTestCase("nil key", k1.Public(), nil, false),
		// a value that does not implement PrivateKey is incomparable.
		newValidKeyPairTestCase("non-key", k1.Public(), "not a key", false),
	)
}

func TestValidKeyPair(t *testing.T) {
	core.RunTestCases(t, validKeyPairTestCases(t))
}
