package errors

import "errors"

// New creates a new error with the given text using the standard library errors package.
// It exists because we are shadowing the errors package.
func New(text string) error {
	return errors.New(text)
}
