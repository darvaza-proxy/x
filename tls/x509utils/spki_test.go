package x509utils_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = subjectPublicKeyBytesTestCase{}
	_ core.TestCase = subjectPublicKeyHashTestCase{}
)

// spkiBytes returns the SubjectPublicKey bit-string bytes for pub by an
// independent stdlib path, so the expected value never re-derives through
// SubjectPublicKeyBytes (the function under test). Each key type has its own
// encoding: an EC key marshals to the uncompressed point, RSA to the PKCS1
// SEQUENCE, Ed25519 to the raw 32-byte key.
func spkiBytes(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		ecdhPub, err := k.ECDH()
		core.AssertMustNoError(t, err, "ecdh")
		return ecdhPub.Bytes()
	case *rsa.PublicKey:
		return x509.MarshalPKCS1PublicKey(k)
	case ed25519.PublicKey:
		return []byte(k)
	default:
		core.AssertMustNil(t, pub, "unhandled public key type")
		return nil
	}
}

// subjectPublicKeyBytesTestCase exercises SubjectPublicKeyBytes: a supported
// key marshals to its raw SubjectPublicKey bytes, stripping the algorithm
// identifier, whereas one that cannot be encoded surfaces the error instead.
type subjectPublicKeyBytesTestCase struct {
	pub     crypto.PublicKey
	name    string
	wantErr string
	want    []byte
}

func (tc subjectPublicKeyBytesTestCase) Name() string { return tc.name }

func (tc subjectPublicKeyBytesTestCase) Test(t *testing.T) {
	t.Helper()
	got, err := x509utils.SubjectPublicKeyBytes(tc.pub)
	if tc.wantErr != "" {
		core.AssertError(t, err, "error")
		core.AssertContains(t, err.Error(), tc.wantErr, "error")
		core.AssertNil(t, got, "bytes")
		return
	}
	core.AssertMustNoError(t, err, "encode")
	core.AssertSliceEqual(t, tc.want, got, "spki bytes")
}

func newSubjectPublicKeyBytesTestCase(name string, pub crypto.PublicKey,
	want []byte, wantErr string) subjectPublicKeyBytesTestCase {
	return subjectPublicKeyBytesTestCase{
		pub:     pub,
		want:    want,
		name:    name,
		wantErr: wantErr,
	}
}

func subjectPublicKeyBytesTestCases(t *testing.T) []subjectPublicKeyBytesTestCase {
	t.Helper()

	ecPub := mkKey(t).Public()
	rsaPub := mkRSAKey(t).Public()
	edPub := mkEd25519Key(t).Public()

	// SubjectPublicKeyBytes asserts the asn1.Unmarshal cannot fail via
	// core.MustNoError: MarshalPKIXPublicKey always emits well-formed SPKI, so
	// re-decoding it never errors. Only the encode arm is reachable here: nil
	// and an unsupported type both fail MarshalPKIXPublicKey.
	return core.S(
		newSubjectPublicKeyBytesTestCase("ECDSA P-256", ecPub,
			spkiBytes(t, ecPub), ""),
		newSubjectPublicKeyBytesTestCase("RSA", rsaPub,
			spkiBytes(t, rsaPub), ""),
		newSubjectPublicKeyBytesTestCase("Ed25519", edPub,
			spkiBytes(t, edPub), ""),
		newSubjectPublicKeyBytesTestCase("nil public key", nil, nil,
			"failed to encode public key"),
		newSubjectPublicKeyBytesTestCase("unsupported type", struct{}{}, nil,
			"failed to encode public key"),
	)
}

func TestSubjectPublicKeyBytes(t *testing.T) {
	core.RunTestCases(t, subjectPublicKeyBytesTestCases(t))
}

// subjectPublicKeyHashTestCase exercises the SHA1/SHA224/SHA256 SubjectPublicKey
// hashers together: each hashes the same SubjectPublicKey bytes, so one row
// pins all three against an independently derived digest and their shared
// error-propagation arm.
type subjectPublicKeyHashTestCase struct {
	pub     crypto.PublicKey
	name    string
	spki    []byte
	wantErr bool
}

func (tc subjectPublicKeyHashTestCase) Name() string { return tc.name }

func (tc subjectPublicKeyHashTestCase) Test(t *testing.T) {
	t.Helper()

	h1, err1 := x509utils.SubjectPublicKeySHA1(tc.pub)
	h224, err224 := x509utils.SubjectPublicKeySHA224(tc.pub)
	h256, err256 := x509utils.SubjectPublicKeySHA256(tc.pub)

	if tc.wantErr {
		core.AssertError(t, err1, "sha1 error")
		core.AssertError(t, err224, "sha224 error")
		core.AssertError(t, err256, "sha256 error")
		return
	}

	core.AssertMustNoError(t, err1, "sha1")
	core.AssertMustNoError(t, err224, "sha224")
	core.AssertMustNoError(t, err256, "sha256")
	core.AssertEqual(t, sha1.Sum(tc.spki), h1, "sha1")
	core.AssertEqual(t, sha256.Sum224(tc.spki), h224, "sha224")
	core.AssertEqual(t, sha256.Sum256(tc.spki), h256, "sha256")
}

func newSubjectPublicKeyHashTestCase(name string, pub crypto.PublicKey,
	spki []byte, wantErr bool) subjectPublicKeyHashTestCase {
	return subjectPublicKeyHashTestCase{
		pub:     pub,
		spki:    spki,
		name:    name,
		wantErr: wantErr,
	}
}

func subjectPublicKeyHashTestCases(t *testing.T) []subjectPublicKeyHashTestCase {
	t.Helper()

	ecPub := mkKey(t).Public()

	return core.S(
		newSubjectPublicKeyHashTestCase("ECDSA P-256", ecPub,
			spkiBytes(t, ecPub), false),
		newSubjectPublicKeyHashTestCase("unsupported type", struct{}{}, nil,
			true),
	)
}

func TestSubjectPublicKeyHashes(t *testing.T) {
	core.RunTestCases(t, subjectPublicKeyHashTestCases(t))
}
