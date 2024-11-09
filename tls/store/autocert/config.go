package autocert

import (
	"context"
	"net/http"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/tls/store/config"
)

const (
	// DefaultDirectoryURL ...
	DefaultDirectoryURL = LetsEncryptStagingURL
)

// StoreOption is a functional option for configuring a Store
type StoreOption func(*Store) error

// Config represents the configuration settings for an autocert store, including ACME directory, authentication, domain validation, and logging options.
type Config struct {
	Logger slog.Logger `json:"-"`

	CacheDIR        string `json:"cache_dir"`
	DisableCacheDIR bool   `json:"cache_disable,omitempty"`

	DirectoryURL string      `json:"acme_url"`
	BearerToken  string      `json:"acme_token"`
	EMail        string      `json:"acme_email" validate:"required,email"`
	EAS          string      `json:"acme_eas,omitempty"`
	Headers      http.Header `json:"acme_headers,omitempty"`

	AcceptTOS bool                                `json:"acme_accept_tos,omitempty"`
	PromptTOS func(context.Context, string) error `json:"-"`

	ValidDomains []string                            `json:"domains,omitempty"`
	ValidDomain  func(context.Context, string) error `json:"-"`
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

	if cfg.DirectoryURL == "" {
		cfg.DirectoryURL = DefaultDirectoryURL
	}

	if cfg.Headers == nil {
		cfg.Headers = make(http.Header)
	}

	return nil
}

// New creates a new [Store] from the config.
func (cfg *Config) New(ctx context.Context) (*Store, error) {
	return New(ctx, cfg)
}
