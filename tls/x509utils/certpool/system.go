package certpool

import (
	"crypto/x509"
	"sync"
)

var (
	systemMutex    sync.Mutex
	systemCerts    *CertPool
	systemCertsErr error

	// SystemCAOnly indicates if SystemCertPool show ignore
	// non-CA certificates.
	//
	// Changing this after SystemCertPool() has been called
	// has no effect.
	SystemCAOnly bool
)

// SystemCertPool returns a Pool populated with the
// system's valid certificates.
func SystemCertPool() (*CertPool, error) {
	var cond func(*x509.Certificate) bool

	systemMutex.Lock()
	defer systemMutex.Unlock()

	if SystemCAOnly {
		cond = func(c *x509.Certificate) bool {
			return c.IsCA
		}
	}

	switch {
	case systemCertsErr != nil:
		return nil, systemCertsErr
	case systemCerts != nil:
		return systemCerts.Copy(nil, cond), nil
	}

	// first call
	roots, err := NewSystemCertPool()
	if roots.Count() > 0 {
		// remember roots
		systemCerts = roots
		return roots.Copy(nil, cond), nil
	}

	if err == nil {
		err = ErrNoCertificatesFound
	}

	// remember error
	systemCertsErr = err
	return nil, err
}
