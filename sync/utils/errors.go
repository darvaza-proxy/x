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
