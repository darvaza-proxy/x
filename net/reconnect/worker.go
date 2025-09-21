package reconnect

import (
	"context"
	"time"

	"darvaza.org/core"
)

// WorkerFunc is a run function for core.ErrGroup's GoCatch.
type WorkerFunc func(context.Context) error

// CatcherFunc is a catch function for core.ErrGroup's GoCatch.
type CatcherFunc func(context.Context, error) error

// A WorkGroup is an error group interface.
type WorkGroup interface {
	Go(...WorkerFunc)
	GoCatch(WorkerFunc, CatcherFunc)

	Shutdown(context.Context) error

	Wait() error
	Done() <-chan struct{}
	Err() error
}

// A Shutdowner is an object that provides a Shutdown method
// that takes a context with deadline to shut down all associated
// workers.
type Shutdowner interface {
	Shutdown(context.Context) error
}

// NewShutdownFunc creates a shutdown [WorkerFunc], optionally
// with a deadline.
func NewShutdownFunc(s Shutdowner, tio time.Duration) WorkerFunc {
	switch {
	case s == nil:
		// nothing to call
		return nil
	case tio > 0:
		// graceful shutdown timeout
		return func(ctx context.Context) error {
			<-ctx.Done()

			deadline := time.Now().Add(tio)
			ctx2, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()

			return s.Shutdown(ctx2)
		}
	default:
		// just wait
		return func(ctx context.Context) error {
			<-ctx.Done()
			return s.Shutdown(context.Background())
		}
	}
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
