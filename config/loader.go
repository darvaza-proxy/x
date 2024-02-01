package config

import (
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

// An Option if a function used to configure an object after decoding.
type Option[T any] func(*T) error

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

	// Options are applied to objects after decoding and
	// before Load() returns.
	Options []Option[T]
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
			// success
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
			return l.applyOptions(v)
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

// New creates a new object
func (l *Loader[T]) New() (*T, error) {
	l.remember(nil, "")
	v := new(T)
	return l.applyOptions(v)
}

func (l *Loader[T]) applyOptions(v *T) (*T, error) {
	for _, opt := range l.Options {
		if err := opt(v); err != nil {
			return nil, NewPathError("", "init", err)
		}
	}

	if err := SetDefaults(v); err != nil {
		return nil, NewPathError("", "init", err)
	}

	if err := Validate(v); err != nil {
		return nil, NewPathError("", "validate", err)
	}

	return v, nil
}

// Join combines a list of directories with a name and an optional list of extensions.
// This are `/` separated, absolute, but without the initial `/`.
// Following [fs.ValidPath] rules. Final result is cleaned.
func Join(directories []string, base string, extensions []string) ([]string, error) {
	l := len(directories)
	names, err := joinedNames(base, extensions)

	if err != nil || l == 0 {
		return names, err
	}

	out := make([]string, 0, len(names)*l)
	for _, dir := range directories {
		switch {
		case dir == "":
			dir = "."
		case fs.ValidPath(dir):
			dir = path.Clean(dir)
		default:
			return out, joinInvalid(dir)
		}

		for _, name := range names {
			out = append(out, path.Join(dir, name))
		}
	}

	return out, nil
}

func joinedNames(base string, extensions []string) ([]string, error) {
	if base == "" || strings.Contains(base, "/") {
		return nil, joinInvalid(base)
	}

	l := len(extensions)
	if l == 0 {
		return []string{base}, nil
	}

	out := make([]string, 0, l)
	for _, ext := range extensions {
		s := joinName(base, ext)
		if !fs.ValidPath(s) {
			return out, joinInvalid(s)
		}

		out = append(out, s)
	}

	return out, nil
}

func joinName(base, ext string) string {
	switch {
	case ext == "":
		return base
	case ext[0] == '.':
		return base + ext
	default:
		return base + "." + ext
	}
}

func joinInvalid(name string) *fs.PathError {
	return NewPathError(name, "join", fs.ErrInvalid)
}
