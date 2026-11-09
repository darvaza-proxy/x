package net

import (
	"context"
	"net"
)

var _ Dialer = &net.Dialer{}

// Dialer establishes a connection to a remote.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
