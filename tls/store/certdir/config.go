package certdir

import (
	"darvaza.org/core"
	"darvaza.org/slog"
)

const (
	// LiveDirectory ...
	LiveDirectory = "live"
	// WorkDirectory ...
	WorkDirectory = "work"

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

type Config struct {
	Logger        slog.Logger
	BaseDirectory string

	LRUSize int
}

func (c *Config) SetDefaults() error {
	if c == nil {
		return core.ErrNilReceiver
	}

	return core.ErrTODO
}
