package config

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils"
)

// Config provides an easy way to populate a [tls.Store]
type Config struct {
	Logger slog.Logger

	Roots []string
	Keys  []string
	Certs []string
}

func (cfg *Config) getLogger() slog.Logger {
	if cfg == nil {
		return nil
	}

	return cfg.Logger
}

// AddCACerts adds all given certificates as trusted roots to the given [tls.Store].
// PEM content, a PEM fileName, or a directory containing PEM files.
func (cfg *Config) AddCACerts(ctx context.Context, out tls.StoreX509Writer, value string) (int, error) {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddCertsCallback()
	if err := x509utils.ReadStringPEM(value, fn); err != nil {
		errs.AppendError(err)
	}

	n, err := buf.AddCACerts(ctx, out)
	if err != nil {
		errs.AppendError(err)
	}

	return n, errs.AsError()
}

func (cfg *Config) applyRoots(ctx context.Context, out tls.StoreX509Writer) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddCertsCallback()

	for _, v := range cfg.Roots {
		if v != "" {
			if err := x509utils.ReadStringPEM(v, fn); err != nil {
				errs.Append(err, "Roots")
			}
		}
	}

	if _, err := buf.AddCACerts(ctx, out); err != nil {
		errs.Append(err, "AddCACerts")
	}

	return errs.AsError()
}

// AddCert adds all given certificates to the specified [tls.Store].
func (cfg *Config) AddCert(ctx context.Context, out tls.StoreX509Writer, value string) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddCertsCallback()
	if err := ReadStringPEM(value, fn); err != nil {
		errs.AppendError(err)
	}

	if _, err := buf.AddCert(ctx, out); err != nil {
		errs.AppendError(err)
	}

	return errs.AsError()
}

func (cfg *Config) applyCerts(ctx context.Context, out tls.StoreX509Writer) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddCertsCallback()
	for _, v := range cfg.Certs {
		if v != "" {
			if err := x509utils.ReadStringPEM(v, fn); err != nil {
				errs.Append(err, "Certs")
			}
		}
	}

	if _, err := buf.AddCert(ctx, out); err != nil {
		errs.Append(err, "AddCert")
	}

	return errs.AsError()
}

// AddCertPair adds a cert-key pair to the specified [tls.Store].
func (cfg *Config) AddCertPair(ctx context.Context, out tls.StoreX509Writer, key, cert string) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn1 := buf.NewAddPrivateKeysCallback()
	if err := ReadStringPEM(key, fn1); err != nil {
		errs.AppendError(err)
	}

	fn2 := buf.NewAddCertsCallback()
	if err := ReadStringPEM(cert, fn2); err != nil {
		errs.AppendError(err)
	}

	if _, err := buf.AddCertPair(ctx, out); err != nil {
		errs.AppendError(err)
	}

	return errs.AsError()
}

// AddPrivateKey adds a private key to the specified [tls.Store].
func (cfg *Config) AddPrivateKey(ctx context.Context, out tls.StoreX509Writer, key string) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddPrivateKeysCallback()

	if err := ReadStringPEM(key, fn); err != nil {
		errs.AppendError(err)
	}

	if _, err := buf.AddPrivateKey(ctx, out); err != nil {
		errs.AppendError(err)
	}

	return errs.AsError()
}

func (cfg *Config) applyKeys(ctx context.Context, out tls.StoreX509Writer) error {
	var errs core.CompoundError

	buf := buffer.New(ctx, cfg.getLogger())
	fn := buf.NewAddPrivateKeysCallback()

	for _, v := range cfg.Keys {
		if v != "" {
			if err := ReadStringPEM(v, fn); err != nil {
				errs.Append(err, "Keys")
			}
		}
	}

	if _, err := buf.AddPrivateKey(ctx, out); err != nil {
		errs.Append(err, "AddPrivateKey")
	}

	return errs.AsError()
}

// Apply adds roots, keys and certs to the given [tls.Store].
func (cfg *Config) Apply(ctx context.Context, s tls.StoreX509Writer) error {
	switch {
	case cfg == nil:
		return core.ErrNilReceiver
	case s == nil:
		return tls.ErrNoStore
	default:
		var errs core.CompoundError

		for _, fn := range []func(context.Context, tls.StoreX509Writer) error{
			cfg.applyRoots,
			cfg.applyKeys,
			cfg.applyCerts,
		} {
			if err := fn(ctx, s); err != nil {
				errs.AppendError(err)
			}
		}

		return errs.AsError()
	}
}
