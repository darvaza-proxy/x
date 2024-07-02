package assets

import (
	"encoding/base64"
	"io"

	"lukechampine.com/blake3"
)

// BLAKE3SumFile calculates a BLAKE3-256 checksum for the file.
// If io.Seeker is implemented, the file will be rewind.
func BLAKE3SumFile(file io.Reader) (string, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	if f, ok := file.(io.ReadSeeker); ok {
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return "", err
		}
	}

	return BLAKE3Sum(b), nil
}

// BLAKE3Sum calculates a BLAKE3-256 check from
// the given content and returns its [base64] representation.
func BLAKE3Sum(data []byte) string {
	hash := blake3.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}
