package autocert

import (
	"crypto/tls"

	"darvaza.org/core"
)

// NewConfig ...
//
// revive:disable:flag-parameter
func (s *Store) NewConfig(zeroTrust bool) (*tls.Config, error) {
	// revive:enable:flag-parameter
	if s == nil {
		return nil, core.ErrNilReceiver
	}

	cfg := &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: s.GetCertificate,
		RootCAs:        s.GetCAPool(),
	}

	if zeroTrust {
		cfg.ClientCAs = cfg.RootCAs
	}

	return cfg, nil
}
