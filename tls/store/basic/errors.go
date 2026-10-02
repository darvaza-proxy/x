package basic

import (
	"darvaza.org/core"

	"darvaza.org/x/container/set"
	"darvaza.org/x/tls/x509utils/certpool"
)

var (
	// ErrInvalid is an alias of [core.ErrInvalid].
	ErrInvalid = core.ErrInvalid
	// ErrExist is an alias of [set.ErrExist]
	ErrExist = set.ErrExist
	// ErrNotExist is an alias of [set.ErrNotExist].
	ErrNotExist = set.ErrNotExist

	// ErrNoCert indicates no certificate was provided.
	ErrNoCert = core.Wrap(ErrInvalid, "no certificate provided")
	// ErrNoKey indicates no private key was provided.
	ErrNoKey = core.Wrap(ErrInvalid, "no key provided")
	// ErrBadCert indicates an invalid certificate was provided.
	ErrBadCert = core.Wrap(ErrInvalid, "invalid certificate provided")
	// ErrBadKey indicates an invalid private key was provided.
	ErrBadKey = core.Wrap(ErrInvalid, "invalid key provided")
	// ErrCertNotFound indicates no matching certificate was found.
	ErrCertNotFound = certpool.ErrNoCertificatesFound
)
