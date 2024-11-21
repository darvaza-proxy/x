// Package buffer provides helpers to decode PEM files, populate a [tls.StoreWriter],
// and work with key and cert sets
package buffer

import (
	"context"
	"sync"

	"darvaza.org/slog"

	"darvaza.org/x/tls/x509utils/certpool"
)

// Buffer is a PEM decoding buffer to populate a [tls.StoreWriter].
type Buffer struct {
	mu     sync.Mutex
	ctx    context.Context
	logger slog.Logger // TODO: use

	keySet  *KeySet
	certSet *certpool.CertSet
	sources map[SourceName]*Source
}

// New creates a PEM decoding Buffer to populate a [tls.StoreWriter].
func New(ctx context.Context, logger slog.Logger) *Buffer {
	return &Buffer{
		ctx:    ctx,
		logger: logger,
	}
}

func (buf *Buffer) unsafeInit() {
	if buf.ctx == nil {
		buf.ctx = context.Background()
	}
	if buf.keySet == nil {
		buf.keySet = MustKeySet()
	}
	if buf.certSet == nil {
		buf.certSet = certpool.MustCertSet()
	}
	if buf.sources == nil {
		buf.sources = make(map[SourceName]*Source)
	}
}

// Clone creates a copy of the [Buffer]. It returns nil if the receiver is
// nil of if it fails to initialize.
func (buf *Buffer) Clone() *Buffer {
	if buf == nil {
		return nil
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.unsafeInit()

	out := &Buffer{
		ctx:     buf.ctx,
		logger:  buf.logger,
		keySet:  buf.keySet.Clone(),
		certSet: buf.certSet.Clone(),
		sources: make(map[SourceName]*Source, len(buf.sources)),
	}

	for _, e := range buf.sources {
		out.sources[e.SourceName] = e.Clone()
	}

	return out
}

// Certs returns the [certpool.CertSet] containing all X.509 certificates in the [Buffer].
func (buf *Buffer) Certs() *certpool.CertSet {
	if buf == nil {
		return nil
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.unsafeInit()
	return buf.certSet
}

// Keys returns the [basic.KeySet] containing all private keys in the [Buffer].
func (buf *Buffer) Keys() *KeySet {
	if buf == nil {
		return nil
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.unsafeInit()
	return buf.keySet
}
