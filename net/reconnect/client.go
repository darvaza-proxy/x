// Package reconnect implement a generic retrying TCP client
package reconnect

import (
	"context"
	"sync"

	"darvaza.org/core"
)

// Client is a reconnecting TCP Client
type Client struct {
	mu     sync.Mutex
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelCauseFunc
	err    error
}

// New creates a new [Client].
func New(ctx context.Context) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancelCause(ctx)
	c := &Client{
		ctx:    ctx,
		cancel: cancel,
	}
	return c, nil
}

// Must is like [New] but it panics on errors.
func Must(ctx context.Context) *Client {
	c, err := New(ctx)
	if err != nil {
		core.Panic(err)
	}
	return c
}
