package resource

// Code generated by ./methods.sh; DO NOT EDIT.

//go:generate ./methods.sh

import (
	"net/http"

	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
)

// HandlerFunc represents a function [web.HandlerFunc] but taking a data
// parameter.
type HandlerFunc[T any] func(http.ResponseWriter, *http.Request, T) error

// AsHandlerFunc wraps a [web.HandlerFunc] into a [HandlerFunc], discarding
// the data pointer.
func AsHandlerFunc[T any](fn web.HandlerFunc) HandlerFunc[T] {
	return func(rw http.ResponseWriter, req *http.Request, _ T) error {
		return fn(rw, req)
	}
}

// Getter represents a resource that handles GET requests.
type Getter interface {
	Get(http.ResponseWriter, *http.Request) error
}

// TGetter represents a resource that handles GET requests with a data field.
type TGetter[T any] interface {
	Get(http.ResponseWriter, *http.Request, T) error
}

func getterOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TGetter[T]:
		return v.Get, true
	case Getter:
		return AsHandlerFunc[T](v.Get), true
	default:
		return nil, false
	}
}

// Peeker represents a resource that handles HEAD requests.
type Peeker interface {
	Head(http.ResponseWriter, *http.Request) error
}

// TPeeker represents a resource that handles HEAD requests with a data field.
type TPeeker[T any] interface {
	Head(http.ResponseWriter, *http.Request, T) error
}

func peekerOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TPeeker[T]:
		return v.Head, true
	case Peeker:
		return AsHandlerFunc[T](v.Head), true
	default:
		return nil, false
	}
}

// Poster represents a resource that handles POST requests.
type Poster interface {
	Post(http.ResponseWriter, *http.Request) error
}

// TPoster represents a resource that handles POST requests with a data field.
type TPoster[T any] interface {
	Post(http.ResponseWriter, *http.Request, T) error
}

func posterOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TPoster[T]:
		return v.Post, true
	case Poster:
		return AsHandlerFunc[T](v.Post), true
	default:
		return nil, false
	}
}

// Putter represents a resource that handles PUT requests.
type Putter interface {
	Put(http.ResponseWriter, *http.Request) error
}

// TPutter represents a resource that handles PUT requests with a data field.
type TPutter[T any] interface {
	Put(http.ResponseWriter, *http.Request, T) error
}

func putterOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TPutter[T]:
		return v.Put, true
	case Putter:
		return AsHandlerFunc[T](v.Put), true
	default:
		return nil, false
	}
}

// Deleter represents a resource that handles DELETE requests.
type Deleter interface {
	Delete(http.ResponseWriter, *http.Request) error
}

// TDeleter represents a resource that handles DELETE requests with a data field.
type TDeleter[T any] interface {
	Delete(http.ResponseWriter, *http.Request, T) error
}

func deleterOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TDeleter[T]:
		return v.Delete, true
	case Deleter:
		return AsHandlerFunc[T](v.Delete), true
	default:
		return nil, false
	}
}

// Optioner represents a resource that handles OPTIONS requests.
type Optioner interface {
	Options(http.ResponseWriter, *http.Request) error
}

// TOptioner represents a resource that handles OPTIONS requests with a data field.
type TOptioner[T any] interface {
	Options(http.ResponseWriter, *http.Request, T) error
}

func optionerOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TOptioner[T]:
		return v.Options, true
	case Optioner:
		return AsHandlerFunc[T](v.Options), true
	default:
		return nil, false
	}
}

// Patcher represents a resource that handles PATCH requests.
type Patcher interface {
	Patch(http.ResponseWriter, *http.Request) error
}

// TPatcher represents a resource that handles PATCH requests with a data field.
type TPatcher[T any] interface {
	Patch(http.ResponseWriter, *http.Request, T) error
}

func patcherOf[T any](x any) (HandlerFunc[T], bool) {
	switch v := x.(type) {
	case TPatcher[T]:
		return v.Patch, true
	case Patcher:
		return AsHandlerFunc[T](v.Patch), true
	default:
		return nil, false
	}
}

func addHandlers[T any](h *Resource[T], x any) {
	// GET
	if fn, ok := getterOf[T](x); ok {
		h.h[consts.GET] = fn
	}
	// HEAD
	if fn, ok := peekerOf[T](x); ok {
		h.h[consts.HEAD] = fn
	}
	// POST
	if fn, ok := posterOf[T](x); ok {
		h.h[consts.POST] = fn
	}
	// PUT
	if fn, ok := putterOf[T](x); ok {
		h.h[consts.PUT] = fn
	}
	// DELETE
	if fn, ok := deleterOf[T](x); ok {
		h.h[consts.DELETE] = fn
	}
	// OPTIONS
	if fn, ok := optionerOf[T](x); ok {
		h.h[consts.OPTIONS] = fn
	}
	// PATCH
	if fn, ok := patcherOf[T](x); ok {
		h.h[consts.PATCH] = fn
	}
}
