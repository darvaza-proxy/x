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

// AttosecondTimestamp represents a timestamp split into
// nanoseconds and attoseconds to avoid int64 overflow issues with
// UnixAttosecond calculations
type AttosecondTimestamp struct {
	// Nanoseconds since Unix epoch (from UnixNano())
	UnixNanoseconds int64
	// Additional attoseconds within that nanosecond [0, 999,999,999]
	Attoseconds uint32
}

// AttosecondDuration represents a time difference split into
// nanoseconds and attoseconds to avoid int64 overflow: one second is
// 1e18 attoseconds, so a plain attosecond count would overflow int64
// beyond ±9.2 seconds. The total difference is Nanoseconds*1e9 +
// Attoseconds attoseconds; Attoseconds is always in [0, 999999999],
// so negative differences carry the sign in Nanoseconds alone.
type AttosecondDuration struct {
	// Whole nanoseconds of the difference (signed)
	Nanoseconds int64
	// Additional attoseconds within that nanosecond [0, 999,999,999]
	Attoseconds uint32
}

// TAINA represents a TAI timestamp with attosecond precision (TAI64NA)
//
//revive:disable-next-line:exported
type TAINA struct {
	sec  uint64
	nano uint32
	atto uint32
}

var (
	_ encoding.BinaryMarshaler   = TAINA{}
	_ encoding.BinaryUnmarshaler = (*TAINA)(nil)
	_ encoding.TextMarshaler     = TAINA{}
	_ encoding.TextUnmarshaler   = (*TAINA)(nil)
	_ json.Marshaler             = TAINA{}
	_ json.Unmarshaler           = (*TAINA)(nil)
)

// NowTAINA returns the current timestamp as a TAINA
func NowTAINA() TAINA {
	return TAINAFromTime(time.Now())
}

// UnixTAINA returns the TAI time corresponding to the given Unix time
// with attoseconds.
// UnixTAINA panics if asec is out of the range [0, 999999999].
func UnixTAINA(sec int64, nsec int64, asec uint32) TAINA {
	if asec > 999999999 {
		core.PanicWrapf(ErrInvalidAttosecondRange, "got %d", asec)
	}
	result := TAINAFromTime(time.Unix(sec, nsec))
	result.atto = asec
	return result
}

// DateTAINA returns the TAI time corresponding to the given date
// configuration with attoseconds
// DateTAINA panics if asec is out of the range [0, 999999999].
func DateTAINA(cfg DateConfig, asec uint32) TAINA {
	if asec > 999999999 {
		core.PanicWrapf(ErrInvalidAttosecondRange, "got %d", asec)
	}
	result := TAINAFromTime(time.Date(cfg.Year, cfg.Month, cfg.Day, cfg.Hour, cfg.Min, cfg.Sec, cfg.Nsec, cfg.Loc))
	result.atto = asec
	return result
}

// ParseTAINA parses a TAI64NA formatted string.
func ParseTAINA(value string) (TAINA, error) {
	if len(value) != 2*TAINALength+1 {
		return TAINA{}, core.Wrapf(ErrInvalidTAINALength, "got %d, want %d", len(value), 2*TAINALength+1)
	}
	if value[0] != '@' {
		return TAINA{}, ErrInvalidTAINAFormat
	}

	var buf [16]byte
	_, err := hex.Decode(buf[:], []byte(value[1:]))
	if err != nil {
		return TAINA{}, core.Wrap(err, "invalid TAI64NA string format")
	}

	atto := binary.BigEndian.Uint32(buf[12:16])
	if atto > 999999999 {
		return TAINA{}, core.Wrapf(ErrInvalidAttosecondRange, "got %d", atto)
	}
	nano := binary.BigEndian.Uint32(buf[8:12])
	if nano > 999999999 {
		return TAINA{}, core.Wrapf(ErrInvalidNanosecondRange, "got %d", nano)
	}

	return TAINA{
		sec:  binary.BigEndian.Uint64(buf[:8]),
		nano: nano,
		atto: atto,
	}, nil
}

func (t TAINA) addPositiveDuration(addSecs int64, addNanos int64) (TAINA, error) {
	var result TAINA
	var err error

	result.sec, err = safeAddUint64(t.sec, uint64(addSecs))
	if err != nil {
		return TAINA{}, err
	}

	newNanos := int64(t.nano) + addNanos
	if newNanos >= 1e9 {
		result.sec, err = safeAddUint64(result.sec, 1)
		if err != nil {
			return TAINA{}, err
		}
		result.nano = uint32(newNanos - 1e9)
	} else {
		result.nano = uint32(newNanos)
	}
	result.atto = t.atto
	return result, nil
}

func (t TAINA) addNegativeDuration(addSecs int64, addNanos int64) (TAINA, error) {
	var result TAINA
	var err error

	result.sec, err = safeSubUint64(t.sec, uint64(-addSecs))
	if err != nil {
		return TAINA{}, err
	}

	newNanos := int64(t.nano) + addNanos
	if newNanos < 0 {
		result.sec, err = safeSubUint64(result.sec, 1)
		if err != nil {
			return TAINA{}, err
		}
		result.nano = uint32(newNanos + 1e9)
	} else {
		result.nano = uint32(newNanos)
	}
	result.atto = t.atto
	return result, nil
}

func (t TAINA) addSafe(d time.Duration) (TAINA, error) {
	totalNanos := int64(d)
	addSecs := totalNanos / 1e9
	addNanos := totalNanos % 1e9

	if addSecs >= 0 && addNanos >= 0 {
		return t.addPositiveDuration(addSecs, addNanos)
	}
	return t.addNegativeDuration(addSecs, addNanos)
}

// Add adds a time.Duration to a TAINA timestamp
// Returns error on overflow/underflow instead of panicking
func (t TAINA) Add(d time.Duration) (TAINA, error) {
	return t.addSafe(d)
}

func (t TAINA) addAttosecondsCarry(newAtto int64) (TAINA, error) {
	carryNanos := newAtto / 1e9
	result := t
	result.atto = uint32(newAtto % 1e9)
	return result.addSafe(time.Duration(carryNanos) * time.Nanosecond)
}

func (t TAINA) addAttosecondsBorrow(newAtto int64) (TAINA, error) {
	borrowNanos, atto := floorDivMod(newAtto, 1e9)
	result := t
	result.atto = uint32(atto)
	return result.addSafe(time.Duration(borrowNanos) * time.Nanosecond)
}

// AddAttoseconds adds attoseconds to a TAINA timestamp.
// It returns ErrOverflow or ErrUnderflow if the result doesn't fit.
func (t TAINA) AddAttoseconds(asec int64) (TAINA, error) {
	newAtto := int64(t.atto) + asec
	if asec > 0 && newAtto < asec {
		// int64(t.atto) + asec wrapped around
		return TAINA{}, ErrOverflow
	}

	if newAtto >= 1e9 {
		return t.addAttosecondsCarry(newAtto)
	}
	if newAtto < 0 {
		return t.addAttosecondsBorrow(newAtto)
	}

	result := t
	result.atto = uint32(newAtto)
	return result, nil
}

// Sub subtracts two TAINA timestamps, at nanosecond precision (attoseconds
// are not included; see SubAttoseconds for the full-precision difference).
// The result is clamped to the representable time.Duration range rather
// than being allowed to wrap or change sign.
func (t TAINA) Sub(u TAINA) time.Duration {
	hiT, loT := toNanos128(t.sec, t.nano)
	hiU, loU := toNanos128(u.sec, u.nano)

	hiD, loD := sub128(hiT, loT, hiU, loU)
	return durationFromSigned128(int64(hiD), loD)
}

// SubAttoseconds returns the difference between two TAINA timestamps as
// an AttosecondDuration. Nanoseconds shares time.Duration's ±292-year
// range (and is clamped the same way by Sub); the split only adds
// attosecond-level precision within that span, not a wider total range.
func (t TAINA) SubAttoseconds(u TAINA) AttosecondDuration {
	durNanos := t.Sub(u).Nanoseconds()
	attoDiff := int64(t.atto) - int64(u.atto)

	if attoDiff < 0 {
		if durNanos == math.MinInt64 {
			// durNanos is already saturated at the most negative
			// representable value borrowing one more nanosecond
			// would wrap int64 and silently flip the sign.
			// Attosecond precision is meaningless at this magnitude,
			// so drop the borrow instead.
			attoDiff = 0
		} else {
			durNanos--
			attoDiff += 1e9
		}
	}

	return AttosecondDuration{
		Nanoseconds: durNanos,
		Attoseconds: uint32(attoDiff),
	}
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t TAINA) Unix() int64 {
	return int64(t.sec - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
// !!! The result is undefined if it does not fit in an int64.
func (t TAINA) UnixMilli() int64 {
	return t.Unix()*1000 + int64(t.nano)/1e6
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
// !!! The result is undefined if it does not fit in an int64.
func (t TAINA) UnixMicro() int64 {
	return t.Unix()*1e6 + int64(t.nano)/1e3
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
// !!! The result is undefined if it does not fit in an int64
// (a date before the year 1678 or after 2262).
func (t TAINA) UnixNano() int64 {
	return t.Unix()*1e9 + int64(t.nano)
}

// UnixAttosecondSplit returns the timestamp as nanoseconds + attoseconds
// to avoid int64 overflow issues with large timestamps.
func (t TAINA) UnixAttosecondSplit() AttosecondTimestamp {
	return AttosecondTimestamp{
		UnixNanoseconds: t.UnixNano(),
		Attoseconds:     t.atto,
	}
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t TAINA) Nanosecond() int {
	return int(t.nano)
}

// Attosecond returns the attosecond offset within the nanosecond
// specified by t in the range [0, 999999999].
func (t TAINA) Attosecond() int {
	return int(t.atto)
}

// Before reports whether the time instant t is before u.
func (t TAINA) Before(u TAINA) bool {
	if t.sec != u.sec {
		return t.sec < u.sec
	}
	if t.nano != u.nano {
		return t.nano < u.nano
	}
	return t.atto < u.atto
}

// After reports whether the time instant t is after u.
func (t TAINA) After(u TAINA) bool {
	if t.sec != u.sec {
		return t.sec > u.sec
	}
	if t.nano != u.nano {
		return t.nano > u.nano
	}
	return t.atto > u.atto
}

// Equal reports whether t and u represent the same time instant.
func (t TAINA) Equal(u TAINA) bool {
	return t.sec == u.sec && t.nano == u.nano && t.atto == u.atto
}

// IsZero reports whether t represents the zero time instant.
func (t TAINA) IsZero() bool {
	return t.sec == 0 && t.nano == 0 && t.atto == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t TAINA) Compare(u TAINA) int {
	switch {
	case t.sec < u.sec:
		return -1
	case t.sec > u.sec:
		return 1
	case t.nano < u.nano:
		return -1
	case t.nano > u.nano:
		return 1
	case t.atto < u.atto:
		return -1
	case t.atto > u.atto:
		return 1
	default:
		return 0
	}
}

// GoTime returns a time.Time representation of the TAINA timestamp.
// Note: Go's time.Time doesn't support attosecond precision,
// so attoseconds are lost.
// Instants inside an inserted leap second cannot be represented by
// time.Time and map to the first second after it.
func (t TAINA) GoTime() time.Time {
	tm := time.Unix(int64(t.sec-TAICONST), int64(t.nano)).UTC()
	return utcFromTAI(tm)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
// !!! Note: Attosecond precision is lost in the conversion.
func (t TAINA) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// TAIN converts TAINA to TAIN by truncating attoseconds.
func (t TAINA) TAIN() TAIN {
	return TAIN{sec: t.sec, nano: t.nano}
}

// TAI converts TAINA to TAI by truncating nanoseconds and attoseconds.
func (t TAINA) TAI() TAI {
	return TAI{x: t.sec}
}

// Truncate returns the result of rounding t down to a multiple of d.
// Attoseconds are always truncated away.
// !!! If d <= 0, Truncate returns t unchanged.
func (t TAINA) Truncate(d time.Duration) TAINA {
	h, _, ok := baselineCalculation(t.sec, t.nano, d)
	if !ok {
		return TAINA{sec: t.sec, nano: t.nano, atto: t.atto}
	}
	a, b := fromNanos128(h.hi, h.lo)
	return TAINA{sec: a, nano: b}
}

// Round returns the result of rounding t to the nearest multiple of d.
// Halfway values round up. Attoseconds are always truncated away.
// !!! If d <= 0, Round returns t unchanged.
func (t TAINA) Round(d time.Duration) TAINA {
	h, r, ok := baselineCalculation(t.sec, t.nano, d)
	if !ok {
		return TAINA{sec: t.sec, nano: t.nano, atto: t.atto}
	}
	if 2*r.remainder >= r.step {
		h.hi, h.lo = add128(h.hi, h.lo, r.step)
	}
	a, b := fromNanos128(h.hi, h.lo)
	return TAINA{sec: a, nano: b}
}

// String returns the TAI64NA string representation
func (t TAINA) String() string {
	var buf [33]byte
	var binBuf [16]byte
	buf[0] = '@'
	binary.BigEndian.PutUint64(binBuf[:8], t.sec)
	binary.BigEndian.PutUint32(binBuf[8:12], t.nano)
	binary.BigEndian.PutUint32(binBuf[12:16], t.atto)
	hex.Encode(buf[1:], binBuf[:])
	return string(buf[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t TAINA) MarshalBinary() ([]byte, error) {
	result := make([]byte, TAINALength)
	binary.BigEndian.PutUint64(result[:], t.sec)
	binary.BigEndian.PutUint32(result[TAILength:], t.nano)
	binary.BigEndian.PutUint32(result[TAINLength:], t.atto)
	return result, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *TAINA) UnmarshalBinary(data []byte) error {
	if len(data) != TAINALength {
		return core.Wrapf(ErrInvalidTAINABinaryLength, "got %d, want %d", len(data), TAINALength)
	}

	atto := binary.BigEndian.Uint32(data[TAINLength:])
	if atto > 999999999 {
		return core.Wrapf(ErrInvalidAttosecondRange, "got %d", atto)
	}

	nano := binary.BigEndian.Uint32(data[TAILength:])
	if nano > 999999999 {
		return core.Wrapf(ErrInvalidNanosecondRange, "got %d", nano)
	}
	t.sec = binary.BigEndian.Uint64(data[:])
	t.nano = nano
	t.atto = atto
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t TAINA) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TAINA) UnmarshalText(text []byte) error {
	taina, err := ParseTAINA(string(text))
	if err != nil {
		return err
	}
	*t = taina
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t TAINA) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TAINA) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	taina, err := ParseTAINA(s)
	if err != nil {
		return err
	}
	*t = taina
	return nil
}

// Since returns the time elapsed since u, as observed at t.
func (t TAINA) Since(u TAINA) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until u, as observed at t.
func (t TAINA) Until(u TAINA) time.Duration {
	return u.Sub(t)
}

// TAINAFromTime returns a TAINA from time.Time
// !!! Note: Attoseconds are set to 0 since time.Time doesn't support
// attosecond precision
//
//revive:disable-next-line:exported
func TAINAFromTime(t time.Time) TAINA {
	return TAINA{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
	}
}
