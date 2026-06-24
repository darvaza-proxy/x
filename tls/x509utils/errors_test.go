package x509utils_test

import (
	"crypto/x509"
	"errors"
	"fmt"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var _ core.TestCase = errInvalidCertTestCase{}

// errInvalidCertTestCase exercises the message and core.ErrInvalid membership
// of an error built by NewErrInvalidCert.
type errInvalidCertTestCase struct {
	inErr         error
	name          string
	reason        string
	wantMsg       string
	wantIsInvalid bool
}

func (tc errInvalidCertTestCase) Name() string { return tc.name }

func (tc errInvalidCertTestCase) Test(t *testing.T) {
	t.Helper()

	err := x509utils.NewErrInvalidCert(nil, tc.inErr, tc.reason)

	core.AssertEqual(t, tc.wantMsg, err.Error(), "message")
	core.AssertEqual(t, tc.wantIsInvalid,
		errors.Is(err, core.ErrInvalid), "is invalid")
}

func newErrInvalidCertTestCase(name string, inErr error, reason,
	wantMsg string, wantIsInvalid bool) errInvalidCertTestCase {
	return errInvalidCertTestCase{
		inErr:         inErr,
		name:          name,
		reason:        reason,
		wantMsg:       wantMsg,
		wantIsInvalid: wantIsInvalid,
	}
}

func errInvalidCertTestCases() []errInvalidCertTestCase {
	boom := errors.New("boom")
	wrapped := fmt.Errorf("deeper: %w", core.ErrInvalid)

	return []errInvalidCertTestCase{
		// A nil cause defaults to core.ErrInvalid: hidden from the message
		// but enough to satisfy errors.Is, with or without a reason.
		newErrInvalidCertTestCase("reason only", nil,
			"none provided", "invalid certificate: none provided", true),
		newErrInvalidCertTestCase("no reason", nil,
			"", "invalid certificate", true),
		// The sentinel passed explicitly behaves like the default: hidden but
		// invalid.
		newErrInvalidCertTestCase("explicit sentinel", core.ErrInvalid,
			"bad", "invalid certificate: bad", true),
		// An unrelated explicit cause is shown and is not promoted to
		// core.ErrInvalid, whether or not a reason is given.
		newErrInvalidCertTestCase("explicit cause", boom,
			"failed to encode certificate",
			"invalid certificate: failed to encode certificate: boom", false),
		newErrInvalidCertTestCase("explicit cause, no reason", boom,
			"", "invalid certificate: boom", false),
		// An explicit cause that itself wraps core.ErrInvalid is shown and
		// still satisfies errors.Is.
		newErrInvalidCertTestCase("explicit wrapped invalid", wrapped,
			"bad", "invalid certificate: bad: deeper: invalid argument", true),
	}
}

func TestNewErrInvalidCert(t *testing.T) {
	core.RunTestCases(t, errInvalidCertTestCases())
}

// TestNewErrInvalidCertFields confirms the constructor keeps the certificate
// and the unwrap target: the default core.ErrInvalid when no cause is given,
// and the caller's cause otherwise.
func TestNewErrInvalidCertFields(t *testing.T) {
	cert := &x509.Certificate{}
	boom := errors.New("boom")

	def := x509utils.NewErrInvalidCert(cert, nil, "x")
	core.AssertSame(t, cert, def.Cert, "cert")
	core.AssertSame(t, core.ErrInvalid, def.Unwrap(), "default cause")

	wrap := x509utils.NewErrInvalidCert(nil, boom, "x")
	core.AssertSame(t, boom, wrap.Unwrap(), "explicit cause")
	core.AssertErrorIs(t, wrap, boom, "wraps cause")
}

// TestErrInvalidCertDirect covers the zero-cause path the constructor never
// produces: a struct built by hand with a nil Err keeps the message clean,
// unwraps to nil, and does not satisfy core.ErrInvalid.
func TestErrInvalidCertDirect(t *testing.T) {
	raw := x509utils.ErrInvalidCert{Reason: "raw"}

	core.AssertEqual(t, "invalid certificate: raw", raw.Error(), "message")
	core.AssertNil(t, raw.Unwrap(), "unwrap")
	core.AssertFalse(t, errors.Is(&raw, core.ErrInvalid), "not invalid")
}
