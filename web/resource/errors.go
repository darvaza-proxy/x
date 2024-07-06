package resource

import (
	"darvaza.org/core"
	"darvaza.org/x/web"
)

func (*Resource[T]) err400(err error) *web.HTTPError {
	return web.NewStatusBadRequest(err)
}

func (r *Resource[T]) err405() *web.HTTPError {
	return web.NewStatusMethodNotAllowed(r.methods...)
}

func (*Resource[T]) err406() *web.HTTPError {
	return web.NewStatusNotAcceptable()
}

func (*Resource[T]) wrap400(err error, msg string, args ...any) *web.HTTPError {
	return web.NewStatusBadRequest(core.QuietWrap(err, msg, args...))
}
