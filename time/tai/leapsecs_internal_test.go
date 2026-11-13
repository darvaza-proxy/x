package tai

import (
	"testing"
	"time"

	"darvaza.org/core"
)

var _ core.TestCase = lsOffsetTestCase{}

type lsOffsetTestCase struct {
	input time.Time
	name  string
	want  uint64
}

func newLSOffsetTestCase(name string, input time.Time, want uint64) lsOffsetTestCase {
	return lsOffsetTestCase{name: name, input: input, want: want}
}

func (tc lsOffsetTestCase) Name() string { return tc.name }

func (tc lsOffsetTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, lsoffset(tc.input), "additional leap seconds")
}

func TestLSOffset(t *testing.T) {
	core.RunTestCases(t, core.S(
		newLSOffsetTestCase("Before Unix epoch", time.Unix(-1, 0), 0),
		newLSOffsetTestCase("Unix epoch", time.Unix(0, 0), 0),
		newLSOffsetTestCase("Before first leap", time.Unix(78796800, -1), 0),
		newLSOffsetTestCase("At first leap", time.Unix(78796800, 0), 1),
		newLSOffsetTestCase("Before second leap", time.Unix(94694400, -1), 1),
		newLSOffsetTestCase("At second leap", time.Unix(94694400, 0), 2),
		newLSOffsetTestCase("Before latest leap", time.Unix(1483228800, -1), 26),
		newLSOffsetTestCase("At latest leap", time.Unix(1483228800, 0), currentLeapSeconds),
		newLSOffsetTestCase("After latest leap", time.Unix(1483228801, 0), currentLeapSeconds),
	))
}
