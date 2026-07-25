package num_test

import (
	"darvaza.org/x/time/num"
)

// Parse-only rows for the text suite. They share the textCase type and its
// textType/textPtr constraints with the round-trip suite in text_test.go;
// an OK row leaves text empty so only the UnmarshalText direction runs, and
// an Err row carries the sentinel UnmarshalText must report.

func newUint128ParseCase(name, input string,
	want num.Uint128) textCase[num.Uint128, *num.Uint128] {
	return textCase[num.Uint128, *num.Uint128]{
		name: name, input: input, val: want}
}

func newUint128ParseErrCase(name, input string,
	wantErr error) textCase[num.Uint128, *num.Uint128] {
	return textCase[num.Uint128, *num.Uint128]{
		name: name, input: input, wantErr: wantErr}
}

func newInt128ParseCase(name, input string,
	want num.Int128) textCase[num.Int128, *num.Int128] {
	return textCase[num.Int128, *num.Int128]{
		name: name, input: input, val: want}
}

func newInt128ParseErrCase(name, input string,
	wantErr error) textCase[num.Int128, *num.Int128] {
	return textCase[num.Int128, *num.Int128]{
		name: name, input: input, wantErr: wantErr}
}

func newInt32ParseCase(name, input string,
	want num.Int32) textCase[num.Int32, *num.Int32] {
	return textCase[num.Int32, *num.Int32]{
		name: name, input: input, val: want}
}

func newInt32ParseErrCase(name, input string,
	wantErr error) textCase[num.Int32, *num.Int32] {
	return textCase[num.Int32, *num.Int32]{
		name: name, input: input, wantErr: wantErr}
}

func newInt64ParseCase(name, input string,
	want num.Int64) textCase[num.Int64, *num.Int64] {
	return textCase[num.Int64, *num.Int64]{
		name: name, input: input, val: want}
}

func newInt64ParseErrCase(name, input string,
	wantErr error) textCase[num.Int64, *num.Int64] {
	return textCase[num.Int64, *num.Int64]{
		name: name, input: input, wantErr: wantErr}
}

func uint128ParseCases() []textCase[num.Uint128, *num.Uint128] {
	return []textCase[num.Uint128, *num.Uint128]{
		newUint128ParseCase("leading zeros", "007", num.NewUint128(0, 7)),
		newUint128ParseErrCase("empty", "", num.ErrSyntax),
		newUint128ParseErrCase("non-digit", "12x3", num.ErrSyntax),
		newUint128ParseErrCase("sign rejected", "+7", num.ErrSyntax),
		newUint128ParseErrCase("overflow 2^128",
			"340282366920938463463374607431768211456", num.ErrRange),
		newUint128ParseErrCase("far overflow",
			"999999999999999999999999999999999999999999", num.ErrRange),
	}
}

func int128ParseCases() []textCase[num.Int128, *num.Int128] {
	return []textCase[num.Int128, *num.Int128]{
		newInt128ParseCase("negative", "-42", num.NewInt128(-42)),
		newInt128ParseCase("plus sign", "+42", num.NewInt128(42)),
		newInt128ParseCase("negative zero", "-0", num.NewInt128(0)),
		newInt128ParseCase("min", "-170141183460469231731687303715884105728",
			num.MinInt128),
		newInt128ParseCase("max", "170141183460469231731687303715884105727",
			num.MaxInt128),
		newInt128ParseErrCase("empty", "", num.ErrSyntax),
		newInt128ParseErrCase("lone minus", "-", num.ErrSyntax),
		newInt128ParseErrCase("non-digit", "1.5", num.ErrSyntax),
		newInt128ParseErrCase("overflow +2^127",
			"170141183460469231731687303715884105728", num.ErrRange),
		newInt128ParseErrCase("overflow -(2^127+1)",
			"-170141183460469231731687303715884105729", num.ErrRange),
	}
}

func int32ParseCases() []textCase[num.Int32, *num.Int32] {
	return []textCase[num.Int32, *num.Int32]{
		newInt32ParseCase("negative", "-5", num.Int32(-5)),
		newInt32ParseErrCase("empty", "", num.ErrSyntax),
		newInt32ParseErrCase("non-digit", "5x", num.ErrSyntax),
		newInt32ParseErrCase("overflow high", "2147483648", num.ErrRange),
		newInt32ParseErrCase("overflow low", "-2147483649", num.ErrRange),
	}
}

func int64ParseCases() []textCase[num.Int64, *num.Int64] {
	return []textCase[num.Int64, *num.Int64]{
		newInt64ParseCase("negative", "-5", num.Int64(-5)),
		newInt64ParseErrCase("empty", "", num.ErrSyntax),
		newInt64ParseErrCase("non-digit", "5x", num.ErrSyntax),
		newInt64ParseErrCase("overflow high", "9223372036854775808", num.ErrRange),
		newInt64ParseErrCase("overflow low", "-9223372036854775809", num.ErrRange),
	}
}
