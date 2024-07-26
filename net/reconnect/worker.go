package reconnect

import (
	"context"

	"darvaza.org/core"
)

// WorkerFunc is a run function for core.ErrGroup's GoCatch
type WorkerFunc func(context.Context) error

// CatcherFunc is a catch function for core.ErrGroup's GoCatch
type CatcherFunc func(context.Context, error) error

// A WorkGroup is an error group interface
type WorkGroup interface {
	Go(...WorkerFunc)
	GoCatch(WorkerFunc, CatcherFunc)

	Shutdown(context.Context) error

	Wait() error
	Done() <-chan struct{}
	Err() error
}

// NewCatchFunc creates a [CatcherFunc] turning any of the given
// errors into nil.
func NewCatchFunc(nonErrors ...error) CatcherFunc {
	if len(nonErrors) == 0 {
		// NO-OP
		return func(_ context.Context, err error) error {
			return err
		}
	}

	return func(_ context.Context, err error) error {
		switch {
		case err == nil:
			return nil
		case core.IsError(err, nonErrors...):
			return nil
		default:
			return err
		}
	}
}
