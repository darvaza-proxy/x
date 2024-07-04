package web

import "net/http"

// A Handler is like [http.Handler] but doesn't render or handle errors
// by itself.
type Handler interface {
	TryServeHTTP(http.ResponseWriter, *http.Request) error
}

var (
	_ http.Handler = HandlerFunc(nil)
	_ Handler      = HandlerFunc(nil)
)

// HandlerFunc is a function that implements [http.Handler] and our [Handler].
// [HandleError] will be called when using the [http.Handler] interface and
// the function returns an error.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (fn HandlerFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := fn(rw, req); err != nil {
		HandleError(rw, req, err)
	}
}

// TryServeHTTP implements the [Handler] interface.
func (fn HandlerFunc) TryServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	return fn(rw, req)
}
