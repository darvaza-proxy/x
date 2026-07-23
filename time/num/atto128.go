package num

var (
	_ Signed[Atto128]    = Atto128{}
	_ Euclidean[Atto128] = Atto128{}
)

// atto128Scale carries the atto (10^-18) resolution as an Int128, the
// backing of Atto128.
type atto128Scale struct{}

func (atto128Scale) Scale() Int128 {
	return Int128{lo: attoScale}
}

// Atto128 is a signed fixed-point number with 18 fractional digits,
// backed by a 128-bit two's-complement integer counting atto-units
// (10^-18). It is the widest instantiation of [Decimal]: the Int128
// backing forms its MulDivMod product in a 256-bit intermediate, and a
// wider backing would need a wider intermediate still. The zero value is
// numeric zero.
type Atto128 = Decimal[Int128, atto128Scale]

// NewAtto128 builds an Atto128 from a whole-unit count and an atto-unit
// fraction (10^-18 units). The magnitudes combine as
// |whole|*10^18 + |atto|, with the sign taken from whole, or from atto
// when whole is zero. So NewAtto128(0, -123e15) yields -0.123, and
// both NewAtto128(-1, 500e15) and NewAtto128(-1, -500e15) yield
// -1.5.
//
// atto need not stay below one whole unit: it carries, so
// NewAtto128(1, 2e18) yields 3.0. The combined magnitude wraps if it
// exceeds the 128-bit range.
func NewAtto128(whole, atto int64) Atto128 {
	return newDecimal[Int128, atto128Scale](NewInt128(whole), NewInt128(atto))
}
