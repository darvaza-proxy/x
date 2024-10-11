//go:build !linux

package certpool

import "darvaza.org/core"

func NewSystemCertPool() (*CertPool, error) {
	return nil, core.ErrTODO
}
