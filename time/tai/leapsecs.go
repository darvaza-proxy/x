package tai

import (
	"slices"
	"time"
)

const (
	// currentTAIOffset is the total TAI-UTC offset in the latest table entry,
	// effective from 2017-01-01 UTC, in seconds.
	currentTAIOffset = 37

	// currentLeapSeconds excludes the baseline already included in TAICONST.
	currentLeapSeconds = currentTAIOffset - taiBaseOffset
)

// leapsecond represents a leap second event in the TAI-UTC offset table.
// Each leap second increases the difference between TAI and UTC by one second.
type leapsecond struct {
	begin  int64 // Unix timestamp when this leap second offset becomes effective
	offset int   // TAI-UTC offset in seconds (cumulative number of leap seconds + 10)
}

var leapseconds = []*leapsecond{
	{78796800, 11},                 // 1972-07-01 00:00:00 UTC
	{94694400, 12},                 // 1973-01-01 00:00:00 UTC
	{126230400, 13},                // 1974-01-01 00:00:00 UTC
	{157766400, 14},                // 1975-01-01 00:00:00 UTC
	{189302400, 15},                // 1976-01-01 00:00:00 UTC
	{220924800, 16},                // 1977-01-01 00:00:00 UTC
	{252460800, 17},                // 1978-01-01 00:00:00 UTC
	{283996800, 18},                // 1979-01-01 00:00:00 UTC
	{315532800, 19},                // 1980-01-01 00:00:00 UTC
	{362793600, 20},                // 1981-07-01 00:00:00 UTC
	{394329600, 21},                // 1982-07-01 00:00:00 UTC
	{425865600, 22},                // 1983-07-01 00:00:00 UTC
	{489024000, 23},                // 1985-07-01 00:00:00 UTC
	{567993600, 24},                // 1988-01-01 00:00:00 UTC
	{631152000, 25},                // 1990-01-01 00:00:00 UTC
	{662688000, 26},                // 1991-01-01 00:00:00 UTC
	{709948800, 27},                // 1992-07-01 00:00:00 UTC
	{741484800, 28},                // 1993-07-01 00:00:00 UTC
	{773020800, 29},                // 1994-07-01 00:00:00 UTC
	{820454400, 30},                // 1996-01-01 00:00:00 UTC
	{867715200, 31},                // 1997-07-01 00:00:00 UTC
	{915148800, 32},                // 1999-01-01 00:00:00 UTC
	{1136073600, 33},               // 2006-01-01 00:00:00 UTC
	{1230768000, 34},               // 2009-01-01 00:00:00 UTC
	{1341100800, 35},               // 2012-07-01 00:00:00 UTC
	{1435708800, 36},               // 2015-07-01 00:00:00 UTC
	{1483228800, currentTAIOffset}, // 2017-01-01 00:00:00 UTC
}

// lsoffset returns the additional leap seconds beyond the baseline in TAICONST.
// It returns zero before the first table entry.
func lsoffset(t time.Time) uint64 {
	unixSec := t.Unix()
	for _, leap := range slices.Backward(leapseconds) {
		if unixSec >= leap.begin {
			return uint64(leap.offset - taiBaseOffset)
		}
	}

	return 0
}
