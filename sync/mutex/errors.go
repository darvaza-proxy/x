package mutex

import (
	"errors"
)

// ErrNilMutex is returned when a nil mutex is encountered and cannot be used.
var ErrNilMutex = errors.New("nil mutex not allowed")
