package x509utils

import (
	"crypto/x509"
	"fmt"
	"io/fs"
	"strings"

	"darvaza.org/core"
)

var (
	_ error            = (*ErrInvalidCert)(nil)
	_ core.Unwrappable = (*ErrInvalidCert)(nil)
)

// ErrInvalidCert indicates the certificate wasn't acceptable.
type ErrInvalidCert struct {
	Cert   *x509.Certificate
	Err    error
	Reason string
}

// NewErrInvalidCert builds an [ErrInvalidCert] reporting that cert was not
// acceptable for the stated reason, optionally wrapping the underlying err.
// Both cert and err may be nil; when err is nil the error defaults to wrapping
// [core.ErrInvalid], so every [ErrInvalidCert] from this constructor satisfies
// errors.Is(err, core.ErrInvalid).
func NewErrInvalidCert(cert *x509.Certificate, err error, reason string) *ErrInvalidCert {
	if err == nil {
		err = core.ErrInvalid
	}
	return &ErrInvalidCert{
		Cert:   cert,
		Err:    err,
		Reason: reason,
	}
}

func (err ErrInvalidCert) Error() string {
	s := make([]string, 0, 3)
	s = append(s, "invalid certificate")

	if err.Reason != "" {
		s = append(s, err.Reason)
	}

	// core.ErrInvalid is the default cause and adds nothing beyond the
	// "invalid certificate" prefix, so it is left out of the message; a
	// caller-supplied cause is always shown.
	if err.Err != nil && err.Err != core.ErrInvalid {
		s = append(s, err.Err.Error())
	}

	return strings.Join(s, ": ")
}

func (err ErrInvalidCert) Unwrap() error {
	return err.Err
}

// newInvalidPathError reports a path as unusable for the stated
// reason, wrapping fs.ErrInvalid so errors.Is matches against the
// stdlib sentinel regardless of the core release in use.
func newInvalidPathError(op, path, reason string) *fs.PathError {
	return &fs.PathError{
		Op:   op,
		Path: clampErrPath(path),
		Err:  core.Wrap(fs.ErrInvalid, reason),
	}
}

// maxErrPathLen bounds how much of an offending path an error echoes. A
// candidate longer than a real file name is a raw blob mistaken for one, and
// storing it whole would flood the error; any genuine path sits well under it.
const maxErrPathLen = 256

// clampErrPath keeps an over-long path from filling the error with the entire
// input, showing a leading slice and the original byte count instead.
func clampErrPath(path string) string {
	if len(path) <= maxErrPathLen {
		return path
	}
	return fmt.Sprintf("%s… (%d bytes)", path[:maxErrPathLen], len(path))
}
