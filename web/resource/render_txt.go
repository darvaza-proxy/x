package resource

import (
	"fmt"
	"net/http"

	"darvaza.org/x/web/consts"
)

// TXTRenderer represents the legacy TXT renderer interface
type TXTRenderer[T any] interface {
	RenderTXT(http.ResponseWriter, *http.Request, T) error
}

// TXTRendererWithCode represents the code-aware TXT renderer interface
type TXTRendererWithCode[T any] interface {
	RenderTXT(http.ResponseWriter, *http.Request, int, T) error
}

// trySetTXTForResource tries both code-aware and legacy interfaces for TXT
func trySetTXTForResource[T any](r *Resource[T], x any) {
	// Try code-aware interface first (with status code parameter)
	if v, ok := x.(TXTRendererWithCode[T]); ok {
		_ = r.addRendererWithCode(consts.TXT, v.RenderTXT)
		return
	}

	// Fallback to legacy interface (without status code parameter)
	if v, ok := x.(TXTRenderer[T]); ok {
		_ = r.addRenderer(consts.TXT, v.RenderTXT)
	}
}

// RenderTXT renders plain text with the specified HTTP status code.
func RenderTXT(rw http.ResponseWriter, req *http.Request, code int,
	data any) error {
	SetHeaderUnlessExists(rw, consts.ContentType, consts.TXT)

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

	text := fmt.Sprintf("%v", data)
	SetHeader(rw, consts.ContentLength, "%v", len(text))
	rw.WriteHeader(code)

	_, err := rw.Write([]byte(text))
	return err
}

// WithTXT is a shortcut for [WithRendererCode] for [consts.TXT].
func WithTXT[T any](fn RendererFunc[T]) OptionFunc[T] {
	if fn == nil {
		fn = func(rw http.ResponseWriter, req *http.Request, code int, data T) error {
			return RenderTXT(rw, req, code, data)
		}
	}
	return WithRendererCode(consts.TXT, fn)
}
