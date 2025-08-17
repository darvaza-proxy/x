package resource

import (
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web/qlist"
)

// RendererFunc represents a renderer function that can handle custom HTTP status codes.
type RendererFunc[T any] func(http.ResponseWriter, *http.Request, int, T) error

// RenderFunc converts a generic renderer taking any as data type into
// a stricter one taking *T instead.
func RenderFunc[T any](fn func(http.ResponseWriter, *http.Request, any) error) HandlerFunc[T] {
	if fn == nil {
		return nil
	}

	return func(rw http.ResponseWriter, req *http.Request, data T) error {
		return fn(rw, req, data)
	}
}

// Render uses the Accept header to choose what renderer to use. If nothing acceptable
// is supported, but an "identity" type has been set, that will be used.
func (r *Resource[T]) Render(rw http.ResponseWriter, req *http.Request, data T) error {
	preferred, err := r.PreferredMediaType(req)
	if err != nil {
		return err
	}

	// Try code-aware renderer first
	if h := r.getRendererWithCode(preferred); h != nil {
		return h(rw, req, http.StatusOK, data)
	}

	// Fallback to legacy renderer
	if h := r.getRenderer(preferred); h != nil {
		return h(rw, req, data)
	}

	return r.err406()
}

// RenderWithCode uses the Accept header to choose what renderer to use and
// writes the specified HTTP status code. Only works with code-aware renderers.
func (r *Resource[T]) RenderWithCode(rw http.ResponseWriter, req *http.Request, code int, data T) error {
	preferred, err := r.PreferredMediaType(req)
	if err != nil {
		return err
	}

	if h := r.getRendererWithCode(preferred); h != nil {
		return h(rw, req, code, data)
	}

	return r.err406()
}

// PreferredMediaType identifies the best Media Type to serve to a particular
// request. If nothing is acceptable, but an "identity" type has been set,
// that will be returned instead of a 406 error.
func (r *Resource[T]) PreferredMediaType(req *http.Request) (string, error) {
	accepted, err := qlist.ParseMediaRangeHeader(req.Header)
	switch {
	case err != nil:
		// 400
		err = core.Wrap(err, "invalid accept header")
		return "", r.err400(err)
	case len(r.ql) == 0:
		return "", r.err406()
	}

	preferred, _, _ := qlist.BestQualityParsed(r.ql, accepted)
	switch {
	case preferred != "":
		return preferred, nil
	case r.identity != "":
		return r.identity, nil
	default:
		return "", r.err406()
	}
}

func (r *Resource[T]) getRenderer(mediaType string) HandlerFunc[T] {
	return r.r[mediaType]
}

func (r *Resource[T]) getRendererWithCode(mediaType string) RendererFunc[T] {
	return r.rc[mediaType]
}

func addRenderers[T any](r *Resource[T], x any) {
	// JSON - use auto-detection pattern
	trySetJSONForResource(r, x)
	// HTML - use auto-detection pattern
	trySetHTMLForResource(r, x)
	// TXT - use auto-detection pattern
	trySetTXTForResource(r, x)
}
