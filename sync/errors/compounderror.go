package errors

import (
	"slices"
	"sync"

	"darvaza.org/core"
)

// CompoundError is the concurrency-safe counterpart of
// [core.CompoundError]. It accumulates errors reported from several
// goroutines and presents them as a single error. The zero value is ready
// to use and must not be copied after first use.
//
// Errors and Unwrap hand out snapshot copies so callers can iterate them
// without holding the lock; AsError returns the receiver itself, whose own
// accessors remain synchronised.
type CompoundError struct {
	errs core.CompoundError
	mu   sync.RWMutex
}

var _ core.Errors = (*CompoundError)(nil)

// AppendError records the non-nil errors, unwrapping nested [core.Errors]
// and Unwrap implementers as [core.CompoundError.AppendError] does, and
// returns the receiver for chaining.
func (w *CompoundError) AppendError(errs ...error) *CompoundError {
	if w == nil {
		return w
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.errs.AppendError(errs...)
	return w
}

// Append records err, optionally annotated with a formatted note, as
// [core.CompoundError.Append] does, and returns the receiver for chaining.
// With a nil err and a non-empty note it records a new error built from the
// note; with both nil and empty it is a no-op.
func (w *CompoundError) Append(err error, note string, args ...any) *CompoundError {
	if w == nil {
		return w
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.errs.Append(err, note, args...)
	return w
}

// Errors returns a snapshot copy of the recorded errors. The caller owns
// the returned slice; later appends do not affect it.
func (w *CompoundError) Errors() []error {
	if w == nil {
		return nil
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	return slices.Clone(w.errs.Errs)
}

// Unwrap returns a snapshot copy of the recorded errors so errors.Is and
// errors.As can traverse them.
func (w *CompoundError) Unwrap() []error {
	return w.Errors()
}

// Error renders the recorded errors as a newline-separated string.
func (w *CompoundError) Error() string {
	if w == nil {
		return ""
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.errs.Error()
}

// OK reports whether no errors have been recorded.
func (w *CompoundError) OK() bool {
	if w == nil {
		return true
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.errs.OK()
}

// AsError returns the receiver when errors have been recorded, and nil
// otherwise, mirroring [core.CompoundError.AsError]. The returned error is
// the live receiver, whose accessors remain synchronised.
func (w *CompoundError) AsError() error {
	if w == nil {
		return nil
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.errs.OK() {
		return nil
	}
	return w
}
