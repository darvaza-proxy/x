package autocert

import (
	"context"

	"darvaza.org/core"
)

// Start starts the worker loop.
func (*Store) Start(_ context.Context) error { return core.ErrTODO }

// Close stops the worker loop.
func (*Store) Close() error { return core.ErrTODO }

// Shutdown gracefully shuts down the worker loop.
func (*Store) Shutdown(_ context.Context) error { return core.ErrTODO }
