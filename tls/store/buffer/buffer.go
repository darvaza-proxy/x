// Package buffer provides a [Buffer] to help decoding PEM files
// and populating a [tls.StoreWriter].
package buffer

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// Buffer is a PEM decoding buffer to populate a [tls.StoreWriter].
type Buffer struct {
	mu     sync.Mutex
	ctx    context.Context
	logger slog.Logger

	entries map[bufferEntryKey]*bufferEntry
}

// New creates a PEM decoding Buffer to populate a [tls.StoreWriter].
func New(ctx context.Context, logger slog.Logger) *Buffer {
	return &Buffer{
		ctx:     ctx,
		logger:  logger,
		entries: make(map[bufferEntryKey]*bufferEntry),
	}
}

// Clone creates a copy of the [Buffer]
func (buf *Buffer) Clone() *Buffer {
	if buf == nil {
		return nil
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	out := &Buffer{
		ctx:     buf.ctx,
		logger:  buf.logger,
		entries: make(map[bufferEntryKey]*bufferEntry, len(buf.entries)),
	}

	for _, e := range buf.entries {
		out.entries[e.bufferEntryKey] = e.Clone()
	}

	return out
}

// NewAddCallback returns a callback that adds all certificates and private keys
// to the [Buffer].
func (buf *Buffer) NewAddCallback() x509utils.DecodePEMBlockFunc {
	return buf.onAdd
}

func (buf *Buffer) onAdd(fSys fs.FS, fileName string, block *pem.Block) bool {
	// cert?
	cert, err := x509utils.BlockToCertificate(block)
	switch {
	case cert != nil:
		// cert
		buf.pushCert(fSys, fileName, cert)
	case err == x509utils.ErrIgnored:
		// key?
		var key x509utils.PrivateKey

		key, err = x509utils.BlockToPrivateKey(block)
		if key != nil {
			// key
			buf.pushKey(fSys, fileName, key)
		}
	}

	if err != nil {
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() != nil
}

// NewAddCertsCallback returns a callback that adds all certificates to the [Buffer].
func (buf *Buffer) NewAddCertsCallback() x509utils.DecodePEMBlockFunc {
	return buf.onAddCert
}

func (buf *Buffer) onAddCert(fSys fs.FS, fileName string, block *pem.Block) bool {
	cert, err := x509utils.BlockToCertificate(block)
	switch {
	case cert != nil:
		buf.pushCert(fSys, fileName, cert)
	case err != nil:
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() != nil
}

// NewAddPrivateKeysCallback returns a callback that adds private keys to the [Buffer].
func (buf *Buffer) NewAddPrivateKeysCallback() x509utils.DecodePEMBlockFunc {
	return buf.onAddPrivateKeys
}

func (buf *Buffer) onAddPrivateKeys(fSys fs.FS, fileName string, block *pem.Block) bool {
	key, err := x509utils.BlockToPrivateKey(block)
	switch {
	case key != nil:
		buf.pushKey(fSys, fileName, key)
	case err != nil:
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() != nil
}

func (buf *Buffer) pushCert(fSys fs.FS, fileName string, cert *x509.Certificate) {
	if cert != nil {
		bek := bufferEntryKey{
			fSys:     fSys,
			fileName: fileName,
		}

		buf.mu.Lock()
		defer buf.mu.Unlock()

		e, ok := buf.entries[bek]
		if !ok {
			e = &bufferEntry{
				bufferEntryKey: bek,
			}
			buf.entries[bek] = e
		}

		e.certs = append(e.certs, cert)
	}
}

func (buf *Buffer) pushKey(fSys fs.FS, fileName string, key x509utils.PrivateKey) {
	if key != nil {
		bek := bufferEntryKey{
			fSys:     fSys,
			fileName: fileName,
		}

		buf.mu.Lock()
		defer buf.mu.Unlock()

		e, ok := buf.entries[bek]
		if !ok {
			e = &bufferEntry{
				bufferEntryKey: bek,
			}
			buf.entries[bek] = e
		}

		e.keys = append(e.keys, key)
	}
}

func (buf *Buffer) pushErr(fSys fs.FS, fileName string, err error) {
	if err != nil {
		bek := bufferEntryKey{
			fSys:     fSys,
			fileName: fileName,
		}

		buf.mu.Lock()
		defer buf.mu.Unlock()

		e, ok := buf.entries[bek]
		if !ok {
			e = &bufferEntry{
				bufferEntryKey: bek,
			}
			buf.entries[bek] = e
		}

		e.errs = append(e.errs, err)
	}
}

// AddCACerts ...
func (buf *Buffer) AddCACerts(s tls.StoreX509Writer) (int, error) {
	// validate
	if err := buf.canAdd(s); err != nil {
		return 0, err
	}

	// produce safe iterable list
	cond := func(e *bufferEntry) bool {
		return e != nil && len(e.certs) > 0
	}
	entries := buf.exportEntries(cond)

	return buf.doAddCACerts(s, entries)
}

func (buf *Buffer) doAddCACerts(s tls.StoreX509Writer, entries []*bufferEntry) (int, error) {
	var errs core.CompoundError
	var count int

	for _, e := range entries {
		for _, c := range e.certs {
			if err := buf.doAddCACert(s, e, c); err != nil {
				errs.AppendError(err)
			} else {
				count++
			}
		}
	}

	return returnAdd2(count, errs.AsError())
}

func (*Buffer) doAddCACert(_ tls.StoreX509Writer, _ *bufferEntry, _ *x509.Certificate) error {
	return core.ErrTODO
}

// AddCert ...
func (buf *Buffer) AddCert(s tls.StoreX509Writer) error {
	// validate
	if err := buf.canAdd(s); err != nil {
		return err
	}

	// produce safe iterable list
	cond := func(e *bufferEntry) bool {
		return e != nil && len(e.certs) > 0
	}
	entries := buf.exportEntries(cond)

	return buf.doAddCertPairs(s, nil, entries)
}

func (*Buffer) doAddCertPairs(_ tls.StoreX509Writer, _ []x509utils.PrivateKey, _ []*bufferEntry) error {
	return returnAdd(0, core.ErrTODO)
}

// AddCertPair ...
func (buf *Buffer) AddCertPair(s tls.StoreX509Writer) error {
	// validate
	if err := buf.canAdd(s); err != nil {
		return err
	}

	// produce safe iterable list
	cond := func(e *bufferEntry) bool {
		return e != nil && len(e.certs) > 0
	}
	entries := buf.exportEntries(cond)

	// and keys
	keys := buf.Keys()

	// assemble and add
	return buf.doAddCertPairs(s, keys, entries)
}

// AddPrivateKey adds all private keys in the [Buffer] to the store.
func (buf *Buffer) AddPrivateKey(s tls.StoreX509Writer) error {
	// validate
	if err := buf.canAdd(s); err != nil {
		return err
	}

	// produce safe iterable list
	cond := func(e *bufferEntry) bool {
		return e != nil && len(e.keys) > 0
	}
	entries := buf.exportEntries(cond)

	// add
	return buf.doAddPrivateKeys(s, entries)
}

func (buf *Buffer) doAddPrivateKeys(s tls.StoreX509Writer, entries []*bufferEntry) error {
	var errs core.CompoundError
	var count int

	for _, e := range entries {
		if err := buf.doAddPrivateKey(s, e); err != nil {
			errs.AppendError(err)
		} else {
			count++
		}
	}

	return returnAdd(count, errs.AsError())
}

func (*Buffer) doAddPrivateKey(_ tls.StoreX509Writer, _ *bufferEntry) error {
	return core.ErrTODO
}

// Keys returns all private keys in the [Buffer]
func (buf *Buffer) Keys() []x509utils.PrivateKey {
	var out []x509utils.PrivateKey
	if buf != nil {
		buf.mu.Lock()
		defer buf.mu.Unlock()

		for _, e := range buf.entries {
			if len(e.keys) > 0 {
				out = append(out, e.keys...)
			}
		}
	}
	return out
}

// ForEach calls a function for each processed file.
func (buf *Buffer) ForEach(fn func(fs.FS, string, []x509utils.PrivateKey, []*x509.Certificate, []error) bool) {
	if buf == nil || fn == nil {
		return
	}

	// create safe list to iterate.
	entries := buf.exportEntries(nil)

	for _, e := range entries {
		if !fn(e.fSys, e.fileName, e.keys, e.certs, e.errs) {
			break
		}
	}
}

func (buf *Buffer) exportEntries(cond func(*bufferEntry) bool) []*bufferEntry {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	out := make([]*bufferEntry, 0, len(buf.entries))
	for _, e := range buf.entries {
		if cond == nil || cond(e) {
			out = append(out, e.Clone())
		}
	}
	return out
}

func (buf *Buffer) canAdd(s tls.StoreX509Writer) error {
	switch {
	case buf == nil:
		return core.ErrNilReceiver
	case s == nil:
		return tls.ErrNoStore
	default:
		return buf.ctx.Err()
	}
}

func returnAdd(count int, err error) error {
	if count == 0 && err == nil {
		err = x509utils.ErrEmpty
	}
	return err
}

func returnAdd2(count int, err error) (int, error) {
	if count == 0 && err == nil {
		err = x509utils.ErrEmpty
	}
	return count, err
}

type bufferEntryKey struct {
	fSys     fs.FS
	fileName string
}

type bufferEntry struct {
	bufferEntryKey

	certs []*x509.Certificate
	keys  []x509utils.PrivateKey
	errs  []error
}

func (e *bufferEntry) Clone() *bufferEntry {
	if e == nil {
		return nil
	}

	return &bufferEntry{
		bufferEntryKey: e.bufferEntryKey,
		certs:          core.SliceCopy(e.certs),
		keys:           core.SliceCopy(e.keys),
		errs:           core.SliceCopy(e.errs),
	}
}
