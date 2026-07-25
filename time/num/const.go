package num

// Unexported word primitives.
const (
	// maxUint64 is the all-ones 64-bit word, the largest uint64.
	maxUint64 = ^uint64(0)
	// maxInt64 is the largest value of a signed 64-bit word.
	maxInt64 uint64 = maxUint64 >> 1
)

// Fixed-point scale factors: the number of sub-units in one whole unit
// at each resolution.
const (
	// milliScale is the milli (10^-3) resolution as a plain factor,
	// narrow enough to convert into either backing width.
	milliScale = 1e3
	// attoScale is the atto (10^-18) resolution as a plain factor.
	attoScale = 1e18
)

// Sentinel bounds. These are effectively constants, held as var only
// because a struct cannot be a Go const.
var (
	// MaxUint128 is the largest representable Uint128.
	MaxUint128 = Uint128{hi: maxUint64, lo: maxUint64}
	// ZeroUint128 is the Uint128 zero value.
	ZeroUint128 = Uint128{hi: 0, lo: 0}

	// MaxInt128 is the largest representable Int128.
	MaxInt128 = Int128{hi: maxInt64, lo: maxUint64}
	// MinInt128 is the smallest (most negative) representable Int128.
	MinInt128 = Int128{hi: 1 << 63, lo: 0}
	// ZeroInt128 is the Int128 zero value.
	ZeroInt128 = Int128{hi: 0, lo: 0}
)
