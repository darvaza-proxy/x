# Agent Documentation for x/web

## Overview

The `web` package provides comprehensive HTTP handler utilities that extend
Go's standard net/http package. It includes middleware patterns, content
negotiation, error handling, RESTful resource management, and static asset
serving with a focus on clean, composable web development.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Features

- **Middleware**: Composable middleware with error handling support.
- **Content Negotiation**: Quality-based content type selection.
- **Error Handling**: Structured HTTP error management.
- **Resource Framework**: RESTful resource handler generation.
- **Asset Management**: Static file serving with caching.

### Main Types

- **`Handler`**: Enhanced http.Handler with error returns.
- **`MiddlewareFunc`**: Middleware function signature.
- **`HTTPError`**: Error type that implements http.Handler.
- **`Resource[T]`**: Generic RESTful resource wrapper.

### Subpackages

#### assets Package

- Static file serving with embedded support.
- Checksum-based caching.
- MIME type handling.
- Filesystem abstractions.

#### forms Package

- Form parsing with JSON support.
- Type-safe value extraction.
- Validation helpers.

#### qlist Package

- Quality list parsing for Accept headers.
- Content negotiation algorithms.
- Media range matching.

#### resource Package

- RESTful resource handler generation.
- Method-based routing.
- Content type rendering.
- Automatic OPTIONS handling.

#### respond Package

- Response builders for common patterns.
- Success/error responses.
- Redirect helpers.
- Content type registry.

#### html Package

- HTML-specific utilities.
- Template integration.
- HTML filesystem serving.

## Architecture Notes

The package follows several design principles:

1. **Composability**: All components work together seamlessly.
2. **Type Safety**: Generic types for compile-time safety.
3. **Error First**: Explicit error handling throughout.
4. **Context Aware**: Full context.Context integration.
5. **Standards Compliant**: Follows HTTP RFCs closely.

Key patterns:

- Middleware functions receive "next" handler for chaining.
- Resources use reflection to discover handler methods.
- Content negotiation uses quality values from Accept headers.
- Errors implement http.Handler for direct serving.

## Development Commands

For common development commands and workflow, see the [root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:

- HTTP handler behaviour.
- Content negotiation accuracy.
- Error response formatting.
- Middleware composition.
- Form parsing edge cases.

## Common Usage Patterns

### Basic Middleware

```go
// Create middleware
logMiddleware := web.NewMiddleware(
    func(w http.ResponseWriter, r *http.Request, next http.Handler) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })

// Apply to handler
handler := logMiddleware(myHandler)
```

### Error Handling

```go
// Create error handler middleware
errorMiddleware := web.NewErrorHandlerMiddleware(
    func(w http.ResponseWriter, r *http.Request, err error) {
        web.HandleError(w, r, err)
    })

// Return errors from handlers
handler := web.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
    if err := validateRequest(r); err != nil {
        return web.NewStatusBadRequest(err)
    }
    return nil
})
```

### RESTful Resources

```go
type UserResource struct {
    db *Database
}

func (u *UserResource) Check(r *http.Request) (*http.Request, *User, error) {
    // Validate and load user
    user, err := u.db.GetUser(r.URL.Path)
    return r, user, err
}

func (u *UserResource) Get(w http.ResponseWriter, r *http.Request,
    user *User) error {
    return resource.RenderJSON(w, r, user)
}

func (u *UserResource) Put(w http.ResponseWriter, r *http.Request,
    user *User) error {
    // Update user
    return nil
}

// Create resource handler
res := resource.New(&UserResource{db: db},
    resource.WithRenderer("application/json", resource.RenderJSON),
    resource.WithIdentity("application/json"),
)
```

### Content Negotiation

```go
// Parse Accept header
accepts := qlist.ParseMediaRanges(r.Header.Get("Accept"))

// Find best match
supported := []string{"application/json", "text/html"}
best := qlist.BestQuality(accepts, supported...)

switch best {
case "application/json":
    respond.JSON(w, r, data)
case "text/html":
    respond.HTML(w, r, data)
default:
    http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
}
```

### Static Assets

```go
// Embed static files
//go:embed static/*
var staticFS embed.FS

// Create asset handler
assets := assets.New(staticFS,
    assets.WithPrefix("/static/"),
    assets.WithCache(24 * time.Hour),
)

// Serve assets
http.Handle("/static/", assets)
```

### Form Handling

```go
// Parse form with JSON support
form, err := forms.Parse(r)
if err != nil {
    return web.NewStatusBadRequest(err)
}

// Extract values
name := form.String("name")
age := form.Int("age", 0)
active := form.Bool("active")
```

### HTTP Errors

```go
// Create typed errors
notFound := web.NewStatusNotFound()
badRequest := web.NewStatusBadRequest(err)
redirect := web.NewStatusFound("/login")

// Custom error
customErr := web.NewHTTPErrorf(http.StatusTeapot, "I'm a teapot: %v", reason)

// Handle in middleware
if err != nil {
    if httpErr, ok := err.(web.HTTPError); ok {
        httpErr.ServeHTTP(w, r)
    } else {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

## Performance Characteristics

- **Middleware**: Minimal overhead per layer.
- **Content Negotiation**: O(n*m) for n accepts and m supported.
- **Resource Routing**: O(1) method lookup.
- **Asset Serving**: Efficient caching with ETags.

## Dependencies

- `darvaza.org/core`: Core utilities.
- Standard library (net/http, encoding/json, mime).

## See Also

- [resource README](resource/README.md) for RESTful resource framework details.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
- HTTP RFCs for standards compliance.
