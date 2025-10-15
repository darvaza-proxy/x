package web

import (
	"net/http"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/web/consts"
)

// TestCase interface validation
var _ core.TestCase = setRetryAfterTestCase{}

// setRetryAfterTestCase tests SetRetryAfter function
type setRetryAfterTestCase struct {
	name           string
	retryAfter     time.Duration
	expectedHeader string
}

func (tc setRetryAfterTestCase) Name() string {
	return tc.name
}

func (tc setRetryAfterTestCase) Test(t *testing.T) {
	t.Helper()

	hdr := make(http.Header)
	SetRetryAfter(hdr, tc.retryAfter)

	actual := hdr.Get(consts.RetryAfter)
	core.AssertEqual(t, tc.expectedHeader, actual, "Retry-After header")
}

func newSetRetryAfterTestCase(name string, retryAfter time.Duration,
	expectedHeader string) setRetryAfterTestCase {
	return setRetryAfterTestCase{
		name:           name,
		retryAfter:     retryAfter,
		expectedHeader: expectedHeader,
	}
}

func TestSetRetryAfter(t *testing.T) {
	testCases := []setRetryAfterTestCase{
		newSetRetryAfterTestCase("zero duration", 0, "0"),
		newSetRetryAfterTestCase("1 second", 1*time.Second, "1"),
		newSetRetryAfterTestCase("60 seconds", 60*time.Second, "60"),
		newSetRetryAfterTestCase("1 minute", 1*time.Minute, "60"),
		newSetRetryAfterTestCase("2 hours", 2*time.Hour, "7200"),
		newSetRetryAfterTestCase("rounds up 500ms", 500*time.Millisecond, "1"),
		newSetRetryAfterTestCase("rounds up 1ms", 1*time.Millisecond, "1"),
		newSetRetryAfterTestCase("rounds up 1.5s", 1500*time.Millisecond, "2"),
		newSetRetryAfterTestCase("rounds up 2.1s", 2100*time.Millisecond, "3"),
		newSetRetryAfterTestCase("negative becomes 0", -10*time.Second, "0"),
		newSetRetryAfterTestCase("large duration", 24*time.Hour, "86400"),
	}

	core.RunTestCases(t, testCases)
}
