package forms_test

import (
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web/forms"
)

// newParseRangeTestCase expects fn(s, minV, maxV) to return expected with
// no error.
//
//revive:disable-next-line:argument-limit
func newParseRangeTestCase[T any](name string,
	fn func(string, T, T) (T, error),
	s string, minV, maxV, expected T) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseRange(fn, s, minV, maxV),
		expected: expected,
	}
}

// newParseRangeErrorTestCase expects fn(s, minV, maxV) to return expected
// with an error matching target.
//
//revive:disable-next-line:argument-limit
func newParseRangeErrorTestCase[T any](name string,
	fn func(string, T, T) (T, error),
	s string, minV, maxV, expected T, target error) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseRange(fn, s, minV, maxV),
		expected: expected,
		target:   target,
	}
}

// newParseRangeNaNErrorTestCase expects fn(s, minV, maxV) to return NaN
// with an error matching target.
//
//revive:disable-next-line:argument-limit
func newParseRangeNaNErrorTestCase[T core.Float](name string,
	fn func(string, T, T) (T, error),
	s string, minV, maxV T, target error) parseNaNTestCase {
	return parseNaNTestCase{
		name: name,
		parse: func() (float64, error) {
			v, err := fn(s, minV, maxV)
			return float64(v), err
		},
		target: target,
	}
}

func bindParseRange[T any](fn func(string, T, T) (T, error),
	s string, minV, maxV T) func() (any, error) {
	return func() (any, error) {
		v, err := fn(s, minV, maxV)
		return v, err
	}
}

// newParseBaseRangeTestCase expects fn(s, base, minV, maxV) to return
// expected with no error.
//
//revive:disable-next-line:argument-limit
func newParseBaseRangeTestCase[T any](name string,
	fn func(string, int, T, T) (T, error),
	s string, base int, minV, maxV, expected T) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseBaseRange(fn, s, base, minV, maxV),
		expected: expected,
	}
}

// newParseBaseRangeErrorTestCase expects fn(s, base, minV, maxV) to return
// expected with an error matching target.
//
//revive:disable-next-line:argument-limit
func newParseBaseRangeErrorTestCase[T any](name string,
	fn func(string, int, T, T) (T, error),
	s string, base int, minV, maxV, expected T,
	target error) parseTestCase {
	return parseTestCase{
		name:     name,
		parse:    bindParseBaseRange(fn, s, base, minV, maxV),
		expected: expected,
		target:   target,
	}
}

func bindParseBaseRange[T any](fn func(string, int, T, T) (T, error),
	s string, base int, minV, maxV T) func() (any, error) {
	return func() (any, error) {
		v, err := fn(s, base, minV, maxV)
		return v, err
	}
}

func TestParseSignedInRange(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseBaseRangeTestCase("within",
			forms.ParseSignedInRange[int], "5", 10, -50, 50,
			5),
		newParseBaseRangeTestCase("at min",
			forms.ParseSignedInRange[int], "-50", 10, -50, 50,
			-50),
		newParseBaseRangeTestCase("at max",
			forms.ParseSignedInRange[int], "50", 10, -50, 50,
			50),
		newParseBaseRangeErrorTestCase("below min",
			forms.ParseSignedInRange[int], "-51", 10, -50, 50,
			-51, forms.ErrRange),
		newParseBaseRangeErrorTestCase("above max",
			forms.ParseSignedInRange[int], "51", 10, -50, 50,
			51, forms.ErrRange),
		newParseBaseRangeErrorTestCase("false range",
			forms.ParseSignedInRange[int], "5", 10, 50, -50,
			5, forms.ErrRange),

		newParseBaseRangeTestCase("base 16 within",
			forms.ParseSignedInRange[int], "a", 16, -50, 50,
			10),
		newParseBaseRangeErrorTestCase("base 16 above max",
			forms.ParseSignedInRange[int], "33", 16, -50, 50,
			51, forms.ErrRange),

		newParseBaseRangeErrorTestCase("above the type's max",
			forms.ParseSignedInRange[int8], "128", 10,
			math.MinInt8, math.MaxInt8,
			127, forms.ErrRange),
		newParseBaseRangeErrorTestCase("letters",
			forms.ParseSignedInRange[int], "x", 10, -50, 50,
			0, forms.ErrSyntax),
		newParseBaseRangeErrorTestCase("empty",
			forms.ParseSignedInRange[int], "", 10, -50, 50,
			0, forms.ErrSyntax),
	))
}

func TestParseUnsignedInRange(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseBaseRangeTestCase("within",
			forms.ParseUnsignedInRange[uint], "5", 10, 2, 50,
			5),
		newParseBaseRangeTestCase("at min",
			forms.ParseUnsignedInRange[uint], "2", 10, 2, 50,
			2),
		newParseBaseRangeTestCase("at max",
			forms.ParseUnsignedInRange[uint], "50", 10, 2, 50,
			50),
		newParseBaseRangeErrorTestCase("below min",
			forms.ParseUnsignedInRange[uint], "1", 10, 2, 50,
			1, forms.ErrRange),
		newParseBaseRangeErrorTestCase("above max",
			forms.ParseUnsignedInRange[uint], "51", 10, 2, 50,
			51, forms.ErrRange),
		newParseBaseRangeErrorTestCase("false range",
			forms.ParseUnsignedInRange[uint], "5", 10, 50, 2,
			5, forms.ErrRange),

		newParseBaseRangeTestCase("base 16 within",
			forms.ParseUnsignedInRange[uint], "a", 16, 2, 50,
			10),

		newParseBaseRangeErrorTestCase("above the type's max",
			forms.ParseUnsignedInRange[uint8], "256", 10,
			0, math.MaxUint8,
			255, forms.ErrRange),
		newParseBaseRangeErrorTestCase("negative",
			forms.ParseUnsignedInRange[uint], "-1", 10, 2, 50,
			0, forms.ErrSyntax),
		newParseBaseRangeErrorTestCase("letters",
			forms.ParseUnsignedInRange[uint], "x", 10, 2, 50,
			0, forms.ErrSyntax),
		newParseBaseRangeErrorTestCase("empty",
			forms.ParseUnsignedInRange[uint], "", 10, 2, 50,
			0, forms.ErrSyntax),
	))
}

func TestParseFloatInRange(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseRangeTestCase("within",
			forms.ParseFloatInRange[float64], "0.5", 0, 1,
			0.5),
		newParseRangeTestCase("at min",
			forms.ParseFloatInRange[float64], "0", 0, 1,
			0),
		newParseRangeTestCase("at max",
			forms.ParseFloatInRange[float64], "1", 0, 1,
			1),
		newParseRangeErrorTestCase("below min",
			forms.ParseFloatInRange[float64], "-0.5", 0, 1,
			-0.5, forms.ErrRange),
		newParseRangeErrorTestCase("above max",
			forms.ParseFloatInRange[float64], "1.5", 0, 1,
			1.5, forms.ErrRange),
		newParseRangeErrorTestCase("false range",
			forms.ParseFloatInRange[float64], "0.5", 1, 0,
			0.5, forms.ErrRange),

		newParseRangeErrorTestCase("infinity",
			forms.ParseFloatInRange[float64], "Inf", 0, 1,
			math.Inf(1), forms.ErrRange),
		newParseRangeErrorTestCase("negative infinity",
			forms.ParseFloatInRange[float64], "-Inf", 0, 1,
			math.Inf(-1), forms.ErrRange),
		newParseRangeTestCase("infinity up to infinity",
			forms.ParseFloatInRange[float64], "Inf", 0, math.Inf(1),
			math.Inf(1)),

		newParseRangeErrorTestCase("above the type's max",
			forms.ParseFloatInRange[float32], "1e39",
			0, float32(math.Inf(1)),
			float32(math.Inf(1)), forms.ErrRange),
		newParseRangeErrorTestCase("letters",
			forms.ParseFloatInRange[float64], "x", 0, 1,
			0, forms.ErrSyntax),
		newParseRangeErrorTestCase("empty",
			forms.ParseFloatInRange[float64], "", 0, 1,
			0, forms.ErrSyntax),
	))
}

func TestParseFloatInRangeNaN(t *testing.T) {
	core.RunTestCases(t, core.S(
		newParseRangeNaNErrorTestCase("finite bounds",
			forms.ParseFloatInRange[float64], "NaN", 0, 1,
			forms.ErrRange),
		newParseRangeNaNErrorTestCase("infinite bounds",
			forms.ParseFloatInRange[float64], "NaN",
			math.Inf(-1), math.Inf(1),
			forms.ErrRange),
		newParseRangeNaNErrorTestCase("float32",
			forms.ParseFloatInRange[float32], "NaN", 0, 1,
			forms.ErrRange),
	))
}
