package resource

import (
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web/qlist"
)

// Render uses the Accept header to choose what renderer to use. If nothing acceptable
// is supported, but an "identity" type has been set, that will be used.
func (r *Resource[T]) Render(rw http.ResponseWriter, req *http.Request, data *T) error {
	h, err := r.getRendererForRequest(req)
	if err != nil {
		return err
	}

	return h(rw, req, data)
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

func (r *Resource[T]) getRendererForRequest(req *http.Request) (HandlerFunc[T], error) {
	preferred, err := r.PreferredMediaType(req)
	if err != nil {
		return nil, err
	}

	if h := r.getRenderer(preferred); h != nil {
		return h, nil
	}

	return nil, r.err406()
}

func (r *Resource[T]) getRenderer(mediaType string) HandlerFunc[T] {
	return r.r[mediaType]
}
