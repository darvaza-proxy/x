package resource

import (
	"bytes"
	"net/http"
	"text/template"
)

func htmlRendererOf[T any](x any) (HandlerFunc[T], bool) {
	if v, ok := x.(interface {
		RenderHTML(http.ResponseWriter, *http.Request, *T) error
	}); ok {
		return v.RenderHTML, true
	}

	return nil, false
}

// RenderHTML compiles an html/template and sends it to the client after setting
// Content-Type and Content-Length.  For HEAD only Content-Type is set.
func RenderHTML(rw http.ResponseWriter, req *http.Request, tmpl *template.Template, data any) error {
	SetHeader(rw, ContentType, HTML)

	if req.Method == "HEAD" {
		// done
		return nil
	}

	return doRenderHTML(rw, tmpl, data)
}

func doRenderHTML(rw http.ResponseWriter, tmpl *template.Template, data any) error {
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	SetHeader(rw, ContentLength, "%v", buf.Len())
	_, err := buf.WriteTo(rw)
	return err
}

// WithHTML is a shortcut for [WithRenderer] for [HTML].
func WithHTML[T any](fn HandlerFunc[T]) OptionFunc[T] {
	return WithRenderer(HTML, fn)
}
