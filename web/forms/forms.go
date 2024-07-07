// Package forms assists working with web forms and Request.Form
package forms

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"darvaza.org/x/web"
)

// DefaultFormMaxMemory indicates the memory limit when parsing a form
// used when ParseForm is called without a positive number.
const DefaultFormMaxMemory = 1 << 20 // 1MiB

// ParseForm is similar to the standard request.ParseForm() but it
// handles urlencoded, multipart and JSON.
// For nested JSON objects ParseForm uses dots to join keys.
func ParseForm(req *http.Request, maxMemory int64) (err error) {
	if req.Form != nil {
		// ready
		return nil
	}

	ct := getContentType(req)
	switch ct {
	case "application/x-www-form-urlencoded":
		err = req.ParseForm()
	case "multipart/form-data":
		err = req.ParseMultipartForm(maxMemory)
	case "application/json":
		err = parseFormJSON(req, maxMemory)
	default:
		return &web.HTTPError{Code: http.StatusUnsupportedMediaType}
	}

	return web.AsErrorWithCode(err, http.StatusBadRequest)
}

// ReadAll read the whole body of a request but fails if it exceeds
// the given limit.
// If no limit is provided, DefaultFormMaxMemory will be used.
func ReadAll(body io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultFormMaxMemory
	}

	r := &io.LimitedReader{
		R: body,
		N: maxBytes + 1,
	}

	b, err := io.ReadAll(r)
	switch {
	case err != nil:
		return b, err
	case int64(len(b)) > maxBytes:
		return nil, errors.New("size limit exceeded")
	default:
		return b, nil
	}
}

func parseFormJSON(req *http.Request, maxMemory int64) error {
	b, err := ReadAll(req.Body, maxMemory)
	if err != nil {
		return err
	}

	values, err := UnmarshalFormJSON(b)
	if err != nil {
		return err
	}

	switch strings.ToUpper(req.Method) {
	case "POST", "PUT", "PATCH":
		req.PostForm = cloneValues(values)
	default:
		req.PostForm = make(url.Values)
	}

	// TODO: add query string values
	req.Form = values

	return nil
}

func cloneValues(orig url.Values) url.Values {
	out := make(url.Values)
	for k, s := range orig {
		s2 := make([]string, len(s))
		copy(s2, s)
		out[k] = s2
	}
	return out
}

func getContentType(req *http.Request) string {
	s := strings.Split(req.Header.Get("Content-Type"), ";")[0]
	return strings.ToLower(strings.TrimSpace(s))
}
