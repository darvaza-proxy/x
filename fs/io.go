package fs

import "io"

type (
	// Reader is an alias of the standard [io.Reader] interface.
	Reader = io.Reader
	// ReadSeeker is an alias of the standard [io.ReadSeeker] interface.
	ReadSeeker = io.ReadSeeker
	// Writer is an alias of the standard [io.Writer] interface.
	Writer = io.Writer
	// Closer is an alias of the standard [io.Closer] interface.
	Closer = io.Closer
)

// A Flusher implements the Flush() error interface to write whatever is left on the buffer.
type Flusher interface {
	Flush() error
}

// A WriteFlusher implements [io.Writer] and [Flusher].
type WriteFlusher interface {
	Writer
	Flusher
}

// A WriteCloseFlusher implements [io.Writer], [io.Closer] and [Flusher].
type WriteCloseFlusher interface {
	Writer
	Closer
	Flusher
}
