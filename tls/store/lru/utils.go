package lru

import (
	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// sanitizeName validates and sanitizes a server name, returning a cleaned version or an error if invalid.
func sanitizeName(name string) (string, error) {
	s, ok := x509utils.SanitizeName(name)
	if !ok {
		return "", core.Wrapf(core.ErrInvalid, "serverName: %q", name)
	}
	return s, nil
}
