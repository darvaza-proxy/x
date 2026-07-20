//go:build darwin

package certpool

// cspell:words macos

import (
	"crypto/x509"

	"darvaza.org/core"

	"darvaza.org/x/tls/internal/macos"
)

// trustSettingsDomains are the trust-settings domains [NewSystemCertPool]
// walks, listed from most to least authoritative.
var trustSettingsDomains = []macos.Domain{
	macos.DomainSystem,
	macos.DomainAdmin,
	macos.DomainUser,
}

// NewSystemCertPool returns a [CertPool] populated from the macOS system
// trust store, read directly through Security.framework rather than cgo.
//
// It enumerates the candidate roots that carry trust settings across the
// system, admin and user domains, then asks the OS to evaluate each one and
// keeps only those it currently honours. Delegating the decision to
// SecTrustEvaluateWithError means an administrator's explicit distrust and
// Apple's dynamic blocklists are respected without re-implementing Apple's
// trust-settings interpretation.
func NewSystemCertPool() (*CertPool, error) {
	b, err := macos.New()
	if err != nil {
		return nil, err
	}

	l := &darwinLoader{
		pool:   New(),
		b:      b,
		policy: b.NewBasicX509Policy(),
	}
	defer l.policy.Release()

	for _, domain := range trustSettingsDomains {
		l.loadDomain(domain)
	}

	err = l.errs.AsError()
	switch {
	case l.pool.Count() > 0:
		// success, possibly alongside per-certificate errors
		return l.pool, err
	case err == nil:
		// no certs and no errors... don't bother again.
		return nil, ErrNoCertificatesFound
	default:
		// no cert, but we got errors to report.
		return nil, err
	}
}

// darwinLoader carries the state threaded through the per-domain,
// per-certificate walk: the pool being filled, the bindings and shared
// policy, and the accumulated non-fatal errors.
type darwinLoader struct {
	pool   *CertPool
	b      *macos.Bindings
	policy macos.Policy
	errs   core.CompoundError
}

// loadDomain evaluates every candidate root in one trust-settings domain and
// adds the OS-confirmed-good certificates to the pool.
func (l *darwinLoader) loadDomain(domain macos.Domain) {
	arr, err := l.b.CopyTrustSettingsCertificates(domain)
	if err != nil {
		_ = l.errs.AppendError(err)
		return
	}
	defer arr.Release()

	for i := range arr.Count() {
		l.addIfTrusted(arr.At(i))
	}
}

// addIfTrusted evaluates one candidate certificate against the shared policy
// and, if the system currently trusts it, adds it to the pool. A "not
// trusted" verdict — an explicit Deny or a dynamic block — is the expected
// outcome for a rejected root, so it is skipped silently rather than
// recorded as an error.
func (l *darwinLoader) addIfTrusted(cert macos.Certificate) {
	trust, err := l.b.NewTrust(cert, l.policy)
	if err != nil {
		_ = l.errs.AppendError(err)
		return
	}

	ok := trust.EvaluateWithError()
	trust.Release()
	if ok {
		l.addCert(cert)
	}
}

// addCert copies the DER out of a trusted certificate, parses it and adds it
// to the pool. A blob Go's parser rejects is recorded against its hash, so
// one bad entry is reported by name without sinking the rest of the load.
func (l *darwinLoader) addCert(cert macos.Certificate) {
	der, err := cert.CopyData()
	if err != nil {
		_ = l.errs.AppendError(err)
		return
	}

	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		hash := Sum(der)
		_ = l.errs.AppendError(core.Wrapf(err, "certificate %x", hash[:]))
		return
	}

	l.pool.AddCert(parsed)
}
