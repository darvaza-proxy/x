package num_test

import "darvaza.org/x/time/num"

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
