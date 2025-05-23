package errors

import "errors"

// New creates a new error with the given text using the standard library
// errors package.
// This function exists because we shadow the errors package.
func New(text string) error {
	return errors.New(text)
}

// Is reports whether err is or wraps target, using the standard library
// errors package.
// This function exists because we shadow the errors package.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As reports whether err is or wraps a value of target type, using the
// standard library errors package.
// This function exists because we shadow the errors package.
func As(err error, target any) bool {
	return errors.As(err, target)
}
