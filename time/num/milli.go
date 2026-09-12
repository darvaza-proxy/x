package num

import (
	"encoding"
	"fmt"
)

var (
	_ Signed[Milli32] = Milli32{}
	_ Signed[Milli64] = Milli64{}

	_ Euclidean[Milli32] = Milli32{}
	_ Euclidean[Milli64] = Milli64{}

	_ fmt.Formatter  = Milli32{}
	_ fmt.Formatter  = Milli64{}
	_ fmt.GoStringer = Milli32{}
	_ fmt.GoStringer = Milli64{}
	_ fmt.Stringer   = Milli32{}
	_ fmt.Stringer   = Milli64{}

	_ encoding.TextAppender  = Milli32{}
	_ encoding.TextAppender  = Milli64{}
	_ encoding.TextMarshaler = Milli32{}
	_ encoding.TextMarshaler = Milli64{}
)

// milli32Scale carries the milli (10^-3) resolution as an Int32, the
// backing of Milli32.
type milli32Scale struct{}

func (milli32Scale) Scale() Int32 {
	return Int32(milliScale)
}

func (milli32Scale) name() string {
	return "Milli32"
}

func (milli32Scale) asInt64(v Int32) (int64, bool) {
	return int64(v), true
}

func (milli32Scale) doAppendText(dst []byte, v Int32) []byte {
	return v.doAppendText(dst)
}

// milli64Scale carries the milli (10^-3) resolution as an Int64, the
// backing of Milli64.
type milli64Scale struct{}

func (milli64Scale) Scale() Int64 {
	return Int64(milliScale)
}

func (milli64Scale) name() string {
	return "Milli64"
}

func (milli64Scale) asInt64(v Int64) (int64, bool) {
	return v.sys(), true
}

func (milli64Scale) doAppendText(dst []byte, v Int64) []byte {
	return v.doAppendText(dst)
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

// AsMilli32 takes an Int32 as a count of milli-units (10^-3), so
// AsMilli32(1500) is 1.5.
func AsMilli32(milli Int32) Milli32 {
	return Milli32{milli}
}

// AsMilli64 takes an Int64 as a count of milli-units (10^-3), so
// AsMilli64(1500) is 1.5.
func AsMilli64(milli Int64) Milli64 {
	return Milli64{milli}
}
