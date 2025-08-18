package resource

import (
	"bytes"
	"net/http"
	"text/template"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
)

// TemplateRenderer represents the legacy template renderer interface
type TemplateRenderer[T any] interface {
	RenderTemplate(http.ResponseWriter, *http.Request, T) error
}

// TemplateRendererWithCode represents the code-aware template renderer interface
type TemplateRendererWithCode[T any] interface {
	RenderTemplate(http.ResponseWriter, *http.Request, int, T) error
}

// RenderTemplate compiles a text/template and sends it to the client with a given HTTP status code.
// Content-Type must be set by the caller before calling this function.
// Content-Length is set automatically. For HEAD requests only status code is written.
func RenderTemplate(rw http.ResponseWriter, req *http.Request, code int, tmpl *template.Template, data any) error {
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

	return doRenderTemplate(rw, tmpl, code, data)
}

func doRenderTemplate(rw http.ResponseWriter, tmpl *template.Template, code int, data any) error {
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

// WithTemplate is a helper for registering template renderers with a specific media type.
// The mediaType parameter specifies the Content-Type that the template will output.
func WithTemplate[T any](mediaType string, fn RendererFunc[T]) OptionFunc[T] {
	return WithRendererCode(mediaType, fn)
}
