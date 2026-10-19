package forms_test

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web/forms"
)

var (
	_ core.TestCase = formatTestCase{}
	_ core.TestCase = parseNaNTestCase{}
	_ core.TestCase = parseTestCase{}
)

// Named types check the conversion to a defined type and, for numbers,
// that the size comes from the underlying type.
type (
	namedBool  bool
	namedFloat float32
	namedInt   int8
	namedUint  uint8
)

type parseTestCase struct {
	expected any
	target   error
	parse    func() (any, error)
	name     string
}

func (tc parseTestCase) Name() string {
	return tc.name
}

func (tc parseTestCase) Test(t *testing.T) {
	t.Helper()

	value, err := tc.parse()
	if tc.target == nil {
		core.AssertNoError(t, err, "error")
	} else {
		core.AssertErrorIs(t, err, tc.target, "error")
	}
	core.AssertEqual(t, tc.expected, value, "value")
}

// newParseTestCase expects fn(s) to return expected with no error.
func newParseTestCase[T any](name string, fn func(string) (T, error),
	s string, expected T) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParse(fn, s),
		expected: expected,
	}
}

// newParseErrorTestCase expects fn(s) to return expected with an error
// matching target.
func newParseErrorTestCase[T any](name string, fn func(string) (T, error),
	s string, expected T, target error) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParse(fn, s),
		expected: expected,
		target:   target,
	}
}

func bindParse[T any](fn func(string) (T, error),
	s string) func() (any, error) {
	return func() (any, error) {
		v, err := fn(s)
		return v, err
	}
}

// newParseBaseTestCase expects fn(s, base) to return expected with no
// error.
func newParseBaseTestCase[T any](name string, fn func(string, int) (T, error),
	s string, base int, expected T) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseBase(fn, s, base),
		expected: expected,
	}
}

// newParseBaseErrorTestCase expects fn(s, base) to return expected with
// an error matching target.
//
//revive:disable-next-line:argument-limit
func newParseBaseErrorTestCase[T any](name string,
	fn func(string, int) (T, error),
	s string, base int, expected T, target error) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseBase(fn, s, base),
		expected: expected,
		target:   target,
	}
}

func bindParseBase[T any](fn func(string, int) (T, error),
	s string, base int) func() (any, error) {
	return func() (any, error) {
		v, err := fn(s, base)
		return v, err
	}
}

func TestParseSigned(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseBaseTestCase("int", forms.ParseSigned[int], "42", 10, 42),
		newParseBaseTestCase("int negative",
			forms.ParseSigned[int], "-42", 10, -42),
		newParseBaseTestCase("int8 max",
			forms.ParseSigned[int8], "127", 10, 127),
		newParseBaseTestCase("int8 min",
			forms.ParseSigned[int8], "-128", 10, -128),
		newParseBaseErrorTestCase("int8 above max",
			forms.ParseSigned[int8], "128", 10, 127, forms.ErrRange),
		newParseBaseErrorTestCase("int8 below min",
			forms.ParseSigned[int8], "-129", 10, -128, forms.ErrRange),
		newParseBaseErrorTestCase("int16 above max",
			forms.ParseSigned[int16], "32768", 10, 32767, forms.ErrRange),
		newParseBaseErrorTestCase("int32 above max",
			forms.ParseSigned[int32], "2147483648", 10,
			math.MaxInt32, forms.ErrRange),
		newParseBaseTestCase("int64 max", forms.ParseSigned[int64],
			"9223372036854775807", 10, math.MaxInt64),
		newParseBaseErrorTestCase("named type above max",
			forms.ParseSigned[namedInt], "128", 10, 127, forms.ErrRange),

		newParseBaseTestCase("base 16", forms.ParseSigned[int], "ff", 16, 255),
		newParseBaseTestCase("base 2", forms.ParseSigned[int], "101", 2, 5),
		newParseBaseTestCase("base 0 prefix",
			forms.ParseSigned[int], "0x10", 0, 16),

		newParseBaseErrorTestCase("empty",
			forms.ParseSigned[int], "", 10, 0, forms.ErrSyntax),
		newParseBaseErrorTestCase("letters",
			forms.ParseSigned[int], "abc", 10, 0, forms.ErrSyntax),
		newParseBaseErrorTestCase("decimal",
			forms.ParseSigned[int], "1.5", 10, 0, forms.ErrSyntax),
		newParseBaseErrorTestCase("leading space",
			forms.ParseSigned[int], " 1", 10, 0, forms.ErrSyntax),
	))
}

func TestParseUnsigned(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseBaseTestCase("uint", forms.ParseUnsigned[uint], "42", 10, 42),
		newParseBaseTestCase("uint8 max",
			forms.ParseUnsigned[uint8], "255", 10, 255),
		newParseBaseErrorTestCase("uint8 above max",
			forms.ParseUnsigned[uint8], "256", 10, 255, forms.ErrRange),
		newParseBaseErrorTestCase("uint16 above max",
			forms.ParseUnsigned[uint16], "65536", 10, 65535, forms.ErrRange),
		newParseBaseTestCase("uint64 max", forms.ParseUnsigned[uint64],
			"18446744073709551615", 10, math.MaxUint64),
		newParseBaseErrorTestCase("named type above max",
			forms.ParseUnsigned[namedUint], "256", 10, 255, forms.ErrRange),

		newParseBaseTestCase("base 16",
			forms.ParseUnsigned[uint8], "ff", 16, 255),
		newParseBaseTestCase("base 0 prefix",
			forms.ParseUnsigned[uint], "0b101", 0, 5),

		newParseBaseErrorTestCase("negative",
			forms.ParseUnsigned[uint], "-1", 10, 0, forms.ErrSyntax),
		newParseBaseErrorTestCase("empty",
			forms.ParseUnsigned[uint], "", 10, 0, forms.ErrSyntax),
		newParseBaseErrorTestCase("letters",
			forms.ParseUnsigned[uint], "abc", 10, 0, forms.ErrSyntax),
	))
}

func TestParseFloat(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseTestCase("float64", forms.ParseFloat[float64], "1.5", 1.5),
		newParseTestCase("float64 negative",
			forms.ParseFloat[float64], "-0.25", -0.25),
		newParseTestCase("float64 exponent",
			forms.ParseFloat[float64], "1e3", 1000),
		newParseTestCase("float64 hexadecimal",
			forms.ParseFloat[float64], "0x1p-2", 0.25),
		newParseTestCase("float32", forms.ParseFloat[float32], "0.1", 0.1),

		newParseTestCase("infinity",
			forms.ParseFloat[float64], "Inf", math.Inf(1)),
		newParseTestCase("negative infinity",
			forms.ParseFloat[float64], "-Inf", math.Inf(-1)),
		newParseErrorTestCase("float64 above max",
			forms.ParseFloat[float64], "1e309", math.Inf(1), forms.ErrRange),
		newParseErrorTestCase("float32 above max",
			forms.ParseFloat[float32], "1e39",
			float32(math.Inf(1)), forms.ErrRange),
		newParseErrorTestCase("named type above max",
			forms.ParseFloat[namedFloat], "1e39",
			namedFloat(math.Inf(1)), forms.ErrRange),

		newParseErrorTestCase("empty",
			forms.ParseFloat[float64], "", 0, forms.ErrSyntax),
		newParseErrorTestCase("letters",
			forms.ParseFloat[float64], "abc", 0, forms.ErrSyntax),
	))
}

func TestParseBool(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseTestCase("true", forms.ParseBool[bool], "true", true),
		newParseTestCase("TRUE", forms.ParseBool[bool], "TRUE", true),
		newParseTestCase("false", forms.ParseBool[bool], "false", false),
		newParseTestCase("one", forms.ParseBool[bool], "1", true),
		newParseTestCase("zero", forms.ParseBool[bool], "0", false),
		newParseTestCase("t", forms.ParseBool[bool], "t", true),
		newParseTestCase("f", forms.ParseBool[bool], "f", false),
		newParseTestCase("y", forms.ParseBool[bool], "y", true),
		newParseTestCase("n", forms.ParseBool[bool], "n", false),
		newParseTestCase("yes", forms.ParseBool[bool], "yes", true),
		newParseTestCase("no", forms.ParseBool[bool], "no", false),

		newParseTestCase("mixed case true",
			forms.ParseBool[bool], "tRuE", true),
		newParseTestCase("upper case yes", forms.ParseBool[bool], "YES", true),
		newParseTestCase("upper case n", forms.ParseBool[bool], "N", false),
		newParseTestCase("named type", forms.ParseBool[namedBool], "yes", true),

		newParseErrorTestCase("empty",
			forms.ParseBool[bool], "", false, forms.ErrSyntax),
		newParseErrorTestCase("unknown word",
			forms.ParseBool[bool], "maybe", false, forms.ErrSyntax),
		newParseErrorTestCase("leading space",
			forms.ParseBool[bool], " yes", false, forms.ErrSyntax),
	))
}

type parseNaNTestCase struct {
	target error
	parse  func() (float64, error)
	name   string
}

func (tc parseNaNTestCase) Name() string {
	return tc.name
}

func (tc parseNaNTestCase) Test(t *testing.T) {
	t.Helper()

	value, err := tc.parse()
	if tc.target == nil {
		core.AssertNoError(t, err, "error")
	} else {
		core.AssertErrorIs(t, err, tc.target, "error")
	}
	core.AssertTrue(t, math.IsNaN(value), "NaN")
}

// newParseNaNTestCase expects fn(s) to return NaN with no error.
func newParseNaNTestCase[T core.Float](name string,
	fn func(string) (T, error), s string) parseNaNTestCase {
	return parseNaNTestCase{
		name: name,
		parse: func() (float64, error) {
			v, err := fn(s)
			return float64(v), err
		},
	}
}

func TestParseFloatNaN(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseNaNTestCase("float64", forms.ParseFloat[float64], "NaN"),
		newParseNaNTestCase("float32", forms.ParseFloat[float32], "NaN"),
		newParseNaNTestCase("lower case", forms.ParseFloat[float64], "nan"),
	))
}

type formatTestCase struct {
	format   func() string
	expected string
	name     string
}

func (tc formatTestCase) Name() string {
	return tc.name
}

func (tc formatTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.expected, tc.format(), "text")
}

// newFormatBaseTestCase expects fn(v, base) to return expected.
func newFormatBaseTestCase[T any](name string, fn func(T, int) string,
	v T, base int, expected string) formatTestCase {
	return formatTestCase{
		name:     name,
		format:   func() string { return fn(v, base) },
		expected: expected,
	}
}

// newFormatFloatTestCase expects fn(v, format, prec) to return expected.
//
//revive:disable-next-line:argument-limit
func newFormatFloatTestCase[T core.Float](name string,
	fn func(T, byte, int) string,
	v T, format byte, prec int, expected string) formatTestCase {
	return formatTestCase{
		name:     name,
		format:   func() string { return fn(v, format, prec) },
		expected: expected,
	}
}

func TestFormatSigned(t *testing.T) {
	core.RunTestCases(t, core.S(
		newFormatBaseTestCase("int", forms.FormatSigned[int], 42, 10, "42"),
		newFormatBaseTestCase("int8 min",
			forms.FormatSigned[int8], -128, 10, "-128"),
		newFormatBaseTestCase("int64 min", forms.FormatSigned[int64],
			math.MinInt64, 10, "-9223372036854775808"),
		newFormatBaseTestCase("named type",
			forms.FormatSigned[namedInt], -1, 10, "-1"),
		newFormatBaseTestCase("base 16",
			forms.FormatSigned[int], 255, 16, "ff"),
		newFormatBaseTestCase("base 2", forms.FormatSigned[int], -5, 2, "-101"),
	))
}

func TestFormatUnsigned(t *testing.T) {
	core.RunTestCases(t, core.S(
		newFormatBaseTestCase("uint", forms.FormatUnsigned[uint], 42, 10, "42"),
		newFormatBaseTestCase("uint8 max",
			forms.FormatUnsigned[uint8], 255, 10, "255"),
		newFormatBaseTestCase("uint64 max", forms.FormatUnsigned[uint64],
			math.MaxUint64, 10, "18446744073709551615"),
		newFormatBaseTestCase("named type",
			forms.FormatUnsigned[namedUint], 7, 10, "7"),
		newFormatBaseTestCase("base 16",
			forms.FormatUnsigned[uint], 255, 16, "ff"),
	))
}

func TestFormatFloat(t *testing.T) {
	core.RunTestCases(t, core.S(
		newFormatFloatTestCase("float64",
			forms.FormatFloat[float64], 1.5, 'f', -1, "1.5"),
		newFormatFloatTestCase("float32 shortest",
			forms.FormatFloat[float32], 0.1, 'f', -1, "0.1"),
		newFormatFloatTestCase("named type shortest",
			forms.FormatFloat[namedFloat], 0.1, 'f', -1, "0.1"),
		newFormatFloatTestCase("large value in full",
			forms.FormatFloat[float64], 1e21, 'f', -1,
			"1000000000000000000000"),
		newFormatFloatTestCase("exponent format",
			forms.FormatFloat[float64], 1234.5, 'e', 2, "1.23e+03"),
		newFormatFloatTestCase("infinity",
			forms.FormatFloat[float64], math.Inf(1), 'f', -1, "+Inf"),
		newFormatFloatTestCase("NaN",
			forms.FormatFloat[float64], math.NaN(), 'f', -1, "NaN"),
	))
}
