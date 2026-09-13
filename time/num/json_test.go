package num_test

import (
	"encoding"
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var _ core.TestCase = jsonCase{}

// jsonValue is the JSON surface every type carries, beside the text
// it is built from.
type jsonValue interface {
	json.Marshaler
	encoding.TextMarshaler
}

// jsonCase pins the JSON form of a value: the MarshalText text, as a
// JSON number when the value is safe for a float64 consumer and as a
// JSON string otherwise. The text is declared; the number row and the
// string row differ only in the quoting the factory adds.
type jsonCase struct {
	in     jsonValue
	text   string
	name   string
	number bool
}

func newJSONNumberCase(name string, in jsonValue, text string) jsonCase {
	return jsonCase{name: name, in: in, text: text, number: true}
}

func newJSONStringCase(name string, in jsonValue, text string) jsonCase {
	return jsonCase{name: name, in: in, text: text}
}

func (tc jsonCase) Name() string { return tc.name }

func (tc jsonCase) Test(t *testing.T) {
	t.Helper()
	text, err := tc.in.MarshalText()
	core.AssertNoError(t, err, "MarshalText")
	core.AssertEqual(t, tc.text, string(text), "text")

	want := tc.text
	if !tc.number {
		want = strconv.Quote(tc.text)
	}
	got, err := tc.in.MarshalJSON()
	core.AssertNoError(t, err, "MarshalJSON")
	core.AssertEqual(t, want, string(got), "MarshalJSON")

	got, err = json.Marshal(tc.in)
	core.AssertNoError(t, err, "json.Marshal")
	core.AssertEqual(t, want, string(got), "json.Marshal")

	// the token kind, read back by the standard decoder rather than
	// by our own quoting.
	var v any
	core.AssertNoError(t, json.Unmarshal(got, &v), "json.Unmarshal")
	_, isNumber := v.(float64)
	core.AssertEqual(t, tc.number, isNumber, "number token")
}

func jsonInt32Cases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("zero", num.AsInt32(0), "0"),
		newJSONNumberCase("negative", num.AsInt32(-42), "-42"),
		newJSONNumberCase("min", num.AsInt32(math.MinInt32), "-2147483648"),
		newJSONNumberCase("max", num.AsInt32(math.MaxInt32), "2147483647"),
	}
}

func jsonInt64Cases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("zero", num.AsInt64(0), "0"),
		newJSONNumberCase("two to the 53", num.AsInt64(1<<53),
			"9007199254740992"),
		newJSONStringCase("past two to the 53", num.AsInt64(1<<53+1),
			"9007199254740993"),
		newJSONNumberCase("minus two to the 53", num.AsInt64(-1<<53),
			"-9007199254740992"),
		newJSONStringCase("below minus two to the 53", num.AsInt64(-1<<53-1),
			"-9007199254740993"),
		newJSONStringCase("max", num.AsInt64(math.MaxInt64),
			"9223372036854775807"),
		newJSONStringCase("min", num.AsInt64(math.MinInt64),
			"-9223372036854775808"),
	}
}

func jsonInt128Cases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("negative", num.AsInt128(-42), "-42"),
		newJSONNumberCase("two to the 53", num.AsInt128(1<<53),
			"9007199254740992"),
		newJSONStringCase("past two to the 53", num.AsInt128(1<<53+1),
			"9007199254740993"),
		newJSONStringCase("two to the 64", num.NewInt128(1, 0),
			"18446744073709551616"),
		newJSONStringCase("max", num.MaxInt128,
			"170141183460469231731687303715884105727"),
		newJSONStringCase("min", num.MinInt128,
			"-170141183460469231731687303715884105728"),
	}
}

func jsonUint128Cases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("zero", num.ZeroUint128, "0"),
		newJSONNumberCase("low word", u(42), "42"),
		newJSONNumberCase("two to the 53", u(1<<53), "9007199254740992"),
		newJSONStringCase("past two to the 53", u(1<<53+1),
			"9007199254740993"),
		newJSONStringCase("two to the 64", num.NewUint128(1, 0),
			"18446744073709551616"),
		newJSONStringCase("max", num.MaxUint128,
			"340282366920938463463374607431768211455"),
	}
}

func jsonMilliCases() []jsonCase {
	return []jsonCase{
		newJSONNumberCase("milli32 zero", num.NewMilli32(0, 0), "0.000"),
		newJSONNumberCase("milli32", num.NewMilli32(1, 500), "1.500"),
		newJSONNumberCase("milli32 min", num.AsMilli32(math.MinInt32),
			"-2147483.648"),
		newJSONNumberCase("milli32 max", num.AsMilli32(math.MaxInt32),
			"2147483.647"),
		newJSONNumberCase("milli64", num.NewMilli64(86400, 5), "86400.005"),
		newJSONNumberCase("milli64 count below ten to the 15",
			num.AsMilli64(999999999999999), "999999999999.999"),
		newJSONStringCase("milli64 count at ten to the 15",
			num.AsMilli64(1e15), "1000000000000.000"),
		newJSONNumberCase("milli64 count above minus ten to the 15",
			num.AsMilli64(-999999999999999), "-999999999999.999"),
		newJSONStringCase("milli64 count at minus ten to the 15",
			num.AsMilli64(-1e15), "-1000000000000.000"),
		newJSONStringCase("milli64 max", num.AsMilli64(math.MaxInt64),
			"9223372036854775.807"),
		newJSONStringCase("milli64 min", num.AsMilli64(math.MinInt64),
			"-9223372036854775.808"),
	}
}

func jsonAtto128Cases() []jsonCase {
	return []jsonCase{
		newJSONStringCase("zero", num.NewAtto128(0, 0), "0.000000000000000000"),
		newJSONStringCase("one and a half", num.NewAtto128(1, 500e15),
			"1.500000000000000000"),
		newJSONStringCase("one atto", num.NewAtto128(0, 1),
			"0.000000000000000001"),
		newJSONStringCase("min", num.AsAtto128(num.MinInt128),
			"-170141183460469231731.687303715884105728"),
	}
}

func TestJSON(t *testing.T) {
	t.Run("int32", runTestJSONInt32)
	t.Run("int64", runTestJSONInt64)
	t.Run("int128", runTestJSONInt128)
	t.Run("uint128", runTestJSONUint128)
	t.Run("milli", runTestJSONMilli)
	t.Run("atto128", runTestJSONAtto128)
}

func runTestJSONInt32(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonInt32Cases())
}

func runTestJSONInt64(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonInt64Cases())
}

func runTestJSONInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonInt128Cases())
}

func runTestJSONUint128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonUint128Cases())
}

func runTestJSONMilli(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonMilliCases())
}

func runTestJSONAtto128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, jsonAtto128Cases())
}
