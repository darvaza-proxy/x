package appdir

// StubSysPrefix overrides the default Prefix directly, bypassing the
// filesystem validation performed by [SetSysPrefix], and returns a
// function restoring the previous value, keeping tests isolated.
func StubSysPrefix(p Prefix) func() {
	prev := prefix.Load()
	prefix.Store(&p)
	return func() {
		prefix.Store(prev)
	}
}
