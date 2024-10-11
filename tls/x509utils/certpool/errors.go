package certpool

import "darvaza.org/core"

var (
	// ErrNoCertificatesFound indicates we didn't find any certificate.
	ErrNoCertificatesFound = core.Wrap(core.ErrNotExists, "no certificates found")
)
