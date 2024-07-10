package net

import (
	"context"
	"net"
)

var _ Dialer = &net.Dialer{}

// Dialer establishes a TCP connection to a remote.
type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}
