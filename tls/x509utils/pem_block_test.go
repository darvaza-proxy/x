package x509utils_test

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var _ core.TestCase = blockToPrivateKeyTestCase{}

// mkRSAKey generates a fresh RSA key.
func mkRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	core.AssertMustNoError(t, err, "generate RSA key")
	return key
}

// mkEd25519Key generates a fresh Ed25519 key.
func mkEd25519Key(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	core.AssertMustNoError(t, err, "generate Ed25519 key")
	return key
}

// pkcs1Block wraps an RSA key as a PKCS1 "RSA PRIVATE KEY" block.
func pkcs1Block(t *testing.T, key *rsa.PrivateKey) *pem.Block {
	t.Helper()
	return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
}

// pkcs8Block wraps any key as a PKCS8 "PRIVATE KEY" block.
func pkcs8Block(t *testing.T, key x509utils.PrivateKey) *pem.Block {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	core.AssertMustNoError(t, err, "marshal PKCS8")
	return &pem.Block{Type: "PRIVATE KEY", Bytes: der}
}

// sec1Block wraps an ECDSA key as a SEC1 "EC PRIVATE KEY" block.
func sec1Block(t *testing.T, key *ecdsa.PrivateKey) *pem.Block {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	core.AssertMustNoError(t, err, "marshal SEC1")
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
}

// blockToPrivateKeyTestCase feeds a PEM block to BlockToPrivateKey and checks
// the returned error and, on success, that the key round-trips to the original.
type blockToPrivateKeyTestCase struct {
	want    x509utils.PrivateKey
	wantErr error
	block   *pem.Block
	name    string
}

func newBlockToPrivateKeyTestCase(name string, block *pem.Block,
	want x509utils.PrivateKey, wantErr error) blockToPrivateKeyTestCase {
	return blockToPrivateKeyTestCase{
		want:    want,
		wantErr: wantErr,
		block:   block,
		name:    name,
	}
}

func (tc blockToPrivateKeyTestCase) Name() string {
	return tc.name
}

func (tc blockToPrivateKeyTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := x509utils.BlockToPrivateKey(tc.block)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, "error")
		core.AssertNil(t, got, "key")
		return
	}

	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "key")
	core.AssertTrue(t, got.Equal(tc.want), "key round-trip")
}

func blockToPrivateKeyTestCases(t *testing.T) []blockToPrivateKeyTestCase {
	t.Helper()

	rsaKey := mkRSAKey(t)
	ecKey := mkKey(t)
	edKey := mkEd25519Key(t)

	return core.S(
		newBlockToPrivateKeyTestCase("RSA PKCS1",
			pkcs1Block(t, rsaKey), rsaKey, nil),
		newBlockToPrivateKeyTestCase("EC PKCS8",
			pkcs8Block(t, ecKey), ecKey, nil),
		newBlockToPrivateKeyTestCase("EC SEC1",
			sec1Block(t, ecKey), ecKey, nil),
		newBlockToPrivateKeyTestCase("Ed25519 PKCS8",
			pkcs8Block(t, edKey), edKey, nil),
		newBlockToPrivateKeyTestCase("certificate block",
			&pem.Block{Type: "CERTIFICATE"}, nil, x509utils.ErrIgnored),
	)
}

func TestBlockToPrivateKey(t *testing.T) {
	core.RunTestCases(t, blockToPrivateKeyTestCases(t))
}
