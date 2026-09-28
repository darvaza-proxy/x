package num_test

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

// outsideSigned is a signed integer defined outside the package, each
// method delegating to the Int64 it wraps. It wraps rather than
// embeds, so it inherits none of the package's unexported methods, and
// the assertion below pins that [num.SignedEuclidean] asks for none.
//
// It is none of the package's integers either, so [num.Decimal], with
// it as the backing of OutsideCenti, reaches it through its
// [num.Signed] methods alone.
type outsideSigned struct {
	v num.Int64
}

var _ num.SignedEuclidean[outsideSigned] = outsideSigned{}

func (o outsideSigned) IsZero() bool                      { return o.v.IsZero() }
func (o outsideSigned) Equal(v outsideSigned) bool        { return o.v.Equal(v.v) }
func (o outsideSigned) Cmp(v outsideSigned) int           { return o.v.Cmp(v.v) }
func (o outsideSigned) IsNegative() bool                  { return o.v.IsNegative() }
func (o outsideSigned) Neg() outsideSigned                { return outsideSigned{o.v.Neg()} }
func (o outsideSigned) Abs() outsideSigned                { return outsideSigned{o.v.Abs()} }
func (o outsideSigned) Add(v outsideSigned) outsideSigned { return outsideSigned{o.v.Add(v.v)} }
func (o outsideSigned) Sub(v outsideSigned) outsideSigned { return outsideSigned{o.v.Sub(v.v)} }
func (o outsideSigned) Mul(v outsideSigned) outsideSigned { return outsideSigned{o.v.Mul(v.v)} }
func (o outsideSigned) Div(v outsideSigned) outsideSigned { return outsideSigned{o.v.Div(v.v)} }
func (o outsideSigned) Mod(v outsideSigned) outsideSigned { return outsideSigned{o.v.Mod(v.v)} }
func (o outsideSigned) One() outsideSigned                { return outsideSigned{o.v.One()} }
func (o outsideSigned) ULP() outsideSigned                { return outsideSigned{o.v.ULP()} }

func (o outsideSigned) DivMod(v outsideSigned) (q, r outsideSigned) {
	vq, vr := o.v.DivMod(v.v)
	return outsideSigned{vq}, outsideSigned{vr}
}

func (o outsideSigned) MulDivMod(v, d outsideSigned) (q, r outsideSigned) {
	vq, vr := o.v.MulDivMod(v.v, d.v)
	return outsideSigned{vq}, outsideSigned{vr}
}

// asOutsideSigned builds an outsideSigned from a native count.
func asOutsideSigned(x int64) outsideSigned {
	return outsideSigned{num.AsInt64(x)}
}

// outsideInt is a type defined outside the package that meets
// [num.Number] on its own methods, each delegating to the Int64 it
// wraps. It wraps rather than embeds, so it lacks the Euclidean steps
// One and ULP, and the assertion below pins that Number asks for
// neither.
type outsideInt struct {
	v num.Int64
}

var _ num.Number[outsideInt] = outsideInt{}

func (o outsideInt) IsZero() bool                { return o.v.IsZero() }
func (o outsideInt) Equal(v outsideInt) bool     { return o.v.Equal(v.v) }
func (o outsideInt) Cmp(v outsideInt) int        { return o.v.Cmp(v.v) }
func (o outsideInt) IsNegative() bool            { return o.v.IsNegative() }
func (o outsideInt) Abs() outsideInt             { return outsideInt{o.v.Abs()} }
func (o outsideInt) Add(v outsideInt) outsideInt { return outsideInt{o.v.Add(v.v)} }
func (o outsideInt) Sub(v outsideInt) outsideInt { return outsideInt{o.v.Sub(v.v)} }
func (o outsideInt) Mul(v outsideInt) outsideInt { return outsideInt{o.v.Mul(v.v)} }
func (o outsideInt) Div(v outsideInt) outsideInt { return outsideInt{o.v.Div(v.v)} }
func (o outsideInt) Mod(v outsideInt) outsideInt { return outsideInt{o.v.Mod(v.v)} }

func (o outsideInt) DivMod(v outsideInt) (q, r outsideInt) {
	vq, vr := o.v.DivMod(v.v)
	return outsideInt{vq}, outsideInt{vr}
}

func (o outsideInt) MulDivMod(v, d outsideInt) (q, r outsideInt) {
	vq, vr := o.v.MulDivMod(v.v, d.v)
	return outsideInt{vq}, outsideInt{vr}
}

func (o outsideInt) AppendText(b []byte) ([]byte, error) { return o.v.AppendText(b) }
func (o outsideInt) MarshalText() ([]byte, error)        { return o.v.MarshalText() }
func (o outsideInt) MarshalJSON() ([]byte, error)        { return o.v.MarshalJSON() }

func (o outsideInt) Format(f fmt.State, verb rune) { o.v.Format(f, verb) }
func (o outsideInt) GoString() string              { return o.v.GoString() }
func (o outsideInt) String() string                { return o.v.String() }

func (o outsideInt) Int32() (num.Int32, bool)     { return o.v.Int32() }
func (o outsideInt) Int64() (num.Int64, bool)     { return o.v.Int64() }
func (o outsideInt) Int128() (num.Int128, bool)   { return o.v.Int128() }
func (o outsideInt) Uint128() (num.Uint128, bool) { return o.v.Uint128() }
func (o outsideInt) Milli32() (num.Milli32, bool) { return o.v.Milli32() }
func (o outsideInt) Milli64() (num.Milli64, bool) { return o.v.Milli64() }
func (o outsideInt) Atto128() (num.Atto128, bool) { return o.v.Atto128() }

// outsideCentiScale is a scaler defined outside the package, two
// fraction digits over the outside backing.
type outsideCentiScale struct{}

var _ num.DecimalScaler[outsideSigned] = outsideCentiScale{}

func (outsideCentiScale) Scale() outsideSigned { return asOutsideSigned(100) }
func (outsideCentiScale) Name() string         { return "num_test.OutsideCenti" }

// OutsideCenti is a signed fixed-point number with two fraction
// digits over a backing and a scaler both defined outside the package.
type OutsideCenti = num.Decimal[outsideSigned, outsideCentiScale]

// NewOutsideCenti builds an OutsideCenti from a whole-unit count and a
// centi-unit fraction; it is the constructor GoString names.
func NewOutsideCenti(whole, centi int64) OutsideCenti {
	return num.NewDecimal[outsideSigned, outsideCentiScale](
		asOutsideSigned(whole), asOutsideSigned(centi))
}

// AsOutsideCenti takes a count of centi-units, so AsOutsideCenti(150)
// is 1.5.
func AsOutsideCenti(centi int64) OutsideCenti {
	return num.AsDecimal[outsideSigned, outsideCentiScale](
		asOutsideSigned(centi))
}

// TestOutsideCenti runs the shared Decimal and Euclidean suites over
// the outside instantiation, and its text, GoString, Format, JSON,
// conversion, count and factory rows.
func TestOutsideCenti(t *testing.T) {
	t.Run("decimal", runTestOutsideCentiDecimal)
	t.Run("euclidean", runTestOutsideCentiEuclidean)
	t.Run("text", runTestOutsideCentiText)
	t.Run("go string", runTestOutsideCentiGoString)
	t.Run("format", runTestOutsideCentiFormat)
	t.Run("json", runTestOutsideCentiJSON)
	t.Run("convert", runTestOutsideCentiConvert)
	t.Run("count", runTestOutsideCentiCount)
	t.Run("from int128", runTestOutsideCentiFromInt128)
}

func runTestOutsideCentiDecimal(t *testing.T) {
	t.Helper()
	runDecimalTests(t, decimalType[OutsideCenti]{
		mk:    NewOutsideCenti,
		as:    AsOutsideCenti,
		scale: 100,
		// 1844674407370956e4 is the first multiple of 10^4 past 2^64.
		big:       1844674407370956,
		wrapWhole: 83,
		wrapFrac:  84,
	})
}

func runTestOutsideCentiEuclidean(t *testing.T) {
	t.Helper()
	runEuclideanDecimalTests(t, euclideanDecimalSuite[OutsideCenti]{
		mk:    NewOutsideCenti,
		scale: 100,
	})
}

func runTestOutsideCentiText(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []textCase{
		newTextCase("zero", NewOutsideCenti(0, 0), "0.00"),
		newTextCase("one and a half", NewOutsideCenti(1, 50), "1.50"),
		newTextCase("negative fraction", NewOutsideCenti(0, -5), "-0.05"),
		newTextCase("min", AsOutsideCenti(math.MinInt64),
			"-92233720368547758.08"),
		newTextCase("max", AsOutsideCenti(math.MaxInt64),
			"92233720368547758.07"),
	})
}

func runTestOutsideCentiGoString(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []goStringCase{
		newGoStringCase("half", NewOutsideCenti(1, 50),
			"num_test.NewOutsideCenti(1, 50)"),
		newGoStringCase("negative", NewOutsideCenti(-1, -5),
			"num_test.NewOutsideCenti(-1, -5)"),
		newGoStringCase("min", AsOutsideCenti(math.MinInt64),
			"num_test.NewOutsideCenti(-92233720368547758, -8)"),
	})
}

func runTestOutsideCentiFormat(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []formatCase{
		newFormatCase("f round", NewOutsideCenti(1, 55), "%.1f", "1.6"),
		newFormatCase("f negative round", NewOutsideCenti(-1, -55), "%.1f",
			"-1.6"),
		newFormatCase("f half up", NewOutsideCenti(0, 50), "%.0f", "1"),
	})
}

func runTestOutsideCentiJSON(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []jsonCase{
		newJSONNumberCase("half", NewOutsideCenti(1, 50), "1.50"),
		newJSONStringCase("min", AsOutsideCenti(math.MinInt64),
			"-92233720368547758.08"),
	})
}

// outsideCentiOf converts v to OutsideCenti, which no method of the
// package targets, for the way back of a conversion row: through
// Atto128, which holds every OutsideCenti, its count divided down to
// centi-units towards zero. It answers false when the value misses
// either.
func outsideCentiOf[S num.Number[S]](v S) (OutsideCenti, bool) {
	a, aok := v.Atto128()
	count, _ := a.AsInt128()
	c, cok := count.Div(num.AsInt128(1e16)).Int64()
	return AsOutsideCenti(int64(c)), aok && cok
}

// convOutsideCentiCases pin every cell of the conversion matrix out of
// the outside instantiation, which widens the outside backing to reach
// the package's types.
func convOutsideCentiCases() []core.TestCase {
	return core.S[core.TestCase](
		newConvCaseTruncated("to int32", NewOutsideCenti(1, 50), num.AsInt32(1)),
		// 92233720368547758 whole units, past the 32-bit range.
		newConvCaseOverflow("max to int32", AsOutsideCenti(math.MaxInt64),
			num.AsInt32(2061584302)),
		newConvCaseTruncated("max to int64", AsOutsideCenti(math.MaxInt64),
			num.AsInt64(92233720368547758)),
		newConvCaseTruncated("min to int64", AsOutsideCenti(math.MinInt64),
			num.AsInt64(-92233720368547758)),
		newConvCase("to int128", NewOutsideCenti(-5, 0), num.AsInt128(-5)),
		newConvCase("to uint128", NewOutsideCenti(5, 0), u(5)),
		newConvCaseOverflow("negative to uint128", NewOutsideCenti(-5, 0),
			num.NewUint128(maxWord, maxWord-4)),
		newConvCase("to milli32", NewOutsideCenti(-1, -55),
			num.NewMilli32(-1, -550)),
		newConvCaseOverflow("past milli32 whole", NewOutsideCenti(2147484, 0),
			num.AsMilli32(-2147483296)),
		newConvCase("to milli64", NewOutsideCenti(-1, -55),
			num.NewMilli64(-1, -550)),
		newConvCaseOverflow("max to milli64", AsOutsideCenti(math.MaxInt64),
			num.AsMilli64(-10)),
		newConvCase("to atto128", NewOutsideCenti(-1, -55),
			num.NewAtto128(-1, -550e15)),
		// -92233720368547758.08 in atto-units, inside the 128-bit range.
		newConvCase("min to atto128", AsOutsideCenti(math.MinInt64),
			num.AsAtto128(num.NewInt128(0xffee3c86c81f8000, 0))),
	)
}

func runTestOutsideCentiConvert(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, convOutsideCentiCases())
}

// runTestOutsideCentiCount pins the count accessors over the outside
// backing, the minimum included, which the fallback reads without
// wrapping.
func runTestOutsideCentiCount(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, core.S[core.TestCase](
		newCountCase("as int32", NewOutsideCenti(-1, -55), num.AsInt32(-155)),
		newCountCaseOverflow("past int32", AsOutsideCenti(1<<31),
			num.AsInt32(math.MinInt32)),
		newCountCase("min as int64", AsOutsideCenti(math.MinInt64),
			num.AsInt64(math.MinInt64)),
		newCountCase("min as int128", AsOutsideCenti(math.MinInt64),
			num.AsInt128(math.MinInt64)),
	))
}

// outsidePlain is a signed integer defined outside the package that
// meets [num.Signed] and lacks the Euclidean steps One and ULP, each
// method delegating to the Int128 it wraps, so [num.Decimal], with it as
// the backing of OutsidePlainCenti, reads a backing two words wide
// across its whole range.
type outsidePlain struct {
	v num.Int128
}

var _ num.Signed[outsidePlain] = outsidePlain{}

func (o outsidePlain) IsZero() bool                    { return o.v.IsZero() }
func (o outsidePlain) Equal(v outsidePlain) bool       { return o.v.Equal(v.v) }
func (o outsidePlain) Cmp(v outsidePlain) int          { return o.v.Cmp(v.v) }
func (o outsidePlain) IsNegative() bool                { return o.v.IsNegative() }
func (o outsidePlain) Neg() outsidePlain               { return outsidePlain{o.v.Neg()} }
func (o outsidePlain) Abs() outsidePlain               { return outsidePlain{o.v.Abs()} }
func (o outsidePlain) Add(v outsidePlain) outsidePlain { return outsidePlain{o.v.Add(v.v)} }
func (o outsidePlain) Sub(v outsidePlain) outsidePlain { return outsidePlain{o.v.Sub(v.v)} }
func (o outsidePlain) Mul(v outsidePlain) outsidePlain { return outsidePlain{o.v.Mul(v.v)} }
func (o outsidePlain) Div(v outsidePlain) outsidePlain { return outsidePlain{o.v.Div(v.v)} }
func (o outsidePlain) Mod(v outsidePlain) outsidePlain { return outsidePlain{o.v.Mod(v.v)} }

func (o outsidePlain) DivMod(v outsidePlain) (q, r outsidePlain) {
	vq, vr := o.v.DivMod(v.v)
	return outsidePlain{vq}, outsidePlain{vr}
}

func (o outsidePlain) MulDivMod(v, d outsidePlain) (q, r outsidePlain) {
	vq, vr := o.v.MulDivMod(v.v, d.v)
	return outsidePlain{vq}, outsidePlain{vr}
}

// asOutsidePlain builds an outsidePlain from a native count.
func asOutsidePlain(x int64) outsidePlain {
	return outsidePlain{num.AsInt128(x)}
}

// outsidePlainCentiScale is a scaler defined outside the package, two
// fraction digits over the plain backing.
type outsidePlainCentiScale struct{}

var _ num.DecimalScaler[outsidePlain] = outsidePlainCentiScale{}

func (outsidePlainCentiScale) Scale() outsidePlain { return asOutsidePlain(100) }
func (outsidePlainCentiScale) Name() string        { return "num_test.OutsidePlainCenti" }

// OutsidePlainCenti is a signed fixed-point number with two fraction
// digits over a backing without the Euclidean steps.
type OutsidePlainCenti = num.Decimal[outsidePlain, outsidePlainCentiScale]

// NewOutsidePlainCenti builds an OutsidePlainCenti from a whole-unit
// count and a centi-unit fraction.
func NewOutsidePlainCenti(whole, centi int64) OutsidePlainCenti {
	return num.NewDecimal[outsidePlain, outsidePlainCentiScale](
		asOutsidePlain(whole), asOutsidePlain(centi))
}

// AsOutsidePlainCenti takes a count of centi-units.
func AsOutsidePlainCenti(centi num.Int128) OutsidePlainCenti {
	return num.AsDecimal[outsidePlain, outsidePlainCentiScale](
		outsidePlain{centi})
}

// TestOutsidePlainCenti runs the Euclidean suite over a backing without
// the steps, whose ULP Decimal derives from the scale, and the text,
// count and factory rows across the backing's two words.
func TestOutsidePlainCenti(t *testing.T) {
	t.Run("euclidean", runTestOutsidePlainCentiEuclidean)
	t.Run("text", runTestOutsidePlainCentiText)
	t.Run("count", runTestOutsidePlainCentiCount)
	t.Run("from int128", runTestOutsidePlainCentiFromInt128)
}

func runTestOutsidePlainCentiEuclidean(t *testing.T) {
	t.Helper()
	runEuclideanDecimalTests(t, euclideanDecimalSuite[OutsidePlainCenti]{
		mk:    NewOutsidePlainCenti,
		scale: 100,
	})
}

func runTestOutsidePlainCentiText(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []textCase{
		newTextCase("zero", NewOutsidePlainCenti(0, 0), "0.00"),
		newTextCase("one and a half", NewOutsidePlainCenti(1, 50), "1.50"),
		newTextCase("negative fraction", NewOutsidePlainCenti(0, -5), "-0.05"),
		newTextCase("two to the 64", AsOutsidePlainCenti(num.NewInt128(1, 0)),
			"184467440737095516.16"),
		newTextCase("min", AsOutsidePlainCenti(num.MinInt128),
			"-1701411834604692317316873037158841057.28"),
		newTextCase("max", AsOutsidePlainCenti(num.MaxInt128),
			"1701411834604692317316873037158841057.27"),
	})
}

// runTestOutsidePlainCentiCount pins the count accessors over the
// backing two words wide, the bits of both words and the minimum
// included, which the fallback reads without wrapping.
func runTestOutsidePlainCentiCount(t *testing.T) {
	t.Helper()
	high := num.NewInt128(0x7fff000000000001, 0x8000000000000001)
	core.RunTestCases(t, core.S[core.TestCase](
		newCountCase("zero", NewOutsidePlainCenti(0, 0), num.AsInt128(0)),
		newCountCase("ulp", AsOutsidePlainCenti(num.AsInt128(1)),
			num.AsInt128(1)),
		newCountCase("negative ulp", AsOutsidePlainCenti(num.AsInt128(-1)),
			num.AsInt128(-1)),
		newCountCase("high bits", AsOutsidePlainCenti(high), high),
		newCountCase("negative high bits", AsOutsidePlainCenti(high.Neg()),
			high.Neg()),
		newCountCase("min", AsOutsidePlainCenti(num.MinInt128), num.MinInt128),
		newCountCase("max", AsOutsidePlainCenti(num.MaxInt128), num.MaxInt128),
		newCountCaseOverflow("two to the 64 as int64",
			AsOutsidePlainCenti(num.NewInt128(1, 0)), num.AsInt64(0)),
	))
}

func outsideCentiFromInt128Cases() []core.TestCase {
	// the whole units at the edge of the resolution.
	const whole = math.MaxInt64 / 100
	return core.S[core.TestCase](
		newDecimalFromInt128Case[outsideSigned, outsideCentiScale]("one",
			num.AsInt128(1), NewOutsideCenti(1, 0)),
		newDecimalFromInt128Case[outsideSigned, outsideCentiScale]("negative",
			num.AsInt128(-5), NewOutsideCenti(-5, 0)),
		newDecimalFromInt128Case[outsideSigned, outsideCentiScale]("whole max",
			num.AsInt128(whole), NewOutsideCenti(whole, 0)),
		newDecimalFromInt128Case[outsideSigned, outsideCentiScale]("whole min",
			num.AsInt128(-whole), NewOutsideCenti(-whole, 0)),
		newDecimalFromInt128CaseRange[outsideSigned, outsideCentiScale]("past whole",
			num.AsInt128(whole+1), AsOutsideCenti(math.MaxInt64)),
		newDecimalFromInt128CaseRange[outsideSigned, outsideCentiScale]("past whole min",
			num.AsInt128(-whole-1), AsOutsideCenti(math.MinInt64)),
		// the count wraps past 128 bits before it reaches the backing.
		newDecimalFromInt128CaseRange[outsideSigned, outsideCentiScale]("int128 max",
			num.MaxInt128, AsOutsideCenti(math.MaxInt64)),
		newDecimalFromInt128CaseRange[outsideSigned, outsideCentiScale]("int128 min",
			num.MinInt128, AsOutsideCenti(math.MinInt64)),
	)
}

// runTestOutsideCentiFromInt128 builds the outside instantiation from
// whole units through the Decimal factory, which reaches the backing
// through its type arguments, and checks the family factory refuses
// it, since its switch knows only the seven types of the package.
func runTestOutsideCentiFromInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, outsideCentiFromInt128Cases())

	got, err := num.NewFromInt128[OutsideCenti](num.AsInt128(1))
	assertUnsupported[OutsideCenti](t, err)
	core.AssertEqual(t, OutsideCenti{}, got, "value")
}

func outsidePlainCentiFromInt128Cases() []core.TestCase {
	// the whole units at the edge of the resolution, and a count past
	// the low word.
	whole := num.MaxInt128.Div(num.AsInt128(100))
	centi := num.AsInt128(100)
	one := num.AsInt128(1)
	two64 := num.NewInt128(1, 0)
	return core.S[core.TestCase](
		newDecimalFromInt128Case[outsidePlain, outsidePlainCentiScale]("one",
			one, NewOutsidePlainCenti(1, 0)),
		newDecimalFromInt128Case[outsidePlain, outsidePlainCentiScale]("negative",
			num.AsInt128(-5), NewOutsidePlainCenti(-5, 0)),
		newDecimalFromInt128Case[outsidePlain, outsidePlainCentiScale]("two to the 64",
			two64, AsOutsidePlainCenti(two64.Mul(centi))),
		newDecimalFromInt128Case[outsidePlain, outsidePlainCentiScale]("whole max",
			whole, AsOutsidePlainCenti(whole.Mul(centi))),
		newDecimalFromInt128Case[outsidePlain, outsidePlainCentiScale]("whole min",
			whole.Neg(), AsOutsidePlainCenti(whole.Neg().Mul(centi))),
		// the count wraps past 128 bits, and the Int128 bound on the
		// side of the units fits the backing.
		newDecimalFromInt128CaseRange[outsidePlain, outsidePlainCentiScale]("past whole",
			whole.Add(one), AsOutsidePlainCenti(num.MaxInt128)),
		newDecimalFromInt128CaseRange[outsidePlain, outsidePlainCentiScale]("past whole min",
			whole.Neg().Sub(one), AsOutsidePlainCenti(num.MinInt128)),
		newDecimalFromInt128CaseRange[outsidePlain, outsidePlainCentiScale]("int128 max",
			num.MaxInt128, AsOutsidePlainCenti(num.MaxInt128)),
		newDecimalFromInt128CaseRange[outsidePlain, outsidePlainCentiScale]("int128 min",
			num.MinInt128, AsOutsidePlainCenti(num.MinInt128)),
	)
}

// runTestOutsidePlainCentiFromInt128 builds the instantiation over the
// backing two words wide from whole units through the Decimal factory,
// its bounds included.
func runTestOutsidePlainCentiFromInt128(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, outsidePlainCentiFromInt128Cases())
}
