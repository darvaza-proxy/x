package x509utils_test

import (
	"crypto"
	"crypto/ecdh"
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

var (
	_ core.TestCase = blockToCertificateTestCase{}
	_ core.TestCase = blockToPrivateKeyTestCase{}
	_ core.TestCase = blockToRSAPrivateKeyTestCase{}
)

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

// mkX25519Key generates a fresh X25519 key. It marshals to PKCS8 but is a
// key-agreement key, not a crypto.Signer, so it is not an x509utils.PrivateKey.
func mkX25519Key(t *testing.T) *ecdh.PrivateKey {
	t.Helper()
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	core.AssertMustNoError(t, err, "generate X25519 key")
	return key
}

// pkcs1Block wraps an RSA key as a PKCS1 "RSA PRIVATE KEY" block.
func pkcs1Block(t *testing.T, key *rsa.PrivateKey) *pem.Block {
	t.Helper()
	return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
}

// pkcs8Block wraps any PKCS8-marshallable key as a "PRIVATE KEY" block,
// including key types (such as X25519) that are not x509utils.PrivateKey.
func pkcs8Block(t *testing.T, key crypto.PrivateKey) *pem.Block {
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

func (tc blockToPrivateKeyTestCase) Name() string { return tc.name }

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
		// X25519 parses as PKCS8 but is not a crypto.Signer, so the
		// PrivateKey assertion fails: ErrNotSupported, not a defensive arm.
		newBlockToPrivateKeyTestCase("X25519 PKCS8 unsupported",
			pkcs8Block(t, mkX25519Key(t)), nil, x509utils.ErrNotSupported),
		newBlockToPrivateKeyTestCase("certificate block",
			&pem.Block{Type: "CERTIFICATE"}, nil, x509utils.ErrIgnored),
	)
}

func TestBlockToPrivateKey(t *testing.T) {
	core.RunTestCases(t, blockToPrivateKeyTestCases(t))
}

// TestBlockToPrivateKeyMalformed covers parsePKCS8PrivateKey's parse-error arm:
// a PRIVATE KEY block whose bytes are rejected by the PKCS1, SEC1 and PKCS8
// decoders in turn surfaces x509's parse error.
func TestBlockToPrivateKeyMalformed(t *testing.T) {
	got, err := x509utils.BlockToPrivateKey(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte("not der"),
	})
	core.AssertError(t, err, "parse error")
	core.AssertNil(t, got, "key")
}

// TestBlockNilGuards checks the public block parsers reject a nil *pem.Block
// with core.ErrInvalid instead of dereferencing it and panicking (F63).
func TestBlockNilGuards(t *testing.T) {
	t.Run("BlockToPrivateKey", func(t *testing.T) {
		_, err := x509utils.BlockToPrivateKey(nil)
		core.AssertErrorIs(t, err, core.ErrInvalid, "error")
	})
	t.Run("BlockToRSAPrivateKey", func(t *testing.T) {
		_, err := x509utils.BlockToRSAPrivateKey(nil)
		core.AssertErrorIs(t, err, core.ErrInvalid, "error")
	})
	t.Run("BlockToCertificate", func(t *testing.T) {
		_, err := x509utils.BlockToCertificate(nil)
		core.AssertErrorIs(t, err, core.ErrInvalid, "error")
	})
}

// blockToCertificateTestCase feeds a PEM block to BlockToCertificate and checks
// the returned error and, on success, that it parses back to the original cert.
type blockToCertificateTestCase struct {
	want    *x509.Certificate
	wantErr error
	block   *pem.Block
	name    string
}

func newBlockToCertificateTestCase(name string, block *pem.Block,
	want *x509.Certificate, wantErr error) blockToCertificateTestCase {
	return blockToCertificateTestCase{
		want:    want,
		wantErr: wantErr,
		block:   block,
		name:    name,
	}
}

func (tc blockToCertificateTestCase) Name() string { return tc.name }

func (tc blockToCertificateTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := x509utils.BlockToCertificate(tc.block)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, "error")
		core.AssertNil(t, got, "cert")
		return
	}

	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "cert")
	core.AssertTrue(t, got.Equal(tc.want), "round-trip")
}

func blockToCertificateTestCases(t *testing.T) []blockToCertificateTestCase {
	t.Helper()

	cert := certSpec{cn: "block.test"}.build(t).cert

	return core.S(
		newBlockToCertificateTestCase("certificate",
			&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}, cert, nil),
		newBlockToCertificateTestCase("wrong type",
			&pem.Block{Type: "PRIVATE KEY"}, nil, x509utils.ErrIgnored),
	)
}

func TestBlockToCertificate(t *testing.T) {
	core.RunTestCases(t, blockToCertificateTestCases(t))
}

// TestBlockToCertificateMalformed covers the parse-error arm: a CERTIFICATE
// block whose bytes are not valid DER surfaces x509's parse error.
func TestBlockToCertificateMalformed(t *testing.T) {
	got, err := x509utils.BlockToCertificate(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("not der"),
	})
	core.AssertError(t, err, "parse error")
	core.AssertNil(t, got, "cert")
}

// blockToRSAPrivateKeyTestCase feeds a PEM block to BlockToRSAPrivateKey and
// checks the type-assertion arms after a successful parse: a genuine RSA key
// round-trips, any other key type is ErrIgnored.
type blockToRSAPrivateKeyTestCase struct {
	want    *rsa.PrivateKey
	wantErr error
	block   *pem.Block
	name    string
}

func newBlockToRSAPrivateKeyTestCase(name string, block *pem.Block,
	want *rsa.PrivateKey, wantErr error) blockToRSAPrivateKeyTestCase {
	return blockToRSAPrivateKeyTestCase{
		want:    want,
		wantErr: wantErr,
		block:   block,
		name:    name,
	}
}

func (tc blockToRSAPrivateKeyTestCase) Name() string { return tc.name }

func (tc blockToRSAPrivateKeyTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := x509utils.BlockToRSAPrivateKey(tc.block)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, "error")
		core.AssertNil(t, got, "key")
		return
	}

	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "key")
	core.AssertTrue(t, got.Equal(tc.want), "round-trip")
}

func blockToRSAPrivateKeyTestCases(t *testing.T) []blockToRSAPrivateKeyTestCase {
	t.Helper()

	rsaKey := mkRSAKey(t)

	return core.S(
		newBlockToRSAPrivateKeyTestCase("RSA PKCS1",
			pkcs1Block(t, rsaKey), rsaKey, nil),
		newBlockToRSAPrivateKeyTestCase("EC ignored",
			pkcs8Block(t, mkKey(t)), nil, x509utils.ErrIgnored),
	)
}

func TestBlockToRSAPrivateKey(t *testing.T) {
	core.RunTestCases(t, blockToRSAPrivateKeyTestCases(t))
}

// TestEncodePKCS1PrivateKey covers the nil-key arm (empty output) and the
// round-trip: the PEM it produces parses back to the original RSA key.
func TestEncodePKCS1PrivateKey(t *testing.T) {
	core.AssertNil(t, x509utils.EncodePKCS1PrivateKey(nil), "nil key")

	key := mkRSAKey(t)
	block, _ := pem.Decode(x509utils.EncodePKCS1PrivateKey(key))
	core.AssertMustNotNil(t, block, "block")
	got, err := x509utils.BlockToPrivateKey(block)
	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "key")
	core.AssertTrue(t, got.Equal(key), "round-trip")
}

// TestEncodePKCS8PrivateKey covers the nil-key arm (empty output, no error),
// the marshal-error arm (a key type PKCS8 rejects) and the round-trip.
func TestEncodePKCS8PrivateKey(t *testing.T) {
	out, err := x509utils.EncodePKCS8PrivateKey(nil)
	core.AssertNoError(t, err, "nil key")
	core.AssertNil(t, out, "nil output")

	_, err = x509utils.EncodePKCS8PrivateKey(unsupportedKey{})
	core.AssertError(t, err, "unsupported")

	key := mkKey(t)
	out, err = x509utils.EncodePKCS8PrivateKey(key)
	core.AssertMustNoError(t, err, "encode")
	block, _ := pem.Decode(out)
	core.AssertMustNotNil(t, block, "block")
	got, err := x509utils.BlockToPrivateKey(block)
	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "key")
	core.AssertTrue(t, got.Equal(key), "round-trip")
}

// TestEncodeCertificate covers the certificate encode path, and EncodeBytes
// underneath: the PEM it produces is a CERTIFICATE block that parses back to
// the original certificate.
func TestEncodeCertificate(t *testing.T) {
	cert := certSpec{cn: "encode.test"}.build(t).cert

	block, _ := pem.Decode(x509utils.EncodeCertificate(cert.Raw))
	core.AssertMustNotNil(t, block, "block")
	core.AssertEqual(t, "CERTIFICATE", block.Type, "type")

	got, err := x509utils.BlockToCertificate(block)
	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "cert")
	core.AssertTrue(t, got.Equal(cert), "round-trip")
}
