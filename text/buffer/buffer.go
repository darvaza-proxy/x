package buffer

import (
	"fmt"
	"io"
	"strings"
	"unsafe"
)

// Buffer is a [strings.Builder] whose chainable helpers return
// *Buffer instead of (int, error), so writes compose without
// intermediate error checks. The zero value is ready to use.
//
// Inherits [strings.Builder]'s no-copy contract: do not copy a
// Buffer after first use.
type Buffer strings.Builder

var (
	_ io.Writer       = (*Buffer)(nil)
	_ io.WriterTo     = (*Buffer)(nil)
	_ io.StringWriter = (*Buffer)(nil)
	_ io.ByteWriter   = (*Buffer)(nil)
)

// sys returns the underlying [strings.Builder]. Nil-safe: returns
// nil on nil receiver so callers don't synthesise typed nils via
// the [Buffer] → [strings.Builder] conversion.
func (buf *Buffer) sys() *strings.Builder {
	if buf == nil {
		return nil
	}
	return (*strings.Builder)(buf)
}

// Len reports the number of bytes currently stored.
func (buf *Buffer) Len() int { return buf.sys().Len() }

// String returns the stored data as a string. The conversion is
// zero-allocation; see [strings.Builder.String].
func (buf *Buffer) String() string { return buf.sys().String() }

// Bytes returns the stored data as a byte slice aliasing the
// buffer's internal storage. Subsequent writes may grow or
// reallocate the underlying array, invalidating the slice.
// Always returns a non-nil slice; emptiness is signalled by
// len(b) == 0, never by a nil return.
//
// WARNING: The bytes MUST be treated as read-only. Mutating them
// also mutates every string previously returned by [Buffer.String]
// — [strings.Builder]'s zero-allocation String relies on the same
// storage and Go strings are required to be immutable. Make a copy
// (`append([]byte(nil), buf.Bytes()...)`) before mutating.
func (buf *Buffer) Bytes() []byte {
	s := buf.String()
	if len(s) == 0 {
		return []byte{}
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Write implements the [io.Writer] interface.
func (buf *Buffer) Write(b []byte) (int, error) { return buf.sys().Write(b) }

// WriteTo implements the [io.WriterTo] interface. On success the
// buffer is drained and ready for reuse; the underlying storage is
// discarded (see [strings.Builder.Reset]). On error the buffer is
// left intact so the caller can retry or inspect the unsent payload.
func (buf *Buffer) WriteTo(out io.Writer) (int64, error) {
	n, err := io.WriteString(out, buf.String())
	if err == nil {
		buf.sys().Reset()
	}
	return int64(n), err
}

// WriteString implements the [io.StringWriter] interface.
// For the chainable form, use [Buffer.WriteStrings].
func (buf *Buffer) WriteString(s string) (int, error) { return buf.sys().WriteString(s) }

// WriteByte implements the [io.ByteWriter] interface.
func (buf *Buffer) WriteByte(c byte) error { return buf.sys().WriteByte(c) }

// WriteRune appends the UTF-8 encoding of r and reports the number
// of bytes written and a nil error.
// For the chainable form, use [Buffer.WriteRunes].
func (buf *Buffer) WriteRune(r rune) (int, error) { return buf.sys().WriteRune(r) }

// Print formats operands with [fmt.Print] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *Buffer) Print(a ...any) *Buffer {
	_, _ = fmt.Fprint(buf.sys(), a...)
	return buf
}

// Println formats operands with [fmt.Println] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *Buffer) Println(a ...any) *Buffer {
	_, _ = fmt.Fprintln(buf.sys(), a...)
	return buf
}

// Printf formats operands with [fmt.Printf] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *Buffer) Printf(format string, a ...any) *Buffer {
	_, _ = fmt.Fprintf(buf.sys(), format, a...)
	return buf
}

// WriteRunes appends the given runes as UTF-8 characters to the buffer.
// Returns the buffer for method chaining.
func (buf *Buffer) WriteRunes(runes ...rune) *Buffer {
	b := buf.sys()
	for _, r := range runes {
		_, _ = b.WriteRune(r)
	}
	return buf
}

// WriteBytes writes the given byte slices to the buffer.
// Returns the buffer for method chaining.
func (buf *Buffer) WriteBytes(ss ...[]byte) *Buffer {
	b := buf.sys()
	for _, s := range ss {
		_, _ = b.Write(s)
	}
	return buf
}

// WriteStrings writes the given strings to the buffer.
// Returns the buffer for method chaining.
func (buf *Buffer) WriteStrings(ss ...string) *Buffer {
	b := buf.sys()
	for _, s := range ss {
		_, _ = b.WriteString(s)
	}
	return buf
}

// Grow guarantees space for n more bytes in the buffer.
// Negative or zero n is a no-op. Returns the buffer for method chaining.
func (buf *Buffer) Grow(n int) *Buffer {
	if n > 0 {
		buf.sys().Grow(n)
	}
	return buf
}

// Reset clears the buffer for reuse. The underlying storage is
// discarded; subsequent writes reallocate (see [strings.Builder.Reset]).
// Returns the buffer for method chaining.
func (buf *Buffer) Reset() *Buffer {
	buf.sys().Reset()
	return buf
}

// New creates a new Buffer with pre-allocated capacity.
// Negative or zero capacity yields an empty buffer.
func New(capacity int) *Buffer {
	return (&Buffer{}).Grow(capacity)
}
