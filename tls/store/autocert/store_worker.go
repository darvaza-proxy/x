package autocert

import (
	"context"

	"darvaza.org/core"
)

func (*Store) Start(_ context.Context) error    { return core.ErrTODO }
func (*Store) Close() error                     { return core.ErrTODO }
func (*Store) Shutdown(_ context.Context) error { return core.ErrTODO }
