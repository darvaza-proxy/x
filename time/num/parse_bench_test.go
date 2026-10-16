package num_test

// cspell:words benchstat

import (
	"strconv"
	"testing"

	"darvaza.org/x/time/num"
)

// BenchmarkParse measures each text parser on a short text, and
// UnmarshalText on its bytes, under impl=num, beside strconv on the
// same text under impl=strconv, so benchstat -col /impl compares the
// two. On the bytes, strconv reads them converted to a string, as an
// UnmarshalText over strconv would. The natives are measured on their
// bounds as well, and the 128-bit integers on their bounds other than
// zero, which strconv does not read.
func BenchmarkParse(b *testing.B) {
	b.Run("Int32/impl=num", runBenchmarkParse(num.ParseInt32, "-1234567"))
	b.Run("Int32/impl=strconv", runBenchmarkParse(strconvParseInt(32),
		"-1234567"))
	b.Run("Int32Min/impl=num", runBenchmarkParse(num.ParseInt32,
		"-2147483648"))
	b.Run("Int32Min/impl=strconv", runBenchmarkParse(strconvParseInt(32),
		"-2147483648"))
	b.Run("Int32Max/impl=num", runBenchmarkParse(num.ParseInt32,
		"2147483647"))
	b.Run("Int32Max/impl=strconv", runBenchmarkParse(strconvParseInt(32),
		"2147483647"))
	b.Run("Int32.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Int32]([]byte("-1234567")))
	b.Run("Int32.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseIntBytes(32), []byte("-1234567")))
	b.Run("Int64/impl=num", runBenchmarkParse(num.ParseInt64, "-1234567"))
	b.Run("Int64/impl=strconv", runBenchmarkParse(strconvParseInt(64),
		"-1234567"))
	b.Run("Int64Min/impl=num", runBenchmarkParse(num.ParseInt64,
		"-9223372036854775808"))
	b.Run("Int64Min/impl=strconv", runBenchmarkParse(strconvParseInt(64),
		"-9223372036854775808"))
	b.Run("Int64Max/impl=num", runBenchmarkParse(num.ParseInt64,
		"9223372036854775807"))
	b.Run("Int64Max/impl=strconv", runBenchmarkParse(strconvParseInt(64),
		"9223372036854775807"))
	b.Run("Int64.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Int64]([]byte("-1234567")))
	b.Run("Int64.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseIntBytes(64), []byte("-1234567")))
	b.Run("Uint128/impl=num", runBenchmarkParse(num.ParseUint128, "1234567"))
	b.Run("Uint128/impl=strconv", runBenchmarkParse(strconvParseUint(64),
		"1234567"))
	b.Run("Uint128.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Uint128]([]byte("1234567")))
	b.Run("Uint128.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseUintBytes(64), []byte("1234567")))
	b.Run("Uint128Max/impl=num", runBenchmarkParse(num.ParseUint128,
		maxUint128Text))
	b.Run("Int128/impl=num", runBenchmarkParse(num.ParseInt128, "-1234567"))
	b.Run("Int128/impl=strconv", runBenchmarkParse(strconvParseInt(64),
		"-1234567"))
	b.Run("Int128.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Int128]([]byte("-1234567")))
	b.Run("Int128.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseIntBytes(64), []byte("-1234567")))
	b.Run("Int128Min/impl=num", runBenchmarkParse(num.ParseInt128,
		minInt128Text))
	b.Run("Int128Max/impl=num", runBenchmarkParse(num.ParseInt128,
		maxInt128Text))
}

// runBenchmarkParse returns a benchmark of parse on s, which must
// parse.
func runBenchmarkParse[S string | []byte, T any](parse func(S) (T, error),
	s S) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := parse(s); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// runBenchmarkUnmarshalText returns a benchmark of UnmarshalText on *T
// over text, which must parse. The receiver is declared once, outside
// the loop, so its escape through the call on PT is not counted per
// call.
func runBenchmarkUnmarshalText[T any, PT textUnmarshaler[T]](
	text []byte) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		var v T
		for b.Loop() {
			if err := PT(&v).UnmarshalText(text); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// strconvParseInt returns strconv.ParseInt in base 10 at bitSize, the
// baseline of the signed integer parsers.
func strconvParseInt(bitSize int) func(string) (int64, error) {
	return func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, bitSize)
	}
}

// strconvParseIntBytes returns strconv.ParseInt in base 10 at bitSize
// over the bytes of a text, converted to a string, the baseline of
// UnmarshalText on the signed integers.
func strconvParseIntBytes(bitSize int) func([]byte) (int64, error) {
	return func(text []byte) (int64, error) {
		return strconv.ParseInt(string(text), 10, bitSize)
	}
}

// strconvParseUint returns strconv.ParseUint in base 10 at bitSize, the
// baseline of the unsigned integer parser.
func strconvParseUint(bitSize int) func(string) (uint64, error) {
	return func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, bitSize)
	}
}

// strconvParseUintBytes returns strconv.ParseUint in base 10 at bitSize
// over the bytes of a text, converted to a string, the baseline of
// UnmarshalText on the unsigned integer.
func strconvParseUintBytes(bitSize int) func([]byte) (uint64, error) {
	return func(text []byte) (uint64, error) {
		return strconv.ParseUint(string(text), 10, bitSize)
	}
}
