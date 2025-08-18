package resource

import (
	"bytes"
	"html/template"
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
)

// HTMLRenderer represents the legacy HTML renderer interface
type HTMLRenderer[T any] interface {
	RenderHTML(http.ResponseWriter, *http.Request, T) error
}

// HTMLRendererWithCode represents the code-aware HTML renderer interface
type HTMLRendererWithCode[T any] interface {
	RenderHTML(http.ResponseWriter, *http.Request, int, T) error
}

// trySetHTMLForResource tries both code-aware and legacy interfaces for HTML
func trySetHTMLForResource[T any](r *Resource[T], x any) {
	// Try code-aware interface first (with status code parameter)
	if v, ok := x.(HTMLRendererWithCode[T]); ok {
		_ = r.addRendererWithCode(consts.HTML, v.RenderHTML)
		return
	}

	// Fallback to legacy interface (without status code parameter)
	if v, ok := x.(HTMLRenderer[T]); ok {
		_ = r.addRenderer(consts.HTML, v.RenderHTML)
	}
}

// RenderHTML compiles an html/template and sends it to the client after setting
// Content-Type (if not already set) and Content-Length with a given HTTP status code.
// For HEAD only Content-Type is set.
func RenderHTML(rw http.ResponseWriter, req *http.Request, code int, tmpl *template.Template, data any) error {
	SetHeaderUnlessExists(rw, consts.ContentType, consts.HTML)

	switch {
	case code < 0:
		code = http.StatusInternalServerError
	case code == 0:
		code = http.StatusOK
	}

	if req.Method == consts.HEAD {
		// done
		if code == http.StatusOK {
			code = http.StatusNoContent
		}

		rw.WriteHeader(code)
		return nil
	}

	return doRenderHTML(rw, tmpl, code, data)
}

func doRenderHTML(rw http.ResponseWriter, tmpl *template.Template, code int, data any) error {
	if tmpl == nil {
		return core.ErrInvalid
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	SetHeader(rw, consts.ContentLength, "%v", buf.Len())
	rw.WriteHeader(code)

	_, err := buf.WriteTo(rw)
	return err
}

// WithHTML is a shortcut for [WithRendererCode] for [consts.HTML].
func WithHTML[T any](fn RendererFunc[T]) OptionFunc[T] {
	return WithRendererCode(consts.HTML, fn)
}
