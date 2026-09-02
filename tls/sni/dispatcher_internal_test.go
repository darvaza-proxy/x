package sni

import (
	"testing"
	"time"

	"darvaza.org/core"
)

var _ core.TestCase = nextAcceptDelayTestCase{}

// nextAcceptDelayTestCase feeds one delay to nextAcceptDelay and
// expects the next, both in milliseconds.
type nextAcceptDelayTestCase struct {
	name  string
	delay time.Duration
	want  time.Duration
}

func newNextAcceptDelayTestCase(name string, delayMs, wantMs int) nextAcceptDelayTestCase {
	return nextAcceptDelayTestCase{
		name:  name,
		delay: time.Duration(delayMs) * time.Millisecond,
		want:  time.Duration(wantMs) * time.Millisecond,
	}
}

func (tc nextAcceptDelayTestCase) Name() string {
	return tc.name
}

func (tc nextAcceptDelayTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, nextAcceptDelay(tc.delay), "delay")
}

// TestNextAcceptDelay pins the pace of accept retries: 5 ms first,
// doubling, capped at a second.
func TestNextAcceptDelay(t *testing.T) {
	core.RunTestCases(t, core.S(
		newNextAcceptDelayTestCase("first", 0, 5),
		newNextAcceptDelayTestCase("doubles", 5, 10),
		newNextAcceptDelayTestCase("caps", 600, 1000),
		newNextAcceptDelayTestCase("holds the cap", 1000, 1000),
	))
}
