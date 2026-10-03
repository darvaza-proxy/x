package num

import "darvaza.org/core"

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

// pow10Table lists the powers of ten a Uint128 holds, 10^0 to 10^38,
// indexed by exponent.
var pow10Table = [...]Uint128{
	{lo: 1e0},
	{lo: 1e1},
	{lo: 1e2},
	{lo: 1e3},
	{lo: 1e4},
	{lo: 1e5},
	{lo: 1e6},
	{lo: 1e7},
	{lo: 1e8},
	{lo: 1e9},
	{lo: 1e10},
	{lo: 1e11},
	{lo: 1e12},
	{lo: 1e13},
	{lo: 1e14},
	{lo: 1e15},
	{lo: 1e16},
	{lo: 1e17},
	{lo: 1e18},
	{lo: 1e19},
	{hi: 0x5, lo: 0x6bc75e2d63100000},        // 10^20
	{hi: 0x36, lo: 0x35c9adc5dea00000},       // 10^21
	{hi: 0x21e, lo: 0x19e0c9bab2400000},      // 10^22
	{hi: 0x152d, lo: 0x02c7e14af6800000},     // 10^23
	{hi: 0xd3c2, lo: 0x1bcecceda1000000},     // 10^24
	{hi: 0x84595, lo: 0x161401484a000000},    // 10^25
	{hi: 0x52b7d2, lo: 0xdcc80cd2e4000000},   // 10^26
	{hi: 0x33b2e3c, lo: 0x9fd0803ce8000000},  // 10^27
	{hi: 0x204fce5e, lo: 0x3e25026110000000}, // 10^28
	{hi: 0x1431e0fae, lo: 0x6d7217caa0000000},        // 10^29
	{hi: 0xc9f2c9cd0, lo: 0x4674edea40000000},        // 10^30
	{hi: 0x7e37be2022, lo: 0xc0914b2680000000},       // 10^31
	{hi: 0x4ee2d6d415b, lo: 0x85acef8100000000},      // 10^32
	{hi: 0x314dc6448d93, lo: 0x38c15b0a00000000},     // 10^33
	{hi: 0x1ed09bead87c0, lo: 0x378d8e6400000000},    // 10^34
	{hi: 0x13426172c74d82, lo: 0x2b878fe800000000},   // 10^35
	{hi: 0xc097ce7bc90715, lo: 0xb34b9f1000000000},   // 10^36
	{hi: 0x785ee10d5da46d9, lo: 0x00f436a000000000},  // 10^37
	{hi: 0x4b3b4ca85a86c47a, lo: 0x098a224000000000}, // 10^38
}

// Pow10 returns 10^n as a Uint128, for 0 <= n <= 38, the powers of ten
// a Uint128 holds; an n outside that range panics with [ErrPow10Range].
func Pow10(n int) Uint128 {
	if n < 0 || n >= len(pow10Table) {
		core.PanicFrom(1, ErrPow10Range)
	}
	return pow10Table[n]
}

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
