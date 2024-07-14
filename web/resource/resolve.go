package resource

import (
	"context"
	"io/fs"
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web"
)

// Resolve expands on [web.Resolve] using RouteParams() to get the URL params
// from the HTTP Router additionally to the Route Path.
// It will fallback to [web.Resolve] if a RouteParams handlers isn't present,
// and also validates the path.
func Resolve(req *http.Request) (path string, params RouteParamsTable, err error) {
	ctx := req.Context()

	if h, ok := RouteParams(ctx); ok {
		path, params, err = h(ctx)
	} else if h, ok := web.Resolver(ctx); ok {
		path, err = h(req)
	} else {
		path = req.URL.Path
	}

	if err == nil {
		s, ok := web.CleanPath(path)
		if !ok {
			err = web.NewStatusBadRequest(core.QuietWrap(fs.ErrInvalid, "%q: invalid path", path))
		}
		path = s
	}

	if err != nil {
		return "", nil, err
	}

	return path, params, nil
}

// RouteParamsTable is a list a URL params
type RouteParamsTable map[string][]any

// Params lists the names of all stored parameters
func (params RouteParamsTable) Params() []string {
	return core.SortedKeys(params)
}

// Add appends a value for the given parameter
func (params RouteParamsTable) Add(key string, value any) {
	params[key] = append(params[key], value)
}

// Len tells how many parameters are stored in the table
func (params RouteParamsTable) Len(key string) (int, bool) {
	if m, ok := params[key]; ok {
		return len(m), true
	}
	return 0, false
}

// First returns the first stored value for a parameter
func (params RouteParamsTable) First(key string) (any, bool) {
	if m, ok := params[key]; ok {
		if len(m) > 0 {
			return m[0], true
		}
	}
	return nil, false
}

// Last returns the last stored value for a parameter
func (params RouteParamsTable) Last(key string) (any, bool) {
	if m, ok := params[key]; ok {
		if l := len(m); l > 0 {
			return m[l-1], true
		}
	}
	return nil, false
}

// All returns all stored values for a parameter
func (params RouteParamsTable) All(key string) ([]any, bool) {
	m, ok := params[key]
	return m, ok
}

// RouteParamsFunc represents a helper that extracts the request parameters
// as provided by the
type RouteParamsFunc func(context.Context) (path string, params RouteParamsTable, err error)

// WithRouteParams attaches a RouteParams handler to the context
func WithRouteParams(ctx context.Context, h RouteParamsFunc) context.Context {
	return routeParamsCtxKey.WithValue(ctx, h)
}

// RouteParams extracts a RouteParams handlers from the context
func RouteParams(ctx context.Context) (RouteParamsFunc, bool) {
	return routeParamsCtxKey.Get(ctx)
}

// NewRouteParamsMiddleware creates a standard HTTP middleware to attach
// a RouteParams handler to all requests.
func NewRouteParamsMiddleware(h RouteParamsFunc) func(http.Handler) http.Handler {
	if h == nil {
		return web.NoMiddleware
	}

	return web.NewMiddleware(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		ctx := WithRouteParams(req.Context(), h)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

var (
	routeParamsCtxKey = core.NewContextKey[RouteParamsFunc]("RouteParams")
)
