package buffer_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// certSpec describes a certificate to generate for tests. A nil parent
// produces a self-signed certificate; otherwise the parent signs the leaf.
type certSpec struct {
	parent *tls.Certificate
	cn     string
	dns    []string
	isCA   bool
}

// build generates the certificate described by the spec, returning a
// [tls.Certificate] with Leaf, Certificate chain and PrivateKey populated.
func (spec certSpec) build(t *testing.T) *tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	core.AssertMustNoError(t, err, "generate key")

	tmpl := spec.template(t)
	parentLeaf, parentKey := spec.parentFor(t, tmpl, key)

	der, err := x509.CreateCertificate(rand.Reader, tmpl, parentLeaf,
		key.Public(), parentKey)
	core.AssertMustNoError(t, err, "create certificate")

	leaf, err := x509.ParseCertificate(der)
	core.AssertMustNoError(t, err, "parse certificate")

	chain := [][]byte{der}
	if spec.parent != nil {
		chain = append(chain, spec.parent.Certificate...)
	}

	return &tls.Certificate{
		Certificate: chain,
		PrivateKey:  key,
		Leaf:        leaf,
	}
}

// template builds the x509 template, filling sensible validity defaults.
func (spec certSpec) template(t *testing.T) *x509.Certificate {
	t.Helper()

	notBefore := time.Now().Add(-time.Hour)
	notAfter := time.Now().Add(time.Hour)

	tmpl := &x509.Certificate{
		SerialNumber: mkSerial(t),
		Subject:      pkix.Name{CommonName: spec.cn},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     spec.dns,
	}

	if spec.isCA {
		tmpl.IsCA = true
		tmpl.BasicConstraintsValid = true
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	}

	return tmpl
}

// parentFor resolves the signing certificate and key: the leaf itself for a
// self-signed spec, or the configured parent otherwise.
func (spec certSpec) parentFor(t *testing.T, leafTmpl *x509.Certificate,
	leafKey crypto.Signer) (*x509.Certificate, crypto.Signer) {
	t.Helper()
	if spec.parent == nil {
		return leafTmpl, leafKey
	}
	signer := core.AssertMustTypeIs[crypto.Signer](t,
		spec.parent.PrivateKey, "parent signer")
	return spec.parent.Leaf, signer
}

func mkSerial(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	core.AssertMustNoError(t, err, "serial")
	return n
}

// mkCA returns a self-signed CA certificate.
func mkCA(t *testing.T, cn string) *tls.Certificate {
	t.Helper()
	return certSpec{cn: cn, isCA: true}.build(t)
}

// mkSubCA returns an intermediate CA certificate signed by parent.
func mkSubCA(t *testing.T, parent *tls.Certificate, cn string) *tls.Certificate {
	t.Helper()
	return certSpec{parent: parent, cn: cn, isCA: true}.build(t)
}

// mkLeaf returns a non-CA leaf certificate signed by parent, with the given
// DNS SANs.
func mkLeaf(t *testing.T, parent *tls.Certificate, cn string,
	dns ...string) *tls.Certificate {
	t.Helper()
	return certSpec{parent: parent, cn: cn, dns: dns}.build(t)
}

// mkSelfSignedLeaf returns a self-signed non-CA leaf certificate.
func mkSelfSignedLeaf(t *testing.T, cn string) *tls.Certificate {
	t.Helper()
	return certSpec{cn: cn}.build(t)
}

// keyOf returns the private key as an [x509utils.PrivateKey].
func keyOf(t *testing.T, c *tls.Certificate) x509utils.PrivateKey {
	t.Helper()
	return core.AssertMustTypeIs[x509utils.PrivateKey](t, c.PrivateKey, "key")
}

// certBlock returns the leaf as a PEM CERTIFICATE block.
func certBlock(c *tls.Certificate) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE", Bytes: c.Certificate[0]}
}

// keyBlock returns the private key as a PEM PRIVATE KEY (PKCS#8) block.
func keyBlock(t *testing.T, c *tls.Certificate) *pem.Block {
	t.Helper()
	body, err := x509.MarshalPKCS8PrivateKey(c.PrivateKey)
	core.AssertMustNoError(t, err, "marshal key")
	return &pem.Block{Type: "PRIVATE KEY", Bytes: body}
}

// pairRec records one AddCertPair call.
type pairRec struct {
	key   crypto.Signer
	cert  *x509.Certificate
	inter []*x509.Certificate
}

// fakeStore is a recording [tls.StoreX509Writer]. Each add is appended to the
// matching slice; a repeated add of the same subject is the [core.ErrExists]
// no-op real stores return, so duplicate handling and idempotence are
// exercised. A non-nil err* field forces that method to fail.
type fakeStore struct {
	seen map[string]bool

	errAddCACerts    error
	errAddPrivateKey error
	errAddCert       error
	errAddCertPair   error

	caCerts []*x509.Certificate
	keys    []crypto.Signer
	certs   []*x509.Certificate
	pairs   []pairRec
}

func newFakeStore() *fakeStore {
	return &fakeStore{seen: make(map[string]bool)}
}

// once reports the first sighting of key, recording it; later sightings yield
// false so the caller can return the ErrExists no-op.
func (s *fakeStore) once(key string) bool {
	if s.seen[key] {
		return false
	}
	s.seen[key] = true
	return true
}

func (*fakeStore) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return nil, nil
}

func (*fakeStore) GetCAPool() *x509.CertPool { return x509.NewCertPool() }

func (s *fakeStore) AddCACerts(_ context.Context, roots ...*x509.Certificate) error {
	if s.errAddCACerts != nil {
		return s.errAddCACerts
	}
	for _, root := range roots {
		if !s.once("ca:" + string(root.Raw)) {
			return core.ErrExists
		}
		s.caCerts = append(s.caCerts, root)
	}
	return nil
}

func (s *fakeStore) AddPrivateKey(_ context.Context, key crypto.Signer) error {
	if s.errAddPrivateKey != nil {
		return s.errAddPrivateKey
	}
	der, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return err
	}
	if !s.once("k:" + string(der)) {
		return core.ErrExists
	}
	s.keys = append(s.keys, key)
	return nil
}

func (s *fakeStore) AddCert(_ context.Context, cert *x509.Certificate) error {
	if s.errAddCert != nil {
		return s.errAddCert
	}
	if !s.once("c:" + string(cert.Raw)) {
		return core.ErrExists
	}
	s.certs = append(s.certs, cert)
	return nil
}

func (s *fakeStore) AddCertPair(_ context.Context, key crypto.Signer,
	cert *x509.Certificate, intermediates []*x509.Certificate) error {
	if s.errAddCertPair != nil {
		return s.errAddCertPair
	}
	if !s.once("p:" + string(cert.Raw)) {
		return core.ErrExists
	}
	s.pairs = append(s.pairs, pairRec{key: key, cert: cert, inter: intermediates})
	return nil
}

func (s *fakeStore) DeleteCert(_ context.Context, cert *x509.Certificate) error {
	if !s.seen["c:"+string(cert.Raw)] {
		return core.ErrNotExists
	}
	delete(s.seen, "c:"+string(cert.Raw))
	return nil
}

// compile-time guard: fakeStore satisfies the interface the buffer writes to.
var _ tls.StoreX509Writer = (*fakeStore)(nil)
