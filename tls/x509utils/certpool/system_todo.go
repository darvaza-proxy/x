//go:build !linux

package certpool

import "darvaza.org/core"

func loadSystemCerts() (*CertPool, error) {
	return nil, core.ErrTODO
}
