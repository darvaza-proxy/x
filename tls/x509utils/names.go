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
// supports. A nil certificate yields no names.
func Names(cert *x509.Certificate) (names, patterns []string) {
	if cert == nil {
		return nil, nil
	}

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
		// AddrFromSlice reports ok only for 4- or 16-byte slices, both of
		// which yield a valid address; Unmap drops any 4-in-6 mapping so the
		// stored key matches the form the lookup paths produce.
		if addr, ok := netip.AddrFromSlice(ip); ok {
			name := fmt.Sprintf("[%s]", addr.Unmap())
			names = append(names, name)
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
	if addr, err := sanitizeAddr(name); err == nil {
		name = addr.String()
	}
	// A non-IP name arrives already validated by SplitHostPort, whose idna
	// check rejects '%', so it cannot carry a scope zone — nothing to strip.
	return name, len(name) > 0
}

// sanitizeAddr parses an IP address and canonicalises it into the single form
// the certificate name keys use — dropping any 4-in-6 mapping and scope zone —
// so the storage (Names) and lookup (SanitizeName/NameAsIP) paths agree.
func sanitizeAddr(name string) (netip.Addr, error) {
	addr, err := core.ParseAddr(name)
	if err != nil {
		return netip.Addr{}, err
	}
	return addr.Unmap().WithZone(""), nil
}

// NameAsIP prepares a sanitised IP address name for matching certificates
func NameAsIP(name string) (string, bool) {
	if addr, err := sanitizeAddr(name); err == nil {
		return fmt.Sprintf("[%s]", addr), true
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
