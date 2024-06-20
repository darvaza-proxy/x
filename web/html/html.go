// Package html facilitates use of [html/template.Template]
package html

import (
	"html/template"
	"io"
	"io/fs"
)

// Template is a type-safe variant of [template.Template]
type Template[T any] template.Template

// Execute writes the result of combining the provided data
// with the [Template].
func (t *Template[T]) Execute(w io.Writer, data T) error {
	return t.Sys().Execute(w, data)
}

// Clone makes a copy of a typed template.
func (t *Template[T]) Clone() (*Template[T], error) {
	t2, err := t.Sys().Clone()
	if err != nil {
		return nil, err
	}

	return (*Template[T])(t2), nil
}

// Sys returns the underlying [template.Template]
func (t *Template[T]) Sys() *template.Template {
	if t == nil {
		return nil
	}

	return (*template.Template)(t)
}

// Lookup attempts to find a named template and returns the type-safe
// variant.
func Lookup[T any](t *template.Template, name string) (*Template[T], error) {
	if t != nil {
		if t2 := t.Lookup(name); t2 != nil {
			return (*Template[T])(t2), nil
		}
	}

	err := &fs.PathError{
		Op:   "lookup",
		Path: name,
		Err:  fs.ErrNotExist,
	}
	return nil, err
}
