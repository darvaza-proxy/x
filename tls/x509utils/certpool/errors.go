package certpool

import "darvaza.org/core"

var (
	// ErrNilReceiver indicates a method was called over an undefined object.
	//
	// TODO: move to core.
	ErrNilReceiver = core.Wrap(core.ErrInvalid, "nil receiver")
)
