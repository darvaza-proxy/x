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
// UnmarshalText over strconv would. UnmarshalJSON is measured under
// impl=num alone, on the same digits as a JSON number and as a JSON
// string, against the UnmarshalText row beside it. The natives and the
// fixed-point types are measured on their bounds as well, the 128-bit
// integers on their bounds other than zero, which strconv does not
// read, and the fixed-point types beside ParseFloat on the same text.
func BenchmarkParse(b *testing.B) {
	benchmarkParseInt32(b)
	benchmarkParseInt64(b)
	benchmarkParseUint128(b)
	benchmarkParseInt128(b)
	benchmarkParseMilli32(b)
	benchmarkParseMilli64(b)
	benchmarkParseAtto128(b)
}

// benchmarkParseInt32 runs the Int32 rows of BenchmarkParse.
func benchmarkParseInt32(b *testing.B) {
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
	b.Run("Int32.UnmarshalJSON.number/impl=num",
		runBenchmarkUnmarshalJSON[num.Int32]([]byte("-1234567")))
	b.Run("Int32.UnmarshalJSON.string/impl=num",
		runBenchmarkUnmarshalJSON[num.Int32]([]byte(`"-1234567"`)))
}

// benchmarkParseInt64 runs the Int64 rows of BenchmarkParse.
func benchmarkParseInt64(b *testing.B) {
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
	b.Run("Int64.UnmarshalJSON.number/impl=num",
		runBenchmarkUnmarshalJSON[num.Int64]([]byte("-1234567")))
	b.Run("Int64.UnmarshalJSON.string/impl=num",
		runBenchmarkUnmarshalJSON[num.Int64]([]byte(`"-1234567"`)))
}

// benchmarkParseUint128 runs the Uint128 rows of BenchmarkParse.
func benchmarkParseUint128(b *testing.B) {
	b.Run("Uint128/impl=num", runBenchmarkParse(num.ParseUint128, "1234567"))
	b.Run("Uint128/impl=strconv", runBenchmarkParse(strconvParseUint(64),
		"1234567"))
	b.Run("Uint128.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Uint128]([]byte("1234567")))
	b.Run("Uint128.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseUintBytes(64), []byte("1234567")))
	b.Run("Uint128.UnmarshalJSON.number/impl=num",
		runBenchmarkUnmarshalJSON[num.Uint128]([]byte("1234567")))
	b.Run("Uint128.UnmarshalJSON.string/impl=num",
		runBenchmarkUnmarshalJSON[num.Uint128]([]byte(`"1234567"`)))
	b.Run("Uint128Max/impl=num", runBenchmarkParse(num.ParseUint128,
		maxUint128Text))
}

// benchmarkParseInt128 runs the Int128 rows of BenchmarkParse.
func benchmarkParseInt128(b *testing.B) {
	b.Run("Int128/impl=num", runBenchmarkParse(num.ParseInt128, "-1234567"))
	b.Run("Int128/impl=strconv", runBenchmarkParse(strconvParseInt(64),
		"-1234567"))
	b.Run("Int128.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Int128]([]byte("-1234567")))
	b.Run("Int128.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseIntBytes(64), []byte("-1234567")))
	b.Run("Int128.UnmarshalJSON.number/impl=num",
		runBenchmarkUnmarshalJSON[num.Int128]([]byte("-1234567")))
	b.Run("Int128.UnmarshalJSON.string/impl=num",
		runBenchmarkUnmarshalJSON[num.Int128]([]byte(`"-1234567"`)))
	b.Run("Int128Min/impl=num", runBenchmarkParse(num.ParseInt128,
		minInt128Text))
	b.Run("Int128Max/impl=num", runBenchmarkParse(num.ParseInt128,
		maxInt128Text))
}

// benchmarkParseMilli32 runs the Milli32 rows of BenchmarkParse.
func benchmarkParseMilli32(b *testing.B) {
	b.Run("Milli32/impl=num", runBenchmarkParse(num.ParseMilli32, "1234.567"))
	b.Run("Milli32/impl=strconv", runBenchmarkParse(strconvParseFloat(32),
		"1234.567"))
	b.Run("Milli32Min/impl=num", runBenchmarkParse(num.ParseMilli32,
		"-2147483.648"))
	b.Run("Milli32Min/impl=strconv", runBenchmarkParse(strconvParseFloat(32),
		"-2147483.648"))
	b.Run("Milli32Max/impl=num", runBenchmarkParse(num.ParseMilli32,
		"2147483.647"))
	b.Run("Milli32Max/impl=strconv", runBenchmarkParse(strconvParseFloat(32),
		"2147483.647"))
	b.Run("Milli32.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Milli32]([]byte("1234.567")))
	b.Run("Milli32.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseFloatBytes(32), []byte("1234.567")))
}

// benchmarkParseMilli64 runs the Milli64 rows of BenchmarkParse.
func benchmarkParseMilli64(b *testing.B) {
	b.Run("Milli64/impl=num", runBenchmarkParse(num.ParseMilli64, "1234.567"))
	b.Run("Milli64/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"1234.567"))
	b.Run("Milli64Min/impl=num", runBenchmarkParse(num.ParseMilli64,
		"-9223372036854775.808"))
	b.Run("Milli64Min/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"-9223372036854775.808"))
	b.Run("Milli64Max/impl=num", runBenchmarkParse(num.ParseMilli64,
		"9223372036854775.807"))
	b.Run("Milli64Max/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"9223372036854775.807"))
	b.Run("Milli64.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Milli64]([]byte("1234.567")))
	b.Run("Milli64.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseFloatBytes(64), []byte("1234.567")))
}

// benchmarkParseAtto128 runs the Atto128 rows of BenchmarkParse.
func benchmarkParseAtto128(b *testing.B) {
	b.Run("Atto128/impl=num", runBenchmarkParse(num.ParseAtto128,
		"1234.567890123456789012"))
	b.Run("Atto128/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"1234.567890123456789012"))
	b.Run("Atto128Min/impl=num", runBenchmarkParse(num.ParseAtto128,
		"-170141183460469231731.687303715884105728"))
	b.Run("Atto128Min/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"-170141183460469231731.687303715884105728"))
	b.Run("Atto128Max/impl=num", runBenchmarkParse(num.ParseAtto128,
		"170141183460469231731.687303715884105727"))
	b.Run("Atto128Max/impl=strconv", runBenchmarkParse(strconvParseFloat(64),
		"170141183460469231731.687303715884105727"))
	b.Run("Atto128.UnmarshalText/impl=num",
		runBenchmarkUnmarshalText[num.Atto128](
			[]byte("1234.567890123456789012")))
	b.Run("Atto128.UnmarshalText/impl=strconv",
		runBenchmarkParse(strconvParseFloatBytes(64),
			[]byte("1234.567890123456789012")))
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
// over text, which must parse. The receiver is declared outside the
// loop: the call through PT moves it to the heap, which inside the
// loop would show as an allocation on every call.
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

// runBenchmarkUnmarshalJSON returns a benchmark of UnmarshalJSON on *T
// over data, which must read. The receiver is declared once, outside
// the loop, as for runBenchmarkUnmarshalText.
func runBenchmarkUnmarshalJSON[T any, PT jsonUnmarshaler[T]](
	data []byte) func(*testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		var v T
		for b.Loop() {
			if err := PT(&v).UnmarshalJSON(data); err != nil {
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
// baseline of the unsigned integer parsers.
func strconvParseUint(bitSize int) func(string) (uint64, error) {
	return func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, bitSize)
	}
}

// strconvParseUintBytes returns strconv.ParseUint in base 10 at bitSize
// over the bytes of a text, converted to a string, the baseline of
// UnmarshalText on the unsigned integers.
func strconvParseUintBytes(bitSize int) func([]byte) (uint64, error) {
	return func(text []byte) (uint64, error) {
		return strconv.ParseUint(string(text), 10, bitSize)
	}
}

// strconvParseFloat returns strconv.ParseFloat at bitSize, the baseline
// of the fixed-point parsers.
func strconvParseFloat(bitSize int) func(string) (float64, error) {
	return func(s string) (float64, error) {
		return strconv.ParseFloat(s, bitSize)
	}
}

// strconvParseFloatBytes returns strconv.ParseFloat at bitSize over the
// bytes of a text, converted to a string, the baseline of UnmarshalText
// on the fixed-point types.
func strconvParseFloatBytes(bitSize int) func([]byte) (float64, error) {
	return func(text []byte) (float64, error) {
		return strconv.ParseFloat(string(text), bitSize)
	}
}
