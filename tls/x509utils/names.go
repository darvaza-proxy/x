package x509utils

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"darvaza.org/core"
)

// Names returns a list of exact names and patterns the certificate
// supports
func Names(cert *x509.Certificate) ([]string, []string) {
	names, patterns := splitDNSNames(cert.DNSNames)
	names = appendIPAddresses(names, cert.IPAddresses)

	// deduplicate
	names = core.SliceUnique(names)
	patterns = core.SliceUnique(patterns)

	return names, patterns
}

func splitDNSNames(dnsNames []string) (names []string, patterns []string) {
	for _, s := range dnsNames {
		s = strings.ToLower(s)

		if strings.HasPrefix(s, "*.") {
			// pattern
			patterns = append(patterns, s[1:])
		} else if s != "" {
			// literal
			names = append(names, s)
		}
	}

	return names, patterns
}

func appendIPAddresses(names []string, addrs []net.IP) []string {
	for _, ip := range addrs {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			if addr.IsValid() {
				name := fmt.Sprintf("[%s]", addr.String())
				names = append(names, name)
			}
		}
	}
	return names
}

// SanitizeName takes a Hostname and returns the name (or address)
// we will use for matching certificates
func SanitizeName(name string) (string, bool) {
	if name != "" {
		// validate and remove port and brackets if present
		if host, _, err := core.SplitHostPort(name); err == nil {
			return doSanitizeName(host)
		}
	}
	return "", false
}

func doSanitizeName(name string) (string, bool) {
	if addr, err := core.ParseAddr(name); err == nil {
		// IP
		addr = addr.Unmap()
		addr = addr.WithZone("")
		name = addr.String()
	} else {
		// Name
		name = removeZone(name)
	}
	return name, len(name) > 0
}

func removeZone(name string) string {
	idx := strings.LastIndexFunc(name, func(r rune) bool {
		return r == '%'
	})
	if idx < 0 {
		return name
	}
	return name[:idx]
}
