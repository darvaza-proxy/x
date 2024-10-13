package certbotdir

import (
	"path/filepath"

	"darvaza.org/slog"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// https://eff-certbot.readthedocs.io/en/latest/using.html#where-are-my-certificates

const (
	// LiveDirectory ...
	LiveDirectory = "live"

	// PrivateKeyFile ...
	PrivateKeyFile = "privkey.pem"
	// FullChainFile ...
	FullChainFile = "fullchain.pem"
	// CertFile ...
	CertFile = "cert.pem"
	// ChainFile ...
	ChainFile = "chain.pem"

	// SelfSignedCertFile ...
	SelfSignedCertFile = "self-signed-" + CertFile
	// SelfSignedKeyFile ...
	SelfSignedKeyFile = "self-signed-" + PrivateKeyFile
	// SelfSignedID ...
	SelfSignedID = "."
)

// Config provides parameters to assemble the [Store].
type Config struct {
	Logger slog.Logger
	Roots  x509utils.CertPool

	BaseDirectory string
	LiveDirectory string
}

// SetDefaults fills gaps in the [Config].
func (cfg *Config) SetDefaults() error {
	if cfg.BaseDirectory == "" {
		cfg.BaseDirectory = DefaultBaseDirectory
	}

	if cfg.LiveDirectory == "" {
		cfg.LiveDirectory = filepath.Join(cfg.BaseDirectory, LiveDirectory)
	}

	if cfg.Roots == nil {
		p, err := certpool.SystemCertPool()
		if err != nil {
			return err
		}
		cfg.Roots = p
	}

	return nil
}

// FilePair returns fullchain.pem and privkey.pem file paths for the specified domain.
// This function doesn't clean the output. domain, BaseDirectory and LiveDirectory
// are expected to be sanitized and cleaned.
func (cfg *Config) FilePair(domain string) (certFile, keyFile string) {
	if domain == "" || domain == SelfSignedID {
		certFile = unsafeJoin(cfg.BaseDirectory, SelfSignedCertFile)
		keyFile = unsafeJoin(cfg.BaseDirectory, SelfSignedKeyFile)
	} else {
		domDir := unsafeJoin(cfg.LiveDirectory, domain)
		certFile = unsafeJoin(domDir, FullChainFile)
		keyFile = unsafeJoin(domDir, PrivateKeyFile)
	}
	return certFile, keyFile
}

// New uses the [Config] to create a new [Store].
func (cfg *Config) New() (*Store, error) {
	return New(cfg)
}
