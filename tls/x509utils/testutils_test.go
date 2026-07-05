package x509utils_test

import (
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

// testCert holds a minted, parsed certificate.
type testCert struct {
	cert *x509.Certificate
}

// certSpec describes a self-signed certificate to mint, valid for a one-hour
// window around now.
type certSpec struct {
	cn  string
	dns []string
	ips []net.IP
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

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl,
		key.Public(), key)
	core.AssertMustNoError(t, err, "create certificate")

	cert, err := x509.ParseCertificate(der)
	core.AssertMustNoError(t, err, "parse certificate")
	return testCert{cert: cert}
}

// template builds the x509 template with a one-hour validity window
// around now.
func (spec certSpec) template(t *testing.T) *x509.Certificate {
	t.Helper()

	now := time.Now()
	return &x509.Certificate{
		SerialNumber: mkSerial(t),
		Subject:      pkix.Name{CommonName: spec.cn},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     spec.dns,
		IPAddresses:  spec.ips,
	}
}
