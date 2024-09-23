package x509utils

import (
	"encoding/pem"
	"io/fs"

	"darvaza.org/core"
)

// DecodePEMBlockFunc is called for each PEM block coded. it returns false
// to terminate the loop
type DecodePEMBlockFunc func(fSys fs.FS, filename string, block *pem.Block) bool

// ReadPEM invoques a callback for each PEM block found
// it can receive raw PEM data
func ReadPEM(b []byte, cb DecodePEMBlockFunc) error {
	if len(b) == 0 || cb == nil {
		// nothing do
		return nil
	}

	if block, rest := pem.Decode(b); block != nil {
		// PEM chain
		_ = readBlock(nil, "", block, rest, cb)
		return nil
	}

	// Not PEM
	return core.ErrInvalid
}

func readBlock(fSys fs.FS, filename string, block *pem.Block, rest []byte, cb DecodePEMBlockFunc) bool {
	for block != nil {
		if !cb(fSys, filename, block) {
			// cascade termination request
			return false
		} else if len(rest) == 0 {
			// EOF
			break
		}

		// next
		block, rest = pem.Decode(rest)
	}

	return true
}
