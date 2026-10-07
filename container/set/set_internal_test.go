package set

import (
	"testing"

	"darvaza.org/core"
)

// TestInitNilReceiver covers Set.init's nil-receiver guard. The public
// Config.Init rejects a nil set before delegating, so this branch is only
// reachable white-box.
func TestInitNilReceiver(t *testing.T) {
	var s *Set[int, int, int]
	core.AssertErrorIs(t, s.init(Config[int, int, int]{}), core.ErrNilReceiver,
		"init on nil receiver")
}
