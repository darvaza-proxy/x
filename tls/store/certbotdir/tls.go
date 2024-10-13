package certbotdir

import "crypto/tls"

// NewConfig returns a fresh [tls.Config] linked to this [Store].
func (s *Store) NewConfig() (*tls.Config, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: s.GetCertificate,
		RootCAs:        s.GetCAPool(),
	}

	return cfg, nil
}
