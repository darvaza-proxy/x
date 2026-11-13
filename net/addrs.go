// Package net provides generic network helpers and proxies
// to some useful standard types and functions
package net

import (
	"net"
	"net/netip"
	"strconv"

	"darvaza.org/core"
)

// SplitHostPort splits a network address into host and port,
// validating the host and port in the process, and returns the host in
// the canonical ASCII form core.SplitHostPortASCII gives it.
func SplitHostPort(hostport string) (string, uint16, error) {
	host, port, err := core.SplitHostPortASCII(hostport)
	switch {
	case err != nil:
		return "", 0, err
	case port == "":
		return host, 0, nil
	default:
		// core.SplitHostPortASCII returns a port of decimal digits up to 65535
		u := core.Must(strconv.ParseUint(port, 10, 16))
		return host, uint16(u), nil
	}
}

// CleanDialAddress validates a host:port address to dial, and returns
// it normalised, in the canonical ASCII form core.SplitHostPortASCII
// gives the host and port, and an IPv4-mapped IPv6 address without a
// zone as the IPv4 address it wraps. An empty host, which dials the local
// system, is kept, and an unspecified address keeps its family. It
// refuses an address with no port, and port 0.
func CleanDialAddress(hostport string) (string, error) {
	host, port, err := core.SplitHostPortASCII(hostport)
	switch {
	case err != nil:
		// bad host or port
		return "", err
	case port == "":
		// missing port
		return "", addrErr(hostport, "missing port")
	case port == "0":
		// port 0, which the split returns without its padding
		return "", addrErr(hostport, "invalid port")
	default:
		// the address checked, normalised
		return net.JoinHostPort(unmapHost(host), port), nil
	}
}

// unmapHost returns a host the split gave as an IP address unmapped as
// unmapAddr does, in its canonical text, and any other host as it is.
func unmapHost(host string) string {
	if addr, err := netip.ParseAddr(host); err == nil {
		// IP address
		return unmapAddr(addr).String()
	}
	// name, or empty
	return host
}

// unmapAddr returns an IPv4-mapped IPv6 address without a zone as the
// IPv4 address it wraps, and any other address as it is, its zone kept.
func unmapAddr(addr netip.Addr) netip.Addr {
	if addr.Zone() == "" {
		// Unmap leaves an address alone unless IPv4-mapped
		return addr.Unmap()
	}
	// zone kept
	return addr
}

// addrErr returns a *net.AddrError naming the address refused and why.
func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}

// JoinHostPort joins a host and a port, the host in the canonical ASCII
// form core.SplitHostPortASCII gives it, and an IPv4-mapped IPv6 address
// without a zone as the IPv4 address it wraps. An empty host is kept, an
// unspecified address keeps its family, and a host carrying a port of
// its own is refused.
func JoinHostPort(host string, port uint16) (string, error) {
	host, err := stringifyAddr(host)
	if err != nil {
		return "", err
	}

	return doJoinHostPort(host, port), nil
}

// doJoinHostPort joins a host stringifyAddr returned and a port.
func doJoinHostPort(host string, port uint16) string {
	return net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
}

// stringifyAddr checks a host to join with a port, and returns it
// without brackets, normalised as JoinHostPort describes. An empty host
// is kept.
func stringifyAddr(host string) (string, error) {
	if host == "" {
		// empty host, kept
		return "", nil
	}

	s, port, err := core.SplitHostPortASCII(host)
	switch {
	case err != nil:
		// bad host
		return "", err
	case port != "":
		// a port where the host goes
		return "", addrErr(host, "unexpected port")
	default:
		// the host checked, normalised
		return unmapHost(s), nil
	}
}

// JoinAllHostPorts joins every host with every port, as JoinHostPort
// does, host by host. It stops at the first host refused, and returns
// the addresses joined before it with the error.
func JoinAllHostPorts(addresses []string, ports []uint16) ([]string, error) {
	var out []string

	for _, s := range addresses {
		addr, err := stringifyAddr(s)
		if err != nil {
			return out, err
		}

		for _, p := range ports {
			out = append(out, doJoinHostPort(addr, p))
		}
	}

	return out, nil
}
