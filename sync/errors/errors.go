// Package errors provides synchronisation-related error definitions.
package errors

import (
	"errors"

	"darvaza.org/core"
)

// ErrAlreadyInitialised indicates initialisation cannot proceed because the
// target is already initialised.
var ErrAlreadyInitialised = errors.New("already initialised")

// ErrNotInitialised indicates operations cannot proceed because the target
// has not been initialised.
var ErrNotInitialised = errors.New("not initialised")

// ErrNilContext indicates operations cannot proceed with a nil context.
var ErrNilContext = errors.New("nil context not allowed")

// ErrNilMutex indicates operations cannot proceed with a nil mutex reference.
var ErrNilMutex = errors.New("nil mutex not allowed")

// ErrNilReceiver is returned when a nil receiver is encountered and cannot be used.
var ErrNilReceiver = core.ErrNilReceiver
