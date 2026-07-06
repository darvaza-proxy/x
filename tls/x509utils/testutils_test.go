package x509utils_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"darvaza.org/core"
)

// testCert bundles a parsed certificate with the key that owns it, so it can
// sign children.
type testCert struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

// certSpec describes a certificate to mint. A nil parent yields a self-signed
// certificate; otherwise parent signs it. Zero validity fields fall back to a
// one-hour window around now.
type certSpec struct {
	notBefore time.Time
	notAfter  time.Time
	parent    *testCert
	cn        string
	dns       []string
	ips       []net.IP
	isCA      bool
}

// mkSerial returns a random 128-bit serial number.
func mkSerial(t *testing.T) *big.Int {
	t.Helper()
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	core.AssertMustNoError(t, err, "serial")
	return serial
}

// build mints and parses the certificate described by the spec.
func (spec certSpec) build(t *testing.T) testCert {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "generate key")

	tmpl := spec.template(t)
	signerCert, signerKey := tmpl, crypto.Signer(key)
	if spec.parent != nil {
		signerCert, signerKey = spec.parent.cert, spec.parent.key
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert,
		key.Public(), signerKey)
	core.AssertMustNoError(t, err, "create certificate")

	cert, err := x509.ParseCertificate(der)
	core.AssertMustNoError(t, err, "parse certificate")
	return testCert{cert: cert, key: key}
}

// template builds the x509 template, filling sensible validity defaults and
// the CA flags when requested.
func (spec certSpec) template(t *testing.T) *x509.Certificate {
	t.Helper()

	now := time.Now()
	notBefore := core.Coalesce(spec.notBefore, now.Add(-time.Hour))
	notAfter := core.Coalesce(spec.notAfter, now.Add(time.Hour))

	tmpl := &x509.Certificate{
		SerialNumber: mkSerial(t),
		Subject:      pkix.Name{CommonName: spec.cn},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     spec.dns,
		IPAddresses:  spec.ips,
	}
	if spec.isCA {
		tmpl.IsCA = true
		tmpl.BasicConstraintsValid = true
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	}
	return tmpl
}

// mkKey generates a fresh ECDSA key for equality tests.
func mkKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "generate key")
	return key
}
