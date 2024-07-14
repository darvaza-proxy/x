package resource

import (
	"bytes"
	"html/template"
	"net/http"

	"darvaza.org/x/web/consts"
)

func htmlRendererOf[T any](x any) (HandlerFunc[T], bool) {
	if v, ok := x.(interface {
		RenderHTML(http.ResponseWriter, *http.Request, T) error
	}); ok {
		return v.RenderHTML, true
	}

	return nil, false
}

// RenderHTML compiles an html/template and sends it to the client after setting
// Content-Type and Content-Length.  For HEAD only Content-Type is set.
func RenderHTML(rw http.ResponseWriter, req *http.Request, tmpl *template.Template, data any) error {
	return RenderHTMLWithCode(rw, req, 0, tmpl, data)
}

// RenderHTMLWithCode compiles an html/template and sends it to the client after setting
// Content-Type and Content-Length with a given HTTP status code.
// For HEAD only Content-Type is set.
func RenderHTMLWithCode(rw http.ResponseWriter, req *http.Request, code int, tmpl *template.Template, data any) error {
	SetHeader(rw, consts.ContentType, consts.HTML)

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
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	SetHeader(rw, consts.ContentLength, "%v", buf.Len())
	rw.WriteHeader(code)

	_, err := buf.WriteTo(rw)
	return err
}

// WithHTML is a shortcut for [WithRenderer] for [consts.HTML].
func WithHTML[T any](fn HandlerFunc[T]) OptionFunc[T] {
	return WithRenderer(consts.HTML, fn)
}
