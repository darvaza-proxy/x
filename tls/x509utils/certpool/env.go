package certpool

import (
	"os"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// NewFromEnv uses [os.LookupEnv] and the given variable name to create a [CertPool]
// using [NewFromStrings].
func NewFromEnv(name string) (*CertPool, error) {
	return NewFromEnvFn(os.LookupEnv, name)
}

// NewFromEnvFn is like [NewFromEnv] but allows the user to provide a function to
// access the environment.
func NewFromEnvFn(getEnv func(string) (string, bool), name string) (*CertPool, error) {
	switch {
	case getEnv == nil:
		return nil, core.Wrapf(core.ErrInvalid, "missing argument: %s", "getEnv")
	case name == "":
		return nil, core.Wrapf(core.ErrInvalid, "missing argument: %s", "name")
	}

	s, ok := getEnv(name)
	switch {
	case !ok:
		return nil, core.Wrapf(core.ErrNotExists, "undefined variable: %q", name)
	case s == "":
		return nil, core.Wrapf(core.ErrNotExists, "empty value for variable: %q", name)
	default:
		return NewFromStrings(s)
	}
}

// NewFromStrings creates a [CertPool] using certificates in a string.
// it could be PEM content, a PEM file, or a directory containing
// PEM files.
// The populated [CertPool] will be returned even if errors were encountered.
// A [core.CompoundError] will be returned if there were errors.
func NewFromStrings(certs ...string) (*CertPool, error) {
	var errs core.CompoundError

	pool := New()
	addFn := newCertAdder(pool, false, &errs)
	for _, s := range certs {
		err := x509utils.ReadStringPEM(s, addFn)
		if err != nil {
			_ = errs.AppendError(err)
		}
	}

	if err := errs.AsError(); err != nil {
		return pool, err
	} else if pool.Count() == 0 {
		return nil, ErrNoCertificatesFound
	}

	return pool, nil
}
