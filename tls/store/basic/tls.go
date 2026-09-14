package basic

import (
	"crypto/tls"

	"darvaza.org/core"
)

// NewConfig creates a new [tls.Config] linked to the [Store].
func (s *Store) NewConfig() (*tls.Config, error) {
	if s == nil {
		return nil, core.ErrNilReceiver
	}

	cfg := &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: s.GetCertificate,
		RootCAs:        s.GetCAPool(),
	}

	return cfg, nil
}
