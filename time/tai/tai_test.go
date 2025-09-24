package tai

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement core.TestCase
var _ core.TestCase = taiFromStringTestCase{}

// taiFromStringTestCase tests ParseTAI parsing.
type taiFromStringTestCase struct {
	input   string
	name    string
	wantErr bool
}

func newTAIFromStringTestCase(name, input string, wantErr bool) taiFromStringTestCase {
	return taiFromStringTestCase{
		name:    name,
		input:   input,
		wantErr: wantErr,
	}
}

// Name returns the test case name for identification.
func (tc taiFromStringTestCase) Name() string {
	return tc.name
}

// Test validates ParseTAI parsing behaviour.
func (tc taiFromStringTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := ParseTAI(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "parse %q", tc.input)
		return
	}
	core.AssertNoError(t, err, "parse %q", tc.input)
}

func TestTAIFromString(t *testing.T) {
	testCases := core.S(
		newTAIFromStringTestCase("valid zero timestamp", "@4000000000000000", false),
		newTAIFromStringTestCase("valid timestamp", "@400000004B05F9FE", false),
		newTAIFromStringTestCase("invalid string", "invalid", true),
		newTAIFromStringTestCase("missing @ prefix", "4000000000000000", true),
		newTAIFromStringTestCase("invalid hex payload", "@invalid", true),
		newTAIFromStringTestCase("invalid hex, correct length",
			"@GGGGGGGGGGGGGGGG", true),
	)

	core.RunTestCases(t, testCases)
}

func TestTAIFromStringSentinels(t *testing.T) {
	_, err := ParseTAI("@40")
	core.AssertErrorIs(t, err, ErrInvalidTAILength, "length error")

	_, err = ParseTAI("X4000000000000000")
	core.AssertErrorIs(t, err, ErrInvalidTAIFormat, "format error")
}

func TestNowTAI(t *testing.T) {
	core.AssertFalse(t, NowTAI().IsZero(), "zero time")
	core.AssertTrue(t, TAI{}.IsZero(), "zero value")
}

func TestTAIComparison(t *testing.T) {
	t1 := TAIFromTime(time.Unix(1000, 0))
	t2 := TAIFromTime(time.Unix(2000, 0))
	t3 := TAIFromTime(time.Unix(1000, 0))

	core.AssertTrue(t, t1.Before(t2), "t1 before t2")
	core.AssertTrue(t, t2.After(t1), "t2 after t1")
	core.AssertFalse(t, t1.Equal(t2), "t1 equal t2")
	core.AssertTrue(t, t1.Equal(t3), "t1 equal t3")

	core.AssertEqual(t, -1, t1.Compare(t2), "t1 compare t2")
	core.AssertEqual(t, 1, t2.Compare(t1), "t2 compare t1")
	core.AssertEqual(t, 0, t1.Compare(t3), "t1 compare t3")
}

func TestTAIArithmetic(t *testing.T) {
	base := TAIFromTime(time.Unix(1000, 0))
	dur := 60 * time.Second

	later, err := base.Add(dur)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, dur, later.Sub(base), "add/sub round trip")

	// Sub-second components are truncated
	truncated, err := base.Add(1500 * time.Millisecond)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, time.Second, truncated.Sub(base), "sub-second truncation")

	earlier, err := base.Add(-30 * time.Second)
	core.AssertMustNoError(t, err, "add negative")
	core.AssertEqual(t, -30*time.Second, earlier.Sub(base), "negative add")
}

func TestTAISubOverflow(t *testing.T) {
	big := TAI{x: math.MaxUint64}
	small := TAI{x: 0}

	core.AssertEqual(t, maxDuration, big.Sub(small), "clamp to max duration")
	core.AssertEqual(t, minDuration, small.Sub(big), "clamp to min duration")
}

func TestTAISubNormal(t *testing.T) {
	t1 := TAI{x: 1005}
	t2 := TAI{x: 1000}
	core.AssertEqual(t, 5*time.Second, t1.Sub(t2), "normal diff")
	core.AssertEqual(t, -5*time.Second, t2.Sub(t1), "reversed diff")
}

func TestTAIArithmeticOverflow(t *testing.T) {
	_, err := TAI{x: math.MaxUint64 - 100}.Add(200 * time.Second)
	core.AssertErrorIs(t, err, ErrOverflow, "overflow")

	_, err = TAI{x: 100}.Add(-200 * time.Second)
	core.AssertErrorIs(t, err, ErrUnderflow, "underflow")
}

func TestTAIMarshalJSON(t *testing.T) {
	tm := TAIFromTime(time.Unix(1000, 0))
	data, err := json.Marshal(tm)
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAI
	core.AssertMustNoError(t, json.Unmarshal(data, &tm2), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "JSON round trip")

	core.AssertError(t, json.Unmarshal([]byte("123"), &tm2), "non-string JSON")
	core.AssertError(t, json.Unmarshal([]byte(`"bogus"`), &tm2), "invalid value")
}

func TestTAIMarshalBinary(t *testing.T) {
	tm := TAIFromTime(time.Unix(1000, 0))
	data, err := tm.MarshalBinary()
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAI
	core.AssertMustNoError(t, tm2.UnmarshalBinary(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "binary round trip")
}

func TestTAIUnmarshalBinaryInvalidLength(t *testing.T) {
	var tm TAI
	err := tm.UnmarshalBinary([]byte{1, 2, 3})
	core.AssertErrorIs(t, err, ErrInvalidTAIBinaryLength, "length error")
	core.AssertErrorIs(t, err, core.ErrInvalid, "core.ErrInvalid chain")
}

func TestConversion(t *testing.T) {
	t.Run("TAI to TAIN", runTestTAIToTAIN)
	t.Run("TAI to TAINA", runTestTAIToTAINA)
	t.Run("TAIN to TAI", runTestTAINToTAI)
}

func runTestTAIToTAINA(t *testing.T) {
	t.Helper()

	tai := TAIFromTime(time.Unix(1000, 0))
	taina := tai.TAINA()
	core.AssertEqual(t, tai.Unix(), taina.Unix(), "unix seconds")
	core.AssertEqual(t, 0, taina.Nanosecond(), "nanoseconds")
	core.AssertEqual(t, 0, taina.Attosecond(), "attoseconds")
}

func runTestTAIToTAIN(t *testing.T) {
	t.Helper()

	tai := TAIFromTime(time.Unix(1000, 0))
	core.AssertEqual(t, 0, tai.TAIN().Nanosecond(), "nanoseconds")
}

func runTestTAINToTAI(t *testing.T) {
	t.Helper()

	tain := TAINFromTime(time.Unix(1000, 123456789))
	core.AssertEqual(t, tain.Unix(), tain.TAI().Unix(), "unix seconds")
}

func TestTAIUnixConversions(t *testing.T) {
	tm := TAIFromTime(time.Unix(1000, 0))
	core.AssertEqual(t, int64(1000), tm.Unix(), "unix")
	core.AssertEqual(t, int64(1000000), tm.UnixMilli(), "milliseconds")
	core.AssertEqual(t, int64(1000000000), tm.UnixMicro(), "microseconds")
	core.AssertEqual(t, int64(1000000000000), tm.UnixNano(), "nano")
}

func TestTAISinceUntil(t *testing.T) {
	t1 := TAIFromTime(time.Unix(1000, 0))
	t2 := TAIFromTime(time.Unix(1060, 0))
	core.AssertEqual(t, time.Minute, t2.Since(t1), "since")
	core.AssertEqual(t, time.Minute, t1.Until(t2), "until")
}

func TestTAIGoTimeRoundTrip(t *testing.T) {
	// A modern date exercises a non-zero leap second offset
	g := time.Date(2024, time.March, 15, 12, 30, 45, 0, time.UTC)
	core.AssertTrue(t, TAIFromTime(g).GoTime().Equal(g), "round trip")

	// Just before a leap second boundary the TAI label lands past the
	// boundary and needs the offset refinement in utcFromTAI
	b := time.Date(2016, time.December, 31, 23, 59, 59, 0, time.UTC)
	core.AssertTrue(t, TAIFromTime(b).GoTime().Equal(b), "before boundary")

	a := time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)
	core.AssertTrue(t, TAIFromTime(a).GoTime().Equal(a), "at boundary")
}

func TestTAIFormat(t *testing.T) {
	g := time.Date(2024, time.March, 15, 12, 30, 45, 0, time.UTC)
	core.AssertEqual(t, g.Format(time.RFC3339), TAIFromTime(g).Format(time.RFC3339),
		"format")
}

func TestTAIMarshalText(t *testing.T) {
	tm := TAIFromTime(time.Unix(1000, 0))
	data, err := tm.MarshalText()
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAI
	core.AssertMustNoError(t, tm2.UnmarshalText(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "text round trip")

	core.AssertError(t, tm2.UnmarshalText([]byte("bogus")), "invalid text")
}

func TestTAIAddSubSecondTruncation(t *testing.T) {
	base := TAI{x: 1000}
	result, err := base.Add(-1500 * time.Millisecond)
	core.AssertMustNoError(t, err, "add")
	// truncates toward zero: -1.5s truncates to -1s, not -2s
	core.AssertEqual(t, TAI{x: 999}, result, "sub-second truncates toward zero")
}
