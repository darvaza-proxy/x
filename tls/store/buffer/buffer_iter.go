package buffer

import (
	"context"
	"crypto/x509"
	"io/fs"

	"darvaza.org/x/tls/x509utils"
)

// ForEachIterFunc represents a callback passed to ForEach and invoked for each source in the Buffer
// until it returns false.
type ForEachIterFunc func(fs.FS, string, []x509utils.PrivateKey, []*x509.Certificate, []error) bool

// ForEach calls a function for each processed source.
func (buf *Buffer) ForEach(fn ForEachIterFunc) {
	if buf == nil || fn == nil {
		return
	}

	// create safe list to iterate.
	sources := buf.exportSources(nil)

	for _, e := range sources {
		if !fn(e.FS, e.FileName, e.Keys, e.Certs, e.Errs) {
			break
		}
	}
}

func (buf *Buffer) exportSources(cond func(*Source) bool) []*Source {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	out := make([]*Source, 0, len(buf.sources))
	for _, e := range buf.sources {
		if cond == nil || cond(e) {
			out = append(out, e.Clone())
		}
	}
	return out
}

// goExportSources emits to a channel all sources meeting the condition or until the given context
// or the embedded one are cancelled.
func (buf *Buffer) goEmitSources(ctx context.Context, cond func(*Source) bool) <-chan *Source {
	ch := make(chan *Source)

	go func() {
		defer close(ch)

		buf.runEmitSources(ctx, cond, ch)
	}()

	return ch
}

func (buf *Buffer) runEmitSources(ctx context.Context, cond func(*Source) bool, out chan<- *Source) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.unsafeInit()

	for _, src := range buf.sources {
		if cond == nil || cond(src) {
			if !buf.emitSource(ctx, out, src) {
				break
			}
		}
	}
}

func (buf *Buffer) emitSource(ctx context.Context, out chan<- *Source, src *Source) bool {
	select {
	case out <- src:
		// emitted
		return true
	case <-ctx.Done():
		// cancelled
		return false
	case <-buf.ctx.Done():
		// cancelled
		return false
	}
}
