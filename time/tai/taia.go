package tai

import (
	"time"

	"darvaza.org/x/time/num"
)

const (
	// taiBaseOffset is the initial TAI-UTC offset included in TAICONST.
	taiBaseOffset = 10

	// TAICONST is the base for mapping Unix seconds to TAI64 labels.
	// It includes the 2^62 label bias and the initial ten-second offset.
	TAICONST = uint64(1<<62 + taiBaseOffset)

	// TAILength is the length of a TAI timestamp in bytes
	TAILength = 8

	// TAINLength is the length of a TAIN timestamp in bytes
	TAINLength = 12

	// TAIALength is the length of a TAIA timestamp in bytes
	TAIALength = 16

	// AttosPerNano is the number of attoseconds in a nanosecond.
	AttosPerNano int64 = 1e9
	// AttosPerSec is the number of attoseconds in a second.
	AttosPerSec int64 = 1e18
)

var (
	attosPerSec  = num.AsInt128(AttosPerSec)
	attosPerNano = num.AsInt128(AttosPerNano)
)

// TAI represents a TAI timestamp with second precision.
// Its signed 128-bit count can exceed the unsigned seconds range of a TAI64
// label. String returns diagnostic text for values outside that range.
type TAI struct {
	x num.Int128
}

// TAIN represents a TAI timestamp with nanosecond precision.
// Its signed 128-bit count can exceed the TAI64N label range. String returns
// diagnostic text when the whole seconds are negative or exceed uint64.
//
//revive:disable-next-line:exported TAIN is the timestamp format's name.
type TAIN struct {
	x num.Int128
}

// TAIA represents a TAI timestamp with attosecond precision.
// Its signed 128-bit count can exceed the TAI64NA label range. String returns
// diagnostic text when the whole seconds are negative or exceed uint64;
// MarshalText and AppendText return ErrTAIARange instead.
//
// Like time.Time, TAIA intentionally mixes receiver types. Read operations use
// value receivers so values implement fmt.Stringer, encoding.TextMarshaler and
// encoding.TextAppender, and methods can be called on function results.
// UnmarshalText uses a pointer receiver because it replaces the destination.
//
//revive:disable-next-line:exported TAIA is the timestamp format's name.
type TAIA struct {
	x num.Int128
}

// Split returns whole seconds, nanoseconds and the remaining attoseconds.
// Seconds round towards negative infinity; both sub-second components are
// non-negative and less than 1e9.
func (t TAIA) Split() (s, n, a num.Int128) {
	s, subAttos := num.EuclideanDivMod(t.x, attosPerSec)
	nano, atto := subAttos.DivMod(attosPerNano)
	return s, nano, atto
}

// Combine builds a timestamp from seconds, nanoseconds and attoseconds.
func Combine(s, n, a num.Int128) TAIA {
	secAttos := s.Mul(attosPerSec)
	nanoAttos := n.Mul(attosPerNano)

	totalAttos := secAttos.Add(nanoAttos).Add(a)

	return TAIA{x: totalAttos}
}

// TAIFromTime converts t into a TAI timestamp with attosecond precision.
//
//revive:disable-next-line:exported Keep the existing conversion function's name.
func TAIFromTime(t time.Time) TAIA {
	taiSecs := num.AsInt128(t.Unix()).Add(num.AsInt128(int64(TAICONST)).Add(num.AsInt128(int64(lsoffset(t)))))

	secAttos := taiSecs.Mul(attosPerSec)

	nanoAttos := num.AsInt128(int64(t.Nanosecond())).Mul(attosPerNano)

	total := secAttos.Add(nanoAttos)
	return TAIA{x: total}
}

// Now returns the current TAI timestamp with attosecond precision.
func Now() TAIA {
	return TAIFromTime(time.Now())
}

// TAI converts TAIA to TAI by discarding nanoseconds and attoseconds.
// The result rounds towards negative infinity, as Split does.
func (t TAIA) TAI() TAI {
	s, _ := num.EuclideanDivMod(t.x, attosPerSec)
	return TAI{x: s}
}

// TAIN converts TAIA to TAIN by discarding attoseconds.
// The result rounds towards negative infinity, as Split does.
func (t TAIA) TAIN() TAIN {
	n, _ := num.EuclideanDivMod(t.x, attosPerNano)
	return TAIN{x: n}
}
