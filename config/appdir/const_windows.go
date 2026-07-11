package appdir

const (
	// PrefixSystem composes the system-mode directories under
	// the machine-wide application data directory,
	// %ProgramData%, resolved when composing.
	PrefixSystem Prefix = `%ProgramData%`
	// PrefixLocal is equivalent to [PrefixSystem] on Windows.
	PrefixLocal = PrefixSystem
	// PrefixOptional is equivalent to [PrefixSystem] on Windows.
	PrefixOptional = PrefixSystem
)

// isWellKnown reports whether p is one of the predefined
// prefixes. [PrefixLocal] and [PrefixOptional] alias
// [PrefixSystem], so matching it covers all three.
func (p Prefix) isWellKnown() bool {
	switch p {
	case PrefixUser, PrefixSystem:
		return true
	default:
		return false
	}
}
