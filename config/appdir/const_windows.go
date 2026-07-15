package appdir

const (
	// PrefixUser is the Prefix indicating user mode, where the
	// FooDir methods return the same as UserFooDir(). Windows
	// has no home-relative "~"; its symbolic token is "@user".
	PrefixUser Prefix = "@user"
	// PrefixSystem composes the system-mode directories under
	// the machine-wide application data directory,
	// %ProgramData%, resolved when composing.
	PrefixSystem Prefix = `%ProgramData%`
)

// isWellKnown reports whether p is one of the predefined
// prefixes.
func (p Prefix) isWellKnown() bool {
	switch p {
	case PrefixUser, PrefixSystem:
		return true
	default:
		return false
	}
}
