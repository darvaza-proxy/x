package x509utils

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"

	"darvaza.org/core"
)

// Names returns a list of exact names and patterns the certificate
// supports
func Names(cert *x509.Certificate) (names, patterns []string) {
	names, patterns = splitDNSNames(cert.DNSNames)
	names = appendIPAddresses(names, cert.IPAddresses)

	// deduplicate
	names = core.SliceUnique(names)
	patterns = core.SliceUnique(patterns)

	return names, patterns
}

func splitDNSNames(dnsNames []string) (names, patterns []string) {
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

// Hostname returns a sanitised hostname for a parsed URL
func Hostname(u *url.URL) (string, bool) {
	return SanitizeName(u.Host)
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

// NameAsIP prepares a sanitised IP address name for matching certificates
func NameAsIP(name string) (string, bool) {
	if addr, err := core.ParseAddr(name); err == nil {
		s := fmt.Sprintf("[%s]", addr)
		return s, true
	}
	return "", false
}

// NameAsSuffix prepares a sanitised hostname for matching
// certificate patterns
func NameAsSuffix(name string) (string, bool) {
	if idx := strings.IndexRune(name, '.'); idx > 0 {
		name = name[idx:]
		return name, len(name) > 1
	}
	return "", false
}
