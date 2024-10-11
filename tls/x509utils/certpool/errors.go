package certpool

import "darvaza.org/core"

var (
	// ErrNilReceiver indicates a method was called over an undefined object.
	//
	// TODO: move to core.
	ErrNilReceiver = core.Wrap(core.ErrInvalid, "nil receiver")

	// ErrNoCertificatesFound indicates we didn't find any certificate.
	ErrNoCertificatesFound = core.Wrap(core.ErrNotExists, "no certificates found")
)
