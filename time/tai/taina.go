// Package tai provides TAI (International Atomic Time) and TAIN timestamps
// that follow the Go time package API conventions.
package tai

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Error variables for TaiNA operations
var (
	ErrUnsupportedLayoutTaiNA   = errors.New("unsupported layout")
	ErrInvalidTAINAFormat       = errors.New("invalid TAI64NA format: expected @XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	ErrInvalidTAINABinaryLength = errors.New("invalid TaiNA binary data length")
	ErrInvalidAttosecondRange   = errors.New("attoseconds must be in range 0-999999999")
	ErrTaiNAOverflow            = errors.New("TaiNA arithmetic overflow")
	ErrTaiNAUnderflow           = errors.New("TaiNA arithmetic underflow")
)

// AttosecondTimestamp represents a timestamp split into nanoseconds and attoseconds
// to avoid int64 overflow issues with UnixAttosecond calculations
type AttosecondTimestamp struct {
	UnixNanoseconds int64  // Nanoseconds since Unix epoch (from UnixNano())
	Attoseconds     uint32 // Additional attoseconds within that nanosecond [0, 999,999,999]
}

// TaiNA represents a TAI timestamp with attosecond precision (TAI64NA)
//
//revive:disable-next-line:exported
type TaiNA struct {
	sec  uint64
	nano uint32
	atto uint32
}

// safeAddUint64 performs checked addition of uint64 values
func safeAddUint64(a, b uint64) (uint64, error) {
	if b > 0 && a > (^uint64(0))-b {
		return 0, ErrTaiNAOverflow
	}
	return a + b, nil
}

// safeSubUint64 performs checked subtraction of uint64 values
func safeSubUint64(a, b uint64) (uint64, error) {
	if b > a {
		return 0, ErrTaiNAUnderflow
	}
	return a - b, nil
}

// Now returns the current timestamp as a TaiNA
func Now() TaiNA {
	now := time.Now()
	return TaiNAfromGoTime(now)
}

// NowTaiNA returns the current timestamp as a TaiNA
func NowTaiNA() TaiNA {
	return Now()
}

// UnixTaiNA returns the TAI time corresponding to the given Unix time with attoseconds.
func UnixTaiNA(sec int64, nsec int64, asec uint32) TaiNA {
	if asec > 999999999 {
		panic("attoseconds must be in range 0-999999999")
	}
	t := time.Unix(sec, nsec)
	return TaiNA{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
		atto: asec,
	}
}

// DateTaiNA returns the TAI time corresponding to the given date configuration with attoseconds
func DateTaiNA(cfg DateConfig, asec uint32) TaiNA {
	if asec > 999999999 {
		panic("attoseconds must be in range 0-999999999")
	}
	t := time.Date(cfg.Year, cfg.Month, cfg.Day, cfg.Hour, cfg.Min, cfg.Sec, cfg.Nsec, cfg.Loc)
	return TaiNA{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
		atto: asec,
	}
}

// ParseTaiNA parses a TAI64NA formatted string.
func ParseTaiNA(layout, value string) (TaiNA, error) {
	if layout != "@XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX" && layout != "tai64na" {
		return TaiNA{}, fmt.Errorf("%w: %s", ErrUnsupportedLayoutTaiNA, layout)
	}
	return TaiNAfromString(value)
}

// addPositiveDuration adds a positive duration to TaiNA
func (t TaiNA) addPositiveDuration(addSecs int64, addNanos int64) (TaiNA, error) {
	var result TaiNA
	var err error

	result.sec, err = safeAddUint64(t.sec, uint64(addSecs))
	if err != nil {
		return TaiNA{}, err
	}

	newNanos := int64(t.nano) + addNanos
	if newNanos >= 1e9 {
		result.sec, err = safeAddUint64(result.sec, 1)
		if err != nil {
			return TaiNA{}, err
		}
		result.nano = uint32(newNanos - 1e9)
	} else {
		result.nano = uint32(newNanos)
	}
	result.atto = t.atto
	return result, nil
}

// addNegativeDuration adds a negative duration to TaiNA
func (t TaiNA) addNegativeDuration(addSecs int64, addNanos int64) (TaiNA, error) {
	var result TaiNA
	var err error

	result.sec, err = safeSubUint64(t.sec, uint64(-addSecs))
	if err != nil {
		return TaiNA{}, err
	}

	newNanos := int64(t.nano) + addNanos
	if newNanos < 0 {
		result.sec, err = safeSubUint64(result.sec, 1)
		if err != nil {
			return TaiNA{}, err
		}
		result.nano = uint32(newNanos + 1e9)
	} else {
		result.nano = uint32(newNanos)
	}
	result.atto = t.atto
	return result, nil
}

// addSafe adds a time.Duration to a TaiNA timestamp with overflow checking
func (t TaiNA) addSafe(d time.Duration) (TaiNA, error) {
	totalNanos := int64(d)
	addSecs := totalNanos / 1e9
	addNanos := totalNanos % 1e9

	if addSecs >= 0 && addNanos >= 0 {
		return t.addPositiveDuration(addSecs, addNanos)
	}
	return t.addNegativeDuration(addSecs, addNanos)
}

// Add adds a time.Duration to a TaiNA timestamp
// Returns error on overflow/underflow instead of panicking
func (t TaiNA) Add(d time.Duration) (TaiNA, error) {
	return t.addSafe(d)
}

// addAttosecondsCarry handles positive attosecond carry to nanoseconds
func (t TaiNA) addAttosecondsCarry(newAtto int64) (TaiNA, error) {
	carryNanos := newAtto / 1e9
	result := t
	result.atto = uint32(newAtto % 1e9)
	return result.addSafe(time.Duration(carryNanos) * time.Nanosecond)
}

// addAttosecondsBorrow handles negative attosecond borrow from nanoseconds
func (t TaiNA) addAttosecondsBorrow(newAtto int64) (TaiNA, error) {
	borrowNanos := (-newAtto + 1e9 - 1) / 1e9 // Ceiling division
	result := t
	result.atto = uint32(newAtto + borrowNanos*1e9)
	return result.addSafe(-time.Duration(borrowNanos) * time.Nanosecond)
}

// AddAttoseconds adds attoseconds to a TaiNA timestamp
func (t TaiNA) AddAttoseconds(asec int64) (TaiNA, error) {
	newAtto := int64(t.atto) + asec

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

// Sub subtracts two TaiNA timestamps
func (t TaiNA) Sub(u TaiNA) time.Duration {
	secDiff := int64(t.sec - u.sec)
	nanoDiff := int64(t.nano) - int64(u.nano)

	if nanoDiff < 0 {
		secDiff--
		nanoDiff += 1e9
	}

	return time.Duration(secDiff)*time.Second + time.Duration(nanoDiff)*time.Nanosecond
}

// SubAttoseconds returns the difference in attoseconds between two TaiNA timestamps
func (t TaiNA) SubAttoseconds(u TaiNA) int64 {
	// First get the duration in nanoseconds
	durNanos := t.Sub(u).Nanoseconds()

	// Then add the attosecond difference
	attoDiff := int64(t.atto) - int64(u.atto)

	return durNanos*1e9 + attoDiff
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t TaiNA) Unix() int64 {
	return int64(t.sec - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
func (t TaiNA) UnixMilli() int64 {
	return t.Unix()*1000 + int64(t.nano)/1e6
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
func (t TaiNA) UnixMicro() int64 {
	return t.Unix()*1e6 + int64(t.nano)/1e3
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
func (t TaiNA) UnixNano() int64 {
	return t.Unix()*1e9 + int64(t.nano)
}

// UnixAttosecondSplit returns the timestamp as nanoseconds + attoseconds
// to avoid int64 overflow issues with large timestamps.
func (t TaiNA) UnixAttosecondSplit() AttosecondTimestamp {
	return AttosecondTimestamp{
		UnixNanoseconds: t.UnixNano(),
		Attoseconds:     t.atto,
	}
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t TaiNA) Nanosecond() int {
	return int(t.nano)
}

// Attosecond returns the attosecond offset within the nanosecond specified by t,
// in the range [0, 999999999].
func (t TaiNA) Attosecond() int {
	return int(t.atto)
}

// Before reports whether the time instant t is before u.
func (t TaiNA) Before(u TaiNA) bool {
	if t.sec != u.sec {
		return t.sec < u.sec
	}
	if t.nano != u.nano {
		return t.nano < u.nano
	}
	return t.atto < u.atto
}

// After reports whether the time instant t is after u.
func (t TaiNA) After(u TaiNA) bool {
	if t.sec != u.sec {
		return t.sec > u.sec
	}
	if t.nano != u.nano {
		return t.nano > u.nano
	}
	return t.atto > u.atto
}

// Equal reports whether t and u represent the same time instant.
func (t TaiNA) Equal(u TaiNA) bool {
	return t.sec == u.sec && t.nano == u.nano && t.atto == u.atto
}

// IsZero reports whether t represents the zero time instant.
func (t TaiNA) IsZero() bool {
	return t.sec == 0 && t.nano == 0 && t.atto == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t TaiNA) Compare(u TaiNA) int {
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

// GoTime returns a time.Time representation of the TaiNA timestamp.
// Note: Go's time.Time doesn't support attosecond precision, so attoseconds are lost.
func (t TaiNA) GoTime() time.Time {
	tm := time.Unix(int64(t.sec-TAICONST), int64(t.nano)).UTC()
	return tm.Add(-time.Duration(lsoffset(tm)) * time.Second)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
// Note: Attosecond precision is lost in the conversion.
func (t TaiNA) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// TaiN converts TaiNA to TaiN by truncating attoseconds.
func (t TaiNA) TaiN() TaiN {
	return TaiN{sec: t.sec, nano: t.nano}
}

// Tai converts TaiNA to Tai by truncating nanoseconds and attoseconds.
func (t TaiNA) Tai() Tai {
	return Tai{x: t.sec}
}

// Truncate returns the result of rounding t down to a multiple of d (since the zero time).
// If d <= 0, Truncate returns t stripped of any monotonic clock reading but otherwise unchanged.
func (t TaiNA) Truncate(d time.Duration) TaiNA {
	if d <= 0 {
		return TaiNA{sec: t.sec, nano: t.nano, atto: t.atto}
	}

	// Convert to nanoseconds since Unix epoch
	totalNanos := t.UnixNano()

	// Truncate to multiple of d
	dNanos := int64(d)
	truncatedNanos := (totalNanos / dNanos) * dNanos

	// Convert back to TaiNA (attoseconds preserved if truncating to >= nanosecond precision)
	result := TaiNAfromGoTime(time.Unix(0, truncatedNanos))
	if d >= time.Nanosecond {
		result.atto = 0 // Clear attoseconds if truncating to nanosecond or larger
	} else {
		result.atto = t.atto // Preserve attoseconds for sub-nanosecond truncation
	}

	return result
}

// Round returns the result of rounding t to the nearest multiple of d (since the zero time).
// The rounding behaviour for halfway values is to round up.
// If d <= 0, Round returns t stripped of any monotonic clock reading but otherwise unchanged.
func (t TaiNA) Round(d time.Duration) TaiNA {
	if d <= 0 {
		return TaiNA{sec: t.sec, nano: t.nano, atto: t.atto}
	}

	// Convert to nanoseconds since Unix epoch
	totalNanos := t.UnixNano()

	// Round to nearest multiple of d
	dNanos := int64(d)
	roundedNanos := ((totalNanos + dNanos/2) / dNanos) * dNanos

	// Convert back to TaiNA (attoseconds preserved if rounding to >= nanosecond precision)
	result := TaiNAfromGoTime(time.Unix(0, roundedNanos))
	if d >= time.Nanosecond {
		result.atto = 0 // Clear attoseconds if rounding to nanosecond or larger
	} else {
		result.atto = t.atto // Preserve attoseconds for sub-nanosecond rounding
	}

	return result
}

// String returns the TAI64NA string representation
func (t TaiNA) String() string {
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
func (t TaiNA) MarshalBinary() []byte {
	result := make([]byte, TAINALength)
	binary.BigEndian.PutUint64(result[:], t.sec)
	binary.BigEndian.PutUint32(result[TAILength:], t.nano)
	binary.BigEndian.PutUint32(result[TAINLength:], t.atto)
	return result
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *TaiNA) UnmarshalBinary(data []byte) error {
	if len(data) != TAINALength {
		return fmt.Errorf("%w: got %d, want %d", ErrInvalidTAINABinaryLength, len(data), TAINALength)
	}
	t.sec = binary.BigEndian.Uint64(data[:])
	t.nano = binary.BigEndian.Uint32(data[TAILength:])
	t.atto = binary.BigEndian.Uint32(data[TAINLength:])

	if t.atto > 999999999 {
		return ErrInvalidAttosecondRange
	}

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t TaiNA) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TaiNA) UnmarshalText(text []byte) error {
	taina, err := TaiNAfromString(string(text))
	if err != nil {
		return err
	}
	*t = taina
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t TaiNA) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TaiNA) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	taina, err := TaiNAfromString(s)
	if err != nil {
		return err
	}
	*t = taina
	return nil
}

// Since returns the time elapsed since t.
func (t TaiNA) Since(u TaiNA) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until t.
func (t TaiNA) Until(u TaiNA) time.Duration {
	return u.Sub(t)
}

// TaiNAfromString returns a TaiNA from an ASCII TAI64NA representation
//
//revive:disable-next-line:exported
func TaiNAfromString(str string) (TaiNA, error) {
	if len(str) != 33 || str[0] != '@' {
		return TaiNA{}, ErrInvalidTAINAFormat
	}

	var buf [16]byte
	_, err := hex.Decode(buf[:], []byte(str[1:]))
	if err != nil {
		return TaiNA{}, fmt.Errorf("invalid TAI64NA string format: %w", err)
	}

	sec := binary.BigEndian.Uint64(buf[:8])
	nano := binary.BigEndian.Uint32(buf[8:12])
	atto := binary.BigEndian.Uint32(buf[12:16])

	if atto > 999999999 {
		return TaiNA{}, ErrInvalidAttosecondRange
	}

	return TaiNA{
		sec:  sec,
		nano: nano,
		atto: atto,
	}, nil
}

// TaiNAfromGoTime returns a TaiNA from time.Time
// Note: Attoseconds are set to 0 since time.Time doesn't support attosecond precision
//
//revive:disable-next-line:exported
func TaiNAfromGoTime(t time.Time) TaiNA {
	return TaiNA{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
		atto: 0, // time.Time doesn't support attosecond precision
	}
}

// TaiNAfromTaiN returns a TaiNA from TaiN with zero attoseconds
//
//revive:disable-next-line:exported
func TaiNAfromTaiN(t TaiN) TaiNA {
	return TaiNA{
		sec:  t.sec,
		nano: t.nano,
		atto: 0,
	}
}
