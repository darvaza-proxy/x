package resource

import (
	"bytes"
	"encoding/json"
	"net/http"

	"darvaza.org/x/web/consts"
)

func jsonRendererOf[T any](x any) (HandlerFunc[T], bool) {
	if v, ok := x.(interface {
		RenderJSON(http.ResponseWriter, *http.Request, T) error
	}); ok {
		return v.RenderJSON, true
	}

	return nil, false
}

// RenderJSON encodes the data as JSON and sends it to the client after setting
// Content-Type and Content-Length.  For HEAD only Content-Type is set.
func RenderJSON(rw http.ResponseWriter, req *http.Request, data any) error {
	SetHeader(rw, consts.ContentType, consts.JSON)

	if req.Method == consts.HEAD {
		// done
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	SetHeader(rw, consts.ContentLength, "%v", len(b))

	buf := bytes.NewBuffer(b)
	_, err = buf.WriteTo(rw)
	return err
}

// WithJSON is a shortcut for [WithRenderer] for [JSON].
// If no custom handler is provided, the generic [RenderJSON] will
// be used.
func WithJSON[T any](fn HandlerFunc[T]) OptionFunc[T] {
	if fn == nil {
		fn = RenderFunc[T](RenderJSON)
	}
	return WithRenderer(consts.JSON, fn)
}
