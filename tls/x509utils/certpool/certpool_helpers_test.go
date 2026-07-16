package certpool_test

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// mkIPLeaf builds a self-signed leaf whose only SAN is an IP address, so it is
// reachable through the IP name key rather than a DNS name.
func mkIPLeaf(t *testing.T, ip string) testCert {
	t.Helper()
	parsed := net.ParseIP(ip)
	core.AssertMustNotNil(t, parsed, "parse ip")

	tmpl := baseTemplate(t, ip, nil)
	tmpl.IPAddresses = []net.IP{parsed}
	return signAndParse(t, tmpl, nil)
}

// mkPopulatedPool builds a pool holding a root CA and two leaves it signed,
// returning the pool and the certificates it contains in insertion order.
func mkPopulatedPool(t *testing.T) (*certpool.CertPool, []*x509.Certificate) {
	t.Helper()
	root := mkTestCA(t, "Pool Root", nil)
	one := mkTestLeaf(t, "one.example.com", &root)
	two := mkTestLeaf(t, "two.example.com", &root)

	pool := certpool.New()
	certs := core.S(root.cert, one.cert, two.cert)
	for _, c := range certs {
		core.AssertMustTrue(t, pool.AddCert(c), "add")
	}
	return pool, certs
}

// certPEM encodes certificates as concatenated PEM CERTIFICATE blocks.
func certPEM(t *testing.T, certs ...*x509.Certificate) []byte {
	t.Helper()
	var buf bytes.Buffer
	for _, c := range certs {
		err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
		core.AssertMustNoError(t, err, "encode pem")
	}
	return buf.Bytes()
}
