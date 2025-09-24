// Package tai provides TAI (International Atomic Time) and TAIN timestamps
// that follow the Go time package API conventions.
package tai

import (
	"encoding"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	"darvaza.org/core"
)

const (
	nsPerSec    = uint64(1e9)
	maxDuration = time.Duration(math.MaxInt64)
	minDuration = time.Duration(math.MinInt64)
)

// TAIN represents a TAI timestamp with nanosecond precision (TAIN)
//
//revive:disable-next-line:exported
type TAIN struct {
	sec  uint64
	nano uint32
}

var (
	_ encoding.BinaryMarshaler   = TAIN{}
	_ encoding.BinaryUnmarshaler = (*TAIN)(nil)
	_ encoding.TextMarshaler     = TAIN{}
	_ encoding.TextUnmarshaler   = (*TAIN)(nil)
	_ json.Marshaler             = TAIN{}
	_ json.Unmarshaler           = (*TAIN)(nil)
)

// NowTAIN returns the current timestamp as a TAIN
func NowTAIN() TAIN {
	return TAINFromTime(time.Now())
}

// UnixTAIN returns the TAIN corresponding to the given Unix time.
func UnixTAIN(sec int64, nsec int64) TAIN {
	t := time.Unix(sec, nsec)
	return TAINFromTime(t)
}

// DateConfig holds parameters for creating a date
type DateConfig struct {
	Loc                             *time.Location
	Year, Day, Hour, Min, Sec, Nsec int
	Month                           time.Month
}

// DateTAIN returns the TAIN corresponding to the given date configuration
func DateTAIN(cfg DateConfig) TAIN {
	t := time.Date(cfg.Year, cfg.Month, cfg.Day, cfg.Hour, cfg.Min, cfg.Sec, cfg.Nsec, cfg.Loc)
	return TAINFromTime(t)
}

// ParseTAIN parses a TAI64N formatted string.
func ParseTAIN(value string) (TAIN, error) {
	if len(value) != 2*TAINLength+1 {
		return TAIN{}, core.Wrapf(ErrInvalidTAINLength, "got %d, want %d", len(value), 2*TAINLength+1)
	}
	if value[0] != '@' {
		return TAIN{}, ErrInvalidTAINFormat
	}

	var buf [12]byte
	_, err := hex.Decode(buf[:], []byte(value[1:]))
	if err != nil {
		return TAIN{}, core.Wrap(err, "invalid TAI64N string format")
	}

	nano := binary.BigEndian.Uint32(buf[8:])
	if nano > 999999999 {
		return TAIN{}, core.Wrapf(ErrInvalidNanosecondRange, "got %d", nano)
	}

	return TAIN{
		sec:  binary.BigEndian.Uint64(buf[:8]),
		nano: nano,
	}, nil
}

// Add adds a time.Duration to a TAIN timestamp.
// It returns ErrOverflow or ErrUnderflow if the result doesn't fit.
func (t TAIN) Add(d time.Duration) (TAIN, error) {
	result, err := TAINA{sec: t.sec, nano: t.nano}.addSafe(d)
	if err != nil {
		return TAIN{}, err
	}
	return result.TAIN(), nil
}

// Sub subtracts two TAIN timestamps.
// The result is clamped to the representable time.Duration range rather
// than being allowed to wrap or change sign.
func (t TAIN) Sub(u TAIN) time.Duration {
	hiT, loT := toNanos128(t.sec, t.nano)
	hiU, loU := toNanos128(u.sec, u.nano)

	hiD, loD := sub128(hiT, loT, hiU, loU)
	return durationFromSigned128(int64(hiD), loD)
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t TAIN) Unix() int64 {
	return int64(t.sec - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
// The result is undefined if it does not fit in an int64.
func (t TAIN) UnixMilli() int64 {
	return t.Unix()*1000 + int64(t.nano)/1e6
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
// The result is undefined if it does not fit in an int64.
func (t TAIN) UnixMicro() int64 {
	return t.Unix()*1e6 + int64(t.nano)/1e3
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
// !!! The result is undefined if it does not fit in an int64
// (a date before the year 1678 or after 2262).
func (t TAIN) UnixNano() int64 {
	return t.Unix()*1e9 + int64(t.nano)
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t TAIN) Nanosecond() int {
	return int(t.nano)
}

// Before reports whether the time instant t is before u.
func (t TAIN) Before(u TAIN) bool {
	if t.sec != u.sec {
		return t.sec < u.sec
	}
	return t.nano < u.nano
}

// After reports whether the time instant t is after u.
func (t TAIN) After(u TAIN) bool {
	if t.sec != u.sec {
		return t.sec > u.sec
	}
	return t.nano > u.nano
}

// Equal reports whether t and u represent the same time instant.
func (t TAIN) Equal(u TAIN) bool {
	return t.sec == u.sec && t.nano == u.nano
}

// IsZero reports whether t represents the zero time instant.
func (t TAIN) IsZero() bool {
	return t.sec == 0 && t.nano == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t TAIN) Compare(u TAIN) int {
	if t.sec < u.sec || (t.sec == u.sec && t.nano < u.nano) {
		return -1
	}
	if t.sec > u.sec || (t.sec == u.sec && t.nano > u.nano) {
		return 1
	}
	return 0
}

// GoTime returns a time.Time representation of the TAIN timestamp.
// Instants inside an inserted leap second cannot be represented by
// time.Time and map to the first second after it.
func (t TAIN) GoTime() time.Time {
	tm := time.Unix(int64(t.sec-TAICONST), int64(t.nano)).UTC()
	return utcFromTAI(tm)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
func (t TAIN) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// TAI converts TAIN to TAI by truncating nanoseconds.
func (t TAIN) TAI() TAI {
	return TAI{x: t.sec}
}

// TAINA converts TAIN to TAINA with zero attoseconds.
func (t TAIN) TAINA() TAINA {
	return TAINA{sec: t.sec, nano: t.nano}
}

type nanos128 struct {
	hi uint64
	lo uint64
}

type stepRemainder struct {
	remainder uint64
	step      uint64
}

func baselineCalculation(sec uint64, nano uint32, d time.Duration) (h nanos128, r stepRemainder, ok bool) {
	if d <= 0 {
		return nanos128{}, stepRemainder{}, false
	}

	r.step = uint64(d)
	hi, lo := toNanos128(sec, nano)
	h.hi = hi
	h.lo = lo

	r.remainder = remainder128(h.hi, h.lo, r.step)

	h.hi, h.lo = sub128(h.hi, h.lo, 0, r.remainder)
	return h, r, true
}

// Truncate returns the result of rounding t down to a multiple of d.
// !!! If d <= 0, Truncate returns t unchanged.
func (t TAIN) Truncate(d time.Duration) TAIN {
	h, _, ok := baselineCalculation(t.sec, t.nano, d)
	if !ok {
		return t
	}
	a, b := fromNanos128(h.hi, h.lo)
	return TAIN{sec: a, nano: b}
}

// Round returns the result of rounding t to the nearest multiple of d.
// Halfway values round up.
// !!! If d <= 0, Round returns t unchanged.
func (t TAIN) Round(d time.Duration) TAIN {
	h, r, ok := baselineCalculation(t.sec, t.nano, d)
	if !ok {
		return t
	}
	if 2*r.remainder >= r.step {
		h.hi, h.lo = add128(h.hi, h.lo, r.step)
	}
	a, b := fromNanos128(h.hi, h.lo)
	return TAIN{sec: a, nano: b}
}

// String returns the TAI64N string representation
func (t TAIN) String() string {
	var buf [25]byte
	var binBuf [12]byte
	buf[0] = '@'
	binary.BigEndian.PutUint64(binBuf[:8], t.sec)
	binary.BigEndian.PutUint32(binBuf[8:], t.nano)
	hex.Encode(buf[1:], binBuf[:])
	return string(buf[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t TAIN) MarshalBinary() ([]byte, error) {
	result := make([]byte, TAINLength)
	binary.BigEndian.PutUint64(result[:], t.sec)
	binary.BigEndian.PutUint32(result[TAILength:], t.nano)
	return result, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *TAIN) UnmarshalBinary(data []byte) error {
	if len(data) != TAINLength {
		return core.Wrapf(ErrInvalidTAINBinaryLength, "got %d, want %d", len(data), TAINLength)
	}

	nano := binary.BigEndian.Uint32(data[TAILength:])
	if nano > 999999999 {
		return core.Wrapf(ErrInvalidNanosecondRange, "got %d", nano)
	}

	t.sec = binary.BigEndian.Uint64(data[:])
	t.nano = nano
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t TAIN) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TAIN) UnmarshalText(text []byte) error {
	tain, err := ParseTAIN(string(text))
	if err != nil {
		return err
	}
	*t = tain
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t TAIN) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TAIN) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	tain, err := ParseTAIN(s)
	if err != nil {
		return err
	}
	*t = tain
	return nil
}

// Since returns the time elapsed since u.
func (t TAIN) Since(u TAIN) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until u.
func (t TAIN) Until(u TAIN) time.Duration {
	return u.Sub(t)
}

// TAINFromTime returns a TAIN from time.Time
//
//revive:disable-next-line:exported
func TAINFromTime(t time.Time) TAIN {
	return TAIN{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
	}
}
