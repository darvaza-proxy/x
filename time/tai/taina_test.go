package tai

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement core.TestCase
var _ core.TestCase = tainaFromStringTestCase{}

// tainaFromStringTestCase tests ParseTAINA parsing.
type tainaFromStringTestCase struct {
	input   string
	name    string
	wantErr bool
}

func newTAINAFromStringTestCase(name, input string, wantErr bool) tainaFromStringTestCase {
	return tainaFromStringTestCase{
		name:    name,
		input:   input,
		wantErr: wantErr,
	}
}

// Name returns the test case name for identification.
func (tc tainaFromStringTestCase) Name() string {
	return tc.name
}

// Test validates ParseTAINA parsing behaviour.
func (tc tainaFromStringTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := ParseTAINA(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "parse %q", tc.input)
		return
	}
	core.AssertNoError(t, err, "parse %q", tc.input)
}

func TestTAINAFromString(t *testing.T) {
	testCases := []tainaFromStringTestCase{
		newTAINAFromStringTestCase("valid zero timestamp",
			"@40000000000000000000000000000000", false),
		newTAINAFromStringTestCase("valid timestamp",
			"@400000004B05F9FE0000000000000000", false),
		newTAINAFromStringTestCase("valid with attoseconds",
			"@400000004B05F9FE12345678000F4240", false),
		newTAINAFromStringTestCase("invalid string", "invalid", true),
		newTAINAFromStringTestCase("missing @ prefix",
			"40000000000000000000000000000000", true),
		newTAINAFromStringTestCase("invalid hex payload", "@invalid", true),
		newTAINAFromStringTestCase("invalid hex, correct length",
			"@GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", true),
		newTAINAFromStringTestCase("too short (TAI64N format)",
			"@400000000000000000000000", true),
		newTAINAFromStringTestCase("too long",
			"@40000000000000000000000000000000X", true),
		newTAINAFromStringTestCase("attoseconds out of range",
			"@400000004B05F9FE12345678FFFFFFFF", true),
	}

	core.RunTestCases(t, testCases)
}

func TestTAINAFromStringSentinels(t *testing.T) {
	_, err := ParseTAINA("@40")
	core.AssertErrorIs(t, err, ErrInvalidTAINALength, "length error")

	_, err = ParseTAINA("X40000000000000000000000000000000")
	core.AssertErrorIs(t, err, ErrInvalidTAINAFormat, "format error")
	core.AssertErrorIs(t, err, core.ErrInvalid, "core.ErrInvalid chain")
}

func TestNowTAINA(t *testing.T) {
	core.AssertFalse(t, NowTAINA().IsZero(), "zero time")
}

func TestTAINAComparison(t *testing.T) {
	t.Run("nanosecond precision", runTestTAINANanosecondComparison)
	t.Run("attosecond precision", runTestTAINAAttosecondComparison)
	t.Run("compare method", runTestTAINACompareMethod)
}

func runTestTAINANanosecondComparison(t *testing.T) {
	t.Helper()

	t1 := TAINAFromTime(time.Unix(1000, 500000000))
	t2 := TAINAFromTime(time.Unix(1000, 600000000))
	t3 := TAINAFromTime(time.Unix(1000, 500000000))

	core.AssertTrue(t, t1.Before(t2), "t1 before t2")
	core.AssertTrue(t, t2.After(t1), "t2 after t1")
	core.AssertFalse(t, t1.Equal(t2), "t1 equal t2")
	core.AssertTrue(t, t1.Equal(t3), "t1 equal t3")
}

func runTestTAINAAttosecondComparison(t *testing.T) {
	t.Helper()

	t1 := UnixTAINA(1000, 500000000, 123456789)
	t2 := UnixTAINA(1000, 500000000, 123456790)

	core.AssertTrue(t, t1.Before(t2), "attosecond ordering")
}

func runTestTAINACompareMethod(t *testing.T) {
	t.Helper()

	t1 := TAINAFromTime(time.Unix(1000, 500000000))
	t2 := TAINAFromTime(time.Unix(1000, 600000000))
	t3 := TAINAFromTime(time.Unix(1000, 500000000))
	t4 := UnixTAINA(1000, 500000000, 123456789)
	t5 := UnixTAINA(1000, 500000000, 123456790)

	core.AssertEqual(t, -1, t1.Compare(t2), "t1 compare t2")
	core.AssertEqual(t, 1, t2.Compare(t1), "t2 compare t1")
	core.AssertEqual(t, 0, t1.Compare(t3), "t1 compare t3")
	core.AssertEqual(t, -1, t4.Compare(t5), "attosecond compare")
	core.AssertEqual(t, 1, t5.Compare(t4), "attosecond compare reverse")

	t6 := TAINAFromTime(time.Unix(2000, 0))
	core.AssertEqual(t, -1, t1.Compare(t6), "cross-second compare")
	core.AssertEqual(t, 1, t6.Compare(t1), "cross-second compare reverse")
	core.AssertTrue(t, t1.Before(t6), "cross-second before")
	core.AssertTrue(t, t6.After(t1), "cross-second after")
	core.AssertTrue(t, t5.After(t4), "attosecond after")
}

func TestTAINAArithmetic(t *testing.T) {
	base := TAINAFromTime(time.Unix(1000, 500000000))
	dur := 60*time.Second + 250*time.Millisecond

	later, err := base.Add(dur)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, dur, later.Sub(base), "add/sub round trip")

	// Sub borrows a second when the nanosecond difference is negative
	x := UnixTAINA(1001, 100, 0)
	y := UnixTAINA(1000, 900, 0)
	core.AssertEqual(t, time.Second-800*time.Nanosecond, x.Sub(y), "nanosecond borrow")
}
func TestTAINASubOverflow(t *testing.T) {
	big := TAINA{sec: math.MaxUint64, nano: 0}
	small := TAINA{sec: 0, nano: 0}

	core.AssertEqual(t, maxDuration, big.Sub(small), "clamp to max duration")
	core.AssertEqual(t, minDuration, small.Sub(big), "clamp to min duration")
}

func TestTAINAAttosecondArithmetic(t *testing.T) {
	t.Run("add attoseconds", runTestAddAttoseconds)
	t.Run("carry to nanoseconds", runTestAddAttosecondsCarry)
	t.Run("negative attoseconds", runTestAddAttosecondsNegative)
	t.Run("intake wrap", runTestAddAttosecondsWrap)
	t.Run("min int64 underflow", runTestAddAttosecondsMinInt64Underflow)
	t.Run("min int64 precise", runTestAddAttosecondsMinInt64Precise)
	t.Run("subtract attoseconds", runTestSubAttoseconds)
}

func runTestAddAttosecondsWrap(t *testing.T) {
	t.Helper()

	base := UnixTAINA(1000, 0, 1)
	_, err := base.AddAttoseconds(math.MaxInt64)
	core.AssertErrorIs(t, err, ErrOverflow, "intake wrap")
}

// runTestAddAttosecondsMinInt64Underflow exercises the borrow path at
// math.MinInt64, the one value whose negation overflows int64. Using a
// small raw sec (bypassing TAICONST) forces the multi-second borrow to
// underflow, proving the computation reaches safeSubUint64 with a
// correct (very negative) magnitude rather than a wrapped/bogus one.
func runTestAddAttosecondsMinInt64Underflow(t *testing.T) {
	t.Helper()

	base := TAINA{sec: 5, nano: 0, atto: 0}
	_, err := base.AddAttoseconds(math.MinInt64)
	core.AssertErrorIs(t, err, ErrUnderflow, "min int64 borrow underflow")
}

// runTestAddAttosecondsMinInt64Precise checks the exact result of
// borrowing math.MinInt64 attoseconds against a timestamp with enough
// headroom to avoid underflow, confirming floorDivMod's quotient and
// remainder are both correct at the extreme.
func runTestAddAttosecondsMinInt64Precise(t *testing.T) {
	t.Helper()

	base := UnixTAINA(1000, 500000000, 0)
	result, err := base.AddAttoseconds(math.MinInt64)
	core.AssertMustNoError(t, err, "min int64 add")
	core.AssertTrue(t, result.Equal(UnixTAINA(991, 276627963, 145224192)),
		"min int64 precise result")
}

func runTestAddAttoseconds(t *testing.T) {
	t.Helper()

	base := UnixTAINA(1000, 500000000, 123456789)
	result, err := base.AddAttoseconds(876543210)
	core.AssertMustNoError(t, err, "add")
	core.AssertTrue(t, result.Equal(UnixTAINA(1000, 500000000, 999999999)), "result")
}

func runTestAddAttosecondsCarry(t *testing.T) {
	t.Helper()

	base := UnixTAINA(1000, 500000000, 123456789)
	// 1 billion attoseconds carry over into 1 nanosecond
	result, err := base.AddAttoseconds(1000000000)
	core.AssertMustNoError(t, err, "add")
	core.AssertTrue(t, result.Equal(UnixTAINA(1000, 500000001, 123456789)), "carry")
}

func runTestAddAttosecondsNegative(t *testing.T) {
	t.Helper()

	base := UnixTAINA(1000, 500000000, 123456789)
	result, err := base.AddAttoseconds(-123456789)
	core.AssertMustNoError(t, err, "add")
	core.AssertTrue(t, result.Equal(UnixTAINA(1000, 500000000, 0)), "negative add")
}

func runTestSubAttoseconds(t *testing.T) {
	t.Helper()

	t1 := UnixTAINA(1000, 500000000, 999999999)
	t2 := UnixTAINA(1000, 500000000, 123456789)
	diff := t1.SubAttoseconds(t2)
	core.AssertEqual(t, int64(0), diff.Nanoseconds, "nanoseconds")
	core.AssertEqual(t, uint32(876543210), diff.Attoseconds, "attoseconds")

	// Negative difference: sign carried by Nanoseconds, attoseconds
	// normalized to [0, 1e9)
	neg := t2.SubAttoseconds(t1)
	core.AssertEqual(t, int64(-1), neg.Nanoseconds, "negative nanoseconds")
	core.AssertEqual(t, uint32(123456790), neg.Attoseconds, "negative attoseconds")

	// Differences beyond ±9.2s of attoseconds no longer overflow
	// thanks to the split representation
	far := UnixTAINA(1010, 500000000, 999999999).SubAttoseconds(t2)
	core.AssertEqual(t, int64(10000000000), far.Nanoseconds, "large nanoseconds")
	core.AssertEqual(t, uint32(876543210), far.Attoseconds, "large attoseconds")
}

func TestTAINAAttosecondAccessors(t *testing.T) {
	tm := UnixTAINA(1000, 123456789, 987654321)
	core.AssertEqual(t, 123456789, tm.Nanosecond(), "nanoseconds")
	core.AssertEqual(t, 987654321, tm.Attosecond(), "attoseconds")

	tm2 := UnixTAINA(1, 123456789, 987654321)
	split := tm2.UnixAttosecondSplit()
	core.AssertEqual(t, tm2.UnixNano(), split.UnixNanoseconds, "unix nanoseconds")
	core.AssertEqual(t, uint32(987654321), split.Attoseconds, "split attoseconds")
}

func TestTAINATruncateRound(t *testing.T) {
	t.Run("epoch era", runTestTAINATruncateRoundEpoch)
	t.Run("modern date", runTestTAINATruncateRoundModern)
	t.Run("pre-epoch", runTestTAINATruncatePreEpoch)
}

func runTestTAINATruncateRoundEpoch(t *testing.T) {
	t.Helper()

	tm := UnixTAINA(1000, 567890123, 456789012)

	truncated := tm.Truncate(time.Second)
	core.AssertTrue(t, truncated.Equal(UnixTAINA(1000, 0, 0)), "truncate to second")

	rounded := tm.Round(time.Second)
	core.AssertTrue(t, rounded.Equal(UnixTAINA(1001, 0, 0)), "round to second")

	truncatedNano := tm.Truncate(time.Nanosecond)
	core.AssertTrue(t, truncatedNano.Equal(UnixTAINA(1000, 567890123, 0)),
		"truncate to nanosecond clears attoseconds")

	core.AssertTrue(t, tm.Truncate(0).Equal(tm), "non-positive truncate")
	core.AssertTrue(t, tm.Round(-time.Second).Equal(tm), "non-positive round")
}

// runTestTAINATruncateRoundModern uses a post-2017 date with a non-zero
// leap second offset, guarding against the offset being applied twice
// when the timestamp is reconstructed after truncation or rounding.
func runTestTAINATruncateRoundModern(t *testing.T) {
	t.Helper()

	base := time.Date(2024, time.March, 15, 12, 30, 45, 567890123, time.UTC)
	tm, err := TAINAFromTime(base).AddAttoseconds(456789012)
	core.AssertMustNoError(t, err, "add attoseconds")

	truncated := tm.Truncate(time.Second)
	core.AssertTrue(t, truncated.Equal(TAINAFromTime(base.Truncate(time.Second))),
		"truncate to second")

	rounded := tm.Round(time.Second)
	core.AssertTrue(t, rounded.Equal(TAINAFromTime(base.Round(time.Second))),
		"round to second")

	truncatedNano := tm.Truncate(time.Nanosecond)
	core.AssertTrue(t, truncatedNano.Equal(TAINAFromTime(base)),
		"truncate to nanosecond clears attoseconds")
}

// runTestTAINATruncatePreEpoch covers truncation and rounding of timestamps
// before the Unix epoch. The internal (sec, nano) representation is always
// floor-oriented regardless of epoch, so pre-epoch instants floor away from
// zero (earlier), not toward zero as truncating through signed Unix
// nanoseconds would produce. Attoseconds are always dropped.
func runTestTAINATruncatePreEpoch(t *testing.T) {
	t.Helper()

	tm := TAINAFromTime(time.Unix(-10, 300000000)) // -9.7s
	truncated := tm.Truncate(250 * time.Millisecond)
	core.AssertTrue(t, truncated.Equal(UnixTAINA(-10, 250000000, 0)), // -9.75s
		"pre-epoch truncate floors away from zero")

	half := TAINAFromTime(time.Unix(-10, 125000000)) // -9.875s, exact halfway point
	rounded := half.Round(250 * time.Millisecond)
	core.AssertTrue(t, rounded.Equal(UnixTAINA(-10, 250000000, 0)), // -9.75s
		"pre-epoch round-half-up picks the later instant")
}

func TestTAINAMarshalJSON(t *testing.T) {
	tm := UnixTAINA(1000, 123456789, 987654321)
	data, err := json.Marshal(tm)
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAINA
	core.AssertMustNoError(t, json.Unmarshal(data, &tm2), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "JSON round trip")

	core.AssertError(t, json.Unmarshal([]byte("123"), &tm2), "non-string JSON")
	core.AssertError(t, json.Unmarshal([]byte(`"bogus"`), &tm2), "invalid value")
}

func TestTAINAMarshalBinary(t *testing.T) {
	tm := UnixTAINA(1000, 123456789, 987654321)
	data, err := tm.MarshalBinary()
	core.AssertMustNoError(t, err, "marshal")
	core.AssertEqual(t, TAINALength, len(data), "binary length")

	var tm2 TAINA
	core.AssertMustNoError(t, tm2.UnmarshalBinary(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "binary round trip")
}

func TestTAINAMarshalText(t *testing.T) {
	tm := UnixTAINA(1000, 123456789, 987654321)
	data, err := tm.MarshalText()
	core.AssertMustNoError(t, err, "marshal")

	var tm2 TAINA
	core.AssertMustNoError(t, tm2.UnmarshalText(data), "unmarshal")
	core.AssertTrue(t, tm.Equal(tm2), "text round trip")

	core.AssertError(t, tm2.UnmarshalText([]byte("bogus")), "invalid text")
}

func TestTAINAStringFormat(t *testing.T) {
	tm := UnixTAINA(0, 0, 123456789)
	str := tm.String()

	// @ followed by 32 hex digits
	core.AssertEqual(t, 33, len(str), "string length")
	core.AssertEqual(t, byte('@'), str[0], "prefix")

	parsed, err := ParseTAINA(str)
	core.AssertMustNoError(t, err, "parse")
	core.AssertTrue(t, tm.Equal(parsed), "string round trip")
}

func TestTAINAUnixMethods(t *testing.T) {
	goTime := time.Unix(1234567890, 123456789)
	tm, err := TAINAFromTime(goTime).AddAttoseconds(987654321)
	core.AssertMustNoError(t, err, "add attoseconds")

	// The Unix methods return the TAI time, not UTC time.
	// TAI is ahead of UTC by the leap second offset.
	expectedUnix := goTime.Unix() + int64(lsoffset(goTime))
	expectedUnixNano := expectedUnix*1e9 + int64(tm.Nanosecond())

	core.AssertEqual(t, expectedUnix, tm.Unix(), "unix seconds")
	core.AssertEqual(t, expectedUnixNano, tm.UnixNano(), "unix nanoseconds")

	split := tm.UnixAttosecondSplit()
	core.AssertEqual(t, expectedUnixNano, split.UnixNanoseconds, "split nanoseconds")
	core.AssertEqual(t, tm.atto, split.Attoseconds, "split attoseconds")
}

func TestTAINAConstructors(t *testing.T) {
	t.Run("UnixTAINA", runTestUnixTAINAConstructor)
	t.Run("DateTAINA", runTestDateTAINAConstructor)
}

func runTestUnixTAINAConstructor(t *testing.T) {
	t.Helper()

	taina := UnixTAINA(1234567890, 123456789, 987654321)
	expected, err := TAINAFromTime(time.Unix(1234567890, 123456789)).
		AddAttoseconds(987654321)
	core.AssertMustNoError(t, err, "add attoseconds")
	core.AssertTrue(t, taina.Equal(expected), "UnixTAINA constructor")
}

func runTestDateTAINAConstructor(t *testing.T) {
	t.Helper()

	taina := DateTAINA(DateConfig{
		Year:  2023,
		Month: time.May,
		Day:   15,
		Hour:  10,
		Min:   30,
		Sec:   45,
		Nsec:  123456789,
		Loc:   time.UTC,
	}, 987654321)
	goTime := time.Date(2023, time.May, 15, 10, 30, 45, 123456789, time.UTC)
	expected, err := TAINAFromTime(goTime).AddAttoseconds(987654321)
	core.AssertMustNoError(t, err, "add attoseconds")
	core.AssertTrue(t, taina.Equal(expected), "DateTAINA constructor")
}

func TestTAINAConversions(t *testing.T) {
	taina := UnixTAINA(1000, 123456789, 987654321)

	tain := taina.TAIN()
	core.AssertTrue(t, tain.Equal(TAINFromTime(time.Unix(1000, 123456789))),
		"TAINA to TAIN")

	tai := taina.TAI()
	core.AssertTrue(t, tai.Equal(TAIFromTime(time.Unix(1000, 123456789))),
		"TAINA to TAI")

	tain2 := TAINFromTime(time.Unix(1000, 123456789))
	taina2 := tain2.TAINA()
	core.AssertTrue(t, taina2.Equal(UnixTAINA(1000, 123456789, 0)),
		"TAIN to TAINA with zero attoseconds")
}

func TestTAINAErrorConditions(t *testing.T) {
	t.Run("binary length", runTestTAINABinaryLengthError)
	t.Run("attosecond range", runTestTAINAAttosecondRangeError)
	t.Run("constructor panic", runTestTAINAConstructorPanic)
	t.Run("date constructor panic", runTestDateTAINAConstructorPanic)
}

func runTestDateTAINAConstructorPanic(t *testing.T) {
	t.Helper()

	core.AssertPanic(t, func() {
		DateTAINA(DateConfig{
			Year: 2023, Month: time.May, Day: 15, Loc: time.UTC,
		}, 1000000000) // > 999999999
	}, ErrInvalidAttosecondRange, "DateTAINA panic")
}

func runTestTAINABinaryLengthError(t *testing.T) {
	t.Helper()

	var tm TAINA
	err := tm.UnmarshalBinary([]byte{1, 2, 3})
	core.AssertErrorIs(t, err, ErrInvalidTAINABinaryLength, "length error")
	core.AssertErrorIs(t, err, core.ErrInvalid, "core.ErrInvalid chain")
}

func runTestTAINAAttosecondRangeError(t *testing.T) {
	t.Helper()

	// Set the attosecond field to an invalid value (> 999999999)
	data := make([]byte, TAINALength)
	data[12] = 0xFF
	data[13] = 0xFF
	data[14] = 0xFF
	data[15] = 0xFF

	var tm TAINA
	err := tm.UnmarshalBinary(data)
	core.AssertErrorIs(t, err, ErrInvalidAttosecondRange, "range error")
	core.AssertTrue(t, tm.IsZero(), "receiver unchanged on error")
}

func runTestTAINAConstructorPanic(t *testing.T) {
	t.Helper()

	core.AssertPanic(t, func() {
		UnixTAINA(0, 0, 1000000000) // > 999999999
	}, ErrInvalidAttosecondRange, "UnixTAINA panic")
}

func TestTAINAEdgeCases(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		core.AssertTrue(t, TAINA{}.IsZero(), "zero value")
	})
	t.Run("maximum attoseconds", func(t *testing.T) {
		core.AssertEqual(t, 999999999, UnixTAINA(0, 0, 999999999).Attosecond(), "max")
	})
	t.Run("attosecond borrow", runTestAttosecondBorrow)
}

func runTestAttosecondBorrow(t *testing.T) {
	t.Helper()

	tm := UnixTAINA(1000, 500000000, 100)
	result, err := tm.AddAttoseconds(-200)
	core.AssertMustNoError(t, err, "add")

	// Borrows from nanoseconds: 999999900 attoseconds, 499999999 nanoseconds
	core.AssertEqual(t, 999999900, result.Attosecond(), "attoseconds")
	core.AssertEqual(t, 499999999, result.Nanosecond(), "nanoseconds")
}

func TestTAINAUnixMilliMicro(t *testing.T) {
	tm := UnixTAINA(1000, 123456789, 0)
	core.AssertEqual(t, int64(1000123), tm.UnixMilli(), "milliseconds")
	core.AssertEqual(t, int64(1000123456), tm.UnixMicro(), "microseconds")
}

func TestTAINASinceUntil(t *testing.T) {
	t1 := UnixTAINA(1000, 0, 0)
	t2 := UnixTAINA(1000, 500000000, 0)
	core.AssertEqual(t, 500*time.Millisecond, t2.Since(t1), "since")
	core.AssertEqual(t, 500*time.Millisecond, t1.Until(t2), "until")
}

func TestTAINAGoTimeRoundTrip(t *testing.T) {
	// A modern date exercises a non-zero leap second offset
	g := time.Date(2024, time.March, 15, 12, 30, 45, 123456789, time.UTC)
	core.AssertTrue(t, TAINAFromTime(g).GoTime().Equal(g), "round trip")

	// Just before a leap second boundary the TAI label lands past the
	// boundary and needs the offset refinement in utcFromTAI
	b := time.Date(2016, time.December, 31, 23, 59, 59, 500000000, time.UTC)
	core.AssertTrue(t, TAINAFromTime(b).GoTime().Equal(b), "before boundary")

	a := time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)
	core.AssertTrue(t, TAINAFromTime(a).GoTime().Equal(a), "at boundary")
}

func TestTAINAFormat(t *testing.T) {
	g := time.Date(2024, time.March, 15, 12, 30, 45, 0, time.UTC)
	core.AssertEqual(t, g.Format(time.RFC3339), TAINAFromTime(g).Format(time.RFC3339),
		"format")
}

func TestTAINASubAttosecondsSaturationBorrow(t *testing.T) {
	big := TAINA{sec: math.MaxUint64, nano: 0, atto: 1}
	small := TAINA{sec: 0, nano: 0, atto: 0}

	diff := small.SubAttoseconds(big)
	core.AssertEqual(t, int64(math.MinInt64), diff.Nanoseconds,
		"saturated negative nanoseconds must not wrap")
	core.AssertEqual(t, uint32(0), diff.Attoseconds,
		"borrow skipped at saturation boundary")
}
