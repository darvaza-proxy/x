# RESTful resource handler

## Resource

`Resource` wraps an object that implements method handlers as actual methods of a class.
`Resource` will provide the `http.Handler` and `web.Handler` entry points, and handle
errors and `OPTIONS` and 403 errors directly.

`Resource` supports a special `Check()` method to validate a prepare the request
before being passed the the dedicated method handler.

If `Check` isn't implemented, `WithChecker()` can be used during the `New` call to
set one. Otherwise the `DefaultChecker` will be used.

`Resource` will check for `Get`, `Head`, `Post`, `Put`, `Delete`, `Options` and
`Patch` methods, but `WithMethod()` can be used during the `New` call for setting
custom ones or replacing/deleting a detected one.

If the `Peek` method doesn't exist, `Get` will be used.
If the `Options` method doesn't exist, a simple implementation will be provided.

`Resource` offers a `Methods()` method with the list of supported HTTP methods. This slice can be safely modified.

`Resource` deals with `403 Method No Allowed` directly, and with bad `req.URL.Path`s
when using the `DefaultChecker`

### Signatures

`Resource` supports two signatures for `Check` and also two signatures for the method handlers.

The recommended workflow is to use `Check` to return a resource data pointer using
```go
Check(*http.Request) (*http.Request, *T, error)
```
but one with implicit `nil` data is also tested for.
```go
Check(*http.Request) (*http.Request, error)
```

The data pointer can then be received by the FOO method handler via
```go
Foo(http.ResponseWriter, *http.Request, *T) error
```

but one without data pointer is also tested for.
```go
Foo(http.ResponseWriter, *http.Request) error
```

### JSON

`Resource` extends the standard `req.Form` handling with it's own `ParseForm()` method that reads
the content-type, handles JSON content, and returns an HTTP 400 error in case of problems.

### Renderers

`Resource` allows the registration of supported Media Types using the `WithRenderer()` option,
and then the `Render()` method will call the correct one after checking the request's preference.

If one wants to return a particular type when none of the supported media types are acceptable,
it can be specified using the `WithIdentity()` option.

For convenience a `RenderJSON` and `RenderHTML` helpers are provided.

`New()` will automatically test for `RenderJSON` and `RenderHTML` methods in the object
and registered `JSON` and `HTML` renderers for the resource, but the _identity_
representation won't be assumed.

## Helpers

* `RenderJSON`
* `RenderHTML`
* `SetHeader`
