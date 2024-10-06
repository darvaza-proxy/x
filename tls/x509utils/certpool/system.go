package certpool

import "sync"

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
// system's valid CA certificates.
func SystemCertPool() (*CertPool, error) {
	systemMutex.Lock()
	defer systemMutex.Unlock()

	switch {
	case systemCertsErr != nil:
		return nil, systemCertsErr
	case systemCerts != nil:
		return systemCerts.Copy(nil, false), nil
	}

	// first call
	roots, err := loadSystemCerts()
	if err != nil {
		// remember error
		systemCertsErr = err
		return nil, err
	}

	// remember roots
	systemCerts = roots
	return roots.Copy(nil, false), nil
}