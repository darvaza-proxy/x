package forms

import (
	"net/http"
	"net/url"
	"strings"

	"darvaza.org/core"
)

// FormValue reads a field from req.Form, after populating it if needed,
// and returning the trimmed string, an indicator saying if it was actually
// present, or an error if ParseForm failed.
func FormValue[T core.String](req *http.Request, field string) (value T, found bool, err error) {
	if err = ParseForm(req, 0); err == nil {
		value, found = doFormValue[T](req.Form, field)
	}
	return value, found, err
}

func doFormValue[T core.String](values url.Values, field string) (T, bool) {
	ss, ok := values[field]
	if !ok || len(ss) == 0 {
		return "", false
	}

	s := strings.TrimSpace(ss[0])
	return T(s), true
}

// FormValues reads a field from req.Form, after populating it if needed,
// and returning the trimmed strings, an indicator saying if it was actually
// present, or an error if ParseForm failed. Empty values will be omitted.
func FormValues[T core.String](req *http.Request, field string) (values []T, found bool, err error) {
	if err = ParseForm(req, 0); err == nil {
		values, found = doFormValues[T](req.Form, field)
	}
	return values, found, err
}

func doFormValues[T core.String](values url.Values, field string) ([]T, bool) {
	ss, ok := values[field]
	if !ok {
		return []T{}, false
	}

	out := make([]T, 0, len(ss))
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, T(s))
		}
	}
	return out, true
}

// PostFormValue reads a field from req.PostForm, after populating it if needed,
// and returning the trimmed string, an indicator saying if it was actually
// present, or an error if ParseForm failed.
func PostFormValue[T core.String](req *http.Request, field string) (value T, found bool, err error) {
	if err = ParseForm(req, 0); err == nil {
		value, found = doFormValue[T](req.PostForm, field)
	}
	return value, found, err
}

// PostFormValues reads a field from req.PostForm, after populating it if needed,
// and returning the trimmed strings, an indicator saying if it was actually
// present, or an error if ParseForm failed. Empty values will be omitted.
func PostFormValues[T core.String](req *http.Request, field string) (values []T, found bool, err error) {
	if err = ParseForm(req, 0); err == nil {
		values, found = doFormValues[T](req.PostForm, field)
	}
	return values, found, err
}

// FormValueFn reads a field from req.Form, and processes it through a helper function
// to get the value.
// FormValueFn also indicates if the field is missing, or if ParseForm or the conversion
// failed.
func FormValueFn[T any, S core.String](req *http.Request, field string,
	fn func(S) (T, error)) (T, bool, error) {
	//
	if err := ParseForm(req, 0); err != nil {
		var zero T
		return zero, false, err
	}

	return doFormValueFn[T, S](req.Form, field, fn)
}

func doFormValueFn[T any, S core.String](values url.Values, field string,
	fn func(S) (T, error)) (T, bool, error) {
	//
	s, found := doFormValue[string](values, field)
	if !found {
		var zero T
		return zero, false, nil
	}

	v, err := fn(S(s))
	if err != nil {
		return v, true, core.Wrap(err, field)
	}

	return v, true, nil
}

// FormValuesFn reads a field from req.Form, and processes all values through a helper function
// to get the values.
// FormValuesFn also indicates if the field is missing, or if ParseForm or the conversion
// failed.
func FormValuesFn[T any, S core.String](req *http.Request, field string,
	fn func(S) (T, error)) ([]T, bool, error) {
	//
	if err := ParseForm(req, 0); err != nil {
		return nil, false, err
	}

	return doFormValuesFn[T, S](req.Form, field, fn)
}

func doFormValuesFn[T any, S core.String](values url.Values, field string,
	fn func(S) (T, error)) ([]T, bool, error) {
	//
	ss, found := doFormValues[string](values, field)
	if !found {
		return nil, false, nil
	}

	out := make([]T, 0, len(ss))
	for _, s := range ss {
		v, err := fn(S(s))
		if err != nil {
			return out, true, core.Wrap(err, field)
		}

		out = append(out, v)
	}

	return out, true, nil
}

// PostFormValueFn reads a field from req.PostForm, and processes it through a helper function
// to get the value.
// PostFormValueFn also indicates if the field is missing, or if ParseForm or the conversion
// failed.
func PostFormValueFn[T any, S core.String](req *http.Request, field string,
	fn func(S) (T, error)) (T, bool, error) {
	//
	if err := ParseForm(req, 0); err != nil {
		var zero T
		return zero, false, err
	}

	return doFormValueFn[T, S](req.PostForm, field, fn)
}

// PostFormValuesFn reads a field from req.PostForm, and processes all values through a helper function
// to get the values.
// PostFormValuesFn also indicates if the field is missing, or if ParseForm or the conversion
// failed.
func PostFormValuesFn[T any, S core.String](req *http.Request, field string,
	fn func(S) (T, error)) ([]T, bool, error) {
	//
	if err := ParseForm(req, 0); err != nil {
		return nil, false, err
	}

	return doFormValuesFn[T, S](req.PostForm, field, fn)
}
