package x509utils_test

import (
	"bytes"
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var _ core.TestCase = writeCertTestCase{}

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
	if !core.AssertError(t, err, "error") {
		return
	}
	core.AssertEqual(t, tc.wantMsg, err.Error(), "message")
	// every WriteCert rejection defaults to core.ErrInvalid (F44).
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid")

	// the rejection threads the offending cert through unchanged.
	ec, ok := err.(*x509utils.ErrInvalidCert)
	core.AssertMustTrue(t, ok, "type")
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
	return []writeCertTestCase{
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
	}
}

func TestWriteCert(t *testing.T) {
	core.RunTestCases(t, writeCertTestCases())
}
