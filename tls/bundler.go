package tls

import (
	"crypto/tls"
	"crypto/x509"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// Bundler uses two CertPools to bundle keys and certificates
type Bundler struct {
	// Root Certificates. Defaults to system's.
	Roots x509utils.CertPool
	// Intermediate Certificates.
	Inter x509utils.CertPool
	// Quality comparison function. Defaults to shorter-chain.
	Less func(a, b []*x509.Certificate) bool

	mu   sync.Mutex
	opts x509.VerifyOptions
}

// Bundle bundles a key and a certificate into a [tls.Certificate] using the
// specified roots, intermediates and quality function.
func (s *Bundler) Bundle(cert *x509.Certificate, key x509utils.PrivateKey) (*tls.Certificate, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	return BundleFn(s.opts, s.Less, cert, key)
}

func (s *Bundler) init() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.unsafeInit()
}

func (s *Bundler) unsafeInit() error {
	if s.opts.Roots == nil {
		if s.Roots == nil {
			pool, err := certpool.SystemCertPool()
			if err != nil {
				return err
			}
			s.Roots = pool
		}
		s.opts.Roots = s.Roots.Export()
	}

	switch {
	case s.Inter == nil:
		s.opts.Intermediates = nil
	case s.opts.Intermediates == nil:
		s.opts.Intermediates = s.Inter.Export()
	}

	return nil
}

// Reset drops any cached information.
func (s *Bundler) Reset() {
	if s != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.opts.Roots = nil
		s.opts.Intermediates = nil
	}
}

// Bundle assembles a verified [tls.Certificate], choosing the shortest trust chain.
func Bundle(opt x509.VerifyOptions, cert *x509.Certificate, key x509utils.PrivateKey) (*tls.Certificate, error) {
	return BundleFn(opt, nil, cert, key)
}

// BundleFn assembles a verified [tls.Certificate], using the given quality function.
func BundleFn(opt x509.VerifyOptions, less func(a, b []*x509.Certificate) bool, //
	cert *x509.Certificate, key x509utils.PrivateKey) (*tls.Certificate, error) {
	//
	switch {
	case cert == nil:
		return nil, core.QuietWrap(core.ErrInvalid, "certificate not provided")
	case key != nil && !x509utils.ValidCertKeyPair(cert, key):
		return nil, core.QuietWrap(core.ErrInvalid, "key doesn't match certificate")
	case opt.Roots == nil:
		pool, err := certpool.SystemCertPool()
		if err != nil {
			return nil, err
		}
		opt.Roots = pool.Export()
	}

	if less == nil {
		less = defaultCertChainLess
	}

	return unsafeBundleFn(opt, less, cert, key)
}

func defaultCertChainLess(a, b []*x509.Certificate) bool {
	return len(a) < len(b)
}

func unsafeBundleFn(opt x509.VerifyOptions, //
	less func(a, b []*x509.Certificate) bool, //
	cert *x509.Certificate, key x509utils.PrivateKey) (*tls.Certificate, error) {
	//
	chains, err := cert.Verify(opt)
	switch {
	case err != nil:
		return nil, err
	case len(chains) == 0:
		return nil, core.QuietWrap(core.ErrInvalid, "failed to build trust chain")
	default:
		core.SliceSortFn(chains, less)
	}

	chain := core.SliceAsFn(func(cert *x509.Certificate) ([]byte, bool) {
		if cert != nil && len(cert.Raw) > 0 {
			return cert.Raw, true
		}
		return nil, false
	}, chains[0])

	out := &tls.Certificate{
		Certificate: chain,
		PrivateKey:  key,
		Leaf:        cert,
	}

	return out, nil
}
