package autocert

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/tls/store/config"
)

const (
	// DefaultDirectoryURL ...
	DefaultDirectoryURL = LetsEncryptStagingURL
)

// Config ...
type Config struct {
	Logger   slog.Logger
	CacheDir string

	DirectoryURL string
	EMail        string
	BearerToken  string
	ClientCert   *tls.Certificate
	EAS          string

	TrustedCAs *x509.CertPool

	AcceptTOS bool
	PromptTOS func(context.Context, string) error

	ValidDomains []string
	ValidDomain  func(context.Context, string) error
}

func (cfg *Config) export() *config.Config {
	return &config.Config{
		Logger: cfg.Logger,
	}
}

// SetDefaults fills any gap in the [Config].
func (cfg *Config) SetDefaults() error {
	if cfg == nil {
		return core.ErrNilReceiver
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	return nil
}

// New creates a new [Store] from the config.
func (cfg *Config) New() (*Store, error) {
	return New(cfg)
}
