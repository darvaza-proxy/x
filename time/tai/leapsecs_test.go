package tai

import (
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement core.TestCase
var _ core.TestCase = lsoffsetTestCase{}

// lsoffsetTestCase tests the lsoffset UTC to TAI conversion helper.
type lsoffsetTestCase struct {
	date     time.Time
	name     string
	expected uint64
}

func newLsoffsetTestCase(name string, date time.Time, expected uint64) lsoffsetTestCase {
	return lsoffsetTestCase{
		name:     name,
		date:     date,
		expected: expected,
	}
}

// newLsoffsetDateTestCase creates a test case for midnight UTC on the
// first day of the given month, named after the date itself.
func newLsoffsetDateTestCase(year int, month time.Month, expected uint64) lsoffsetTestCase {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return newLsoffsetTestCase(date.Format("2006-01-02"), date, expected)
}

// Name returns the test case name for identification.
func (tc lsoffsetTestCase) Name() string {
	return tc.name
}

// Test validates the leap second offset for the test case date.
func (tc lsoffsetTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.expected, lsoffset(tc.date), "offset")
}

func TestLsoffsetBeforeFirstLeapSecond(t *testing.T) {
	testCases := []lsoffsetTestCase{
		newLsoffsetTestCase("Unix epoch", time.Unix(0, 0), 0),
		newLsoffsetTestCase("one day before first leap second",
			time.Date(1972, time.June, 30, 23, 59, 59, 0, time.UTC), 0),
		newLsoffsetDateTestCase(1970, time.January, 0),
	}

	core.RunTestCases(t, testCases)
}

func TestLsoffsetAtLeapSecondBoundaries(t *testing.T) {
	testCases := []lsoffsetTestCase{
		newLsoffsetDateTestCase(1972, time.July, 11),
		newLsoffsetDateTestCase(1973, time.January, 12),
		newLsoffsetDateTestCase(1974, time.January, 13),
		newLsoffsetDateTestCase(1976, time.January, 15),
		newLsoffsetDateTestCase(1980, time.January, 19),
		newLsoffsetDateTestCase(1990, time.January, 25),
		newLsoffsetDateTestCase(2000, time.January, 32),
		newLsoffsetDateTestCase(2017, time.January, 37),
	}

	core.RunTestCases(t, testCases)
}

func TestLsoffsetBetweenLeapSeconds(t *testing.T) {
	testCases := []lsoffsetTestCase{
		newLsoffsetTestCase("between first and second leap second",
			time.Date(1972, time.December, 31, 23, 59, 59, 0, time.UTC), 11),
		newLsoffsetTestCase("middle of 1973",
			time.Date(1973, time.June, 1, 12, 0, 0, 0, time.UTC), 12),
		newLsoffsetTestCase("gap year with no leap second (1984)",
			time.Date(1984, time.December, 1, 0, 0, 0, 0, time.UTC), 22),
		newLsoffsetTestCase("gap year with no leap second (1987)",
			time.Date(1987, time.June, 1, 0, 0, 0, 0, time.UTC), 23),
	}

	core.RunTestCases(t, testCases)
}

func TestLsoffsetAfterLastLeapSecond(t *testing.T) {
	// The "current time" expectation is derived from the table so the
	// test stays deterministic when new leap seconds are recorded; the
	// fixed dates above pin the historical values.
	lastOffset := uint64(leapseconds[len(leapseconds)-1].offset)

	testCases := []lsoffsetTestCase{
		newLsoffsetDateTestCase(2017, time.June, 37),
		newLsoffsetDateTestCase(2020, time.January, 37),
		newLsoffsetTestCase("current time", time.Now(), lastOffset),
	}

	core.RunTestCases(t, testCases)
}

func TestLsoffsetEdgeCases(t *testing.T) {
	testCases := []lsoffsetTestCase{
		newLsoffsetTestCase("one second before July 1, 1972",
			time.Date(1972, time.June, 30, 23, 59, 59, 0, time.UTC), 0),
		newLsoffsetTestCase("exactly July 1, 1972 00:00:00",
			time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC), 11),
		newLsoffsetTestCase("one second after July 1, 1972",
			time.Date(1972, time.July, 1, 0, 0, 1, 0, time.UTC), 11),
		newLsoffsetTestCase("Y2K boundary",
			time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC), 32),
		newLsoffsetTestCase("Y2K",
			time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), 32),
	}

	core.RunTestCases(t, testCases)
}

// lsoffsetTableTestCases derives a test case from every entry of the
// leap second table, validating lsoffset at each boundary.
func lsoffsetTableTestCases() []lsoffsetTestCase {
	cases := make([]lsoffsetTestCase, 0, len(leapseconds))
	for _, ls := range leapseconds {
		date := time.Unix(ls.begin, 0).UTC()
		cases = append(cases,
			newLsoffsetTestCase(date.Format("2006-01-02"), date, uint64(ls.offset)))
	}
	return cases
}

func TestLsoffsetSequentialIncrease(t *testing.T) {
	core.RunTestCases(t, lsoffsetTableTestCases())
}

func TestLeapSecondDataIntegrity(t *testing.T) {
	core.AssertMustTrue(t, len(leapseconds) > 0, "table not empty")

	t.Run("chronological order", runTestChronologicalOrder)
	t.Run("first leap second", runTestFirstLeapSecond)
	t.Run("last leap second", runTestLastLeapSecond)
}

func runTestChronologicalOrder(t *testing.T) {
	t.Helper()

	for i := 1; i < len(leapseconds); i++ {
		prev := leapseconds[i-1]
		curr := leapseconds[i]

		core.AssertTrue(t, curr.begin > prev.begin, "chronological order at %d", i)
		core.AssertEqual(t, prev.offset+1, curr.offset, "offset increment at %d", i)
	}
}

func runTestFirstLeapSecond(t *testing.T) {
	t.Helper()

	first := leapseconds[0]
	expectedSec := time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC).Unix()
	core.AssertEqual(t, expectedSec, first.begin, "first leap second date")
	core.AssertEqual(t, 11, first.offset, "first leap second offset")
}

func runTestLastLeapSecond(t *testing.T) {
	t.Helper()

	last := leapseconds[len(leapseconds)-1]
	expectedSec := time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	core.AssertEqual(t, expectedSec, last.begin, "last leap second date")
	core.AssertEqual(t, 37, last.offset, "last leap second offset")
}
