package buffer

import (
	"testing"

	"darvaza.org/core"
)

// TestSys exercises the unexported sys() accessor.
func TestSys(t *testing.T) {
	t.Run("typed-nil", runTestSysTypedNil)
	t.Run("identity", runTestSysIdentity)
}

func runTestSysTypedNil(t *testing.T) {
	t.Helper()
	var nilBuf *Buffer
	core.AssertEqual(t, nil, nilBuf.sys(), "nil receiver sys()")
}

func runTestSysIdentity(t *testing.T) {
	t.Helper()
	var buf Buffer
	sys := buf.sys()
	core.AssertMustNotNil(t, sys, "sys")
	_, _ = sys.WriteString("via sys")
	core.AssertEqual(t, "via sys", buf.String(), "shared storage")
}
