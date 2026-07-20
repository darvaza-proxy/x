package num

var (
	_ Signed[Milli32] = Milli32{}
	_ Signed[Milli64] = Milli64{}
)

// milli32Scale carries the milli (10^-3) resolution as an Int32, the
// backing of Milli32.
type milli32Scale struct{}

func (milli32Scale) Scale() Int32 {
	return Int32(milliScale)
}

// milli64Scale carries the milli (10^-3) resolution as an Int64, the
// backing of Milli64.
type milli64Scale struct{}

func (milli64Scale) Scale() Int64 {
	return Int64(milliScale)
}

// Milli32 is a signed fixed-point number with 3 fractional digits,
// backed by a 32-bit integer counting milli-units (10^-3). The zero
// value is numeric zero.
type Milli32 = Decimal[Int32, milli32Scale]

// Milli64 is a signed fixed-point number with 3 fractional digits,
// backed by a 64-bit integer counting milli-units (10^-3). The zero
// value is numeric zero.
type Milli64 = Decimal[Int64, milli64Scale]

// NewMilli32 builds a Milli32 from a whole-unit count and a milli-unit
// fraction (10^-3 units). The magnitudes combine as |whole|*1000 +
// |milli|, with the sign taken from whole, or from milli when whole is
// zero. milli need not stay below one whole unit: it carries. The
// combined magnitude wraps if it exceeds the 32-bit range.
func NewMilli32(whole, milli int32) Milli32 {
	return newDecimal[Int32, milli32Scale](Int32(whole), Int32(milli))
}

// NewMilli64 builds a Milli64 from a whole-unit count and a milli-unit
// fraction (10^-3 units). The magnitudes combine as |whole|*1000 +
// |milli|, with the sign taken from whole, or from milli when whole is
// zero. milli need not stay below one whole unit: it carries. The
// combined magnitude wraps if it exceeds the 64-bit range.
func NewMilli64(whole, milli int64) Milli64 {
	return newDecimal[Int64, milli64Scale](Int64(whole), Int64(milli))
}
