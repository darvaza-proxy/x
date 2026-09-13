package num_test

// cspell:words femto

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

var (
	_ core.TestCase = badVerbCrossCase{}
	_ core.TestCase = decimalCrossCase{}
	_ core.TestCase = fmtBadVerbCase{}
	_ core.TestCase = formatCase{}
	_ core.TestCase = formatCrossCase{}
	_ core.TestCase = goStringCase{}
	_ core.TestCase = goSyntaxCrossCase{}
)

// formatCase pins the text one format string produces for one value,
// through fmt, so the flags, width and precision handling of Format is
// exercised the way a caller reaches it.
type formatCase struct {
	in     any
	format string
	want   string
	name   string
}

func newFormatCase(name string, in any, format, want string) formatCase {
	return formatCase{name: name, in: in, format: format, want: want}
}

func (tc formatCase) Name() string { return tc.name }

func (tc formatCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, fmt.Sprintf(tc.format, tc.in), "text")
}

func formatUint128Cases() []formatCase {
	return []formatCase{
		newFormatCase("zero", num.ZeroUint128, "%d", "0"),
		newFormatCase("low word", u(42), "%d", "42"),
		newFormatCase("v", u(42), "%v", "42"),
		newFormatCase("s", u(42), "%s", "42"),
		// the chunked path: 2^64 straddles the 10^19 chunk, the maximum
		// fills all three chunks.
		newFormatCase("two to the 64", num.NewUint128(1, 0), "%d",
			"18446744073709551616"),
		newFormatCase("max", num.MaxUint128, "%d",
			"340282366920938463463374607431768211455"),
		// 2*10^19 is one chunk of 2 over a chunk of zeros, which must keep
		// its width.
		newFormatCase("zero chunk", u(1e19).Add(u(1e19)), "%d",
			"20000000000000000000"),
		newFormatCase("hex", num.NewUint128(1, 0), "%x", "10000000000000000"),
		newFormatCase("hex upper", u(255), "%X", "FF"),
		newFormatCase("hex prefixed", u(255), "%#x", "0xff"),
		newFormatCase("hex upper prefixed", u(255), "%#X", "0XFF"),
		newFormatCase("octal", u(8), "%o", "10"),
		newFormatCase("octal prefixed", u(8), "%#o", "010"),
		// fmt adds no octal prefix to a leading zero.
		newFormatCase("octal prefixed zero", num.ZeroUint128, "%#o", "0"),
		newFormatCase("octal 0o", u(8), "%O", "0o10"),
		// 128 bits are 42 octal digits and a leading 3.
		newFormatCase("octal max", num.MaxUint128, "%o",
			"3777777777777777777777777777777777777777777"),
		newFormatCase("binary", u(5), "%b", "101"),
		newFormatCase("binary prefixed", u(5), "%#b", "0b101"),
		newFormatCase("binary high word", num.NewUint128(1, 0), "%b",
			"1"+strings.Repeat("0", 64)),
		newFormatCase("width", u(42), "%10d", "        42"),
		newFormatCase("width left", u(42), "%-10d|", "42        |"),
		newFormatCase("width zero", u(42), "%010d", "0000000042"),
		newFormatCase("plus", u(42), "%+d", "+42"),
		newFormatCase("space", u(42), "% d", " 42"),
		newFormatCase("precision", u(42), "%.5d", "00042"),
		// a precision switches the zero flag off, as fmt does.
		newFormatCase("precision and zero width", u(42), "%08.5d", "   00042"),
		// a zero under precision zero prints as padding alone, as fmt does.
		newFormatCase("precision zero of zero", num.ZeroUint128, "%.0d", ""),
		newFormatCase("precision zero of zero width", num.ZeroUint128,
			"%+3.0d|", "   |"),
		newFormatCase("precision zero of one", u(1), "%.0d", "1"),
		// the zero flag is a precision on the digits, so the prefix goes
		// outside the width and the result is two characters longer.
		newFormatCase("hex zero width prefixed", u(255), "%#010x",
			"0x00000000ff"),
		// '#' has no prefix to add in decimal.
		newFormatCase("decimal prefixed", u(42), "%#d", "42"),
		newFormatCase("go syntax", num.NewUint128(1, 0), "%#v",
			"num.NewUint128(0x1, 0x0)"),
		newFormatCase("bad verb", u(42), "%q", "%!q(num.Uint128=42)"),
	}
}

func formatInt128Cases() []formatCase {
	return []formatCase{
		newFormatCase("zero", num.ZeroInt128, "%d", "0"),
		newFormatCase("positive", num.AsInt128(42), "%d", "42"),
		newFormatCase("negative", num.AsInt128(-42), "%d", "-42"),
		newFormatCase("v", num.AsInt128(-42), "%v", "-42"),
		newFormatCase("s", num.AsInt128(-42), "%s", "-42"),
		newFormatCase("max", num.MaxInt128, "%d",
			"170141183460469231731687303715884105727"),
		// the minimum has no positive counterpart; its magnitude is taken
		// as the unsigned 2^127.
		newFormatCase("min", num.MinInt128, "%d",
			"-170141183460469231731687303715884105728"),
		// the sign precedes the magnitude in every base, as fmt does.
		newFormatCase("hex negative", num.AsInt128(-255), "%x", "-ff"),
		newFormatCase("hex negative prefixed", num.AsInt128(-255), "%#X",
			"-0XFF"),
		newFormatCase("min hex", num.MinInt128, "%x",
			"-8"+strings.Repeat("0", 31)),
		newFormatCase("binary negative", num.AsInt128(-5), "%b", "-101"),
		newFormatCase("plus", num.AsInt128(42), "%+d", "+42"),
		newFormatCase("plus negative", num.AsInt128(-42), "%+d", "-42"),
		newFormatCase("space", num.AsInt128(42), "% d", " 42"),
		newFormatCase("width zero negative", num.AsInt128(-42), "%08d",
			"-0000042"),
		newFormatCase("width left negative", num.AsInt128(-42), "%-8d|",
			"-42     |"),
		newFormatCase("precision negative", num.AsInt128(-42), "%.5d",
			"-00042"),
		newFormatCase("precision zero of zero", num.ZeroInt128, "%.0d", ""),
		newFormatCase("go syntax", num.AsInt128(-42), "%#v",
			"num.AsInt128(-42)"),
		newFormatCase("bad verb", num.AsInt128(-42), "%q",
			"%!q(num.Int128=-42)"),
	}
}

// The native integers hand their verbs to fmt, so the rows pin the
// routing rather than fmt's own arithmetic: the decimal aliases, the
// %#v form and the bad verb.
func formatNativeCases() []formatCase {
	return []formatCase{
		newFormatCase("int32", num.AsInt32(-42), "%d", "-42"),
		newFormatCase("int32 v", num.AsInt32(-42), "%v", "-42"),
		newFormatCase("int32 s", num.AsInt32(-42), "%s", "-42"),
		newFormatCase("int32 hex", num.AsInt32(-255), "%#x", "-0xff"),
		newFormatCase("int32 octal", num.AsInt32(8), "%O", "0o10"),
		newFormatCase("int32 binary width", num.AsInt32(5), "%08b", "00000101"),
		newFormatCase("int32 plus precision", num.AsInt32(42), "%+.5d", "+00042"),
		// the two fmt corners the 128-bit writer mirrors.
		newFormatCase("int32 precision zero of zero", num.AsInt32(0), "%.0d", ""),
		newFormatCase("int32 octal prefixed zero", num.AsInt32(0), "%#o", "0"),
		newFormatCase("int32 min", num.AsInt32(math.MinInt32), "%d",
			"-2147483648"),
		newFormatCase("int32 go syntax", num.AsInt32(-42), "%#v",
			"num.AsInt32(-42)"),
		newFormatCase("int32 bad verb", num.AsInt32(-42), "%q",
			"%!q(num.Int32=-42)"),
		newFormatCase("int64", num.AsInt64(-42), "%d", "-42"),
		newFormatCase("int64 v", num.AsInt64(-42), "%v", "-42"),
		newFormatCase("int64 s", num.AsInt64(-42), "%s", "-42"),
		newFormatCase("int64 hex upper", num.AsInt64(255), "%X", "FF"),
		newFormatCase("int64 space width", num.AsInt64(42), "% 6d", "    42"),
		newFormatCase("int64 min", num.AsInt64(math.MinInt64), "%d",
			"-9223372036854775808"),
		newFormatCase("int64 go syntax", num.AsInt64(-42), "%#v",
			"num.AsInt64(-42)"),
		newFormatCase("int64 bad verb", num.AsInt64(-42), "%q",
			"%!q(num.Int64=-42)"),
	}
}

func formatMilliCases() []formatCase {
	return []formatCase{
		newFormatCase("zero", num.NewMilli32(0, 0), "%v", "0.000"),
		newFormatCase("v", num.NewMilli32(1, 500), "%v", "1.500"),
		newFormatCase("s", num.NewMilli32(1, 500), "%s", "1.500"),
		newFormatCase("negative", num.NewMilli32(-1, -500), "%v", "-1.500"),
		newFormatCase("negative fraction", num.NewMilli32(0, -5), "%v",
			"-0.005"),
		// v and s print every digit the resolution holds and nothing
		// more, whatever precision is asked for.
		newFormatCase("v precision", num.NewMilli32(1, 500), "%.1v", "1.500"),
		newFormatCase("s precision", num.NewMilli32(1, 500), "%8.5s",
			"   1.500"),
		// the minimum backing: Abs would wrap, the parts do not.
		newFormatCase("min", num.AsMilli32(math.MinInt32), "%v",
			"-2147483.648"),
		// f follows fmt: six digits without a precision, zero-filled past
		// the resolution.
		newFormatCase("f", num.NewMilli32(1, 500), "%f", "1.500000"),
		newFormatCase("f wide", num.NewMilli32(1, 500), "%.5f", "1.50000"),
		newFormatCase("f exact", num.NewMilli32(1, 500), "%.3f", "1.500"),
		// below the resolution it rounds half away from zero.
		newFormatCase("f round down", num.NewMilli32(1, 234), "%.2f", "1.23"),
		newFormatCase("f round up", num.NewMilli32(1, 235), "%.2f", "1.24"),
		newFormatCase("f round negative", num.NewMilli32(-1, -235), "%.2f",
			"-1.24"),
		newFormatCase("f half up", num.NewMilli32(2, 500), "%.0f", "3"),
		newFormatCase("f half up negative", num.NewMilli32(-2, -500), "%.0f",
			"-3"),
		// a carry out of the fraction reaches the whole count.
		newFormatCase("f carry", num.NewMilli32(1, 995), "%.2f", "2.00"),
		newFormatCase("f carry from zero", num.NewMilli32(0, 999), "%.0f", "1"),
		newFormatCase("width", num.NewMilli32(1, 500), "%8.1f", "     1.5"),
		newFormatCase("width zero negative", num.NewMilli32(-1, -500), "%08.1f",
			"-00001.5"),
		newFormatCase("width left", num.NewMilli32(1, 500), "%-8.1f|",
			"1.5     |"),
		// the '+' of %+v asks fmt for struct field names, so it signs
		// nothing; everywhere else it is a sign.
		newFormatCase("plus", num.NewMilli32(1, 500), "%+f", "+1.500000"),
		newFormatCase("plus v", num.NewMilli32(1, 500), "%+v", "1.500"),
		newFormatCase("space", num.NewMilli32(1, 500), "% v", " 1.500"),
		newFormatCase("go syntax", num.NewMilli32(1, 500), "%#v",
			"num.NewMilli32(1, 500)"),
		newFormatCase("bad verb", num.NewMilli32(1, 500), "%d",
			"%!d(num.Milli32=1.500)"),
		newFormatCase("bad verb negative", num.NewMilli32(-1, -500), "%d",
			"%!d(num.Milli32=-1.500)"),
		newFormatCase("milli64", num.NewMilli64(86400, 5), "%v", "86400.005"),
		newFormatCase("milli64 min", num.AsMilli64(math.MinInt64), "%v",
			"-9223372036854775.808"),
		newFormatCase("milli64 bad verb", num.NewMilli64(1, 500), "%x",
			"%!x(num.Milli64=1.500)"),
	}
}

func formatAtto128Cases() []formatCase {
	return []formatCase{
		newFormatCase("v", num.NewAtto128(1, 500e15), "%v",
			"1.500000000000000000"),
		newFormatCase("one atto", num.NewAtto128(0, 1), "%v",
			"0.000000000000000001"),
		newFormatCase("negative", num.NewAtto128(-1, -500e15), "%s",
			"-1.500000000000000000"),
		newFormatCase("f", num.NewAtto128(1, 500e15), "%f", "1.500000"),
		newFormatCase("f exact", num.NewAtto128(1, 500e15), "%.18f",
			"1.500000000000000000"),
		newFormatCase("f wide", num.NewAtto128(1, 500e15), "%.20f",
			"1.50000000000000000000"),
		newFormatCase("f round", num.NewAtto128(0, 123_456_789_012_345_678),
			"%.9f", "0.123456789"),
		newFormatCase("f half up", num.NewAtto128(0, 500e15), "%.0f", "1"),
		// '#' keeps the point a zero precision would drop.
		newFormatCase("f sharp", num.NewAtto128(1, 0), "%#.0f", "1."),
		// the whole count past int64 and the minimum backing both print
		// through the backing integer.
		newFormatCase("whole past int64",
			num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0)), "%.0f",
			"9223372036854775808"),
		newFormatCase("min", num.AsAtto128(num.MinInt128), "%v",
			"-170141183460469231731.687303715884105728"),
		newFormatCase("go syntax", num.NewAtto128(1, 500e15), "%#v",
			"num.NewAtto128(1, 500_000_000_000_000_000)"),
		newFormatCase("bad verb", num.NewAtto128(1, 500e15), "%d",
			"%!d(num.Atto128=1.500000000000000000)"),
	}
}

// formatCrossCase asserts every type prints exactly as fmt prints the
// native integer of the same value, so the rows above enforce fmt's
// behaviour rather than transcribe my reading of it. The expectation
// comes from an oracle format, which is the format under test except
// for the s verb, where fmt is asked for the d it prints as.
type formatCrossCase struct {
	format string
	oracle string
	name   string
	value  int64
}

// newFormatCrossCase builds a row fmt answers under the very format
// being tested.
func newFormatCrossCase(format string, value int64) formatCrossCase {
	return newFormatCrossCaseAs(format, format, value)
}

// newFormatCrossCaseAs builds a row whose expectation comes from
// another format, for a verb fmt does not take on an integer.
func newFormatCrossCaseAs(format, oracle string, value int64) formatCrossCase {
	return formatCrossCase{
		name:   fmt.Sprintf("%s of %d", format, value),
		format: format,
		oracle: oracle,
		value:  value,
	}
}

func (tc formatCrossCase) Name() string { return tc.name }

func (tc formatCrossCase) Test(t *testing.T) {
	t.Helper()
	tc.testWide(t)
	tc.testNarrow(t)
	tc.testUnsigned(t)
}

// testWide compares the signed types holding every row against the
// native int64 of the same value.
func (tc formatCrossCase) testWide(t *testing.T) {
	t.Helper()
	want := fmt.Sprintf(tc.oracle, tc.value)
	core.AssertEqual(t, want, fmt.Sprintf(tc.format, num.AsInt64(tc.value)),
		"int64")
	core.AssertEqual(t, want, fmt.Sprintf(tc.format, num.AsInt128(tc.value)),
		"int128")
}

// testNarrow compares Int32 against the native int32, for the rows
// whose value it can hold.
func (tc formatCrossCase) testNarrow(t *testing.T) {
	t.Helper()
	if tc.value < math.MinInt32 || tc.value > math.MaxInt32 {
		return
	}
	v := int32(tc.value)
	core.AssertEqual(t, fmt.Sprintf(tc.oracle, v),
		fmt.Sprintf(tc.format, num.AsInt32(v)), "int32")
}

// testUnsigned compares Uint128 against the native uint64, for the rows
// whose value it can hold.
func (tc formatCrossCase) testUnsigned(t *testing.T) {
	t.Helper()
	if tc.value < 0 {
		return
	}
	v := uint64(tc.value)
	core.AssertEqual(t, fmt.Sprintf(tc.oracle, v),
		fmt.Sprintf(tc.format, num.AsUint128(v)), "uint128")
}

// crossValues are the values the matrix runs: the corners the flags
// and precision reach, and both bounds of every width being compared.
func crossValues() []int64 {
	return core.S[int64](
		0, 1, 7, 8, 42, 255, math.MaxInt32, math.MaxInt64,
		-1, -8, -42, -255, math.MinInt32, math.MinInt64,
	)
}

// crossFormats are the formats fmt answers directly.
func crossFormats() []string {
	return core.S(
		"%d", "%v", "%x", "%X", "%o", "%O", "%b",
		"%+v", "% v", "%06v", "%-6v|", "%.3v", "%8.3v",
		"%#x", "%#X", "%#o", "%#b", "%#d", "%#O",
		"%+d", "% d", "% x", "% +d", "%-6d|", "%-06d|", "%06d", "%6d",
		"%.0d", "%.3d", "%.0x", "%.1o", "%+.0d", "%.20d", "%030d",
		"%#.1o", "%#.3o", "%#.0o", "%#06o", "%#08x", "%#014O",
		"%+8.3d", "%08.3d", "%-8.3d|", "% .3d", "%08.0d", "% .0d",
		"%+#12.4x", "%-#12.4X|", "% #12.4o", "%#12.4b",
	)
}

// crossFlagSets are the flags, widths and precisions the s verb is run
// under. fmt takes no s on an integer, so d answers for it.
func crossFlagSets() []string {
	return core.S("", "+", " ", "#", "8", "-8", "08", ".5", "+.5",
		"+#12.4", "-#12.4")
}

func formatCrossCases() []formatCrossCase {
	return append(verbCrossCases(), stringVerbCrossCases()...)
}

// verbCrossCases runs every format fmt answers directly over every
// value.
func verbCrossCases() []formatCrossCase {
	formats, values := crossFormats(), crossValues()
	cases := make([]formatCrossCase, 0, len(formats)*len(values))
	for _, f := range formats {
		for _, v := range values {
			cases = append(cases, newFormatCrossCase(f, v))
		}
	}
	return cases
}

// stringVerbCrossCases runs the s verb over every value, against the d
// of the same flags.
func stringVerbCrossCases() []formatCrossCase {
	flags, values := crossFlagSets(), crossValues()
	cases := make([]formatCrossCase, 0, len(flags)*len(values))
	for _, f := range flags {
		for _, v := range values {
			cases = append(cases,
				newFormatCrossCaseAs("%"+f+"s", "%"+f+"d", v))
		}
	}
	return cases
}

// decimalCrossCase asserts a Decimal pads, signs and fills as fmt does
// the float64 of the same value, on all three instantiations. Every
// row's value is exact in a float64, so fmt's digits are the value's
// and not an approximation of it, and the formats ask for at least the
// three fraction digits the rows carry, or for none of a whole value,
// so the comparison never reaches the rounding fmt and this package
// deliberately differ on.
type decimalCrossCase struct {
	format string
	name   string
	milli  int32
}

func newDecimalCrossCase(format string, milli int32) decimalCrossCase {
	return decimalCrossCase{
		name:   fmt.Sprintf("%s of %d milli", format, milli),
		format: format,
		milli:  milli,
	}
}

func (tc decimalCrossCase) Name() string { return tc.name }

func (tc decimalCrossCase) Test(t *testing.T) {
	t.Helper()
	want := fmt.Sprintf(tc.format, float64(tc.milli)/1000)
	core.AssertEqual(t, want,
		fmt.Sprintf(tc.format, num.AsMilli32(num.AsInt32(tc.milli))), "milli32")
	core.AssertEqual(t, want,
		fmt.Sprintf(tc.format, num.AsMilli64(num.AsInt64(int64(tc.milli)))), "milli64")
	core.AssertEqual(t, want, fmt.Sprintf(tc.format, attoOf(tc.milli)),
		"atto128")
}

// attoOf returns m milli-units as an Atto128, the same value at the
// finer resolution, the whole count and the fraction both carrying its
// sign.
func attoOf(m int32) num.Atto128 {
	return num.NewAtto128(int64(m/1000), int64(m%1000)*1e15)
}

// decimalCrossFormats are the formats every row is run under, all of
// them asking for the three fraction digits a milli value carries, or
// more.
func decimalCrossFormats() []string {
	return core.S(
		"%f", "%.3f", "%.6f", "%.18f", "%#.3f",
		"%12.3f", "%-12.3f|", "%012.3f", "%20.6f", "%-6.3f|",
		"%+.3f", "% .3f", "%+012.3f", "%-+13.3f|", "% 012.3f",
		// F is f under another name.
		"%F", "%.3F", "%+012.3F", "%-12.4F|",
	)
}

// decimalWholeFormats ask for a zero precision, which only a whole
// value can be asked for without meeting the rounding divergence.
func decimalWholeFormats() []string {
	return core.S("%.0f", "%+.0f", "% .0f", "%08.0f", "%-8.0f|",
		"%#.0f", "%#08.0f", "%.0F", "%#.0F")
}

func decimalCrossCases() []decimalCrossCase {
	values := core.S[int32](0, 500, 1500, -500, -1500, 125, -125, 42000)
	cases := decimalCrossRows(decimalCrossFormats(), values)
	whole := core.S[int32](0, 42000, -42000)
	return append(cases, decimalCrossRows(decimalWholeFormats(), whole)...)
}

func decimalCrossRows(formats []string, values []int32) []decimalCrossCase {
	cases := make([]decimalCrossCase, 0, len(formats)*len(values))
	for _, f := range formats {
		for _, v := range values {
			cases = append(cases, newDecimalCrossCase(f, v))
		}
	}
	return cases
}

// goSyntaxRef carries a GoString form as a plain [fmt.GoStringer],
// which fmt renders through its own %#v path: the oracle for the width
// and precision a Format method must apply to that text, since fmt
// never takes that path for a [fmt.Formatter].
type goSyntaxRef string

// GoString returns the text verbatim.
func (r goSyntaxRef) GoString() string { return string(r) }

// goSyntaxCrossCase asserts the %#v form of a value pads and truncates
// as fmt pads and truncates the same text behind a bare GoStringer.
type goSyntaxCrossCase struct {
	in     any
	format string
	name   string
}

func newGoSyntaxCrossCase(format string, in any) goSyntaxCrossCase {
	return goSyntaxCrossCase{
		name:   fmt.Sprintf("%s of %T", format, in),
		format: format,
		in:     in,
	}
}

func (tc goSyntaxCrossCase) Name() string { return tc.name }

func (tc goSyntaxCrossCase) Test(t *testing.T) {
	t.Helper()
	gs := core.AssertMustTypeIs[fmt.GoStringer](t, tc.in, "go stringer")
	want := fmt.Sprintf(tc.format, goSyntaxRef(gs.GoString()))
	core.AssertEqual(t, want, fmt.Sprintf(tc.format, tc.in), "text")
}

// crossTypes is one value of every type, for the rows that vary the
// format rather than the value.
func crossTypes() []any {
	return core.S[any](
		num.AsInt32(-42), num.AsInt64(-42), num.AsInt128(-42),
		num.MaxUint128, num.NewMilli32(1, 500), num.NewAtto128(1, 500e15),
	)
}

func goSyntaxCrossCases() []goSyntaxCrossCase {
	formats := core.S("%#v", "%#24v", "%-#24v|", "%#024v",
		"%#.6v", "%#12.6v", "%-#12.6v|")
	values := crossTypes()
	cases := make([]goSyntaxCrossCase, 0, len(formats)*len(values))
	for _, f := range formats {
		for _, v := range values {
			cases = append(cases, newGoSyntaxCrossCase(f, v))
		}
	}
	return cases
}

// badVerbSubject is one value with the native that answers for it:
// the type name the bad-verb form carries, and the verb giving the
// type its own full form, d for an integer and f at the resolution for
// a Decimal.
type badVerbSubject struct {
	in       any
	native   any
	typeName string
	verb     string
}

func newBadVerbSubject(typeName string, in, native any,
	verb string) badVerbSubject {
	return badVerbSubject{
		typeName: typeName,
		in:       in,
		native:   native,
		verb:     verb,
	}
}

// badVerbCrossCase asserts the %!verb(type=value) form renders its
// value under the flags and width, as fmt's own does by printing the
// argument again, and pads nothing around it. The flags reach that
// print unmunged, since '+' and '#' take the meanings v gives them
// only under v itself, so the oracle reads them under d or f rather
// than v.
type badVerbCrossCase struct {
	subject badVerbSubject
	flags   string
	name    string
}

func newBadVerbCrossCase(flags string, s badVerbSubject) badVerbCrossCase {
	return badVerbCrossCase{
		name:    fmt.Sprintf("%%%st of %s", flags, s.typeName),
		flags:   flags,
		subject: s,
	}
}

func (tc badVerbCrossCase) Name() string { return tc.name }

func (tc badVerbCrossCase) Test(t *testing.T) {
	t.Helper()
	s := tc.subject
	value := fmt.Sprintf("%"+tc.flags+s.verb, s.native)
	want := fmt.Sprintf("%%!t(%s=%s)", s.typeName, value)
	core.AssertEqual(t, want, fmt.Sprintf("%"+tc.flags+"t", s.in), "text")
}

// badVerbIntegerSubjects are the integers, each beside the native fmt
// renders in its place. Uint128 takes a value a uint64 can hold, since
// no native answers for a wider one.
func badVerbIntegerSubjects() []badVerbSubject {
	return core.S(
		newBadVerbSubject("num.Int32", num.AsInt32(-42), int64(-42), "d"),
		newBadVerbSubject("num.Int64", num.AsInt64(-42), int64(-42), "d"),
		newBadVerbSubject("num.Int128", num.AsInt128(-42), int64(-42), "d"),
		newBadVerbSubject("num.Uint128", num.AsUint128(42), uint64(42), "d"),
	)
}

// badVerbDecimalSubjects are the instantiations, against the float64
// of the same value at their own resolution.
func badVerbDecimalSubjects() []badVerbSubject {
	return core.S(
		newBadVerbSubject("num.Milli32", num.NewMilli32(-1, -500), -1.5,
			".3f"),
		newBadVerbSubject("num.Atto128", num.NewAtto128(1, 500e15), 1.5,
			".18f"),
	)
}

func badVerbCrossCases() []badVerbCrossCase {
	flags := core.S("", "12", "-12", "012", "+", " ", "#")
	subjects := append(badVerbIntegerSubjects(),
		badVerbDecimalSubjects()...)
	cases := badVerbRows(flags, subjects)
	// a precision is the digit count of an integer, which is what the
	// print inside a bad verb gives it. A Decimal prints at its
	// resolution whatever the precision, so no float oracle can be
	// asked for one.
	precision := core.S(".0", ".3", "+14.6")
	return append(cases,
		badVerbRows(precision, badVerbIntegerSubjects())...)
}

func badVerbRows(flags []string, subjects []badVerbSubject) []badVerbCrossCase {
	cases := make([]badVerbCrossCase, 0, len(flags)*len(subjects))
	for _, f := range flags {
		for _, s := range subjects {
			cases = append(cases, newBadVerbCrossCase(f, s))
		}
	}
	return cases
}

// fmtBadVerbCase pins the premise the bad-verb oracle rests on, against
// fmt alone: fmt prints the value inside its own %!verb(type=value)
// form under the flags and width the bad verb was given, so '+' signs
// it there although the reprint runs under v.
type fmtBadVerbCase struct {
	in     any
	format string
	want   string
	name   string
}

func newFmtBadVerbCase(format string, in any, want string) fmtBadVerbCase {
	return fmtBadVerbCase{
		name:   fmt.Sprintf("%s of %v", format, in),
		format: format,
		in:     in,
		want:   want,
	}
}

func (tc fmtBadVerbCase) Name() string { return tc.name }

func (tc fmtBadVerbCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, fmt.Sprintf(tc.format, tc.in), "text")
}

func fmtBadVerbCases() []fmtBadVerbCase {
	return core.S(
		newFmtBadVerbCase("%+t", 42, "%!t(int=+42)"),
		newFmtBadVerbCase("%+t", -42, "%!t(int=-42)"),
		newFmtBadVerbCase("%+t", 1.5, "%!t(float64=+1.5)"),
		newFmtBadVerbCase("% t", 42, "%!t(int= 42)"),
		newFmtBadVerbCase("%6t", 42, "%!t(int=    42)"),
		newFmtBadVerbCase("%-6t", 42, "%!t(int=42    )"),
	)
}

func TestFormatMatchesFmt(t *testing.T) {
	t.Run("integer", runTestIntegerMatchesFmt)
	t.Run("decimal", runTestDecimalMatchesFmt)
	t.Run("go syntax", runTestGoSyntaxMatchesFmt)
	t.Run("bad verb", runTestBadVerbMatchesFmt)
	t.Run("fmt bad verb", runTestFmtBadVerb)
}

func runTestFmtBadVerb(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, fmtBadVerbCases())
}

func runTestIntegerMatchesFmt(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, formatCrossCases())
}

func runTestDecimalMatchesFmt(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, decimalCrossCases())
}

func runTestGoSyntaxMatchesFmt(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, goSyntaxCrossCases())
}

func runTestBadVerbMatchesFmt(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, badVerbCrossCases())
}

func TestFormatUint128(t *testing.T) {
	core.RunTestCases(t, formatUint128Cases())
}

func TestFormatInt128(t *testing.T) {
	core.RunTestCases(t, formatInt128Cases())
}

func TestFormatNative(t *testing.T) {
	core.RunTestCases(t, formatNativeCases())
}

func TestFormatMilli(t *testing.T) {
	core.RunTestCases(t, formatMilliCases())
}

func TestFormatAtto128(t *testing.T) {
	core.RunTestCases(t, formatAtto128Cases())
}

// goStringCase pins the %#v form of a value. Each row builds its value
// with the constructor the output names, so the row shows the round
// trip, and the same text must come back through fmt.
type goStringCase struct {
	in   fmt.GoStringer
	want string
	name string
}

func newGoStringCase(name string, in fmt.GoStringer,
	want string) goStringCase {
	return goStringCase{name: name, in: in, want: want}
}

func (tc goStringCase) Name() string { return tc.name }

func (tc goStringCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.in.GoString(), "GoString")
	core.AssertEqual(t, tc.want, fmt.Sprintf("%#v", tc.in), "fmt")
}

func goStringIntCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("int32 zero", num.AsInt32(0), "num.AsInt32(0)"),
		newGoStringCase("int32 negative", num.AsInt32(-5), "num.AsInt32(-5)"),
		newGoStringCase("int32 min", num.AsInt32(math.MinInt32),
			"num.AsInt32(-2147483648)"),
		newGoStringCase("int64 positive", num.AsInt64(42), "num.AsInt64(42)"),
		newGoStringCase("int64 min", num.AsInt64(math.MinInt64),
			"num.AsInt64(-9223372036854775808)"),
	}
}

func goStringUint128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("uint128 zero", num.ZeroUint128, "num.AsUint128(0)"),
		newGoStringCase("uint128 low word", num.AsUint128(5),
			"num.AsUint128(5)"),
		newGoStringCase("uint128 max word", num.AsUint128(maxWord),
			"num.AsUint128(18446744073709551615)"),
		// the high word set switches to the words form, in hex.
		newGoStringCase("uint128 high word", num.NewUint128(1, 0),
			"num.NewUint128(0x1, 0x0)"),
		newGoStringCase("uint128 max", num.MaxUint128,
			"num.NewUint128(0xffffffffffffffff, 0xffffffffffffffff)"),
	}
}

func goStringInt128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("int128 zero", num.ZeroInt128, "num.AsInt128(0)"),
		newGoStringCase("int128 positive", num.AsInt128(42),
			"num.AsInt128(42)"),
		newGoStringCase("int128 negative", num.AsInt128(-42),
			"num.AsInt128(-42)"),
		// the int64 bounds still fit: the words are all ones for -1 and
		// the sign extension of the top bit for the minimum.
		newGoStringCase("int128 minus one", num.AsInt128(-1),
			"num.AsInt128(-1)"),
		newGoStringCase("int128 max int64", num.AsInt128(math.MaxInt64),
			"num.AsInt128(9223372036854775807)"),
		newGoStringCase("int128 min int64", num.AsInt128(math.MinInt64),
			"num.AsInt128(-9223372036854775808)"),
		// past int64 the words form takes over: 2^63 has the low sign bit
		// set with a zero high word, -2^64 the reverse.
		newGoStringCase("int128 two to the 63", num.NewInt128(0, signBit),
			"num.NewInt128(0x0, 0x8000000000000000)"),
		newGoStringCase("int128 minus two to the 64",
			num.NewInt128(maxWord, 0), "num.NewInt128(0xffffffffffffffff, 0x0)"),
		newGoStringCase("int128 min", num.MinInt128,
			"num.NewInt128(0x8000000000000000, 0x0)"),
		newGoStringCase("int128 max", num.MaxInt128,
			"num.NewInt128(0x7fffffffffffffff, 0xffffffffffffffff)"),
	}
}

func goStringMilliCases() []goStringCase {
	return []goStringCase{
		newGoStringCase("milli32 zero", num.NewMilli32(0, 0),
			"num.NewMilli32(0, 0)"),
		newGoStringCase("milli32 half", num.NewMilli32(1, 500),
			"num.NewMilli32(1, 500)"),
		// both parts carry the sign, so the call rebuilds the value under
		// either sign rule of the constructor.
		newGoStringCase("milli32 negative", num.NewMilli32(-1, -500),
			"num.NewMilli32(-1, -500)"),
		newGoStringCase("milli32 negative fraction", num.NewMilli32(0, -500),
			"num.NewMilli32(0, -500)"),
		newGoStringCase("milli32 as", num.AsMilli32(1500),
			"num.NewMilli32(1, 500)"),
		newGoStringCase("milli32 min", num.AsMilli32(math.MinInt32),
			"num.NewMilli32(-2147483, -648)"),
		newGoStringCase("milli64 day", num.NewMilli64(86400, 5),
			"num.NewMilli64(86400, 5)"),
		newGoStringCase("milli64 min", num.AsMilli64(math.MinInt64),
			"num.NewMilli64(-9223372036854775, -808)"),
	}
}

func goStringAtto128Cases() []goStringCase {
	return []goStringCase{
		newGoStringCase("atto128 zero", num.NewAtto128(0, 0),
			"num.NewAtto128(0, 0)"),
		// the fraction is grouped in thousands from four digits up.
		newGoStringCase("atto128 one atto", num.NewAtto128(0, 1),
			"num.NewAtto128(0, 1)"),
		newGoStringCase("atto128 four digits", num.NewAtto128(2, 1234),
			"num.NewAtto128(2, 1_234)"),
		newGoStringCase("atto128 femto", num.NewAtto128(0, 1e15),
			"num.NewAtto128(0, 1_000_000_000_000_000)"),
		newGoStringCase("atto128 half", num.NewAtto128(1, 500e15),
			"num.NewAtto128(1, 500_000_000_000_000_000)"),
		newGoStringCase("atto128 negative", num.NewAtto128(-1, -500e15),
			"num.NewAtto128(-1, -500_000_000_000_000_000)"),
		newGoStringCase("atto128 negative fraction",
			num.NewAtto128(0, -123e15),
			"num.NewAtto128(0, -123_000_000_000_000_000)"),
		newGoStringCase("atto128 max int64 whole",
			num.NewAtto128(math.MaxInt64, 0),
			"num.NewAtto128(9223372036854775807, 0)"),
		// a whole count past int64 falls back to the backing's words
		// form: 2^63 whole units is 10^18 shifted left by 63, and the
		// minimum backing is the most negative Int128.
		newGoStringCase("atto128 whole past int64",
			num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0)),
			"num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0x0))"),
		newGoStringCase("atto128 min", num.AsAtto128(num.MinInt128),
			"num.AsAtto128(num.NewInt128(0x8000000000000000, 0x0))"),
	}
}

// TestAtto128GoStringFallback pins the fallback row's value: the words
// above are 2^63 whole units, one past the largest int64 whole count.
func TestAtto128GoStringFallback(t *testing.T) {
	past := num.NewAtto128(math.MaxInt64, 0).Add(num.NewAtto128(1, 0))
	core.AssertEqual(t, num.AsAtto128(num.NewInt128(0x6f05b59d3b20000, 0)),
		past, "2^63 whole")
}

func TestGoString(t *testing.T) {
	core.RunTestCases(t, goStringIntCases())
	core.RunTestCases(t, goStringUint128Cases())
	core.RunTestCases(t, goStringInt128Cases())
	core.RunTestCases(t, goStringMilliCases())
	core.RunTestCases(t, goStringAtto128Cases())
}
