// Package buffer provides a [strings.Builder] wrapper whose chainable
// write helpers discard errors, letting callers build up output
// without error-handling boilerplate. Interface-compliant methods
// ([io.Writer], [io.WriterTo], [io.StringWriter], [io.ByteWriter])
// keep their standard signatures.
//
// Buffer is intended for single-use, write-then-consume workflows.
// It inherits [strings.Builder]'s no-copy contract — do not copy a
// Buffer after first use — and is not safe for concurrent use.
package buffer
