package autocert

import (
	"crypto/tls"

	"darvaza.org/core"
)

// ServerConfig prepares the TLS configuration for a server using this [Store].
// If no configuration is provided, a new one is created.
func (s *Store) ServerConfig(cfg *tls.Config) (*tls.Config, error) {
	cfg, err := s.doTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// TODO: enable ALPN-01
	return cfg, nil
}

func (s *Store) doTLSConfig(cfg *tls.Config) (*tls.Config, error) {
	if s == nil {
		return nil, core.ErrNilReceiver
	}

	if cfg == nil {
		cfg = new(tls.Config)
	}

	cfg.MinVersion = core.Coalesce(cfg.MinVersion, tls.VersionTLS12)
	cfg.GetCertificate = s.GetCertificate
	cfg.RootCAs = s.GetCAPool()

	return cfg, nil
}
