package num_test

import (
	"fmt"

	"darvaza.org/x/time/num"
)

// outsideSigned is a signed integer defined outside the package, each
// method delegating to the Int64 it wraps. It wraps rather than
// embeds, so it inherits none of the package's unexported methods, and
// the assertion below pins that [num.SignedEuclidean] asks for none.
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
