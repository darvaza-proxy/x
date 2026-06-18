package tls

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"darvaza.org/core"
)

// ErrNoStore is an error indicating the [Store] wasn't provided.
var ErrNoStore = core.Wrap(core.ErrInvalid, "store not provided")

// IsExists reports whether err is the no-op an add returns when the subject was
// already present, leaving the store unchanged. Use it to tell a redundant
// write from a real failure.
func IsExists(err error) bool {
	return errors.Is(err, core.ErrExists)
}

// IsNotExists reports whether err is the miss [StoreReader.Get] returns, or the
// no-op a delete returns, when the subject is not held. Use it to tell an
// absent subject from a real failure.
func IsNotExists(err error) bool {
	return errors.Is(err, core.ErrNotExists)
}

// IsInvalid reports whether err is the rejection a method returns for a missing
// or malformed argument: a nil or corrupt certificate or key, or a missing
// [Store]. Use it to tell bad input from a real failure.
func IsInvalid(err error) bool {
	return errors.Is(err, core.ErrInvalid)
}

// A Store is the source of certificates and trust a [tls.Config] consults during
// the handshake: it presents a certificate for an incoming connection and offers
// the pool of roots to verify peers against. Implementations are living entities,
// designed to be layered — see the package overview. Bind one to a [tls.Config]
// with [WithStore].
//
// Across these interfaces, any method returning an error reports a nil receiver
// with [core.ErrNilReceiver] and a missing or malformed certificate or key with
// [core.ErrInvalid], before making any change.
type Store interface {
	// GetCertificate selects the certificate to present in response to a
	// ClientHello, keyed on its ServerName (SNI). Both its signature and its
	// return semantics are those of [tls.Config.GetCertificate], so a Store
	// wires straight in.
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)

	// GetCAPool returns the trust anchors used to verify peer certificates,
	// consumed as a [tls.Config]'s RootCAs and ClientCAs under its semantics.
	GetCAPool() *x509.CertPool
}

// The Store methods are guarded here against the [tls.Config] fields they feed —
// GetCertificate to GetCertificate, GetCAPool to RootCAs and ClientCAs — so the
// correspondence is explicit and checked at build time. [WithStore] performs the
// actual wiring.
var _ = func(s Store) {
	_ = tls.Config{
		GetCertificate: s.GetCertificate,
		RootCAs:        s.GetCAPool(),
		ClientCAs:      s.GetCAPool(),
	}
}

// StoreReader adds lookup and iteration over the held certificates to the
// [Store].
type StoreReader interface {
	Store

	// Get returns the certificate matching name (a server name, exact or
	// wildcard). Implementations may serve it from an on-demand source when
	// nothing is held; a miss they cannot resolve is reported with
	// [core.ErrNotExists], which [IsNotExists] tests.
	Get(ctx context.Context, name string) (*tls.Certificate, error)

	// ForEach calls fn for every stored certificate, stopping early if fn
	// returns false or ctx is cancelled.
	ForEach(ctx context.Context, fn func(context.Context, *tls.Certificate) bool)

	// ForEachMatch is ForEach restricted to certificates that match name (a
	// server name, exact or wildcard).
	ForEachMatch(ctx context.Context, name string, fn func(context.Context, *tls.Certificate) bool)
}

// StoreWriter adds [tls.Certificate] write methods to the [Store]. Where
// [StoreX509Writer] takes raw building blocks (keys, certificates, chains),
// StoreWriter works in finished units: a [tls.Certificate] already carrying its
// leaf, private key and chain.
type StoreWriter interface {
	Store

	// Put stores a complete certificate, first filling any gaps it can from
	// material already held and verifying the result. A leaf missing its key or
	// intermediates is completed when the store holds them; one that cannot be
	// verified is rejected. Storing a certificate already held is the
	// [core.ErrExists] no-op ([IsExists]); new or renewed material is an
	// effective change.
	Put(ctx context.Context, cert *tls.Certificate) error

	// Delete removes a previously stored certificate; removing one not held is
	// the no-op reported with [core.ErrNotExists] ([IsNotExists]).
	Delete(ctx context.Context, cert *tls.Certificate) error
}

// StoreX509Writer adds [x509.Certificate] write methods to the [Store]. The
// methods form a toolkit in which trust anchors, private keys, certificates and
// complete bundles each have their own entry point.
//
// Writes report whether they had effect, so layers can react to real changes. A
// write that changes nothing is a no-op rather than a failure, reported with a
// sentinel: a duplicate add with [core.ErrExists], tested with [IsExists]; a
// delete of something absent with [core.ErrNotExists], tested with
// [IsNotExists]; a nil return marks an effective change.
type StoreX509Writer interface {
	Store

	// AddCACerts adds the given certificates as trusted roots. They populate the
	// pool returned by [Store.GetCAPool] and are the anchors verification
	// ultimately chains to. Only certificates marked as a CA are accepted; a
	// non-CA is rejected. Passing none is a no-op; roots already trusted count as
	// no change under the convention above.
	AddCACerts(ctx context.Context, roots ...*x509.Certificate) error

	// AddPrivateKey adds a private key on its own, ahead of the certificate it
	// belongs to. It is the companion of [StoreX509Writer.AddCert]: a key added
	// here is matched by public key when its leaf is added afterwards.
	AddPrivateKey(ctx context.Context, key crypto.Signer) error

	// AddCert adds a single certificate, classifying it by what it is:
	//   - self-signed: trusted as a root, as if passed to AddCACerts;
	//   - CA (not self-signed): held as an intermediate, used to complete
	//     chains when a leaf is later verified;
	//   - leaf (non-CA): its private key must already have been added with
	//     AddPrivateKey, otherwise the certificate is rejected.
	// To add a leaf together with its key and chain in one step, use
	// [StoreX509Writer.AddCertPair].
	AddCert(ctx context.Context, cert *x509.Certificate) error

	// AddCertPair adds a leaf certificate together with its private key and any
	// intermediates in a single call, without a prior AddPrivateKey. It is the
	// bundle counterpart of the piecemeal AddPrivateKey + AddCert path. key may
	// be nil, in which case the store resolves it from keys already held, as
	// AddCert does, and rejects the certificate when none matches; intermediates
	// may be empty.
	AddCertPair(ctx context.Context, key crypto.Signer, cert *x509.Certificate, intermediates []*x509.Certificate) error

	// DeleteCert removes a certificate from the store, identified by its DER
	// contents. It is the inverse of AddCert and AddCertPair; removing one not
	// held is the no-op reported with [core.ErrNotExists] ([IsNotExists]).
	DeleteCert(ctx context.Context, cert *x509.Certificate) error
}

// StoreReadWriter includes read and write methods for the [Store].
type StoreReadWriter interface {
	StoreReader
	StoreWriter
}

// WithStore binds a given [Store] to the [tls.Config]
func WithStore(cfg *tls.Config, store Store) error {
	if cfg == nil {
		return fmt.Errorf("missing argument: %s", "cfg")
	}

	if store == nil {
		return fmt.Errorf("missing argument: %s", "store")
	}

	pool := store.GetCAPool()
	if pool == nil {
		return fmt.Errorf("missing parameter: %s", "CAPool")
	}

	cfg.GetCertificate = store.GetCertificate
	cfg.RootCAs = pool
	cfg.ClientCAs = pool
	return nil
}

// NewConfig returns a basic [tls.Config] optionally configured to use the given Store.
func NewConfig(store Store) (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if store != nil {
		if err := WithStore(cfg, store); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// SplitClientHelloInfo takes the context and server name out of a [tls.ClientHelloInfo].
// If no ServerName is provided, the server's IP address will be used.
func SplitClientHelloInfo(chi *tls.ClientHelloInfo) (ctx context.Context, serverName string, err error) {
	if chi == nil {
		return ctx, "", core.ErrInvalid
	}
	ctx = chi.Context()
	serverName = chi.ServerName
	if serverName == "" {
		host, _, _ := core.SplitHostPort(chi.Conn.LocalAddr().String())
		if host != "" {
			serverName = fmt.Sprintf("[%s]", host)
		}
	}

	return ctx, serverName, nil
}
