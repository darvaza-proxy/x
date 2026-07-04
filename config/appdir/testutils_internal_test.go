package appdir

// StubSysPrefix overrides the system prefix directly, bypassing the
// filesystem validation performed by [SetSysPrefix], and returns a
// function restoring the previous value. Unlike [SetSysPrefix] it can
// restore [PrefixUser] mode, keeping tests isolated.
func StubSysPrefix(dir string) func() {
	prev := prefix
	prefix = dir
	return func() {
		prefix = prev
	}
}
