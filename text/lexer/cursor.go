package lexer

import (
	"strings"
	"unicode/utf8"
)

// Cursor is a UTF-8-aware read cursor over a string source with an
// emit buffer for accumulating output. The API is rune- and
// text-shaped; encoding is an implementation detail.
//
// The zero Cursor is not usable; construct one with [New].
type Cursor struct {
	src string
	out strings.Builder
	pos int
}

// New returns a [Cursor] positioned at the start of src.
//
// src is treated as UTF-8 text; invalid byte sequences decode to
// [utf8.RuneError].
func New(src string) *Cursor {
	return &Cursor{src: src}
}

// Done reports whether the cursor is at end of input.
func (c *Cursor) Done() bool { return c.pos >= len(c.src) }

// Peek returns the next rune without advancing. ok is false at end
// of input.
func (c *Cursor) Peek() (r rune, ok bool) {
	if c.pos >= len(c.src) {
		return 0, false
	}
	r, _ = utf8.DecodeRuneInString(c.src[c.pos:])
	return r, true
}

// Advance moves the cursor forward by n runes, stopping at end of
// input. Negative n is a no-op.
func (c *Cursor) Advance(n int) {
	for i := 0; i < n && c.pos < len(c.src); i++ {
		_, size := utf8.DecodeRuneInString(c.src[c.pos:])
		c.pos += size
	}
}

// Consume returns the next rune and advances past it. ok is false
// at end of input; the cursor is unchanged in that case.
func (c *Cursor) Consume() (r rune, ok bool) {
	if c.pos >= len(c.src) {
		return 0, false
	}
	var size int
	r, size = utf8.DecodeRuneInString(c.src[c.pos:])
	c.pos += size
	return r, true
}

// Keep appends s to the emit buffer.
func (c *Cursor) Keep(s string) {
	_, _ = c.out.WriteString(s)
}

// KeepRune appends r to the emit buffer.
func (c *Cursor) KeepRune(r rune) {
	_, _ = c.out.WriteRune(r)
}

// EmitRest appends the remaining input to the emit buffer and
// advances the cursor to end of input.
func (c *Cursor) EmitRest() {
	c.Keep(c.src[c.pos:])
	c.pos = len(c.src)
}

// Emitted returns the contents of the emit buffer.
func (c *Cursor) Emitted() string { return c.out.String() }

// Reset clears the emit buffer. The cursor position is unchanged.
func (c *Cursor) Reset() { c.out.Reset() }
