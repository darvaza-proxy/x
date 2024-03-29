package bind

import (
	"context"
	"io"
	"net"
	"time"
)

var (
	_ Listener       = (*ListenConfig)(nil)
	_ AllListener    = (*ListenConfig)(nil)
	_ TCPListener    = (*ListenConfig)(nil)
	_ AllTCPListener = (*ListenConfig)(nil)
	_ UDPListener    = (*ListenConfig)(nil)
	_ AllUDPListener = (*ListenConfig)(nil)
)

// ListenConfig extends the standard net.ListeConfig with a central holder
// for the Context bound to the listeners
type ListenConfig struct {
	net.ListenConfig

	// Context used when registering the listeners
	Context context.Context
}

// NewListenConfig assists creating a ListenConfig due to the two-layer definition
// making difficult static declaration when `net` is shadowed
func NewListenConfig(ctx context.Context, keepalive time.Duration) *ListenConfig {
	if ctx == nil {
		ctx = context.Background()
	}

	return &ListenConfig{
		ListenConfig: net.ListenConfig{
			KeepAlive: keepalive,
		},
		Context: ctx,
	}
}

// Listen acts like the standard net.Listen but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) Listen(network, addr string) (net.Listener, error) {
	ctx := lc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return lc.ListenConfig.Listen(ctx, network, addr)
}

// ListenPacket acts like the standard net.ListenPacket but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) ListenPacket(network, addr string) (net.PacketConn, error) {
	ctx := lc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return lc.ListenConfig.ListenPacket(ctx, network, addr)
}

// ListenTCP acts like the standard net.ListenTCP but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error) {
	if laddr == nil {
		laddr = &net.TCPAddr{}
	}

	ln, err := lc.Listen(network, laddr.String())
	if err != nil {
		return nil, err
	}

	if ln, ok := ln.(*net.TCPListener); ok {
		return ln, nil
	}

	panic("unreachable")
}

// ListenUDP acts like the standard net.ListenUDP but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) ListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error) {
	if laddr == nil {
		laddr = &net.UDPAddr{}
	}

	ln, err := lc.ListenPacket(network, laddr.String())
	if err != nil {
		return nil, err
	}

	if ln, ok := ln.(*net.UDPConn); ok {
		return ln, nil
	}

	panic("unreachable")
}

// ListenAll acts like Listen but on a list of addresses
func (lc ListenConfig) ListenAll(network string, addrs []string) ([]net.Listener, error) {
	var ok bool

	out := make([]net.Listener, 0, len(addrs))

	// close all on error
	defer func() {
		if !ok {
			closeAll(out)
		}
	}()

	for _, addr := range addrs {
		lsn, err := lc.Listen(network, addr)
		if err != nil {
			return nil, err
		}
		out = append(out, lsn)
	}

	ok = true
	return out, nil
}

// ListenAllPacket acts like ListenPacket but on a list of addresses
func (lc ListenConfig) ListenAllPacket(network string, addrs []string) ([]net.PacketConn, error) {
	var ok bool
	out := make([]net.PacketConn, 0, len(addrs))

	// close all on error
	defer func() {
		if !ok {
			closeAll(out)
		}
	}()

	for _, addr := range addrs {
		lsn, err := lc.ListenPacket(network, addr)
		if err != nil {
			return nil, err
		}
		out = append(out, lsn)
	}

	ok = true
	return out, nil
}

// ListenAllTCP acts like ListenTCP but on a list of addresses
func (lc ListenConfig) ListenAllTCP(network string, laddrs []*net.TCPAddr) (
	[]*net.TCPListener, error) {
	//
	var ok bool
	out := make([]*net.TCPListener, 0, len(laddrs))

	// close all on error
	defer func() {
		if !ok {
			closeAll(out)
		}
	}()

	for _, addr := range laddrs {
		lsn, err := lc.ListenTCP(network, addr)
		if err != nil {
			return nil, err
		}
		out = append(out, lsn)
	}

	ok = true
	return out, nil
}

// ListenAllUDP acts like ListenUDP but on a list of addresses
func (lc ListenConfig) ListenAllUDP(network string, laddrs []*net.UDPAddr) ([]*net.UDPConn, error) {
	var ok bool
	out := make([]*net.UDPConn, 0, len(laddrs))

	// close all on error
	defer func() {
		if !ok {
			closeAll(out)
		}
	}()

	for _, addr := range laddrs {
		lsn, err := lc.ListenUDP(network, addr)
		if err != nil {
			return nil, err
		}
		out = append(out, lsn)
	}

	ok = true
	return out, nil
}

func closeAll[T io.Closer](a []T) {
	for _, v := range a {
		_ = v.Close()
	}
}
