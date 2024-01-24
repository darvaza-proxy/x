package config

import (
	"io"
	"io/fs"
	"os"
)

// Loader tries to load an object from the first success on a list
// of options.
type Loader[T any] struct {
	lastFS   fs.FS
	lastName string

	// NewDecoder returns a [Decoder] based on the filename
	NewDecoder func(string) (Decoder[T], error)

	// IsSkip checks if the error returned by the [Decoder]
	// indicates we should try the next option instead of
	// failing. [os.IsNotExist] is always tested first.
	IsSkip func(error) bool
}

// Last returns the filename last used. empty if it was the
// Fallback.
func (l *Loader[T]) Last() (fs.FS, string) {
	return l.lastFS, l.lastName
}

func (l *Loader[T]) remember(fSys fs.FS, name string) {
	l.lastFS = fSys
	l.lastName = name
}

// NewFromFile returns the first successfully decoded option.
func (l *Loader[T]) NewFromFile(fSys fs.FS, names ...string) (*T, error) {
	if l.NewDecoder != nil {
		v, err := l.tryLoad(fSys, names)
		switch {
		case err != nil:
			return nil, NewPathError(l.lastName, "load", err)
		case v != nil:
			return v, nil
		}
	}

	return nil, NewPathError("", "load", fs.ErrInvalid)
}

func (l *Loader[T]) tryLoad(fSys fs.FS, names []string) (*T, error) {
	for _, name := range names {
		l.remember(fSys, name)
		v, err := l.doReadDecode(fSys, name)
		switch {
		case err == nil:
			return v, nil
		case os.IsNotExist(err), l.IsSkip != nil && l.IsSkip(err):
			continue
		default:
			return nil, NewPathError(name, "decode", err)
		}
	}
	return nil, nil
}

func (l *Loader[T]) doReadDecode(fSys fs.FS, name string) (*T, error) {
	data, err := fs.ReadFile(fSys, name)
	if err != nil {
		return nil, err
	}

	dec, err := l.NewDecoder(name)
	switch {
	case err != nil:
		return nil, err
	case dec == nil:
		return nil, fs.ErrNotExist
	}

	if d, ok := dec.(io.Closer); ok {
		defer d.Close()
	}

	return dec.Decode(name, data)
}
