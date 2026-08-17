package fs_test

import (
	"os"

	"darvaza.org/x/fs"
)

// Compile-time verification that the file type returned by [os.Create]
// and [os.OpenFile] satisfies WriterFile, and that ClosedFile keeps
// matching it.
var (
	_ fs.WriterFile = (*os.File)(nil)
	_ fs.WriterFile = (*fs.ClosedFile)(nil)
)
