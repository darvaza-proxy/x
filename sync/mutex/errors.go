package mutex

import (
	"errors"
)

// ErrNilContext is returned when a nil context is encountered and cannot be used.
var ErrNilContext = errors.New("nil context not allowed")

// ErrNilMutex is returned when a nil mutex is encountered and cannot be used.
var ErrNilMutex = errors.New("nil mutex not allowed")
