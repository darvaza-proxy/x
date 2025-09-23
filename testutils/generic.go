package testutils

import "darvaza.org/core"

// TypeTestFunc represents a validation function for testing factory-created objects.
type TypeTestFunc[T any] func(core.T, *T) bool
