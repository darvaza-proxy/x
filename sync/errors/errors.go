// Package errors provides synchronisation-related error definitions.
package errors

import "darvaza.org/core"

const (
	// ErrAlreadyInitialised indicates initialisation cannot proceed
	// because the target is already initialised.
	ErrAlreadyInitialised core.StringError = "already initialised"

	// ErrNotInitialised indicates operations cannot proceed because the
	// target has not been initialised.
	ErrNotInitialised core.StringError = "not initialised"

	// ErrClosed indicates operations cannot proceed because the target is
	// closed.
	ErrClosed core.StringError = "closed"

	// ErrNotLocked indicates an unlock cannot proceed because the target
	// is not locked.
	ErrNotLocked core.StringError = "not locked"

	// ErrReadLocked indicates an exclusive unlock cannot proceed because
	// the target is locked for reading.
	ErrReadLocked core.StringError = "read-locked"

	// ErrNilContext indicates operations cannot proceed with a nil
	// context.
	ErrNilContext core.StringError = "nil context not allowed"

	// ErrNilFunction indicates operations cannot proceed with a nil
	// function.
	ErrNilFunction core.StringError = "nil function not allowed"

	// ErrNilMutex indicates operations cannot proceed with a nil mutex
	// reference.
	ErrNilMutex core.StringError = "nil mutex not allowed"
)

// ErrNilReceiver is returned when a nil receiver is encountered and cannot be used.
var ErrNilReceiver = core.ErrNilReceiver
