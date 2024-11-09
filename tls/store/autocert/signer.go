package autocert

import (
	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/signer"
)

// Signer ...
type Signer struct {
	// mu     sync.RWMutex
	logger slog.Logger

	key x509utils.PrivateKey
	// cert  []*x509.Certificate
	// chain [][]byte
}

// SignerOption configures a [Signer] within [NewSigner]
type SignerOption func(*Signer) error

// NewSigner ...
func NewSigner(logger slog.Logger, options ...SignerOption) (*Signer, error) {
	cs := &Signer{
		logger: logger,
	}

	if err := cs.unsafeInit(); err != nil {
		return nil, err
	}

	for _, fn := range options {
		if err := fn(cs); err != nil {
			return nil, err
		}
	}

	if err := cs.unsafeBootstrap(); err != nil {
		return nil, err
	}

	return cs, nil
}

// func (*Signer) isInitialized() bool  { core.Panic(core.ErrTODO); return false }
// func (*Signer) isBootstrapped() bool { core.Panic(core.ErrTODO); return false }
func (*Signer) unsafeInit() error { return core.ErrTODO }

func (cs *Signer) unsafeBootstrap() error {
	if cs.key == nil {
		key, err := signer.GenerateED25519Key()
		if err != nil {
			return core.Wrap(err, "failed to generate private key")
		}
		cs.key = key
	}

	return nil
}
