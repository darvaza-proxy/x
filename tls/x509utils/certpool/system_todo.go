//go:build !linux && !windows

package certpool

import "darvaza.org/core"

// NewSystemCertPool returns a [CertPool] populated with the system's
// trusted certificates. This platform has no native loader yet, so it
// reports [core.ErrTODO], the sentinel callers treat as "not
// implemented".
func NewSystemCertPool() (*CertPool, error) {
	return nil, core.ErrTODO
}
