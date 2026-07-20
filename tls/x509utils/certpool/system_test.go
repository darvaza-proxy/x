package certpool_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

func TestSystemCertPool(t *testing.T) {
	pool, err := certpool.SystemCertPool()
	core.AssertMustNoError(t, err, "system pool")
	core.AssertMustNotNil(t, pool, "pool")

	ctx := context.Background()
	if deadLine, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadLine)
		defer cancel()
	}

	count, minCount := pool.Count(), minSystemCertCount()
	core.AssertTrue(t, count >= minCount, "count %d >= %d", count, minCount)

	i := 1
	pool.ForEach(ctx, func(_ context.Context, cert *x509.Certificate) bool {
		printSystemCertTest(t, i, count, cert)
		i++
		return true
	})
	core.AssertEqual(t, count, i-1, "visited all")
}

// minSystemCertCount returns the fewest roots a healthy system trust store
// is expected to hold on this platform. A loader that returns a non-empty
// but implausibly small pool — a partial walk, an over-strict filter —
// passes a bare non-empty check; this floor catches it. Each value sits
// well below what the platform actually ships, so it only trips on a
// genuinely truncated load.
func minSystemCertCount() int {
	switch runtime.GOOS {
	case "darwin", "linux":
		// Apple ships ~150 built-in roots; Linux distributions carry
		// the Mozilla bundle, of a similar size.
		return 100
	case "windows":
		// A fresh Windows root store holds only a few dozen
		// preinstalled roots and grows on demand.
		return 30
	default:
		return 1
	}
}

// TestSystemCertPoolVerifiesGitHub dials github.com over TLS trusting only
// the certificates the system loader produced. A completed handshake proves
// the loaded pool carries the roots needed to verify a real public
// certificate, so it is the end-to-end check the native platform loaders
// otherwise lack.
func TestSystemCertPoolVerifiesGitHub(t *testing.T) {
	if testing.Short() {
		t.Skip("network test skipped in -short mode")
	}

	pool, err := certpool.SystemCertPool()
	core.AssertMustNoError(t, err, "system pool")
	core.AssertMustNotNil(t, pool, "pool")

	roots := pool.Export()
	core.AssertMustNotNil(t, roots, "exported roots")

	conn := dialGitHub(t, roots)
	defer func() { _ = conn.Close() }()

	certs := conn.ConnectionState().PeerCertificates
	core.AssertTrue(t, len(certs) > 0, "peer certificates")
}

// TestCertVerifyErrorClassification exercises the predicate dialGitHub uses
// to tell a real failure from a skip: only a TLS certificate-verification
// error counts. Dialling 127.0.0.1:0 stays on the loopback and never touches
// the network, yet always fails to connect, so its error must classify as
// skip-worthy; a synthetic verification error must classify as a hard
// failure.
func TestCertVerifyErrorClassification(t *testing.T) {
	_, dialErr := dialTLS("127.0.0.1:0", nil)
	core.AssertMustError(t, dialErr, "dial 127.0.0.1:0")
	core.AssertFalse(t, isCertVerifyError(dialErr), "connect error is verify")

	verifyErr := &tls.CertificateVerificationError{Err: core.ErrInvalid}
	core.AssertTrue(t, isCertVerifyError(verifyErr), "verify error recognised")
}

// dialGitHub completes a TLS handshake to github.com verified against roots
// and returns the connection. The subject under test is whether roots can
// verify a real public certificate, not whether the network is up: a
// [tls.CertificateVerificationError] means the loaded pool lacks a needed
// root and fails the test, while any other error — an unreachable host, a
// connect or handshake timeout, a reset mid-handshake — says nothing about
// the pool and skips instead.
func dialGitHub(t *testing.T, roots *x509.CertPool) *tls.Conn {
	t.Helper()

	conn, err := dialTLS("github.com:443", roots)
	if err != nil {
		if !isCertVerifyError(err) {
			t.Skipf("github.com:443 unreachable: %v", err)
		}
		core.AssertMustNoError(t, err, "tls verify github.com")
	}
	return conn
}

// dialTimeout bounds the whole github.com dial: TCP connect plus TLS
// handshake. The handshake needs only tens of milliseconds, so five seconds
// looks absurd — but this is not sizing the handshake. A timeout skips the
// test rather than failing it (see dialGitHub), so too tight a bound never
// causes a false failure; it silently drops the end-to-end check that the
// loaded roots verify a live certificate. Kept generous to absorb dual-stack
// fallback and cold DNS on slow CI runners, while still bailing well before
// the OS default when the host is truly unreachable.
const dialTimeout = 5 * time.Second

// dialTLS completes a TLS handshake to addr verified against roots, under a
// generous connect timeout that fails fast only against a truly unreachable
// host.
func dialTLS(addr string, roots *x509.CertPool) (*tls.Conn, error) {
	dialer := &net.Dialer{Timeout: dialTimeout}
	return tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		RootCAs:    roots,
		MinVersion: tls.VersionTLS12,
	})
}

// isCertVerifyError reports whether err is a TLS certificate-verification
// failure — the one dial error that means the loaded pool lacks a needed
// root, rather than the host merely being unreachable.
func isCertVerifyError(err error) bool {
	var verifyErr *tls.CertificateVerificationError
	return errors.As(err, &verifyErr)
}

func printSystemCertTest(t *testing.T, i, count int, cert *x509.Certificate) {
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "[%v/%v] ", i, count)
	if cert.IsCA {
		_, _ = buf.WriteString("CA ")
	}
	if len(cert.SubjectKeyId) > 0 {
		_, _ = buf.WriteString(base64.StdEncoding.EncodeToString(cert.SubjectKeyId))
		_, _ = buf.WriteRune(' ')
	}

	_, _ = fmt.Fprintf(&buf, "%q", cert.Subject)

	t.Log(buf.String())
}
