package num

// Euclidean is the constraint met by the types the Euclidean helpers
// can correct, listing just the surface the correction needs rather
// than the full [Signed] interface. The signed integers Int32, Int64
// and Int128 and the Decimal instantiations over them qualify. The
// unexported step methods set how far the quotient moves: the two
// coincide for the integers and differ for Decimal, whose DivMod
// quotient moves in whole units while its MulDivMod quotient moves
// in sub-units.
type Euclidean[T any] interface {
	IsNegative() bool
	Abs() T

	Add(v T) T
	Sub(v T) T
	DivMod(v T) (q, r T)
	MulDivMod(v, d T) (q, r T)

	// one returns the multiplicative unit, the step between
	// consecutive DivMod quotients.
	one() T
	// ulp returns the smallest positive value, the step between
	// consecutive MulDivMod quotients.
	ulp() T
}

// SignedEuclidean is the constraint met by signed types offering both
// the full [Signed] surface and the [Euclidean] correction surface:
// the signed integers Int32, Int64 and Int128 and the Decimal
// instantiations over them. It is the constraint a [Decimal] backing
// must meet.
type SignedEuclidean[T any] interface {
	Signed[T]
	Euclidean[T]
}

// EuclideanDivMod returns the quotient and remainder of v/w with the
// remainder always non-negative, so that v == q*w + r with
// 0 <= r < |w|. The quotient rounds towards negative infinity when w
// is positive and towards positive infinity when w is negative; for
// Decimal it is the whole-multiple count of DivMod, corrected the
// same way. It panics with [ErrDivZero] when w is zero. The bound
// holds even for the most negative w, whose magnitude wraps: the
// corrected remainder is still representable and below it.
func EuclideanDivMod[T Euclidean[T]](v, w T) (q, r T) {
	q, r = v.DivMod(w)
	return toEuclidean(q, r, w, w.one())
}

// EuclideanMulDivMod returns the quotient and remainder of v*w/d with
// the remainder always non-negative and less than |d|, forming the
// product in the same wide intermediate as MulDivMod so it cannot
// overflow before the division. The quotient wraps if it exceeds the
// backing range; v*w == q*d + r holds whenever it does not — for
// Decimal in backing sub-units, as with MulDivMod. It panics with
// [ErrDivZero] when d is zero. As with [EuclideanDivMod], the bound
// survives the most negative d, whose magnitude wraps.
func EuclideanMulDivMod[T Euclidean[T]](v, w, d T) (q, r T) {
	q, r = v.MulDivMod(w, d)
	return toEuclidean(q, r, d, d.ulp())
}

// toEuclidean converts a truncated quotient and remainder into the
// Euclidean pair. A negative remainder gains |d| while the quotient
// gives back one step in the direction of d's sign, preserving the
// underlying identity. When d is the most negative value its
// magnitude wraps, but adding it still applies the required 2^(N-1)
// correction modulo 2^N, and the corrected remainder is
// representable.
func toEuclidean[T Euclidean[T]](uq, ur, d, step T) (q, r T) {
	q, r = uq, ur
	if r.IsNegative() {
		r = r.Add(d.Abs())
		if d.IsNegative() {
			q = q.Add(step)
		} else {
			q = q.Sub(step)
		}
	}
	return q, r
}
