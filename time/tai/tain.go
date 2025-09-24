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

// Error variables for TaiN operations
var (
	ErrUnsupportedLayout       = errors.New("unsupported layout")
	ErrInvalidTAINFormat       = errors.New("invalid TAI64N format: expected @XXXXXXXXXXXXXXXXXXXXXXXX")
	ErrInvalidTAINBinaryLength = errors.New("invalid TaiN binary data length")
)

// TaiN represents a TAI timestamp with nanosecond precision (TAIN)
//
//revive:disable-next-line:exported
type TaiN struct {
	sec  uint64
	nano uint32
}

// NowTaiN returns the current timestamp as a TaiN
func NowTaiN() TaiN {
	now := time.Now()
	return TaiN{
		sec:  TAICONST + lsoffset(now) + uint64(now.Unix()),
		nano: uint32(now.Nanosecond()),
	}
}

// Unix returns the TAI time corresponding to the given Unix time.
func Unix(sec int64, nsec int64) TaiN {
	t := time.Unix(sec, nsec)
	return TaiNfromGoTime(t)
}

// DateConfig holds parameters for creating a date
type DateConfig struct {
	Year, Day, Hour, Min, Sec, Nsec int
	Month                           time.Month
	Loc                             *time.Location
}

// Date returns the TAI time corresponding to the given date configuration
func Date(cfg DateConfig) TaiN {
	t := time.Date(cfg.Year, cfg.Month, cfg.Day, cfg.Hour, cfg.Min, cfg.Sec, cfg.Nsec, cfg.Loc)
	return TaiNfromGoTime(t)
}

// Parse parses a TAI64N formatted string.
func Parse(layout, value string) (TaiN, error) {
	if layout != "@XXXXXXXXXXXXXXXXXXXXXXXX" && layout != "tai64n" {
		return TaiN{}, fmt.Errorf("%w: %s", ErrUnsupportedLayout, layout)
	}
	return TaiNfromString(value)
}

// Add adds a time.Duration to a TaiN timestamp
func (t TaiN) Add(d time.Duration) TaiN {
	totalNanos := int64(d)

	addSecs := totalNanos / 1e9
	addNanos := totalNanos % 1e9

	var result TaiN

	if addSecs >= 0 && addNanos >= 0 {
		// Positive duration
		result.sec = t.sec + uint64(addSecs)
		newNanos := int64(t.nano) + addNanos
		if newNanos >= 1e9 {
			result.sec++
			result.nano = uint32(newNanos - 1e9)
		} else {
			result.nano = uint32(newNanos)
		}
	} else {
		// Negative duration - handle underflow
		result.sec = t.sec - uint64(-addSecs)
		newNanos := int64(t.nano) + addNanos // addNanos is negative
		if newNanos < 0 {
			result.sec-- // Will wrap around if underflows
			result.nano = uint32(newNanos + 1e9)
		} else {
			result.nano = uint32(newNanos)
		}
	}

	return result
}

// Sub subtracts two TimeN timestamps
func (t TaiN) Sub(u TaiN) time.Duration {
	secDiff := int64(t.sec - u.sec)
	nanoDiff := int64(t.nano) - int64(u.nano)

	if nanoDiff < 0 {
		secDiff--
		nanoDiff += 1e9
	}

	return time.Duration(secDiff)*time.Second + time.Duration(nanoDiff)*time.Nanosecond
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t TaiN) Unix() int64 {
	return int64(t.sec - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
func (t TaiN) UnixMilli() int64 {
	return t.Unix()*1000 + int64(t.nano)/1e6
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
func (t TaiN) UnixMicro() int64 {
	return t.Unix()*1e6 + int64(t.nano)/1e3
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
func (t TaiN) UnixNano() int64 {
	return t.Unix()*1e9 + int64(t.nano)
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t TaiN) Nanosecond() int {
	return int(t.nano)
}

// Before reports whether the time instant t is before u.
func (t TaiN) Before(u TaiN) bool {
	if t.sec != u.sec {
		return t.sec < u.sec
	}
	return t.nano < u.nano
}

// After reports whether the time instant t is after u.
func (t TaiN) After(u TaiN) bool {
	if t.sec != u.sec {
		return t.sec > u.sec
	}
	return t.nano > u.nano
}

// Equal reports whether t and u represent the same time instant.
func (t TaiN) Equal(u TaiN) bool {
	return t.sec == u.sec && t.nano == u.nano
}

// IsZero reports whether t represents the zero time instant.
func (t TaiN) IsZero() bool {
	return t.sec == 0 && t.nano == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t TaiN) Compare(u TaiN) int {
	if t.sec < u.sec || (t.sec == u.sec && t.nano < u.nano) {
		return -1
	}
	if t.sec > u.sec || (t.sec == u.sec && t.nano > u.nano) {
		return 1
	}
	return 0
}

// GoTime returns a time.Time representation of the TimeN timestamp.
func (t TaiN) GoTime() time.Time {
	tm := time.Unix(int64(t.sec-TAICONST), int64(t.nano)).UTC()
	return tm.Add(-time.Duration(lsoffset(tm)) * time.Second)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
func (t TaiN) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// Tai converts TaiN to Tai by truncating nanoseconds.
func (t TaiN) Tai() Tai {
	return Tai{x: t.sec}
}

// Truncate returns the result of rounding t down to a multiple of d (since the zero time).
// If d <= 0, Truncate returns t stripped of any monotonic clock reading but otherwise unchanged.
func (t TaiN) Truncate(d time.Duration) TaiN {
	if d <= 0 {
		return t
	}

	// Convert to nanoseconds since Unix epoch
	totalNanos := t.UnixNano()

	// Truncate to multiple of d
	dNanos := int64(d)
	truncatedNanos := (totalNanos / dNanos) * dNanos

	// Convert back to TimeN
	return TaiNfromGoTime(time.Unix(0, truncatedNanos))
}

// Round returns the result of rounding t to the nearest multiple of d (since the zero time).
// The rounding behaviour for halfway values is to round up.
// If d <= 0, Round returns t stripped of any monotonic clock reading but otherwise unchanged.
func (t TaiN) Round(d time.Duration) TaiN {
	if d <= 0 {
		return t
	}

	// Convert to nanoseconds since Unix epoch
	totalNanos := t.UnixNano()

	// Round to nearest multiple of d
	dNanos := int64(d)
	roundedNanos := ((totalNanos + dNanos/2) / dNanos) * dNanos

	// Convert back to TimeN
	return TaiNfromGoTime(time.Unix(0, roundedNanos))
}

// String returns the TAI64N string representation
func (t TaiN) String() string {
	var buf [25]byte
	var binBuf [12]byte
	buf[0] = '@'
	binary.BigEndian.PutUint64(binBuf[:8], t.sec)
	binary.BigEndian.PutUint32(binBuf[8:], t.nano)
	hex.Encode(buf[1:], binBuf[:])
	return string(buf[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t TaiN) MarshalBinary() []byte {
	result := make([]byte, TAINLength)
	binary.BigEndian.PutUint64(result[:], t.sec)
	binary.BigEndian.PutUint32(result[TAILength:], t.nano)
	return result
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *TaiN) UnmarshalBinary(data []byte) error {
	if len(data) != TAINLength {
		return fmt.Errorf("%w: got %d, want %d", ErrInvalidTAINBinaryLength, len(data), TAINLength)
	}
	t.sec = binary.BigEndian.Uint64(data[:])
	t.nano = binary.BigEndian.Uint32(data[TAILength:])
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t TaiN) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TaiN) UnmarshalText(text []byte) error {
	tain, err := TaiNfromString(string(text))
	if err != nil {
		return err
	}
	*t = tain
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t TaiN) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TaiN) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	tain, err := TaiNfromString(s)
	if err != nil {
		return err
	}
	*t = tain
	return nil
}

// Since returns the time elapsed since t.
func (t TaiN) Since(u TaiN) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until t.
func (t TaiN) Until(u TaiN) time.Duration {
	return u.Sub(t)
}

// TaiNfromString returns a TaiN from an ASCII TAI64N representation
//
//revive:disable-next-line:exported
func TaiNfromString(str string) (TaiN, error) {
	if len(str) != 25 || str[0] != '@' {
		return TaiN{}, ErrInvalidTAINFormat
	}

	var buf [12]byte
	_, err := hex.Decode(buf[:], []byte(str[1:]))
	if err != nil {
		return TaiN{}, fmt.Errorf("invalid TAI64N string format: %w", err)
	}

	return TaiN{
		sec:  binary.BigEndian.Uint64(buf[:8]),
		nano: binary.BigEndian.Uint32(buf[8:]),
	}, nil
}

// TaiNfromGoTime returns a TaiN from time.Time
//
//revive:disable-next-line:exported
func TaiNfromGoTime(t time.Time) TaiN {
	return TaiN{
		sec:  TAICONST + lsoffset(t) + uint64(t.Unix()),
		nano: uint32(t.Nanosecond()),
	}
}
