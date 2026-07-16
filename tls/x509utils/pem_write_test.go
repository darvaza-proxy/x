package x509utils_test

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

// Compile-time TestCase interface assertions, kept together as an index.
var (
	_ core.TestCase = writeCertTestCase{}
	_ core.TestCase = writeKeyTestCase{}
)

// errWrite is the failure a failing writer reports.
var errWrite = errors.New("write failed")

// errWriter fails every write, to drive WriteCert's flush-to-destination error
// path after the certificate has already encoded.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errWrite }

// writeCertTestCase exercises WriteCert: its two reachable rejection paths,
// each an ErrInvalidCert built by NewErrInvalidCert, and the encoding path.
type writeCertTestCase struct {
	cert    *x509.Certificate
	name    string
	wantMsg string // empty => expect a PEM block, no error
}

func (tc writeCertTestCase) Name() string { return tc.name }

func (tc writeCertTestCase) Test(t *testing.T) {
	t.Helper()

	var buf bytes.Buffer
	n, err := x509utils.WriteCert(&buf, tc.cert)

	if tc.wantMsg == "" {
		core.AssertNoError(t, err, "error")
		core.AssertEqual(t, int64(buf.Len()), n, "written")
		core.AssertContains(t, buf.String(), "BEGIN CERTIFICATE", "pem")
		return
	}

	core.AssertEqual(t, int64(0), n, "written")
	core.AssertMustError(t, err, "error")
	core.AssertEqual(t, tc.wantMsg, err.Error(), "message")
	// every WriteCert rejection defaults to core.ErrInvalid (F44).
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid")

	// the rejection threads the offending cert through unchanged.
	ec := core.AssertMustTypeIs[*x509utils.ErrInvalidCert](t, err, "error")
	core.AssertSame(t, tc.cert, ec.Cert, "cert")
}

func newWriteCertTestCase(name string, cert *x509.Certificate,
	wantMsg string) writeCertTestCase {
	return writeCertTestCase{
		cert:    cert,
		name:    name,
		wantMsg: wantMsg,
	}
}

func writeCertTestCases() []writeCertTestCase {
	return core.S(
		// nil cert: nothing to thread through, just a reason.
		newWriteCertTestCase("nil", nil,
			"invalid certificate: not provided"),
		// a cert with no DER cannot be encoded; the cert is carried.
		newWriteCertTestCase("empty raw", &x509.Certificate{},
			"invalid certificate: missing Raw DER certificate"),
		// pem.Encode wraps the Raw bytes verbatim; no parsing happens, so any
		// non-empty Raw yields a PEM block.
		newWriteCertTestCase("encodes raw",
			&x509.Certificate{Raw: []byte("not real DER")}, ""),
	)
}

func TestWriteCert(t *testing.T) {
	core.RunTestCases(t, writeCertTestCases())
}

// TestWriteCertWriteError covers the writer-failure branch: the certificate
// encodes into the buffer, but flushing it to the destination fails, so
// WriteCert reports zero bytes written and the writer's error.
func TestWriteCertWriteError(t *testing.T) {
	cert := &x509.Certificate{Raw: []byte("not real DER")}

	n, err := x509utils.WriteCert(errWriter{}, cert)
	core.AssertEqual(t, int64(0), n, "written")
	core.AssertErrorIs(t, err, errWrite, "error")
}

// unsupportedKey is a PrivateKey whose concrete type
// x509.MarshalPKCS8PrivateKey does not recognise, driving WriteKey's
// marshal-error arm.
type unsupportedKey struct{}

func (unsupportedKey) Public() crypto.PublicKey { return nil }

func (unsupportedKey) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, nil
}

func (unsupportedKey) Equal(crypto.PrivateKey) bool { return false }

// writeKeyTestCase exercises WriteKey: the PKCS8 encode-and-round-trip path,
// and the marshal-error arm reached by a nil key or a key type PKCS8
// marshalling rejects.
type writeKeyTestCase struct {
	key     x509utils.PrivateKey
	name    string
	wantErr bool
}

func (tc writeKeyTestCase) Name() string { return tc.name }

func (tc writeKeyTestCase) Test(t *testing.T) {
	t.Helper()

	var buf bytes.Buffer
	n, err := x509utils.WriteKey(&buf, tc.key)

	if tc.wantErr {
		core.AssertEqual(t, int64(0), n, "written")
		core.AssertMustError(t, err, "error")
		core.AssertContains(t, err.Error(), "failed to encode key", "message")
		return
	}

	core.AssertMustNoError(t, err, "error")
	core.AssertEqual(t, int64(buf.Len()), n, "written")

	// the written block decodes back to the original key.
	block, _ := pem.Decode(buf.Bytes())
	core.AssertMustNotNil(t, block, "block")
	core.AssertEqual(t, "PRIVATE KEY", block.Type, "type")

	got, err := x509utils.BlockToPrivateKey(block)
	core.AssertMustNoError(t, err, "parse")
	core.AssertMustNotNil(t, got, "key")
	core.AssertTrue(t, got.Equal(tc.key), "key round-trip")
}

func newWriteKeyTestCase(name string, key x509utils.PrivateKey,
	wantErr bool) writeKeyTestCase {
	return writeKeyTestCase{
		key:     key,
		name:    name,
		wantErr: wantErr,
	}
}

func writeKeyTestCases(t *testing.T) []writeKeyTestCase {
	t.Helper()

	return core.S(
		// a real key round-trips through PKCS8; WriteKey is type-agnostic,
		// so one key type covers the encode-and-flush path.
		newWriteKeyTestCase("ECDSA", mkKey(t), false),
		// a nil key and an unrecognised key type both fail PKCS8 marshalling.
		newWriteKeyTestCase("nil key", nil, true),
		newWriteKeyTestCase("unsupported key", unsupportedKey{}, true),
	)
}

func TestWriteKey(t *testing.T) {
	core.RunTestCases(t, writeKeyTestCases(t))
}

// TestWriteKeyWriteError covers the writer-failure branch: the key encodes
// into the buffer, but flushing it to the destination fails, so WriteKey
// reports zero bytes written and the writer's error.
func TestWriteKeyWriteError(t *testing.T) {
	n, err := x509utils.WriteKey(errWriter{}, mkKey(t))
	core.AssertEqual(t, int64(0), n, "written")
	core.AssertErrorIs(t, err, errWrite, "error")
}
