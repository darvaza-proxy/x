package web

import (
	"context"
	"net/http"
	"strings"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

// A ResolverFunc attempts to extract the path from a [http.Request].
type ResolverFunc func(*http.Request) (string, error)

// WithResolver attaches a [ResolverFunc] to a context for
// later retrieval
func WithResolver(ctx context.Context, h ResolverFunc) context.Context {
	return resolverCtxKey.WithValue(ctx, h)
}

// Resolver attempts to get a [ResolverFunc] from the given context
func Resolver(ctx context.Context) (ResolverFunc, bool) {
	return resolverCtxKey.Get(ctx)
}

// Resolve extracts the routing path from the request using
// the Resolver in the context when available.
//
// The resulting path is cleaned and validated before
// returning it.
//
// The Resolver either returns the route path, or a [web.HTTPError]
// usually of one one of the following codes:
// * 30x (Redirects)
// * 400 (Bad request)
// * 401 (Login Required)
// * 403 (Permission Denied)
// * 404 (Resource Not Found),
// * or 500 (Server Error)
func Resolve(req *http.Request) (path string, err error) {
	// resolve
	r, _ := Resolver(req.Context())
	if r == nil {
		// default resolver
		path = req.URL.Path
	} else if path, err = r(req); err != nil {
		// resolver error.
		return "", AsError(err)
	}

	if s, ok := CleanPath(path); ok {
		// good path
		return s, nil
	}

	// bad path
	err = core.QuietWrap(fs.ErrInvalid, "%q: invalid path", path)
	return "", NewStatusBadRequest(err)
}

// CleanPath cleans and validates a URL.Path
func CleanPath(path string) (string, bool) {
	s, _ := fs.Clean(path)
	switch {
	case s == "" || s[0] != '/':
		return "", false
	case strings.HasSuffix(s, "/.."), strings.HasPrefix(s, "/../"):
		return "", false
	default:
		return s, true
	}
}

var (
	resolverCtxKey = core.NewContextKey[ResolverFunc]("Resolver")
)
