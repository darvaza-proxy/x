package x509utils_test

import (
	"crypto"
	"crypto/x509"
	"io"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = publicKeyFromCertificateTestCase{}
	_ core.TestCase = publicKeyFromPrivateKeyTestCase{}
)

// notPublicKey is a value without an Equal(crypto.PublicKey) bool method, so it
// fails the x509utils.PublicKey assertion. It stands in for a public key the
// helpers must reject.
type notPublicKey struct{}

// fakeSigner satisfies x509utils.PrivateKey (crypto.Signer plus Equal) yet its
// Public returns a notPublicKey, driving the arm where the private key is
// well-typed but its public half is not an x509utils.PublicKey.
type fakeSigner struct{}

func (fakeSigner) Public() crypto.PublicKey { return notPublicKey{} }

func (fakeSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, nil
}

func (fakeSigner) Equal(crypto.PrivateKey) bool { return false }

// publicKeyFromPrivateKeyTestCase exercises PublicKeyFromPrivateKey: a
// well-typed key yields its public half; every other operand yields nil.
type publicKeyFromPrivateKeyTestCase struct {
	key  crypto.PrivateKey
	want crypto.PublicKey
	name string
}

func (tc publicKeyFromPrivateKeyTestCase) Name() string { return tc.name }

func (tc publicKeyFromPrivateKeyTestCase) Test(t *testing.T) {
	got := x509utils.PublicKeyFromPrivateKey(tc.key)
	if tc.want == nil {
		core.AssertNil(t, got, "public key")
		return
	}
	core.AssertMustNotNil(t, got, "public key")
	core.AssertTrue(t, got.Equal(tc.want), "public key")
}

func newPublicKeyFromPrivateKeyTestCase(name string, key crypto.PrivateKey,
	want crypto.PublicKey) publicKeyFromPrivateKeyTestCase {
	return publicKeyFromPrivateKeyTestCase{
		key:  key,
		want: mustNotZero(want, "publicKeyFromPrivateKeyTestCase: every accepted row must state the key it gives"),
		name: name,
	}
}

func newPublicKeyFromPrivateKeyTestCaseRefused(name string,
	key crypto.PrivateKey) publicKeyFromPrivateKeyTestCase {
	return publicKeyFromPrivateKeyTestCase{key: key, want: nil, name: name}
}

func publicKeyFromPrivateKeyTestCases(
	t *testing.T,
) []publicKeyFromPrivateKeyTestCase {
	key := mkKey(t)

	return core.S(
		newPublicKeyFromPrivateKeyTestCase("ECDSA key", key, key.Public()),
		newPublicKeyFromPrivateKeyTestCaseRefused("nil", nil),
		// a value that is not an x509utils.PrivateKey fails the first assertion.
		newPublicKeyFromPrivateKeyTestCaseRefused("non-key value", "not a key"),
		// a well-typed key whose Public is not an x509utils.PublicKey fails the
		// second assertion.
		newPublicKeyFromPrivateKeyTestCaseRefused("public not a key",
			fakeSigner{}),
	)
}

func TestPublicKeyFromPrivateKey(t *testing.T) {
	core.RunTestCases(t, publicKeyFromPrivateKeyTestCases(t))
}

// publicKeyFromCertificateTestCase exercises PublicKeyFromCertificate: a
// certificate carrying a well-typed public key yields it; a nil certificate,
// a nil public key, or a public key of the wrong shape yields nil.
type publicKeyFromCertificateTestCase struct {
	cert *x509.Certificate
	want crypto.PublicKey
	name string
}

func (tc publicKeyFromCertificateTestCase) Name() string { return tc.name }

func (tc publicKeyFromCertificateTestCase) Test(t *testing.T) {
	got := x509utils.PublicKeyFromCertificate(tc.cert)
	if tc.want == nil {
		core.AssertNil(t, got, "public key")
		return
	}
	core.AssertMustNotNil(t, got, "public key")
	core.AssertTrue(t, got.Equal(tc.want), "public key")
}

func newPublicKeyFromCertificateTestCase(name string, cert *x509.Certificate,
	want crypto.PublicKey) publicKeyFromCertificateTestCase {
	return publicKeyFromCertificateTestCase{
		cert: cert,
		want: mustNotZero(want, "publicKeyFromCertificateTestCase: every accepted row must state the key it gives"),
		name: name,
	}
}

func newPublicKeyFromCertificateTestCaseRefused(name string,
	cert *x509.Certificate) publicKeyFromCertificateTestCase {
	return publicKeyFromCertificateTestCase{cert: cert, want: nil, name: name}
}

func publicKeyFromCertificateTestCases(
	t *testing.T,
) []publicKeyFromCertificateTestCase {
	leaf := certSpec{cn: "utils.example.com",
		dns: core.S("utils.example.com")}.build(t)

	return core.S(
		newPublicKeyFromCertificateTestCase("ECDSA cert", leaf.cert,
			leaf.cert.PublicKey),
		newPublicKeyFromCertificateTestCaseRefused("nil cert", nil),
		newPublicKeyFromCertificateTestCaseRefused("nil public key",
			&x509.Certificate{}),
		// a public key without Equal fails the assertion.
		newPublicKeyFromCertificateTestCaseRefused("public not a key",
			&x509.Certificate{PublicKey: notPublicKey{}}),
	)
}

func TestPublicKeyFromCertificate(t *testing.T) {
	core.RunTestCases(t, publicKeyFromCertificateTestCases(t))
}
