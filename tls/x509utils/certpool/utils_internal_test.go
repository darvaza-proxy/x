package certpool

import (
	"crypto/x509"
	"testing"

	"darvaza.org/core"
)

// validCertTestCase exercises validCert: a certificate is valid only when it
// carries DER bytes, a public key, and both a non-empty subject and issuer
// (RFC 5280 makes issuer a mandatory field).
type validCertTestCase struct {
	cert *x509.Certificate
	name string
	want bool
}

// Name returns the test case name for identification.
func (tc validCertTestCase) Name() string { return tc.name }

// Test asserts validCert's verdict for the case's certificate.
func (tc validCertTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, validCert(tc.cert), "validCert")
}

var _ core.TestCase = validCertTestCase{}

func newValidCertTestCase(name string, cert *x509.Certificate,
	want bool) validCertTestCase {
	return validCertTestCase{
		cert: cert,
		name: name,
		want: want,
	}
}

// fullCert returns a certificate satisfying every validCert requirement; each
// negative case blanks exactly one field.
func fullCert() *x509.Certificate {
	return &x509.Certificate{
		Raw:        []byte{0x01},
		RawSubject: []byte{0x02},
		RawIssuer:  []byte{0x03},
		PublicKey:  struct{}{},
	}
}

func withoutPublicKey() *x509.Certificate {
	c := fullCert()
	c.PublicKey = nil
	return c
}

func withoutRaw() *x509.Certificate {
	c := fullCert()
	c.Raw = nil
	return c
}

func withoutSubject() *x509.Certificate {
	c := fullCert()
	c.RawSubject = nil
	return c
}

func withoutIssuer() *x509.Certificate {
	c := fullCert()
	c.RawIssuer = nil
	return c
}

func validCertTestCases() []validCertTestCase {
	return []validCertTestCase{
		newValidCertTestCase("nil cert", nil, false),
		newValidCertTestCase("no public key", withoutPublicKey(), false),
		newValidCertTestCase("no DER bytes", withoutRaw(), false),
		newValidCertTestCase("no subject", withoutSubject(), false),
		newValidCertTestCase("no issuer", withoutIssuer(), false),
		newValidCertTestCase("complete", fullCert(), true),
	}
}

func TestValidCert(t *testing.T) {
	core.RunTestCases(t, validCertTestCases())
}
