package buffer

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
)

// SourceName identifies the [Source].
type SourceName struct {
	FS       fs.FS
	FileName string
}

// IsFile tells if the [Source] was a file.
func (sn SourceName) IsFile() bool {
	return sn.FileName != ""
}

// NewErrorf works like NewError but the note is a formatted string.
func (sn SourceName) NewErrorf(err error, op, format string, args ...any) error {
	note := fmt.Sprintf(format, args...)
	return sn.NewError(err, op, note)
}

// NewError creates a fs.PathError if the source has a name, otherwise
// it wraps and annotates the given error. if no error or annotation is passed
// no error will be returned.
func (sn SourceName) NewError(err error, op, note string) error {
	switch {
	case sn.IsFile():
		return sn.NewPathError(err, op, note)
	case err == nil:
		// no wrapping
		return sn.doNewError2(op, note)
	case op == "" && note == "":
		// pass through
		return err
	case op == "":
		return core.Wrap(err, op)
	case note == "":
		return core.Wrap(err, note)
	default:
		return core.Wrapf(err, "%s: %s", op, note)
	}
}

func (SourceName) doNewError2(op, note string) error {
	switch {
	case op == "" && note == "":
		// no error
		return nil
	case op != "" && note != "":
		return fmt.Errorf("%s: %s", op, note)
	case op == "":
		return errors.New(note)
	default:
		return errors.New(op)
	}
}

// AppendError appends an error annotated by with source details. Compound errors will be
// appended individually to the output.
func (sn SourceName) AppendError(out *core.CompoundError, err error, op, note string) {
	if out == nil {
		core.Panic("nil CompoundError")
	}

	if errs, ok := err.(*core.CompoundError); ok {
		for _, err := range errs.Errs {
			sn.AppendError(out, err, op, note)
		}
	} else {
		_ = out.AppendError(sn.NewError(err, op, note))
	}
}

// NewPathError creates an [fs.PathError] using its fileName and the given Op and
// error. The note can be used to annotate the error. See [NewErrorf] for formatted
// annotations.
func (sn SourceName) NewPathError(err error, op, note string) *fs.PathError {
	if note != "" {
		if err == nil {
			// not wrapping
			err = errors.New(note)
		} else {
			// wrapping
			err = core.Wrap(err, note)
		}
	}

	return &fs.PathError{
		Path: sn.FileName,
		Op:   op,
		Err:  err,
	}
}

// NewSourceName creates a new [SourceName]
func NewSourceName(fSys fs.FS, fileName string) SourceName {
	return SourceName{
		FS:       fSys,
		FileName: fileName,
	}
}

// Source contains certificates, keys and errors collected
// while reading PEM content.
type Source struct {
	SourceName

	Certs []*x509.Certificate
	Keys  []x509utils.PrivateKey
	Errs  []error
}

// Clone creates a copy of the [Source].
func (src *Source) Clone() *Source {
	if src == nil {
		return nil
	}

	return &Source{
		SourceName: src.SourceName,

		Certs: core.SliceCopy(src.Certs),
		Keys:  core.SliceCopy(src.Keys),
		Errs:  core.SliceCopy(src.Errs),
	}
}

// AddCACerts adds all certificates from the [Source] to the [tls.Store] as trusted CAs.
func (src *Source) AddCACerts(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	// validate
	if err := src.canAdd(ctx, out); err != nil {
		return 0, err
	}

	// signature matching.
	addCACert := func(ctx context.Context, cert *x509.Certificate) error {
		return out.AddCACerts(ctx, cert)
	}

	addFn := func(ctx context.Context, _ tls.StoreX509Writer, cert *x509.Certificate) (int, error) {
		return doAddFn(ctx, cert, addCACert)
	}

	return addSourceFn(ctx, out, src, src.Certs, addFn)
}

// AddCert adds all certificates from the [Source] to the [tls.Store].
func (src *Source) AddCert(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	// validate
	if err := src.canAdd(ctx, out); err != nil {
		return 0, err
	}

	fn := func(ctx context.Context, out tls.StoreX509Writer, cert *x509.Certificate) (int, error) {
		return doAddFn(ctx, cert, out.AddCert)
	}

	return addSourceFn(ctx, out, src, src.Certs, fn)
}

// AddCertPair adds the first certificate in the [Source] to the [tls.Store] using
// the rest as intermediate. a key in the same source is preferred but it can also
// use a [KeySet] to find it.
func (src *Source) AddCertPair(ctx context.Context, out tls.StoreX509Writer, keys *KeySet) (int, error) {
	var certs []*x509.Certificate
	var inter []*x509.Certificate

	// validate
	if err := src.canAdd(ctx, out); err != nil {
		return 0, err
	}

	if len(src.Certs) > 0 {
		// TODO: consider exception when source isn't a file.
		// first is the leaf.
		certs = []*x509.Certificate{src.Certs[0]}
		inter = src.Certs[1:]
	}

	addFn := func(ctx context.Context, out tls.StoreX509Writer, leaf *x509.Certificate) (int, error) {
		return src.doAddCertPair(ctx, out, leaf, inter, keys)
	}

	return addSourceFn(ctx, out, src, certs, addFn)
}

func (src *Source) doAddCertPair(ctx context.Context, out tls.StoreX509Writer, leaf *x509.Certificate,
	inter []*x509.Certificate, keys *KeySet) (int, error) {
	//
	key := src.findKeyForCert(leaf, keys)
	err := out.AddCertPair(ctx, key, leaf, inter)
	switch {
	case err == nil:
		// success
		return 1, nil
	case IsExists(err):
		// dupe
		return 0, nil
	default:
		// failed
		return 0, err
	}
}

func (src *Source) findKeyForCert(cert *x509.Certificate, keys *KeySet) x509utils.PrivateKey {
	// same source first
	for _, key := range src.Keys {
		if x509utils.PublicKeyEqual(cert.PublicKey, key.Public) {
			return key
		}
	}
	// then the whole buffer
	if keys != nil {
		key, err := keys.GetFromCertificate(cert)
		if err == nil {
			return key
		}
	}

	return nil
}

// AddPrivateKeys adds all private keys from the [Source] to the [tls.Store].
func (src *Source) AddPrivateKeys(ctx context.Context, out tls.StoreX509Writer) (int, error) {
	// validate
	if err := src.canAdd(ctx, out); err != nil {
		return 0, err
	}

	// signature matching
	addKey := func(ctx context.Context, key x509utils.PrivateKey) error {
		return out.AddPrivateKey(ctx, key)
	}

	addFn := func(ctx context.Context, _ tls.StoreX509Writer, key x509utils.PrivateKey) (int, error) {
		return doAddFn(ctx, key, addKey)
	}

	return addSourceFn(ctx, out, src, src.Keys, addFn)
}

func doAddFn[T any](ctx context.Context, v T, add func(context.Context, T) error) (int, error) {
	err := add(ctx, v)
	switch {
	case err == nil:
		return 1, nil
	case IsExists(err):
		return 0, nil
	default:
		return 0, err
	}
}

func addSourceFn[T any](ctx context.Context, out tls.StoreX509Writer, src *Source, data []T,
	fn func(context.Context, tls.StoreX509Writer, T) (int, error)) (int, error) {
	//
	var errs core.CompoundError

	// run
	count, err := doAddSourceFn(ctx, out, data, fn)
	if err != nil {
		// append add errors
		_ = errs.AppendError(err)
	}

	// append source errors
	_ = errs.AppendError(src.Errs...)

	if err := ctx.Err(); err != nil {
		_ = errs.AppendError(err)
	}

	return returnAdd2(count, errs.AsError())
}

func doAddSourceFn[T any](ctx context.Context, out tls.StoreX509Writer, data []T,
	fn func(context.Context, tls.StoreX509Writer, T) (int, error)) (int, error) {
	//
	var errs core.CompoundError
	var count int

loop:
	for _, v := range data {
		n, err := fn(ctx, out, v)
		switch {
		case err == nil:
			// success
			count += n
		case IsCancelled(err):
			// cancelled
			break loop
		case !IsExists(err):
			// error
			_ = errs.AppendError(err)
		}

		if ctx.Err() != nil {
			// cancelled
			break loop
		}
	}

	return count, errs.AsError()
}

func (src *Source) canAdd(ctx context.Context, out tls.StoreX509Writer) error {
	switch {
	case src == nil:
		return core.ErrNilReceiver
	case out == nil:
		return tls.ErrNoStore
	default:
		return ctx.Err()
	}
}
