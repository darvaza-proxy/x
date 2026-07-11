//go:build !windows

package appdir

const (
	// PrefixLocal represents services installed outside
	// the scope of the package manager.
	PrefixLocal Prefix = "/usr/local"
	// PrefixSystem represents services installed by
	// the package manager.
	PrefixSystem Prefix = "/"
	// PrefixOptional represents services installed outside
	// the scope of the package manager but requiring
	// a complex hierarchy, usually installed by extracting
	// an archive file.
	PrefixOptional Prefix = "/opt"
)

// isWellKnown reports whether p is one of the predefined
// prefixes.
func (p Prefix) isWellKnown() bool {
	switch p {
	case PrefixUser, PrefixSystem, PrefixLocal, PrefixOptional:
		return true
	default:
		return false
	}
}
