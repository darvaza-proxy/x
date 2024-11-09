package autocert

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	storeConfig "darvaza.org/x/tls/store/config"
)

const (
	// DefaultDirectoryURL ...
	DefaultDirectoryURL = LetsEncryptStagingURL
)

// Config ...
type Config struct {
	Logger slog.Logger `json:"-" toml:"-" yaml:"-"`

	DirectoryURL string

	ValidDomains []string
	ValidDomain  func(context.Context, string) error

	AcceptTOS bool
	PromptTOS func(context.Context, string) error

	TrustedCAs *x509.CertPool
	ClientCert *tls.Certificate
}

func (cfg *Config) export() *storeConfig.Config {
	return &storeConfig.Config{
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
