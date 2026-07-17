package num

// Unexported word primitives.
const (
	// maxUint64 is the all-ones 64-bit word, the largest uint64.
	maxUint64 = ^uint64(0)
	// maxInt64 is the largest value of a signed 64-bit word.
	maxInt64 uint64 = maxUint64 >> 1
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
