package tai

import (
	"math"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

const zeroEpochName = "Zero epoch"

var (
	_ core.TestCase = taiaSplitTestCase{}
	_ core.TestCase = taiFromTimeTestCase{}
	_ core.TestCase = taiaToTAITestCase{}
	_ core.TestCase = taiaToTAINTestCase{}
	_ core.TestCase = combineTestCase{}
)

type taiaSplitTestCase struct {
	name                       string
	input, seconds, nano, atto num.Int128
}

func newTAIASplitTestCase(name string, input, seconds num.Int128, nano, atto int64) taiaSplitTestCase {
	return taiaSplitTestCase{name: name, input: input, seconds: seconds,
		nano: num.AsInt128(nano), atto: num.AsInt128(atto)}
}

func (tc taiaSplitTestCase) Name() string { return tc.name }

func (tc taiaSplitTestCase) Test(t *testing.T) {
	s, n, a := (TAIA{x: tc.input}).Split()
	core.AssertEqual(t, tc.seconds, s, "Seconds")
	core.AssertEqual(t, tc.nano, n, "Nanoseconds")
	core.AssertEqual(t, tc.atto, a, "Attoseconds")

	limit := num.AsInt128(AttosPerNano)
	core.AssertFalse(t, n.IsNegative(), "nanoseconds sign")
	core.AssertTrue(t, n.Cmp(limit) < 0, "nanoseconds upper bound")
	core.AssertFalse(t, a.IsNegative(), "attoseconds sign")
	core.AssertTrue(t, a.Cmp(limit) < 0, "attoseconds upper bound")
}

func taiaSplitNativeTestCases() []taiaSplitTestCase {
	return core.S(
		newTAIASplitTestCase(zeroEpochName, num.AsInt128(0), num.AsInt128(0), 0, 0),
		newTAIASplitTestCase("Positive time: 2.500000001000000003 seconds",
			num.AsInt128(2_500_000_001_000_000_003), num.AsInt128(2), 500_000_001, 3),
		newTAIASplitTestCase("Pre-epoch: -0.25 seconds", num.AsInt128(-250_000_000_000_000_000),
			num.AsInt128(-1), 750_000_000, 0),
		newTAIASplitTestCase("Pre-epoch with sub-nanoseconds: -0.5 ns", num.AsInt128(-500_000_000),
			num.AsInt128(-1), 999_999_999, 500_000_000),
		newTAIASplitTestCase("Exact negative second", num.AsInt128(-AttosPerSec), num.AsInt128(-1), 0, 0),
		newTAIASplitTestCase("Just past -1 second", num.AsInt128(-AttosPerSec-1),
			num.AsInt128(-2), 999_999_999, 999_999_999),
		newTAIASplitTestCase("One attosecond below a second", num.AsInt128(AttosPerSec-1),
			num.AsInt128(0), 999_999_999, 999_999_999),
	)
}

func taiaSplitWideTestCases() []taiaSplitTestCase {
	millionYears := num.AsInt128(-31_556_952_000_000)
	thousandYears := num.AsInt128(-31_556_952_000)
	scale := num.AsInt128(AttosPerSec)
	return core.S(
		newTAIASplitTestCase("Deep pre-epoch: -1,000,000 years", millionYears.Mul(scale), millionYears, 0, 0),
		newTAIASplitTestCase("Deep pre-epoch: -1000 years plus one attosecond",
			thousandYears.Mul(scale).Add(num.AsInt128(1)), thousandYears, 0, 1),
		newTAIASplitTestCase("Maximum 128-bit count", num.MaxInt128,
			num.NewInt128(0x9, 0x392ee8e921d5d073), 687_303_715, 884_105_727),
		newTAIASplitTestCase("Minimum 128-bit count", num.MinInt128,
			num.NewInt128(0xfffffffffffffff6, 0xc6d11716de2a2f8c), 312_696_284, 115_894_272),
	)
}

func TestTAIASplit(t *testing.T) {
	core.RunTestCases(t, append(taiaSplitNativeTestCases(), taiaSplitWideTestCases()...))
}

type taiFromTimeTestCase struct {
	input time.Time
	name  string
	want  num.Int128
}

// newTAIFromTimeTestCase takes the expected TAI seconds and nanoseconds separately,
// forming the attosecond count after widening to avoid native integer overflow.
func newTAIFromTimeTestCase(name string, input time.Time, seconds num.Int128, nano int64) taiFromTimeTestCase {
	want := seconds.Mul(num.AsInt128(AttosPerSec)).
		Add(num.AsInt128(nano).Mul(num.AsInt128(AttosPerNano)))
	return taiFromTimeTestCase{name: name, input: input, want: want}
}

func (tc taiFromTimeTestCase) Name() string { return tc.name }

func (tc taiFromTimeTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, TAIFromTime(tc.input).x, "TAIFromTime")
}

func taiFromTimeTestCases() []taiFromTimeTestCase {
	epoch := num.AsInt128(int64(TAICONST))
	modern := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ancient := time.Date(100, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	leap := time.Date(2016, 12, 31, 23, 59, 59, 0, time.UTC)
	return core.S(
		newTAIFromTimeTestCase("Unix epoch", time.Unix(0, 0).UTC(), epoch, 0),
		newTAIFromTimeTestCase("Unix epoch with nanoseconds", time.Unix(0, 500_000_000).UTC(), epoch, 500_000_000),
		newTAIFromTimeTestCase("Pre-1970 date", time.Unix(-1, 750_000_000).UTC(),
			epoch.Sub(num.AsInt128(1)), 750_000_000),
		newTAIFromTimeTestCase("Modern date with leap seconds", modern,
			num.AsInt128(modern.Unix()+int64(TAICONST)+currentLeapSeconds), 0),
		newTAIFromTimeTestCase("Sub-second nanosecond boundary", time.Unix(0, 999_999_999).UTC(), epoch, 999_999_999),
		newTAIFromTimeTestCase("Historical ancient date", ancient, num.AsInt128(ancient.Unix()+int64(TAICONST)), 0),
		newTAIFromTimeTestCase("Far future year 9999", future,
			num.AsInt128(future.Unix()+int64(TAICONST)+currentLeapSeconds), 999_999_999),
		newTAIFromTimeTestCase("Minimum Unix second", time.Unix(math.MinInt64, 0).UTC(),
			num.AsInt128(math.MinInt64+int64(TAICONST)), 0),
		newTAIFromTimeTestCase("Maximum Unix second", time.Unix(math.MaxInt64, 999_999_999).UTC(),
			num.NewInt128(0, uint64(math.MaxInt64)+TAICONST+currentLeapSeconds), 999_999_999),
		newTAIFromTimeTestCase("Leap second boundary", leap, num.AsInt128(leap.Unix()+int64(TAICONST)+26), 0),
	)
}

func taiFromTimeLeapTestCases() []taiFromTimeTestCase {
	first := time.Unix(78796800, 0).UTC()
	latest := time.Unix(1483228800, 0).UTC()
	return core.S(
		// DJB's published UTC-to-TAI64 example: https://cr.yp.to/libtai/tai64.html.
		newTAIFromTimeTestCase("DJB reference label", time.Date(1992, 6, 2, 8, 6, 43, 0, time.UTC),
			num.NewInt128(0, 0x400000002a2b2c2d), 0),
		newTAIFromTimeTestCase("Before first leap", first.Add(-time.Nanosecond),
			num.AsInt128(first.Unix()-1+int64(TAICONST)), 999_999_999),
		newTAIFromTimeTestCase("At first leap", first, num.AsInt128(first.Unix()+int64(TAICONST)+1), 0),
		newTAIFromTimeTestCase("Before latest leap", latest.Add(-time.Nanosecond),
			num.AsInt128(latest.Unix()-1+int64(TAICONST)+26), 999_999_999),
		newTAIFromTimeTestCase("At latest leap", latest,
			num.AsInt128(latest.Unix()+int64(TAICONST)+currentLeapSeconds), 0),
	)
}

func TestTAIFromTime(t *testing.T) {
	core.RunTestCases(t, append(taiFromTimeTestCases(), taiFromTimeLeapTestCases()...))
}

func TestTAIFromTimeAndSplitRoundTrip(t *testing.T) {
	tm := time.Date(2026, 7, 22, 10, 0, 0, 123456789, time.UTC)
	taia := TAIFromTime(tm)
	s, n, a := taia.Split()

	expectedSec := num.AsInt128(tm.Unix() + int64(TAICONST) + currentLeapSeconds)
	expectedNano := num.AsInt128(int64(tm.Nanosecond()))
	expectedAtto := num.AsInt128(0)
	core.AssertEqual(t, expectedSec, s, "Seconds")
	core.AssertEqual(t, expectedNano, n, "Nanoseconds")
	core.AssertEqual(t, expectedAtto, a, "Attoseconds")
}

type taiaToTAITestCase struct {
	name        string
	input, want num.Int128
}

func newTAIAToTAITestCase(name string, input, want int64) taiaToTAITestCase {
	return taiaToTAITestCase{name: name, input: num.AsInt128(input), want: num.AsInt128(want)}
}

func (tc taiaToTAITestCase) Name() string { return tc.name }

func (tc taiaToTAITestCase) Test(t *testing.T) {
	got := (TAIA{x: tc.input}).TAI()
	core.AssertEqual(t, tc.want, got.x, "seconds")
}

type taiaToTAINTestCase struct {
	name        string
	input, want num.Int128
}

func newTAIAToTAINTestCase(name string, input, want int64) taiaToTAINTestCase {
	return taiaToTAINTestCase{name: name, input: num.AsInt128(input), want: num.AsInt128(want)}
}

func (tc taiaToTAINTestCase) Name() string { return tc.name }

func (tc taiaToTAINTestCase) Test(t *testing.T) {
	got := (TAIA{x: tc.input}).TAIN()
	core.AssertEqual(t, tc.want, got.x, "nanoseconds")
}

func TestTAIAToTAI(t *testing.T) {
	core.RunTestCases(t, core.S(
		newTAIAToTAITestCase(zeroEpochName, 0, 0),
		newTAIAToTAITestCase("Positive seconds with exact boundary", 5_000_000_000_000_000_000, 5),
		newTAIAToTAITestCase("Positive seconds with sub-second remainder", 5_999_999_999_999_999_999, 5),
		newTAIAToTAITestCase("Negative seconds with sub-second remainder", -250_000_000_000_000_000, -1),
		newTAIAToTAITestCase("Negative seconds exact boundary", -AttosPerSec, -1),
		newTAIAToTAITestCase("Negative seconds past boundary", -1_500_000_000_000_000_000, -2),
	))
}

func TestTAIAToTAIN(t *testing.T) {
	core.RunTestCases(t, core.S(
		newTAIAToTAINTestCase(zeroEpochName, 0, 0),
		newTAIAToTAINTestCase("Positive conversion with exact nanoseconds", 1_000_000_500_000_000_000, 1_000_000_500),
		newTAIAToTAINTestCase("Positive conversion with attosecond truncation",
			1_000_000_500_000_000_999, 1_000_000_500),
		newTAIAToTAINTestCase("Negative conversion with attosecond remainder", -500_500_000_000, -501),
		newTAIAToTAINTestCase("Negative conversion exact boundary", -1_000_000_000_000, -1_000),
	))
}

// combineTestCase checks the assembled count independently of Split.
type combineTestCase struct {
	name                      string
	seconds, nano, atto, want num.Int128
}

func newCombineTestCase(name string, seconds, nano, atto, want num.Int128) combineTestCase {
	return combineTestCase{name: name, seconds: seconds, nano: nano, atto: atto, want: want}
}

func (tc combineTestCase) Name() string { return tc.name }

func (tc combineTestCase) Test(t *testing.T) {
	got := Combine(tc.seconds, tc.nano, tc.atto)
	core.AssertEqual(t, tc.want, got.x, "attosecond count")
}

// combineTestCases reuses the independently specified split fixtures, reversing
// their inputs and expected parts without calling Split to compute an answer.
func combineTestCases() []combineTestCase {
	parts := append(taiaSplitNativeTestCases(), taiaSplitWideTestCases()...)
	cases := make([]combineTestCase, 0, len(parts))
	for _, tc := range parts {
		cases = append(cases, newCombineTestCase(tc.name, tc.seconds, tc.nano, tc.atto, tc.input))
	}
	return cases
}

func TestCombine(t *testing.T) {
	core.RunTestCases(t, combineTestCases())
}

func TestNow(t *testing.T) {
	before := TAIFromTime(time.Now())
	got := Now()
	after := TAIFromTime(time.Now())
	core.AssertTrue(t, got.x.Cmp(before.x) >= 0, "lower clock bound")
	core.AssertTrue(t, got.x.Cmp(after.x) <= 0, "upper clock bound")
	core.AssertEqual(t, num.ZeroInt128, got.x.Mod(num.AsInt128(AttosPerNano)), "sub-nanosecond count")
}
