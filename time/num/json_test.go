package num_test

import (
	"encoding"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = jsonCase{}
	_ core.TestCase = jsonDecodeCase[num.Int32, *num.Int32]{}
	_ core.TestCase = jsonDirectCase[num.Int32, *num.Int32]{}
	_ core.TestCase = jsonAllocsCase[num.Int32, *num.Int32]{}
	_ core.TestCase = jsonNilCase[num.Int32, *num.Int32]{}
)

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

	want := tc.json()
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

// json returns the JSON the row declares: the text, bare as a number
// or quoted as a string.
func (tc jsonCase) json() string {
	if tc.number {
		return tc.text
	}
	return strconv.Quote(tc.text)
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

// jsonUnmarshaler is the pointer side of a type that reads its JSON
// back: the pointer to T that reads what T's MarshalJSON writes.
type jsonUnmarshaler[T any] interface {
	*T
	json.Unmarshaler
}

// jsonDecodeCase pins one JSON value read into T, by UnmarshalJSON on
// *T and by json.Unmarshal alike: the value stored, and on failure the
// sentinel, the ParseError naming the parser and quoting the text it
// read, and the receiver left as it was. Every row declares the value
// the receiver holds before the call, chosen apart from the value a
// success row expects, so a store, or a missing one, shows; a failing
// row expects no value, since the receiver keeps its seed.
type jsonDecodeCase[T num.Number[T], PT jsonUnmarshaler[T]] struct {
	wantErr error
	seed    T
	want    T
	in      string
	text    string
	name    string
}

func newJSONDecodeCase[T num.Number[T], PT jsonUnmarshaler[T]](name,
	in string, seed, want T) jsonDecodeCase[T, PT] {
	return jsonDecodeCase[T, PT]{name: name, in: in, seed: seed, want: want}
}

// newJSONDecodeCaseErr returns a row failing on in with wantErr, its
// report quoting text, the text the parser was given.
func newJSONDecodeCaseErr[T num.Number[T], PT jsonUnmarshaler[T]](name,
	in string, seed T, text string, wantErr error) jsonDecodeCase[T, PT] {
	return jsonDecodeCase[T, PT]{name: name, in: in, seed: seed, text: text,
		wantErr: wantErr}
}

// newJSONTokenCaseErr returns a row failing with wantErr on a token
// other than a string, which the parser is given as it is.
func newJSONTokenCaseErr[T num.Number[T], PT jsonUnmarshaler[T]](name,
	token string, seed T, wantErr error) jsonDecodeCase[T, PT] {
	return newJSONDecodeCaseErr[T, PT](name, token, seed, token, wantErr)
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonDecodeCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonDecodeCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	t.Run("UnmarshalJSON", func(t *testing.T) { tc.check(t, tc.decodeDirect) })
	t.Run("json.Unmarshal", func(t *testing.T) { tc.check(t, tc.decodeJSON) })
}

// decodeDirect reads the row into p by calling UnmarshalJSON directly.
func (tc jsonDecodeCase[T, PT]) decodeDirect(p PT) error {
	return p.UnmarshalJSON([]byte(tc.in))
}

// decodeJSON reads the row into p through json.Unmarshal, which checks
// the value is valid JSON before it calls UnmarshalJSON.
func (tc jsonDecodeCase[T, PT]) decodeJSON(p PT) error {
	return json.Unmarshal([]byte(tc.in), p)
}

// check reads the row through decode into a receiver holding the
// seed.
func (tc jsonDecodeCase[T, PT]) check(t *testing.T, decode func(PT) error) {
	t.Helper()
	got := tc.seed
	err := decode(PT(&got))
	if tc.wantErr == nil {
		core.AssertMustNotEqual(t, tc.want, tc.seed, "seed")
		core.AssertNoError(t, err, "decode")
		core.AssertEqual(t, tc.want, got, "value")
		return
	}
	core.AssertMustError(t, err, "decode")
	core.AssertContains(t, err.Error(), "UnmarshalJSON", "method name")
	assertParseReport[T](t, err, tc.text, tc.wantErr)
	core.AssertEqual(t, tc.seed, got, "untouched")
}

// jsonDirectCase pins one value read by calling UnmarshalJSON directly,
// a value that is not valid JSON, which json.Unmarshal refuses before
// it calls the method.
type jsonDirectCase[T num.Number[T], PT jsonUnmarshaler[T]] struct {
	jsonDecodeCase[T, PT]
}

func newJSONDirectCase[T num.Number[T], PT jsonUnmarshaler[T]](name,
	in string, seed, want T) jsonDirectCase[T, PT] {
	return jsonDirectCase[T, PT]{newJSONDecodeCase[T, PT](name, in, seed,
		want)}
}

// newJSONDirectCaseErr returns a direct row failing on in with wantErr,
// its report quoting text, the text the parser was given.
func newJSONDirectCaseErr[T num.Number[T], PT jsonUnmarshaler[T]](name,
	in string, seed T, text string, wantErr error) jsonDirectCase[T, PT] {
	return jsonDirectCase[T, PT]{newJSONDecodeCaseErr[T, PT](name, in, seed,
		text, wantErr)}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonDirectCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	tc.check(t, tc.decodeDirect)
}

// jsonDecodeCasesOf turns the rows of a JSON table, every value of
// which is a T, into decode rows seeded with seed, so the JSON written
// for each value, number or string, must read back as that value.
func jsonDecodeCasesOf[T num.Number[T], PT jsonUnmarshaler[T]](
	t *testing.T, cases []jsonCase, seed T) []jsonDecodeCase[T, PT] {
	t.Helper()
	out := make([]jsonDecodeCase[T, PT], 0, len(cases))
	for _, tc := range cases {
		v := core.AssertMustTypeIs[T](t, tc.in, "%s value", tc.name)
		out = append(out, newJSONDecodeCase[T, PT](tc.name, tc.json(),
			seed, v))
	}
	return out
}

// jsonTokenCases returns the rows every integer reads alike, seeded
// with seed: forty-two as a number, as a string and with its digits
// escaped, then the tokens no integer reads, a fraction or an exponent
// in either form, the other JSON types and strings holding other than
// the digits.
func jsonTokenCases[T num.Number[T], PT jsonUnmarshaler[T]](seed,
	fortyTwo T) []jsonDecodeCase[T, PT] {
	return []jsonDecodeCase[T, PT]{
		newJSONDecodeCase[T, PT]("number", "42", seed, fortyTwo),
		newJSONDecodeCase[T, PT]("string", `"42"`, seed, fortyTwo),
		newJSONDecodeCase[T, PT]("escaped digits", "\"\\u0034\\u0032\"", seed,
			fortyTwo),
		newJSONTokenCaseErr[T, PT]("fraction", "4.2", seed, num.ErrSyntax),
		newJSONTokenCaseErr[T, PT]("exponent", "42e0", seed, num.ErrSyntax),
		newJSONDecodeCaseErr[T, PT]("fraction string", `"4.2"`, seed, "4.2",
			num.ErrSyntax),
		newJSONDecodeCaseErr[T, PT]("exponent string", `"42e0"`, seed, "42e0",
			num.ErrSyntax),
		newJSONTokenCaseErr[T, PT]("true", "true", seed, num.ErrSyntax),
		newJSONTokenCaseErr[T, PT]("object", `{"v":42}`, seed, num.ErrSyntax),
		newJSONTokenCaseErr[T, PT]("array", "[42]", seed, num.ErrSyntax),
		newJSONDecodeCaseErr[T, PT]("empty string", `""`, seed, "",
			num.ErrSyntax),
		newJSONDecodeCaseErr[T, PT]("escaped space", "\"\\u002042\"", seed,
			" 42", num.ErrSyntax),
	}
}

// jsonDirectCases returns the direct rows every integer reads alike,
// seeded with seed: a number with a leading zero, parsed as given, and
// strings that are not valid JSON, one with a quote inside, read in
// place between its outer quotes, one left open, one followed by a
// space and a lone quote, none of which ends on its closing quote, and
// one whose closing quote is escaped, which does not unquote, so each
// is parsed as given and fails on its quote.
func jsonDirectCases[T num.Number[T], PT jsonUnmarshaler[T]](seed,
	fortyTwo T) []jsonDirectCase[T, PT] {
	return []jsonDirectCase[T, PT]{
		newJSONDirectCase[T, PT]("leading zero", "042", seed, fortyTwo),
		newJSONDirectCaseErr[T, PT]("quote inside", `"4"2"`, seed, `4"2`,
			num.ErrSyntax),
		newJSONDirectCaseErr[T, PT]("open string", `"42`, seed, `"42`,
			num.ErrSyntax),
		newJSONDirectCaseErr[T, PT]("space after string", `"42" `, seed,
			`"42" `, num.ErrSyntax),
		newJSONDirectCaseErr[T, PT]("lone quote", `"`, seed, `"`,
			num.ErrSyntax),
		newJSONDirectCaseErr[T, PT]("escaped closing quote", `"42\"`, seed,
			`"42\"`, num.ErrSyntax),
	}
}

func jsonDecodeInt32Cases() []jsonDecodeCase[num.Int32, *num.Int32] {
	return append(jsonTokenCases[num.Int32, *num.Int32](seed32,
		num.AsInt32(42)),
		newJSONTokenCaseErr[num.Int32, *num.Int32]("past max", "2147483648",
			seed32, num.ErrRange),
	)
}

func jsonDecodeInt64Cases() []jsonDecodeCase[num.Int64, *num.Int64] {
	return append(jsonTokenCases[num.Int64, *num.Int64](seed64,
		num.AsInt64(42)),
		newJSONDecodeCaseErr[num.Int64, *num.Int64]("past max",
			`"9223372036854775808"`, seed64, "9223372036854775808",
			num.ErrRange),
	)
}

func jsonDecodeUint128Cases() []jsonDecodeCase[num.Uint128, *num.Uint128] {
	return append(jsonTokenCases[num.Uint128, *num.Uint128](seedU128,
		num.AsUint128(42)),
		newJSONTokenCaseErr[num.Uint128, *num.Uint128]("negative", "-1",
			seedU128, num.ErrSyntax),
		newJSONDecodeCaseErr[num.Uint128, *num.Uint128]("past max",
			`"340282366920938463463374607431768211456"`, seedU128,
			"340282366920938463463374607431768211456", num.ErrRange),
	)
}

func jsonDecodeInt128Cases() []jsonDecodeCase[num.Int128, *num.Int128] {
	return append(jsonTokenCases[num.Int128, *num.Int128](seed128,
		num.AsInt128(42)),
		newJSONDecodeCaseErr[num.Int128, *num.Int128]("below min",
			`"-170141183460469231731687303715884105729"`, seed128,
			"-170141183460469231731687303715884105729", num.ErrRange),
	)
}

func TestUnmarshalJSON(t *testing.T) {
	t.Run("int32", func(t *testing.T) {
		runTestUnmarshalJSON(t, jsonInt32Cases, jsonDecodeInt32Cases(),
			seed32, num.AsInt32(42))
	})
	t.Run("int64", func(t *testing.T) {
		runTestUnmarshalJSON(t, jsonInt64Cases, jsonDecodeInt64Cases(),
			seed64, num.AsInt64(42))
	})
	t.Run("uint128", func(t *testing.T) {
		runTestUnmarshalJSON(t, jsonUint128Cases, jsonDecodeUint128Cases(),
			seedU128, num.AsUint128(42))
	})
	t.Run("int128", func(t *testing.T) {
		runTestUnmarshalJSON(t, jsonInt128Cases, jsonDecodeInt128Cases(),
			seed128, num.AsInt128(42))
	})
	t.Run("allocs", runTestUnmarshalJSONAllocs)
}

// jsonAllocsCase pins that UnmarshalJSON on *T stores a value that
// reads, and allocates nothing doing it.
type jsonAllocsCase[T any, PT jsonUnmarshaler[T]] struct {
	want T
	in   string
	name string
}

func newJSONAllocsCase[T any, PT jsonUnmarshaler[T]](name, in string,
	want T) jsonAllocsCase[T, PT] {
	return jsonAllocsCase[T, PT]{name: name, in: in, want: want}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonAllocsCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonAllocsCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	b := []byte(tc.in)
	var got T
	var err error
	allocs := testing.AllocsPerRun(10, func() {
		err = PT(&got).UnmarshalJSON(b)
	})
	core.AssertNoError(t, err, "decode")
	core.AssertEqual(t, tc.want, got, "value")
	core.AssertEqual(t, 0, allocs, "allocs")
}

// jsonAllocsCases returns the allocation rows of T on text, its
// longest, which reads as want: as a number, and as a string with no
// escape, which for the 128-bit integers runs past the 32 bytes the
// runtime converts from bytes to a string on the stack.
func jsonAllocsCases[T any, PT jsonUnmarshaler[T]](text string,
	want T) []jsonAllocsCase[T, PT] {
	return []jsonAllocsCase[T, PT]{
		newJSONAllocsCase[T, PT]("number", text, want),
		newJSONAllocsCase[T, PT]("string", strconv.Quote(text), want),
	}
}

// runTestUnmarshalJSONAllocs checks UnmarshalJSON allocates nothing on
// each integer's longest text.
func runTestUnmarshalJSONAllocs(t *testing.T) {
	t.Helper()
	int32s := jsonAllocsCases[num.Int32, *num.Int32]("-2147483648",
		num.AsInt32(math.MinInt32))
	int64s := jsonAllocsCases[num.Int64, *num.Int64]("-9223372036854775808",
		num.AsInt64(math.MinInt64))
	uint128s := jsonAllocsCases[num.Uint128, *num.Uint128](maxUint128Text,
		num.MaxUint128)
	int128s := jsonAllocsCases[num.Int128, *num.Int128](minInt128Text,
		num.MinInt128)
	t.Run("int32", func(t *testing.T) { core.RunTestCases(t, int32s) })
	t.Run("int64", func(t *testing.T) { core.RunTestCases(t, int64s) })
	t.Run("uint128", func(t *testing.T) { core.RunTestCases(t, uint128s) })
	t.Run("int128", func(t *testing.T) { core.RunTestCases(t, int128s) })
}

// runTestUnmarshalJSON runs the UnmarshalJSON test of T: every row of
// its JSON table read back, its token rows, a null, and direct calls
// json.Unmarshal does not make, on values that are not valid JSON,
// read as forty-two where they read, and on a nil receiver.
func runTestUnmarshalJSON[T num.Number[T], PT jsonUnmarshaler[T]](
	t *testing.T, table func() []jsonCase, tokens []jsonDecodeCase[T, PT],
	seed, fortyTwo T) {
	t.Helper()
	rows := jsonDecodeCasesOf[T, PT](t, table(), seed)
	direct := jsonDirectCases[T, PT](seed, fortyTwo)
	nils := jsonNilCases[T, PT]()
	t.Run("round trip", func(t *testing.T) { core.RunTestCases(t, rows) })
	t.Run("tokens", func(t *testing.T) { core.RunTestCases(t, tokens) })
	t.Run("null", func(t *testing.T) {
		runTestUnmarshalJSONNull[T, PT](t, seed)
	})
	t.Run("direct", func(t *testing.T) { core.RunTestCases(t, direct) })
	t.Run("nil receiver", func(t *testing.T) { core.RunTestCases(t, nils) })
}

// runTestUnmarshalJSONNull checks a null stores nothing, read by
// json.Unmarshal into a receiver holding seed.
func runTestUnmarshalJSONNull[T any, PT jsonUnmarshaler[T]](t *testing.T,
	seed T) {
	t.Helper()
	got := seed
	core.AssertNoError(t, json.Unmarshal([]byte("null"), PT(&got)), "null")
	core.AssertEqual(t, seed, got, "untouched")
}

// jsonNilCase pins one value read by calling UnmarshalJSON directly on
// a nil receiver, which json.Unmarshal does not do: the one cause the
// row expects, of core.ErrNilReceiver and ErrSyntax, and the method's
// name in front. A value that parses fails with core.ErrNilReceiver, a
// null included, and one that does not reports its own failure first.
type jsonNilCase[T any, PT jsonUnmarshaler[T]] struct {
	wantErr error
	in      string
	name    string
}

// newJSONNilCase returns a row on a value that parses, failing on the
// nil receiver.
func newJSONNilCase[T any, PT jsonUnmarshaler[T]](name,
	in string) jsonNilCase[T, PT] {
	return jsonNilCase[T, PT]{name: name, in: in,
		wantErr: core.ErrNilReceiver}
}

// newJSONNilCaseSyntax returns a row on a value that does not parse,
// failing on its text before the nil receiver.
func newJSONNilCaseSyntax[T any, PT jsonUnmarshaler[T]](name,
	in string) jsonNilCase[T, PT] {
	return jsonNilCase[T, PT]{name: name, in: in, wantErr: num.ErrSyntax}
}

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonNilCase[T, PT]) Name() string { return tc.name }

//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (tc jsonNilCase[T, PT]) Test(t *testing.T) {
	t.Helper()
	var p PT
	err := p.UnmarshalJSON([]byte(tc.in))
	core.AssertMustError(t, err, "decode")
	core.AssertErrorIs(t, err, tc.wantErr, "cause")
	core.AssertNotEqual(t, errors.Is(err, core.ErrNilReceiver),
		errors.Is(err, num.ErrSyntax), "one cause")
	core.AssertContains(t, err.Error(), "UnmarshalJSON", "method name")
}

// jsonNilCases returns the nil-receiver rows every integer reads alike:
// a null, a number and a string that parse, and a number and a string
// that do not.
func jsonNilCases[T any, PT jsonUnmarshaler[T]]() []jsonNilCase[T, PT] {
	return []jsonNilCase[T, PT]{
		newJSONNilCase[T, PT]("null", "null"),
		newJSONNilCase[T, PT]("number", "1"),
		newJSONNilCase[T, PT]("string", `"1"`),
		newJSONNilCaseSyntax[T, PT]("bad number", "x"),
		newJSONNilCaseSyntax[T, PT]("bad string", `"x"`),
	}
}
