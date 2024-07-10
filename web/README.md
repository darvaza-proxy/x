# Helpers for implementing http.Handlers

## Response Handlers

Using `respond.WithRequest()` we compute our options and `PreferredContentType()`
tells one how to encode the data.

## Content Negotiation

### QualityList

The QualityList parser allows choosing the best option during Content Negotiation, e.g. accepted `Content-Type`s.

### BestQuality

`qlist` offers two helpers to choose the best option from a QualityList and a list of
supported options, `BestQuality()` and `BestQualityWithIdentity()`. _Identity_ is an special
option we consider unless it's explicitly forbidden.

### BestEncoding

`qlist.BestEncoding()` is a special case of `BestQualityWithIdentity()` using the `Accept`
header, and falling back to `"identity"` as magic type.

### See also

* [Accept](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept)
* [Content Negotiation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation)
* [Quality Values](https://developer.mozilla.org/en-US/docs/Glossary/Quality_values)

## HTTP Errors

### `HTTPError`

`HTTPError{}` is an `http.Handler` that is also an _error_ and can be used to build HTTP errors.

### Error Handlers

`darvaza.org/x/web` provides a mechanism to hook an HTTP error handler to the request Context.

* `WithErrorHandler()` to attach a `func(http.ResponseWriter, *http.Request, error)`
* and `ErrorHandler()` to read it back.

We also provide a basic implementation called `HandleError` which will first attempt
to get a better handler for the context, via `ErrorHandler(req.Context())` and hand it over.
If there is no `ErrorHandlerFunc` in the context it will test if the _error_
itself via the `http.Handler` interface and invoke it.

As last resort `HandleError()` will check if the error provides an `HTTPStatus() int` method
to infer the HTTP status code of the error, and if negative or undefined it will assume
it's a 500, compose a `web.HTTPError` and serve it.

### Error Factories

* `AsError()` that will do the same as `HandleError()` to ensure the given error, if any,
  error is `http.Handler`-able
* and `AsErrorWithCode()` to **suggest** an HTTP status code to be used instead of 500
  when it can't be determined.

There are also `web.HTTPError` factories to create new errors, from a generic:

* `NewHTTPError()` and `NewHTTPErrorf()` and a companion `ErrorText(code)` helper.

to redirect factories from formatted strings:

* `NewStatusMovedPermanently(loc, ...)` (301)
* `NewStatusFound(loc, ...)` (302)
* `NewStatusSeeOther(loc, ...)` (303)
* `NewStatusTemporaryRedirect(loc, ...)` (307)
* `NewStatusPermanentRedirect(loc, ...)` (308)
