package tai_test

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
	"darvaza.org/x/time/tai"
)

const (
	zeroTAIALabel   = "@00000000000000000000000000000000"
	labelTestPrefix = "prefix:"
)

var (
	_ core.TestCase = taiaLabelTestCase{}
	_ core.TestCase = taiaLabelErrorTestCase{}
	_ core.TestCase = taiaLabelWriteErrorTestCase{}
)

type taiaLabelTestCase struct {
	name, text string
	value      tai.TAIA
}

func newTAIALabelTestCase(name string, seconds uint64, nano, atto int64, text string) taiaLabelTestCase {
	value := tai.Combine(num.NewInt128(0, seconds), num.AsInt128(nano), num.AsInt128(atto))
	return taiaLabelTestCase{name: name, text: text, value: value}
}

func (tc taiaLabelTestCase) Name() string { return tc.name }

func (tc taiaLabelTestCase) Test(t *testing.T) {
	t.Run("ParseTAIA", tc.testParse)
	t.Run("MarshalText", tc.testMarshalText)
	t.Run("UnmarshalText", tc.testUnmarshalText)
	t.Run("AppendText", tc.testAppendText)
	t.Run("TAIA.String", tc.testString)
	t.Run("TAI.String", tc.testTAIString)
	t.Run("TAIN.String", tc.testTAINString)
}

func (tc taiaLabelTestCase) testString(t *testing.T) {
	want := strings.ToLower(tc.text)
	core.AssertEqual(t, want, tc.value.String(), "String")
	core.AssertEqual(t, want, fmt.Sprint(tc.value), "fmt.Sprint")
}

func (tc taiaLabelTestCase) testTAIString(t *testing.T) {
	want := strings.ToLower(tc.text[:tai.TAITextLength])
	value := tc.value.TAI()
	core.AssertEqual(t, want, value.String(), "String")
	core.AssertEqual(t, want, fmt.Sprint(value), "fmt.Sprint")
}

func (tc taiaLabelTestCase) testTAINString(t *testing.T) {
	want := strings.ToLower(tc.text[:tai.TAINTextLength])
	value := tc.value.TAIN()
	core.AssertEqual(t, want, value.String(), "String")
	core.AssertEqual(t, want, fmt.Sprint(value), "fmt.Sprint")
}

func (tc taiaLabelTestCase) testParse(t *testing.T) {
	got, err := tai.ParseTAIA(tc.text)
	core.AssertMustNoError(t, err, "parse")
	core.AssertEqual(t, tc.value, got, "value")
}

func (tc taiaLabelTestCase) testMarshalText(t *testing.T) {
	got, err := tc.value.MarshalText()
	core.AssertMustNoError(t, err, "marshal")
	core.AssertEqual(t, strings.ToLower(tc.text), string(got), "label")
	core.AssertEqual(t, tai.TAIATextLength, len(got), "length")
}

func (tc taiaLabelTestCase) testUnmarshalText(t *testing.T) {
	got := labelSeed()
	core.AssertMustNoError(t, got.UnmarshalText([]byte(tc.text)), "unmarshal")
	core.AssertEqual(t, tc.value, got, "value")
}

func (tc taiaLabelTestCase) testAppendText(t *testing.T) {
	b := make([]byte, len(labelTestPrefix), len(labelTestPrefix)+tai.TAIATextLength)
	copy(b, labelTestPrefix)
	got, err := tc.value.AppendText(b)
	core.AssertMustNoError(t, err, "append")
	core.AssertEqual(t, labelTestPrefix+strings.ToLower(tc.text), string(got), "label")
	core.AssertSame(t, &b[0], &got[0], "buffer")
}

// These labels were checked against libtai's taia_pack implementation.
func taiaLabelTestCases() []taiaLabelTestCase {
	return core.S(
		newTAIALabelTestCase("zero", 0, 0, 0, zeroTAIALabel),
		newTAIALabelTestCase("field order and padding", 0x0123456789abcdef, 123456789, 987654321,
			"@0123456789abcdef075bcd153ade68b1"),
		newTAIALabelTestCase("DJB seconds label", 0x400000002a2b2c2d, 0, 0,
			"@400000002a2b2c2d0000000000000000"),
		newTAIALabelTestCase("Unix epoch with fraction", tai.TAICONST, 500000000, 0,
			"@400000000000000a1dcd650000000000"),
		newTAIALabelTestCase("fraction bounds", 0x4000000000000001, 999999999, 999999999,
			"@40000000000000013b9ac9ff3b9ac9ff"),
		newTAIALabelTestCase("unsigned seconds", 1<<63, 0, 1,
			"@80000000000000000000000000000001"),
		maxTAIALabelTestCase(),
		newTAIALabelTestCase("uppercase input", 0x0123456789abcdef, 123456789, 987654321,
			"@0123456789ABCDEF075BCD153ADE68B1"),
	)
}

func maxTAIALabelTestCase() taiaLabelTestCase {
	return newTAIALabelTestCase("maximum fields", math.MaxUint64, 999999999, 999999999,
		"@ffffffffffffffff3b9ac9ff3b9ac9ff")
}

func TestTAIALabel(t *testing.T) {
	core.RunTestCases(t, taiaLabelTestCases())
}

type taiaLabelErrorTestCase struct {
	wantErr    error
	name, text string
}

func newTAIALabelErrorTestCase(name, text string, wantErr error) taiaLabelErrorTestCase {
	return taiaLabelErrorTestCase{name: name, text: text, wantErr: wantErr}
}

func (tc taiaLabelErrorTestCase) Name() string { return tc.name }

func (tc taiaLabelErrorTestCase) Test(t *testing.T) {
	t.Run("ParseTAIA", tc.testParse)
	t.Run("UnmarshalText", tc.testUnmarshalText)
}

func (tc taiaLabelErrorTestCase) testParse(t *testing.T) {
	got, err := tai.ParseTAIA(tc.text)
	assertLabelError(t, err, tc.wantErr)
	core.AssertContains(t, err.Error(), "ParseTAIA", "context")
	core.AssertEqual(t, tai.TAIA{}, got, "zero value")
}

func (tc taiaLabelErrorTestCase) testUnmarshalText(t *testing.T) {
	seed := labelSeed()
	got := seed
	err := got.UnmarshalText([]byte(tc.text))
	assertLabelError(t, err, tc.wantErr)
	core.AssertContains(t, err.Error(), "UnmarshalText", "context")
	core.AssertEqual(t, seed, got, "untouched receiver")
}

func taiaLabelErrorTestCases() []taiaLabelErrorTestCase {
	return core.S(
		newTAIALabelErrorTestCase("empty", "", tai.ErrInvalidTAIALength),
		newTAIALabelErrorTestCase("prefix only", "@", tai.ErrInvalidTAIALength),
		newTAIALabelErrorTestCase("missing prefix", zeroTAIALabel[1:], tai.ErrInvalidTAIALength),
		newTAIALabelErrorTestCase("short", zeroTAIALabel[:len(zeroTAIALabel)-1], tai.ErrInvalidTAIALength),
		newTAIALabelErrorTestCase("trailing newline", zeroTAIALabel+"\n", tai.ErrInvalidTAIALength),
		newTAIALabelErrorTestCase("wrong prefix", "#"+zeroTAIALabel[1:], tai.ErrInvalidTAIAFormat),
		newTAIALabelErrorTestCase("invalid seconds", "@g0000000000000000000000000000000", tai.ErrInvalidTAIAFormat),
		newTAIALabelErrorTestCase("invalid nanoseconds", "@0000000000000000g000000000000000", tai.ErrInvalidTAIAFormat),
		newTAIALabelErrorTestCase("invalid attoseconds", "@000000000000000000000000g0000000", tai.ErrInvalidTAIAFormat),
		newTAIALabelErrorTestCase("signed seconds", "@-0000000000000000000000000000000", tai.ErrInvalidTAIAFormat),
		newTAIALabelErrorTestCase("nanosecond overflow", "@00000000000000003b9aca0000000000", tai.ErrTAIARange),
		newTAIALabelErrorTestCase("attosecond overflow", "@0000000000000000000000003b9aca00", tai.ErrTAIARange),
	)
}

func TestTAIALabelErrors(t *testing.T) {
	core.RunTestCases(t, taiaLabelErrorTestCases())
}

type taiaLabelWriteErrorTestCase struct {
	name  string
	value tai.TAIA
}

func newTAIALabelWriteErrorTestCase(name string, value tai.TAIA) taiaLabelWriteErrorTestCase {
	return taiaLabelWriteErrorTestCase{name: name, value: value}
}

func (tc taiaLabelWriteErrorTestCase) Name() string { return tc.name }

func (tc taiaLabelWriteErrorTestCase) Test(t *testing.T) {
	t.Run("MarshalText", tc.testMarshalText)
	t.Run("AppendText", tc.testAppendText)
}

func (tc taiaLabelWriteErrorTestCase) testMarshalText(t *testing.T) {
	got, err := tc.value.MarshalText()
	assertLabelError(t, err, tai.ErrTAIARange)
	core.AssertNil(t, got, "label")
}

func (tc taiaLabelWriteErrorTestCase) testAppendText(t *testing.T) {
	buf := bytes.Repeat([]byte{'x'}, len(labelTestPrefix)+tai.TAIATextLength)
	copy(buf, labelTestPrefix)
	want := bytes.Clone(buf)
	got, err := tc.value.AppendText(buf[:len(labelTestPrefix)])
	assertLabelError(t, err, tai.ErrTAIARange)
	core.AssertEqual(t, labelTestPrefix, string(got), "prefix")
	core.AssertSliceEqual(t, want, buf, "untouched buffer")
}

func TestTAIALabelWriteErrors(t *testing.T) {
	zero := num.ZeroInt128
	core.RunTestCases(t, core.S(
		newTAIALabelWriteErrorTestCase("negative second", tai.Combine(num.AsInt128(-1), zero, zero)),
		newTAIALabelWriteErrorTestCase("negative attosecond", tai.Combine(zero, zero, num.AsInt128(-1))),
		newTAIALabelWriteErrorTestCase("seconds exceed uint64", tai.Combine(num.NewInt128(1, 0), zero, zero)),
	))
}

func labelSeed() tai.TAIA {
	return tai.Combine(num.AsInt128(42), num.AsInt128(1), num.AsInt128(2))
}

func assertLabelError(t *testing.T, err, want error) {
	t.Helper()
	core.AssertMustError(t, err, "label")
	core.AssertErrorIs(t, err, want, "cause")
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid input")
}

func TestTAIAUnmarshalTextNilReceiver(t *testing.T) {
	t.Run("valid label", runTestTAIANilReceiverValid)
	t.Run("invalid label", runTestTAIANilReceiverInvalid)
}

func runTestTAIANilReceiverValid(t *testing.T) {
	var receiver *tai.TAIA
	err := receiver.UnmarshalText([]byte(zeroTAIALabel))
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "nil receiver")
}

func runTestTAIANilReceiverInvalid(t *testing.T) {
	var receiver *tai.TAIA
	err := receiver.UnmarshalText(nil)
	assertLabelError(t, err, tai.ErrInvalidTAIALength)
	core.AssertNotErrorIs(t, err, core.ErrNilReceiver, "parse error first")
}

func TestTAIALabelAllocs(t *testing.T) {
	t.Run("ParseTAIA", runTestTAIAParseAllocs)
	t.Run("AppendText", runTestTAIAAppendAllocs)
	t.Run("UnmarshalText", runTestTAIAUnmarshalAllocs)
}

func runTestTAIAParseAllocs(t *testing.T) {
	tc := maxTAIALabelTestCase()
	var got tai.TAIA
	var err error
	allocs := testing.AllocsPerRun(100, func() { got, err = tai.ParseTAIA(tc.text) })
	core.AssertNoError(t, err, "parse")
	core.AssertEqual(t, tc.value, got, "value")
	core.AssertEqual(t, 0, allocs, "allocations")
}

func runTestTAIAAppendAllocs(t *testing.T) {
	tc := maxTAIALabelTestCase()
	buf := make([]byte, 0, tai.TAIATextLength)
	var got []byte
	var err error
	allocs := testing.AllocsPerRun(100, func() { got, err = tc.value.AppendText(buf[:0]) })
	core.AssertNoError(t, err, "append")
	core.AssertEqual(t, tc.text, string(got), "label")
	core.AssertEqual(t, 0, allocs, "allocations")
}

func runTestTAIAUnmarshalAllocs(t *testing.T) {
	tc := maxTAIALabelTestCase()
	text := []byte(tc.text)
	var got tai.TAIA
	var err error
	allocs := testing.AllocsPerRun(100, func() { err = got.UnmarshalText(text) })
	core.AssertNoError(t, err, "unmarshal")
	core.AssertEqual(t, tc.value, got, "value")
	core.AssertEqual(t, 0, allocs, "allocations")
}

// The reference seconds label is published at https://cr.yp.to/libtai/tai64.html.
func TestTAIAFromTimeLabel(t *testing.T) {
	instant := time.Date(1992, 6, 2, 8, 6, 43, 0, time.UTC)
	label, err := tai.TAIFromTime(instant).MarshalText()
	core.AssertMustNoError(t, err, "marshal")
	core.AssertEqual(t, "@400000002a2b2c2d0000000000000000", string(label), "DJB label")
}

// ExampleParseTAIA demonstrates parsing a label and writing its canonical text.
func ExampleParseTAIA() {
	value, err := tai.ParseTAIA("@400000002a2b2c2d0000000000000000")
	if err != nil {
		_, _ = fmt.Println(err)
		return
	}
	label, err := value.MarshalText()
	if err != nil {
		_, _ = fmt.Println(err)
		return
	}
	_, _ = fmt.Println(string(label))
	// Output: @400000002a2b2c2d0000000000000000
}
