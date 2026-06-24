package tls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/tls"
)

var _ core.TestCase = verifyRejectTestCase{}

// mkVerifyLeaf mints a self-signed certificate valid across the given window,
// formed so it can act as its own trust root (CA plus server auth), and returns
// it with the key that owns it.
func mkVerifyLeaf(t *testing.T, notBefore, notAfter time.Time) (
	*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "key")

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "verify-test"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	core.AssertMustNoError(t, err, "create")

	cert, err := x509.ParseCertificate(der)
	core.AssertMustNoError(t, err, "parse")
	return cert, key
}

// mkValidCert returns a tls.Certificate whose leaf, chain and key all agree, so
// Verify runs past the rejection arms into the structural checks.
func mkValidCert(t *testing.T) *tls.Certificate {
	t.Helper()

	now := time.Now().UTC()
	leaf, key := mkVerifyLeaf(t, now.Add(-time.Hour), now.Add(time.Hour))
	return &tls.Certificate{
		Certificate: [][]byte{leaf.Raw},
		PrivateKey:  key,
		Leaf:        leaf,
	}
}

// verifyRejectTestCase feeds a tls.Certificate that Verify must reject, checking
// the message and that it reports invalid input.
type verifyRejectTestCase struct {
	cert    *tls.Certificate
	name    string
	wantMsg string
}

func (tc verifyRejectTestCase) Name() string { return tc.name }

func (tc verifyRejectTestCase) Test(t *testing.T) {
	t.Helper()

	err := tls.Verify(tc.cert, nil)
	core.AssertMustError(t, err, "error")
	core.AssertEqual(t, tc.wantMsg, err.Error(), "message")
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid")
	core.AssertTrue(t, tls.IsInvalid(err), "IsInvalid")
}

func newVerifyRejectTestCase(name string, cert *tls.Certificate,
	wantMsg string) verifyRejectTestCase {
	return verifyRejectTestCase{cert: cert, name: name, wantMsg: wantMsg}
}

func verifyRejectTestCases(t *testing.T) []verifyRejectTestCase {
	t.Helper()

	now := time.Now().UTC()
	valid, _ := mkVerifyLeaf(t, now.Add(-time.Hour), now.Add(time.Hour))
	expired, _ := mkVerifyLeaf(t, now.Add(-2*time.Hour), now.Add(-time.Hour))
	future, _ := mkVerifyLeaf(t, now.Add(time.Hour), now.Add(2*time.Hour))

	wrongKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "wrong key")

	return core.S(
		newVerifyRejectTestCase("nil certificate", nil,
			"invalid certificate: none provided"),
		// arm 2, both halves: a nil leaf, and a leaf with an empty chain.
		newVerifyRejectTestCase("nil leaf",
			&tls.Certificate{Certificate: [][]byte{{0x01}}},
			"invalid certificate: missing leaf certificate"),
		newVerifyRejectTestCase("empty chain",
			&tls.Certificate{Leaf: valid},
			"invalid certificate: missing leaf certificate"),
		// arm 3, each guarded condition in turn.
		newVerifyRejectTestCase("empty leaf DER",
			&tls.Certificate{Leaf: &x509.Certificate{}, Certificate: [][]byte{{0x01}}},
			"invalid certificate: invalid leaf certificate"),
		newVerifyRejectTestCase("chain mismatch",
			&tls.Certificate{Leaf: valid, Certificate: [][]byte{{0x01}}},
			"invalid certificate: invalid leaf certificate"),
		newVerifyRejectTestCase("expired leaf",
			&tls.Certificate{Leaf: expired, Certificate: [][]byte{expired.Raw}},
			"invalid certificate: invalid leaf certificate"),
		newVerifyRejectTestCase("not yet valid leaf",
			&tls.Certificate{Leaf: future, Certificate: [][]byte{future.Raw}},
			"invalid certificate: invalid leaf certificate"),
		// arm 4 and arm 5: the key is missing, then present but mismatched.
		newVerifyRejectTestCase("missing private key",
			&tls.Certificate{Leaf: valid, Certificate: [][]byte{valid.Raw}},
			"invalid certificate: missing private key"),
		newVerifyRejectTestCase("wrong private key",
			&tls.Certificate{
				Leaf: valid, Certificate: [][]byte{valid.Raw}, PrivateKey: wrongKey,
			},
			"invalid certificate: invalid private key"),
	)
}

func TestVerify(t *testing.T) {
	core.RunTestCases(t, verifyRejectTestCases(t))
}

// TestVerifyNoRoots covers the default arm: a well-formed certificate with no
// roots passes the structural checks.
func TestVerifyNoRoots(t *testing.T) {
	err := tls.Verify(mkValidCert(t), nil)
	core.AssertNoError(t, err, "verify")
}

// TestVerifyWithRoots covers the roots arm: the chain is validated against a
// pool that trusts the self-signed leaf.
func TestVerifyWithRoots(t *testing.T) {
	cert := mkValidCert(t)

	roots := x509.NewCertPool()
	roots.AddCert(cert.Leaf)

	err := tls.Verify(cert, roots)
	core.AssertNoError(t, err, "verify")
}
