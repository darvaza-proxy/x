# Helpers for implementing http.Handlers

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/web.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/web
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/web
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/web
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=web
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/web
[socket-link]: https://socket.dev/go/package/darvaza.org/x/web

## Requests Handling

### Middleware

To facilitate the implementation of standard `func(http.Handler) http.Handler`
middleware, the `MiddlewareFunc` interface and the `NewMiddleware()` factory
were created.

```go
type MiddlewareFunc func(rw http.ResponseWriter, req *http.Request,
    next http.Handler)
```

The `next` argument is never `nil`, and a do-nothing `NoMiddleware` middleware
was introduced. When the `NoMiddleware()` is called without a handler, it will
return a 404 handler.

Alternatively there is `MiddlewareErrorFunc` and `NewMiddlewareError()` that
allows the handler to return an error that is then passed to `HandleError()`
and then to the registered `ErrorHandler`.

### Resolver

We call _Resolver_ a function that will give us the Path our resource should be
handling, and for this task `darvaza.org/x/web` provides four helpers.

* `WithResolver()` to attach a dedicated _Resolver_ to the request's context.
* `NewResolverMiddleware()` to attach one to every request.
* `Resolver()`, to retrieve a previously attached _Resolver_ from the request's
  context.
* and a `Resolve()` helper that will use the above and call the specified
  _Resolver_, or take the request's `URL.Path`, and then clean it to make sure
  its safe to use.
* `CleanPath()` cleans and validates the path for `URL.Path` handling.

### RESTful Handlers

The `darvaza.org/x/web/resource` sub-package offers a `Resource[T]` wrapper to
implement a RESTful interface to a particular resource.

## Response Handlers

Using `respond.WithRequest()` we compute our options and
`PreferredContentType()` tells one how to encode the data.

## Content Negotiation

### QualityList

The QualityList parser allows choosing the best option during Content
Negotiation, e.g. accepted `Content-Type`s.

### BestQuality

`qlist` offers two helpers to choose the best option from a QualityList and a
list of supported options, `BestQuality()` and `BestQualityWithIdentity()`.
_Identity_ is an special option we consider unless it's explicitly forbidden.

### BestEncoding

`qlist.BestEncoding()` is a special case of `BestQualityWithIdentity()`
using the `Accept` header, and falling back to `"identity"` as magic type.

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).

### See also

* [Accept][mdn-accept]
* [Content Negotiation][mdn-content-negotiation]
* [Quality Values][mdn-quality-values]

[mdn-accept]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
[mdn-content-negotiation]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation
[mdn-quality-values]: https://developer.mozilla.org/en-US/docs/Glossary/Quality_values

## HTTP Errors

### `HTTPError`

`HTTPError{}` is an `http.Handler` that is also an _error_ and can be used to
build HTTP errors.

### Error Handlers

`darvaza.org/x/web` provides a mechanism to hook an HTTP error handler to the
request Context.

* `WithErrorHandler()` to attach a
  `func(http.ResponseWriter, *http.Request, error)`
* `NewErrorHandlerMiddleware()` to attach it to every request,
* and `ErrorHandler()` to read it back.

We also provide a basic implementation called `HandleError` which will first
attempt to get a better handler for the context, via
`ErrorHandler(req.Context())` and hand it over.
If there is no `ErrorHandlerFunc` in the context it will test if the _error_
itself via the `http.Handler` interface and invoke it.

As last resort `HandleError()` will check if the error provides an
`HTTPStatus() int` method to infer the HTTP status code of the error, and if
negative or undefined it will assume it's a 500, compose a `web.HTTPError` and
serve it.

### Error Factories

* `AsError()` that will do the same as `HandleError()` to ensure the given
  error, if any, is `http.Handler`-able
* and `AsErrorWithCode()` to **suggest** an HTTP status code to be used
  instead of 500 when it can't be determined.

There are also `web.HTTPError` factories to create new errors, from a generic:

* `NewHTTPError()` and `NewHTTPErrorf()` and a companion `ErrorText(code)`
  helper.

to redirect factories from formatted strings:

* `NewStatusMovedPermanently(loc, ...)` (301)
* `NewStatusFound(loc, ...)` (302)
* `NewStatusSeeOther(loc, ...)` (303)
* `NewStatusTemporaryRedirect(loc, ...)` (307)
* `NewStatusPermanentRedirect(loc, ...)` (308)

error wrappers:

* `NewStatusBadRequest(err)` (400)

and simple responses:

* `NewStatusNotModified()` (304)
* `NewStatusBadRequest(err)` (400)
* `NewStatusNotFound()` (404)
* `NewStatusMethodNotAllowed(allowed...)` (405)
* `NewStatusNotAcceptable()` (406)
