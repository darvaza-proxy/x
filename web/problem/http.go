package problem

import (
	"bytes"
	"encoding/json"
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
)

// TryServeHTTP returns the Problem if it's an error, or render a 204 if it isn't.
func TryServeHTTP(h web.Error, rw http.ResponseWriter, _ *http.Request) error {
	if err := core.AsError(h); err != nil {
		// pass errors through
		return err
	}

	// 204
	applyHeaders(rw, h)
	rw.WriteHeader(http.StatusNoContent)

	return nil
}

// ServeHTTP renders JSON Problems.
func ServeHTTP(h web.Error, rw http.ResponseWriter, req *http.Request) {
	err := TryServeHTTP(h, rw, req)
	if err == nil {
		return
	} else if h == err {
		// Handler is the error, render directly.
		RenderHTTP(h, rw, req)
	} else {
		// other errors, pass to handler
		web.HandleError(rw, req, err)
	}
}

// RenderHTTP renders a web.Error as JSON using the [RFC7807]
// Content-Type
func RenderHTTP(h web.Error, rw http.ResponseWriter, req *http.Request) {
	statusCode := h.HTTPStatus()

	hdr := applyHeaders(rw, h)
	web.SetHeader(hdr, consts.ContentType, JSON)

	if req.Method == consts.HEAD {
		rw.WriteHeader(statusCode)
		return
	}

	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		web.NewStatusInternalServerError(err).ServeHTTP(rw, req)
		return
	}

	web.SetHeader(hdr, consts.ContentLength, "%v", len(b))
	rw.WriteHeader(statusCode)

	buf := bytes.NewBuffer(b)
	_, _ = buf.WriteTo(rw)
}

func applyHeaders(rw http.ResponseWriter, vv ...any) http.Header {
	hdr := rw.Header()

	for _, v := range vv {
		for k, ee := range getHeaders(v) {
			hdr[k] = append(hdr[k], ee...)
		}
	}

	return hdr
}

func getHeaders(v any) http.Header {
	switch x := v.(type) {
	case interface {
		Header() http.Header
	}:
		return x.Header()
	case interface {
		Headers() http.Header
	}:
		return x.Headers()
	default:
		return nil
	}
}
