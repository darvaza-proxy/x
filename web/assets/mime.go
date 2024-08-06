package assets

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"darvaza.org/core"
)

// TypeByFilename attempts to infer the ContentType from the filename.
// It will return "" if unknown.
func TypeByFilename(fileName string) string {
	if fileName == "" {
		return ""
	}
	return mime.TypeByExtension(filepath.Ext(fileName))
}

// TypeBySniffing looks at the first 512 bytes of data to attempt
// to determine ContentType.
func TypeBySniffing(file io.ReadSeeker) (string, error) {
	var buf [512]byte

	// try reading 512 bytes from the file and rewind.
	n, _ := io.ReadFull(file, buf[:])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", core.Wrap(err, "seek")
	}

	cType := http.DetectContentType(buf[:n])
	return cType, nil
}
