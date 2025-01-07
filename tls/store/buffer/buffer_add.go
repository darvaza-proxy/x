package buffer

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

// AddCACerts adds all certificates in the [Buffer] to the [tls.Store] as trusted CAs.
func (buf *Buffer) AddCACerts(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	fn := func(ctx context.Context, src *Source, out tls.StoreX509Writer) (int, error) {
		return src.AddCACerts(ctx, out)
	}

	return buf.doAddFn(ctx, out, "AddCACerts", fn)
}

// AddPrivateKey adds all private keys in the [Buffer] to the [tls.Store].
func (buf *Buffer) AddPrivateKey(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	fn := func(ctx context.Context, src *Source, out tls.StoreX509Writer) (int, error) {
		return src.AddPrivateKeys(ctx, out)
	}

	return buf.doAddFn(ctx, out, "AddPrivateKeys", fn)
}

// AddCert adds all certificates in the [Buffer] to the [tls.Store].
func (buf *Buffer) AddCert(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	fn := func(ctx context.Context, src *Source, out tls.StoreX509Writer) (int, error) {
		return src.AddCert(ctx, out)
	}

	return buf.doAddFn(ctx, out, "AddCert", fn)
}

// AddCertPair adds all certificates in the [Buffer] to the [tls.Store] considering intermediate
// certificates in the [Source] and a private key anywhere in the [Buffer]
func (buf *Buffer) AddCertPair(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	fn := func(ctx context.Context, src *Source, out tls.StoreX509Writer) (int, error) {
		return src.AddCertPair(ctx, out, buf.keySet)
	}

	return buf.doAddFn(ctx, out, "AddCertPair", fn)
}

func (buf *Buffer) doAddFn(ctx context.Context, out tls.StoreX509Writer, op string,
	fn func(context.Context, *Source, tls.StoreX509Writer) (int, error)) (int, error) {
	// validate
	if err := buf.canAdd(ctx, out); err != nil {
		return 0, err
	}

	// only sources with certificates
	cond := func(e *Source) bool {
		return len(e.Certs) > 0
	}

	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	sources := buf.goEmitSources(ctx2, cond)
	return buf.doAddFnFromChannel(ctx2, out, op, fn, sources)
}

func (buf *Buffer) doAddFnFromChannel(ctx context.Context, out tls.StoreX509Writer, op string,
	fn func(context.Context, *Source, tls.StoreX509Writer) (int, error),
	sources <-chan *Source) (int, error) {
	//
	var errs core.CompoundError
	var count int

	for src := range sources {
		n, cont, err := buf.doAddFnFromSource(ctx, out, op, fn, src)
		if n > 0 {
			count += n
		}
		if err != nil {
			errs.AppendError(err)
		}
		if !cont {
			break
		}
	}

	return returnAdd2(count, errs.AsError())
}

func (buf *Buffer) doAddFnFromSource(ctx context.Context, out tls.StoreX509Writer, op string,
	fn func(context.Context, *Source, tls.StoreX509Writer) (int, error),
	src *Source) (int, bool, error) {
	//
	var errs core.CompoundError

	cont := true
	count, err := fn(ctx, src, out)
	if err != nil {
		src.AppendError(&errs, err, op, "")
	}

	if err := ctx.Err(); err != nil {
		// cancelled
		errs.AppendError(err)
		cont = false
	} else if err := buf.ctx.Err(); err != nil {
		// cancelled
		errs.AppendError(err)
		cont = false
	}

	return count, cont, errs.AsError()
}

func (buf *Buffer) canAdd(ctx context.Context, out tls.StoreX509Writer) error {
	switch {
	case buf == nil:
		return core.ErrNilReceiver
	case out == nil:
		return tls.ErrNoStore
	default:
		if err := ctx.Err(); err != nil {
			return err
		}

		return buf.ctx.Err()
	}
}
