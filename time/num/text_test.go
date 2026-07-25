package num_test

import (
	"encoding"
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// textType is the value-side text surface shared by the numeric types:
// base-10 String and MarshalText, plus Equal for comparing a parsed value
// back against the original.
type textType[T any] interface {
	fmt.Stringer
	encoding.TextMarshaler
	Equal(T) bool
}

// textPtr is the pointer-side surface: UnmarshalText is defined on the
// pointer, so a round-trip needs *T to satisfy TextUnmarshaler.
type textPtr[T any] interface {
	*T
	encoding.TextUnmarshaler
}

var (
	_ core.TestCase = textCase[num.Uint128, *num.Uint128]{}
	_ core.TestCase = textCase[num.Int128, *num.Int128]{}
	_ core.TestCase = textCase[num.Int32, *num.Int32]{}
	_ core.TestCase = textCase[num.Int64, *num.Int64]{}
)

// textCase exercises the base-10 text surface of one numeric value. Every
// row parses input through UnmarshalText and checks the result: val for a
// success row, wantErr for a failure. A round-trip row additionally sets
// text, pinning the canonical rendering that String and MarshalText must
// produce; parse-only rows leave text empty to skip that direction.
type textCase[T textType[T], PT textPtr[T]] struct {
	val     T
	wantErr error
	input   string
	text    string
	name    string
}

func (tc textCase[T, PT]) Name() string { return tc.name }

func (tc textCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	if tc.wantErr != nil {
		tc.testParseError(t)
		return
	}
	if tc.text != "" {
		tc.testMarshal(t)
	}
	tc.testUnmarshal(t)
}

// testMarshal checks that String and MarshalText both render the canonical
// text.
func (tc textCase[T, PT]) testMarshal(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.text, tc.val.String(), "String")
	mt, err := tc.val.MarshalText()
	core.AssertNoError(t, err, "MarshalText")
	core.AssertEqual(t, tc.text, string(mt), "MarshalText")
}

// testUnmarshal checks that UnmarshalText parses input back to val.
func (tc textCase[T, PT]) testUnmarshal(t *testing.T) {
	t.Helper()
	var got T
	err := PT(&got).UnmarshalText([]byte(tc.input))
	core.AssertNoError(t, err, "UnmarshalText")
	core.AssertTrue(t, tc.val.Equal(got), "value %s == %s",
		tc.val.String(), got.String())
}

// testParseError checks that UnmarshalText rejects input with wantErr, over
// core.ErrInvalid.
func (tc textCase[T, PT]) testParseError(t *testing.T) {
	t.Helper()
	var got T
	err := PT(&got).UnmarshalText([]byte(tc.input))
	core.AssertErrorIs(t, err, tc.wantErr, "error")
	core.AssertErrorIs(t, err, core.ErrInvalid, "core.ErrInvalid")
}

// newUint128TextCase builds a round-trip row: text is both the canonical rendering of
// val and the input UnmarshalText must parse back to it.
func newUint128TextCase(name string, val num.Uint128,
	text string) textCase[num.Uint128, *num.Uint128] {
	return textCase[num.Uint128, *num.Uint128]{
		name: name, val: val, text: text, input: text}
}

func newInt128TextCase(name string, val num.Int128,
	text string) textCase[num.Int128, *num.Int128] {
	return textCase[num.Int128, *num.Int128]{
		name: name, val: val, text: text, input: text}
}

func newInt32TextCase(name string, val num.Int32,
	text string) textCase[num.Int32, *num.Int32] {
	return textCase[num.Int32, *num.Int32]{
		name: name, val: val, text: text, input: text}
}

func newInt64TextCase(name string, val num.Int64,
	text string) textCase[num.Int64, *num.Int64] {
	return textCase[num.Int64, *num.Int64]{
		name: name, val: val, text: text, input: text}
}

func TestUint128Text(t *testing.T) {
	core.RunTestCases(t, uint128TextCases())
	core.RunTestCases(t, uint128ParseCases())
}

func uint128TextCases() []textCase[num.Uint128, *num.Uint128] {
	ten19 := num.NewUint128(0, 1e19)
	e20 := ten19.Mul(num.NewUint128(0, 10))                             // 10^20
	padLow := ten19.Mul(num.NewUint128(0, 2)).Add(num.NewUint128(0, 5)) // 2e19+5
	return []textCase[num.Uint128, *num.Uint128]{
		newUint128TextCase("zero", num.NewUint128(0, 0), "0"),
		newUint128TextCase("small", num.NewUint128(0, 42), "42"),
		newUint128TextCase("max uint64", num.NewUint128(0, math.MaxUint64),
			"18446744073709551615"),
		newUint128TextCase("2^64", num.NewUint128(1, 0), "18446744073709551616"),
		newUint128TextCase("10^19", ten19, "10000000000000000000"),
		newUint128TextCase("10^20 zero-padded chunk", e20, "100000000000000000000"),
		newUint128TextCase("2e19+5 zero-padded low", padLow, "20000000000000000005"),
		newUint128TextCase("max", num.MaxUint128,
			"340282366920938463463374607431768211455"),
	}
}

func TestInt128Text(t *testing.T) {
	core.RunTestCases(t, int128TextCases())
	core.RunTestCases(t, int128ParseCases())
}

func int128TextCases() []textCase[num.Int128, *num.Int128] {
	e21 := num.NewInt128(1_000_000_000_000_000_000).Mul(num.NewInt128(1000))
	return []textCase[num.Int128, *num.Int128]{
		newInt128TextCase("zero", num.NewInt128(0), "0"),
		newInt128TextCase("positive", num.NewInt128(42), "42"),
		newInt128TextCase("negative", num.NewInt128(-42), "-42"),
		newInt128TextCase("min int64", num.NewInt128(math.MinInt64),
			"-9223372036854775808"),
		newInt128TextCase("10^21", e21, "1000000000000000000000"),
		newInt128TextCase("max", num.MaxInt128,
			"170141183460469231731687303715884105727"),
		newInt128TextCase("min", num.MinInt128,
			"-170141183460469231731687303715884105728"),
	}
}

func TestInt32Text(t *testing.T) {
	core.RunTestCases(t, int32TextCases())
	core.RunTestCases(t, int32ParseCases())
}

func int32TextCases() []textCase[num.Int32, *num.Int32] {
	return []textCase[num.Int32, *num.Int32]{
		newInt32TextCase("zero", num.Int32(0), "0"),
		newInt32TextCase("positive", num.Int32(12345), "12345"),
		newInt32TextCase("negative", num.Int32(-12345), "-12345"),
		newInt32TextCase("min", num.Int32(math.MinInt32), "-2147483648"),
		newInt32TextCase("max", num.Int32(math.MaxInt32), "2147483647"),
	}
}

func TestInt64Text(t *testing.T) {
	core.RunTestCases(t, int64TextCases())
	core.RunTestCases(t, int64ParseCases())
}

func int64TextCases() []textCase[num.Int64, *num.Int64] {
	return []textCase[num.Int64, *num.Int64]{
		newInt64TextCase("zero", num.Int64(0), "0"),
		newInt64TextCase("positive", num.Int64(987654321), "987654321"),
		newInt64TextCase("negative", num.Int64(-987654321), "-987654321"),
		newInt64TextCase("min", num.Int64(math.MinInt64), "-9223372036854775808"),
		newInt64TextCase("max", num.Int64(math.MaxInt64), "9223372036854775807"),
	}
}

// assertNilReceiver checks UnmarshalText over a nil *T. Valid input reports
// core.ErrNilReceiver rather than dereferencing the nil receiver; malformed
// input reports the data fault ([ErrSyntax]) first, so a broken receiver
// never masks a broken input.
func assertNilReceiver[T textType[T], PT textPtr[T]](t *testing.T,
	name string) {
	t.Helper()
	var p PT
	err := p.UnmarshalText([]byte("1"))
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "%s nil receiver", name)

	err = p.UnmarshalText([]byte("x"))
	core.AssertErrorIs(t, err, num.ErrSyntax, "%s data error precedence", name)
	core.AssertNotErrorIs(t, err, core.ErrNilReceiver, "%s receiver masked", name)
}

func TestTextNilReceiver(t *testing.T) {
	assertNilReceiver[num.Uint128, *num.Uint128](t, "Uint128")
	assertNilReceiver[num.Int128, *num.Int128](t, "Int128")
	assertNilReceiver[num.Int32, *num.Int32](t, "Int32")
	assertNilReceiver[num.Int64, *num.Int64](t, "Int64")
}
