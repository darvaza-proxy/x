package tai

import (
	"testing"
	"time"
)

// testLsoffsetHelper runs a set of lsoffset test cases
func testLsoffsetHelper(t *testing.T, tests []struct {
	name     string
	date     time.Time
	expected uint64
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lsoffset(tt.date)
			if result != tt.expected {
				t.Errorf("lsoffset(%v) = %d, want %d", tt.date, result, tt.expected)
			}
		})
	}
}

func TestLsoffsetBeforeFirstLeapSecond(t *testing.T) {
	// Test dates before the first leap second (July 1, 1972)
	tests := []struct {
		name     string
		date     time.Time
		expected uint64
	}{
		{
			name:     "Unix epoch",
			date:     time.Unix(0, 0),
			expected: 0,
		},
		{
			name:     "June 30, 1972 - one day before first leap second",
			date:     time.Date(1972, time.June, 30, 23, 59, 59, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "January 1, 1970",
			date:     time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}

	testLsoffsetHelper(t, tests)
}

func TestLsoffsetAtLeapSecondBoundaries(t *testing.T) {
	// Test exact leap second boundaries
	tests := []struct {
		name     string
		date     time.Time
		expected uint64
	}{
		{
			name:     "First leap second - July 1, 1972",
			date:     time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC),
			expected: 11,
		},
		{
			name:     "January 1, 1973",
			date:     time.Date(1973, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 12,
		},
		{
			name:     "January 1, 1974",
			date:     time.Date(1974, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 13,
		},
		{
			name:     "July 1, 1981",
			date:     time.Date(1981, time.July, 1, 0, 0, 0, 0, time.UTC),
			expected: 20,
		},
		{
			name:     "January 1, 2017 - most recent leap second",
			date:     time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 37,
		},
	}

	testLsoffsetHelper(t, tests)
}

func TestLsoffsetBetweenLeapSeconds(t *testing.T) {
	// Test dates between leap seconds to ensure correct offset is maintained
	tests := []struct {
		name     string
		date     time.Time
		expected uint64
	}{
		{
			name:     "December 31, 1972 - between first and second leap second",
			date:     time.Date(1972, time.December, 31, 23, 59, 59, 0, time.UTC),
			expected: 11,
		},
		{
			name:     "June 1, 1973 - middle of 1973",
			date:     time.Date(1973, time.June, 1, 12, 0, 0, 0, time.UTC),
			expected: 12,
		},
		{
			name:     "December 1, 1984 - gap year with no leap second",
			date:     time.Date(1984, time.December, 1, 0, 0, 0, 0, time.UTC),
			expected: 22, // Should be same as 1983 leap second
		},
		{
			name:     "June 1, 1987 - gap year with no leap second",
			date:     time.Date(1987, time.June, 1, 0, 0, 0, 0, time.UTC),
			expected: 23, // Should be same as 1985 leap second
		},
	}

	testLsoffsetHelper(t, tests)
}

func TestLsoffsetAfterLastLeapSecond(t *testing.T) {
	// Test dates after the last recorded leap second
	tests := []struct {
		name     string
		date     time.Time
		expected uint64
	}{
		{
			name:     "June 1, 2017 - after last leap second",
			date:     time.Date(2017, time.June, 1, 0, 0, 0, 0, time.UTC),
			expected: 37,
		},
		{
			name:     "January 1, 2020 - future date",
			date:     time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 37,
		},
		{
			name:     "Current time (should be 37)",
			date:     time.Now(),
			expected: 37,
		},
	}

	testLsoffsetHelper(t, tests)
}

func TestLsoffsetSequentialIncrease(t *testing.T) {
	// Test that leap second offsets increase sequentially
	tests := []struct {
		date   time.Time
		offset uint64
	}{
		{time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC), 11},
		{time.Date(1973, time.January, 1, 0, 0, 0, 0, time.UTC), 12},
		{time.Date(1974, time.January, 1, 0, 0, 0, 0, time.UTC), 13},
		{time.Date(1975, time.January, 1, 0, 0, 0, 0, time.UTC), 14},
		{time.Date(1976, time.January, 1, 0, 0, 0, 0, time.UTC), 15},
		{time.Date(1977, time.January, 1, 0, 0, 0, 0, time.UTC), 16},
		{time.Date(1978, time.January, 1, 0, 0, 0, 0, time.UTC), 17},
		{time.Date(1979, time.January, 1, 0, 0, 0, 0, time.UTC), 18},
		{time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC), 19},
		{time.Date(1981, time.July, 1, 0, 0, 0, 0, time.UTC), 20},
		{time.Date(1982, time.July, 1, 0, 0, 0, 0, time.UTC), 21},
		{time.Date(1983, time.July, 1, 0, 0, 0, 0, time.UTC), 22},
		{time.Date(1985, time.July, 1, 0, 0, 0, 0, time.UTC), 23},
		{time.Date(1988, time.January, 1, 0, 0, 0, 0, time.UTC), 24},
		{time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), 25},
		{time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC), 26},
		{time.Date(1992, time.July, 1, 0, 0, 0, 0, time.UTC), 27},
		{time.Date(1993, time.July, 1, 0, 0, 0, 0, time.UTC), 28},
		{time.Date(1994, time.July, 1, 0, 0, 0, 0, time.UTC), 29},
		{time.Date(1996, time.January, 1, 0, 0, 0, 0, time.UTC), 30},
		{time.Date(1997, time.July, 1, 0, 0, 0, 0, time.UTC), 31},
		{time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC), 32},
		{time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC), 33},
		{time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC), 34},
		{time.Date(2012, time.July, 1, 0, 0, 0, 0, time.UTC), 35},
		{time.Date(2015, time.July, 1, 0, 0, 0, 0, time.UTC), 36},
		{time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC), 37},
	}

	testSequentialOffsets(t, tests)
}

func testSequentialOffsets(t *testing.T, tests []struct {
	date   time.Time
	offset uint64
}) {
	for i, tt := range tests {
		t.Run(tt.date.Format("2006-01-02"), func(t *testing.T) {
			validateOffset(t, tt.date, tt.offset)

			if i > 0 {
				validateSequential(t, lsoffset(tt.date), tests[i-1].offset)
			}
		})
	}
}

func validateOffset(t *testing.T, date time.Time, expected uint64) {
	result := lsoffset(date)
	if result != expected {
		t.Errorf("lsoffset(%v) = %d, want %d", date, result, expected)
	}
}

func validateSequential(t *testing.T, current, previous uint64) {
	if current != previous+1 {
		t.Errorf("leap second offset should increase by 1: got %d, previous was %d",
			current, previous)
	}
}

func TestLsoffsetEdgeCases(t *testing.T) {
	// Test edge cases and boundary conditions
	edgeCases := []struct {
		name     string
		date     time.Time
		expected uint64
	}{
		{
			name:     "One second before July 1, 1972",
			date:     time.Date(1972, time.June, 30, 23, 59, 59, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "Exactly July 1, 1972 00:00:00",
			date:     time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC),
			expected: 11,
		},
		{
			name:     "One second after July 1, 1972",
			date:     time.Date(1972, time.July, 1, 0, 0, 1, 0, time.UTC),
			expected: 11,
		},
		{
			name:     "December 31, 1999 23:59:59 (Y2K boundary)",
			date:     time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC),
			expected: 32,
		},
		{
			name:     "January 1, 2000 00:00:00 (Y2K)",
			date:     time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: 32,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lsoffset(tc.date)
			if result != tc.expected {
				t.Errorf("lsoffset(%v) = %d, want %d", tc.date, result, tc.expected)
			}
		})
	}
}

func TestLeapSecondDataIntegrity(t *testing.T) {
	// Test that leap second data is properly structured
	if len(leapseconds) == 0 {
		t.Fatal("leapseconds slice should not be empty")
	}

	testChronologicalOrder(t)
	testFirstLeapSecond(t)
	testLastLeapSecond(t)
}

func testChronologicalOrder(t *testing.T) {
	for i := 1; i < len(leapseconds); i++ {
		prev := leapseconds[i-1]
		curr := leapseconds[i]

		if curr.begin <= prev.begin {
			t.Errorf("leap seconds not in chronological order: %d should be after %d",
				curr.begin, prev.begin)
		}

		if curr.offset != prev.offset+1 {
			t.Errorf("leap second offset should increase by 1: %d -> %d",
				prev.offset, curr.offset)
		}
	}
}

func testFirstLeapSecond(t *testing.T) {
	first := leapseconds[0]
	expectedFirst := time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC).Unix()
	if first.begin != expectedFirst {
		t.Errorf("first leap second date = %d, want %d", first.begin, expectedFirst)
	}
	if first.offset != 11 {
		t.Errorf("first leap second offset = %d, want 11", first.offset)
	}
}

func testLastLeapSecond(t *testing.T) {
	last := leapseconds[len(leapseconds)-1]
	expectedLast := time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	if last.begin != expectedLast {
		t.Errorf("last leap second date = %d, want %d", last.begin, expectedLast)
	}
	if last.offset != 37 {
		t.Errorf("last leap second offset = %d, want 37", last.offset)
	}
}
