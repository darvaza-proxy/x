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
