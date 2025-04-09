package utils

import (
	"errors"

	"darvaza.org/core"
)

// ErrNilContext is returned when a nil context is encountered and cannot be used.
var ErrNilContext = errors.New("nil context not allowed")

// ErrNilMutex is returned when a nil mutex is encountered and cannot be used.
var ErrNilMutex = errors.New("nil mutex not allowed")

// ErrNilReceiver is returned when a nil receiver is encountered and cannot be used.
var ErrNilReceiver = core.ErrNilReceiver

// NewNilReceiverPanic creates a new error wrapping a nil receiver panic with
// an optional note.
// The skip parameter controls the stack trace depth.
func NewNilReceiverPanic(skip int, note string) error {
	return core.NewPanicWrap(skip+1, ErrNilReceiver, note)
}
