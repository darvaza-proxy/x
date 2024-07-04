package resource

import (
	"darvaza.org/core"
	"darvaza.org/x/web"
)

func (r *Resource[T]) err405() *web.HTTPError {
	return web.NewStatusMethodNotAllowed(r.methods...)
}

func (*Resource[T]) wrap400(err error, msg string, args ...any) *web.HTTPError {
	return web.NewStatusBadRequest(core.QuietWrap(err, msg, args...))
}
