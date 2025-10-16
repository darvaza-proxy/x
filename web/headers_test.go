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

var _ core.TestCase = setLastModifiedHeaderTestCase{}

type setLastModifiedHeaderTestCase struct {
	name           string
	lastModified   time.Time
	existingHeader string
	expectSet      bool
	checkTimestamp bool
}

func (tc setLastModifiedHeaderTestCase) Name() string {
	return tc.name
}

func (tc setLastModifiedHeaderTestCase) Test(t *testing.T) {
	t.Helper()

	hdr := make(http.Header)
	if tc.existingHeader != "" {
		hdr.Set(consts.LastModified, tc.existingHeader)
	}

	SetLastModifiedHeader(hdr, tc.lastModified)

	actual := hdr.Get(consts.LastModified)
	if tc.expectSet {
		core.AssertNotEqual(t, "", actual, "Last-Modified header should be set")

		parsed, err := http.ParseTime(actual)
		core.AssertNoError(t, err, "Last-Modified header should be valid HTTP-date")

		if tc.checkTimestamp {
			var expected = tc.lastModified.UTC().Truncate(time.Second)

			core.AssertEqual(t, expected, parsed.UTC().Truncate(time.Second), "parsed time")
		}
	} else {
		core.AssertEqual(t, tc.existingHeader, actual, "Last-Modified header unchanged")
	}
}

func newSetLastModifiedHeaderTestCase(
	name string, lastModified time.Time, existingHeader string, expectSet, checkTimestamp bool,
) setLastModifiedHeaderTestCase {
	return setLastModifiedHeaderTestCase{
		name:           name,
		lastModified:   lastModified,
		existingHeader: existingHeader,
		expectSet:      expectSet,
		checkTimestamp: checkTimestamp,
	}
}

func TestSetLastModifiedHeader(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	existing := "Mon, 01 Jan 2024 00:00:00 GMT"

	testCases := []setLastModifiedHeaderTestCase{
		newSetLastModifiedHeaderTestCase("with time", past, "", true, true),
		newSetLastModifiedHeaderTestCase("with zero time uses current", time.Time{}, "", true, false),
		newSetLastModifiedHeaderTestCase("preserves existing", past, existing, false, false),
	}

	core.RunTestCases(t, testCases)
}

var _ core.TestCase = checkIfModifiedSinceTestCase{}

type checkIfModifiedSinceTestCase struct {
	name           string
	lastModified   time.Time
	headerValue    string
	expectModified bool
}

func (tc checkIfModifiedSinceTestCase) Name() string {
	return tc.name
}

func (tc checkIfModifiedSinceTestCase) Test(t *testing.T) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	core.AssertNoError(t, err, "create request")

	if tc.headerValue != "" {
		req.Header.Set(consts.IfModifiedSince, tc.headerValue)
	}

	result := CheckIfModifiedSince(req, tc.lastModified)
	core.AssertEqual(t, tc.expectModified, result, "modified check")
}

func newCheckIfModifiedSinceTestCase(name string, lastModified time.Time,
	headerValue string, expectModified bool) checkIfModifiedSinceTestCase {
	return checkIfModifiedSinceTestCase{
		name:           name,
		lastModified:   lastModified,
		headerValue:    headerValue,
		expectModified: expectModified,
	}
}

func TestCheckIfModifiedSince(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	testCases := []checkIfModifiedSinceTestCase{
		newCheckIfModifiedSinceTestCase("no header", now, "", true),
		newCheckIfModifiedSinceTestCase("zero lastModified", time.Time{}, now.Format(http.TimeFormat), true),
		newCheckIfModifiedSinceTestCase("malformed header", now, "invalid", true),
		newCheckIfModifiedSinceTestCase("modified after header", now, past.Format(http.TimeFormat), true),
		newCheckIfModifiedSinceTestCase("not modified", past, now.Format(http.TimeFormat), false),
		newCheckIfModifiedSinceTestCase("same time", now, now.Format(http.TimeFormat), false),
		newCheckIfModifiedSinceTestCase("future header", past, future.Format(http.TimeFormat), false),
	}

	core.RunTestCases(t, testCases)
}
