package tai

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement core.TestCase
var _ core.TestCase = tainFromStringTestCase{}

// tainFromStringTestCase tests ParseTAIN parsing.
type tainFromStringTestCase struct {
	input   string
	name    string
	wantErr bool
}

func newTAINFromStringTestCase(name, input string, wantErr bool) tainFromStringTestCase {
	return tainFromStringTestCase{
		name:    name,
		input:   input,
		wantErr: wantErr,
	}
}

// Name returns the test case name for identification.
func (tc tainFromStringTestCase) Name() string {
	return tc.name
}

// Test validates ParseTAIN parsing behaviour.
func (tc tainFromStringTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := ParseTAIN(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "parse %q", tc.input)
		return
	}
	core.AssertNoError(t, err, "parse %q", tc.input)
}

func TestTAINFromString(t *testing.T) {
	testCases := []tainFromStringTestCase{
		newTAINFromStringTestCase("valid zero timestamp", "@400000000000000000000000", false),
		newTAINFromStringTestCase("valid timestamp", "@400000004B05F9FE00000000", false),
		newTAINFromStringTestCase("invalid string", "invalid", true),
		newTAINFromStringTestCase("missing @ prefix", "400000000000000000000000", true),
		newTAINFromStringTestCase("invalid hex payload", "@invalid", true),
		newTAINFromStringTestCase("invalid hex, correct length",
			"@GGGGGGGGGGGGGGGGGGGGGGGG", true),
		newTAINFromStringTestCase("too short (TAI64 format)", "@4000000000000000", true),
	}

	core.RunTestCases(t, testCases)
}

func TestTAINFromStringSentinels(t *testing.T) {
	_, err := ParseTAIN("@40")
	core.AssertErrorIs(t, err, ErrInvalidTAINLength, "length error")

	_, err = ParseTAIN("X400000000000000000000000")
	core.AssertErrorIs(t, err, ErrInvalidTAINFormat, "format error")
}

func TestNowTAIN(t *testing.T) {
	core.AssertFalse(t, NowTAIN().IsZero(), "zero time")
}

func TestTAINComparison(t *testing.T) {
	t1 := TAINFromTime(time.Unix(1000, 500000000))
	t2 := TAINFromTime(time.Unix(1000, 600000000))
	t3 := TAINFromTime(time.Unix(1000, 500000000))

	core.AssertTrue(t, t1.Before(t2), "t1 before t2")
	core.AssertTrue(t, t2.After(t1), "t2 after t1")
	core.AssertFalse(t, t1.Equal(t2), "t1 equal t2")
	core.AssertTrue(t, t1.Equal(t3), "t1 equal t3")

	core.AssertEqual(t, -1, t1.Compare(t2), "t1 compare t2")
	core.AssertEqual(t, 1, t2.Compare(t1), "t2 compare t1")
	core.AssertEqual(t, 0, t1.Compare(t3), "t1 compare t3")

	t4 := TAINFromTime(time.Unix(2000, 0))
	core.AssertTrue(t, t1.Before(t4), "cross-second before")
	core.AssertTrue(t, t4.After(t1), "cross-second after")
}

func TestTAINArithmetic(t *testing.T) {
	base := TAINFromTime(time.Unix(1000, 500000000))
	dur := 60*time.Second + 250*time.Millisecond

	later, err := base.Add(dur)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, dur, later.Sub(base), "add/sub round trip")

	// Sub borrows a second when the nanosecond difference is negative
	x := UnixTAIN(1001, 100)
	y := UnixTAIN(1000, 900)
	core.AssertEqual(t, time.Second-800*time.Nanosecond, x.Sub(y), "nanosecond borrow")
}

func TestTAINArithmeticOverflow(t *testing.T) {
	_, err := TAIN{sec: math.MaxUint64 - 100}.Add(200 * time.Second)
	core.AssertErrorIs(t, err, ErrOverflow, "overflow")

	_, err = TAIN{sec: 100}.Add(-200 * time.Second)
	core.AssertErrorIs(t, err, ErrUnderflow, "underflow")
}

func TestTAINNanoseconds(t *testing.T) {
	tm := TAINFromTime(time.Unix(1000, 123456789))
	core.AssertEqual(t, 123456789, tm.Nanosecond(), "nanoseconds")
}

func TestTAINTruncateRound(t *testing.T) {
	t.Run("epoch era", runTestTAINTruncateRoundEpoch)
	t.Run("modern date", runTestTAINTruncateRoundModern)
	t.Run("pre-epoch", runTestTAINTruncatePreEpoch)
}

func runTestTAINTruncateRoundEpoch(t *testing.T) {
	t.Helper()

	tm := TAINFromTime(time.Unix(1000, 567890123))

	truncated := tm.Truncate(time.Second)
	core.AssertTrue(t, truncated.Equal(TAINFromTime(time.Unix(1000, 0))),
		"truncate to second")

	rounded := tm.Round(time.Second)
	core.AssertTrue(t, rounded.Equal(TAINFromTime(time.Unix(1001, 0))),
		"round to second")

	core.AssertTrue(t, tm.Truncate(0).Equal(tm), "non-positive truncate")
	core.AssertTrue(t, tm.Round(-time.Second).Equal(tm), "non-positive round")
}

// runTestTAINTruncatePreEpoch covers truncation and rounding of timestamps
// before the Unix epoch. The internal (sec, nano) representation is always
// floor-oriented regardless of epoch, so pre-epoch instants floor away from
// zero (earlier), not toward zero as truncating through signed Unix
// nanoseconds would produce.
func runTestTAINTruncatePreEpoch(t *testing.T) {
	t.Helper()

	tm := TAINFromTime(time.Unix(-10, 300000000)) // -9.7s
	truncated := tm.Truncate(250 * time.Millisecond)
	core.AssertTrue(t, truncated.Equal(TAINFromTime(time.Unix(-10, 250000000))), // -9.75s
		"pre-epoch truncate floors away from zero")

	half := TAINFromTime(time.Unix(-10, 125000000)) // -9.875s, exact halfway point
	rounded := half.Round(250 * time.Millisecond)
	core.AssertTrue(t, rounded.Equal(TAINFromTime(time.Unix(-10, 250000000))), // -9.75s
		"pre-epoch round-half-up picks the later instant")
}

// runTestTAINTruncateRoundModern uses a post-2017 date with a non-zero
// leap second offset, guarding against the offset being applied twice
// when the timestamp is reconstructed after truncation or rounding.
func runTestTAINTruncateRoundModern(t *testing.T) {
	t.Helper()

	base := time.Date(2024, time.March, 15, 12, 30, 45, 567890123, time.UTC)
	tm := TAINFromTime(base)

	truncated := tm.Truncate(time.Second)
	core.AssertTrue(t, truncated.Equal(TAINFromTime(base.Truncate(time.Second))),
		"truncate to second")

	rounded := tm.Round(time.Second)
	core.AssertTrue(t, rounded.Equal(TAINFromTime(base.Round(time.Second))),
		"round to second")

	core.AssertEqual(t, tm.Unix(), truncated.Unix(), "truncate keeps the second")
}

func TestTAINMarshalJSON(t *testing.T) {
	tm := TAINFromTime(time.Unix(1000, 123456789))
	data, err := json.Marshal(tm)
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAIN
	core.AssertMustNoError(t, json.Unmarshal(data, &tm2), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "JSON round trip")

	core.AssertError(t, json.Unmarshal([]byte("123"), &tm2), "non-string JSON")
	core.AssertError(t, json.Unmarshal([]byte(`"bogus"`), &tm2), "invalid value")
}

func TestTAINMarshalBinary(t *testing.T) {
	tm := TAINFromTime(time.Unix(1000, 123456789))
	data, err := tm.MarshalBinary()
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAIN
	core.AssertMustNoError(t, tm2.UnmarshalBinary(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "binary round trip")

	core.AssertErrorIs(t, tm2.UnmarshalBinary([]byte{1, 2}),
		ErrInvalidTAINBinaryLength, "length error")
}

func TestUnixMethods(t *testing.T) {
	goTime := time.Unix(1234567890, 123456789)
	tm := TAINFromTime(goTime)

	// The Unix methods return the TAI time, not UTC time.
	// TAI is ahead of UTC by the leap second offset.
	expectedUnix := goTime.Unix() + int64(lsoffset(goTime))

	core.AssertEqual(t, expectedUnix, tm.Unix(), "unix seconds")
	core.AssertEqual(t, expectedUnix*1e9+int64(tm.Nanosecond()), tm.UnixNano(),
		"unix nanoseconds")
}

func TestPackageConstructors(t *testing.T) {
	t.Run("Unix", runTestUnixConstructor)
	t.Run("Date", runTestDateConstructor)
}

func runTestUnixConstructor(t *testing.T) {
	t.Helper()

	tain := UnixTAIN(1234567890, 123456789)
	expected := TAINFromTime(time.Unix(1234567890, 123456789))
	core.AssertTrue(t, tain.Equal(expected), "Unix constructor")
}

func runTestDateConstructor(t *testing.T) {
	t.Helper()

	tain := DateTAIN(DateConfig{
		Year:  2023,
		Month: time.May,
		Day:   15,
		Hour:  10,
		Min:   30,
		Sec:   45,
		Nsec:  123456789,
		Loc:   time.UTC,
	})
	expected := TAINFromTime(time.Date(2023, time.May, 15, 10, 30, 45, 123456789, time.UTC))
	core.AssertTrue(t, tain.Equal(expected), "Date constructor")
}

func TestTAINUnixConversions(t *testing.T) {
	tm := TAINFromTime(time.Unix(1000, 123456789))
	core.AssertEqual(t, int64(1000), tm.Unix(), "unix")
	core.AssertEqual(t, int64(1000123), tm.UnixMilli(), "milliseconds")
	core.AssertEqual(t, int64(1000123456), tm.UnixMicro(), "microseconds")
	core.AssertEqual(t, int64(1000123456789), tm.UnixNano(), "nano")
}

func TestTAINSinceUntil(t *testing.T) {
	t1 := TAINFromTime(time.Unix(1000, 0))
	t2 := TAINFromTime(time.Unix(1000, 500000000))
	core.AssertEqual(t, 500*time.Millisecond, t2.Since(t1), "since")
	core.AssertEqual(t, 500*time.Millisecond, t1.Until(t2), "until")
}

func TestTAINGoTimeRoundTrip(t *testing.T) {
	// A modern date exercises a non-zero leap second offset
	g := time.Date(2024, time.March, 15, 12, 30, 45, 123456789, time.UTC)
	core.AssertTrue(t, TAINFromTime(g).GoTime().Equal(g), "round trip")

	// Just before a leap second boundary the TAI label lands past the
	// boundary and needs the offset refinement in utcFromTAI
	b := time.Date(2016, time.December, 31, 23, 59, 59, 500000000, time.UTC)
	core.AssertTrue(t, TAINFromTime(b).GoTime().Equal(b), "before boundary")

	a := time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)
	core.AssertTrue(t, TAINFromTime(a).GoTime().Equal(a), "at boundary")
}

func TestTAINFormat(t *testing.T) {
	g := time.Date(2024, time.March, 15, 12, 30, 45, 0, time.UTC)
	core.AssertEqual(t, g.Format(time.RFC3339), TAINFromTime(g).Format(time.RFC3339),
		"format")
}

func TestTAINMarshalText(t *testing.T) {
	tm := TAINFromTime(time.Unix(1000, 123456789))
	data, err := tm.MarshalText()
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAIN
	core.AssertMustNoError(t, tm2.UnmarshalText(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "text round trip")

	core.AssertError(t, tm2.UnmarshalText([]byte("bogus")), "invalid text")
}

func TestTAINFromStringInvalidNanosecond(t *testing.T) {
	// nano field = 0x3B9ACA00 (1_000_000_000), one past the valid max
	_, err := ParseTAIN("@40000000000000003B9ACA00")
	core.AssertErrorIs(t, err, ErrInvalidNanosecondRange, "nanosecond range error")
}
