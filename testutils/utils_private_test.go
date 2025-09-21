package testutils

import (
	"testing"

	"darvaza.org/core"
)

func TestMatchStringOnly(t *testing.T) {
	t.Run("wh anything", func(t *testing.T) {
		// the helper is just a dummy that accepts anything.
		// It's needed by testing.RunTests but never called.
		ok, err := matchStringDummy("pat", "str")
		core.AssertTrue(t, ok, "ok")
		core.AssertNoError(t, err, "err")
	})
}
