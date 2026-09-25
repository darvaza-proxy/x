package num

import (
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = appendPaddedCase{}

// appendPaddedCase pins the text appendPadded writes for one value at
// one width.
type appendPaddedCase struct {
	name  string
	want  string
	v     uint64
	width int
}

func newAppendPaddedCase(name string, v uint64, width int,
	want string) appendPaddedCase {
	return appendPaddedCase{name: name, v: v, width: width, want: want}
}

func (tc appendPaddedCase) Name() string { return tc.name }

func (tc appendPaddedCase) Test(t *testing.T) {
	t.Helper()
	got := appendPadded(nil, tc.v, tc.width)
	core.AssertEqual(t, tc.want, string(got), "text")
}

func TestAppendPadded(t *testing.T) {
	core.RunTestCases(t, []appendPaddedCase{
		newAppendPaddedCase("zero at width zero", 0, 0, ""),
		newAppendPaddedCase("zero at negative width", 0, -19, ""),
		newAppendPaddedCase("zero at width three", 0, 3, "000"),
		newAppendPaddedCase("padded", 7, 3, "007"),
		newAppendPaddedCase("wider than width", 1234, 2, "1234"),
	})
}
