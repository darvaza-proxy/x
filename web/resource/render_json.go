package resource

import (
	"bytes"
	"encoding/json"
	"net/http"

	"darvaza.org/x/web/consts"
)

// JSONRenderer represents the legacy JSON renderer interface
type JSONRenderer[T any] interface {
	RenderJSON(http.ResponseWriter, *http.Request, T) error
}

// JSONRendererWithCode represents the code-aware JSON renderer interface
type JSONRendererWithCode[T any] interface {
	RenderJSON(http.ResponseWriter, *http.Request, int, T) error
}

// trySetJSONForResource tries both code-aware and legacy interfaces for JSON
func trySetJSONForResource[T any](r *Resource[T], x any) {
	// Try code-aware interface first (with status code parameter)
	if v, ok := x.(JSONRendererWithCode[T]); ok {
		_ = r.addRendererWithCode(consts.JSON, v.RenderJSON)
		return
	}

	// Fallback to legacy interface (without status code parameter)
	if v, ok := x.(JSONRenderer[T]); ok {
		_ = r.addRenderer(consts.JSON, v.RenderJSON)
	}
}

// RenderJSON encodes the data as JSON and sends it to the client after setting
// Content-Type (if not already set) and Content-Length with the specified HTTP status code.
// For HEAD only Content-Type is set.
func RenderJSON(rw http.ResponseWriter, req *http.Request, code int, data any) error {
	SetHeaderUnlessExists(rw, consts.ContentType, consts.JSON)

	switch {
	case code < 0:
		code = http.StatusInternalServerError
	case code == 0:
		code = http.StatusOK
	}

	if req.Method == consts.HEAD {
		if code == http.StatusOK {
			code = http.StatusNoContent
		}
		rw.WriteHeader(code)
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	SetHeader(rw, consts.ContentLength, "%v", len(b))
	rw.WriteHeader(code)

	buf := bytes.NewBuffer(b)
	_, err = buf.WriteTo(rw)
	return err
}

// WithJSON is a shortcut for [WithRendererCode] for [JSON].
// If no custom handler is provided, the generic [RenderJSON] will
// be used.
func WithJSON[T any](fn RendererFunc[T]) OptionFunc[T] {
	if fn == nil {
		fn = func(rw http.ResponseWriter, req *http.Request, code int, data T) error {
			return RenderJSON(rw, req, code, data)
		}
	}
	return WithRendererCode(consts.JSON, fn)
}
