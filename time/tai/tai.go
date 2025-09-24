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

// TAICONST is 2^62+10 representing the TAI label of the second Unix started
// 1970-01-01 00:00:00 +0000 UTC
const TAICONST = uint64(4611686018427387914)

// TAILength is the length of a TAI timestamp in bytes
const TAILength = 8

// TAINLength is the length of a TAIN timestamp in bytes
const TAINLength = 12

// TAINALength is the length of a TAINA timestamp in bytes
const TAINALength = 16

// Error variables for common error conditions
var (
	ErrInvalidTAIFormat       = errors.New("invalid TAI format: expected @XXXXXXXXXXXXXXXX")
	ErrInvalidTAILength       = errors.New("invalid TAI length")
	ErrInvalidTAIBinaryLength = errors.New("invalid TAI binary data length")
)

// Tai represents a TAI (International Atomic Time) timestamp
type Tai struct {
	x uint64
}

// ParseTai parses a TAI64 formatted string.
func ParseTai(value string) (Tai, error) {
	return TaifromString(value)
}

// Add adds a time.Duration to a TAI timestamp
func (t Tai) Add(d time.Duration) Tai {
	seconds := int64(d / time.Second)
	if seconds >= 0 {
		return Tai{x: t.x + uint64(seconds)}
	}
	return Tai{x: t.x - uint64(-seconds)}
}

// Sub subtracts two TAI timestamps
func (t Tai) Sub(u Tai) time.Duration {
	x := t.x - u.x
	return time.Duration(x) * time.Second
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t Tai) Unix() int64 {
	return int64(t.x - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
func (t Tai) UnixMilli() int64 {
	return t.Unix() * 1000
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
func (t Tai) UnixMicro() int64 {
	return t.Unix() * 1000000
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
func (t Tai) UnixNano() int64 {
	return t.Unix() * 1000000000
}

// Before reports whether the time instant t is before u.
func (t Tai) Before(u Tai) bool {
	return t.x < u.x
}

// After reports whether the time instant t is after u.
func (t Tai) After(u Tai) bool {
	return t.x > u.x
}

// Equal reports whether t and u represent the same time instant.
func (t Tai) Equal(u Tai) bool {
	return t.x == u.x
}

// IsZero reports whether t represents the zero time instant.
func (t Tai) IsZero() bool {
	return t.x == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t Tai) Compare(u Tai) int {
	if t.x < u.x {
		return -1
	}
	if t.x > u.x {
		return 1
	}
	return 0
}

// GoTime returns a time.Time representation of the TAI timestamp.
func (t Tai) GoTime() time.Time {
	tm := time.Unix(int64(t.x-TAICONST), 0).UTC()
	return tm.Add(-time.Duration(lsoffset(tm)) * time.Second)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
func (t Tai) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// TaiN converts TAI to TaiN with zero nanoseconds.
func (t Tai) TaiN() TaiN {
	return TaiN{sec: t.x, nano: 0}
}

// String returns the TAI64 string representation
func (t Tai) String() string {
	var buf [17]byte
	var binBuf [8]byte
	buf[0] = '@'
	binary.BigEndian.PutUint64(binBuf[:], t.x)
	hex.Encode(buf[1:], binBuf[:])
	return string(buf[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t Tai) MarshalBinary() []byte {
	result := make([]byte, TAILength)
	binary.BigEndian.PutUint64(result[:], t.x)
	return result
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *Tai) UnmarshalBinary(data []byte) error {
	if len(data) != TAILength {
		return fmt.Errorf("%w: got %d, want %d", ErrInvalidTAIBinaryLength, len(data), TAILength)
	}
	t.x = binary.BigEndian.Uint64(data[:])
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t Tai) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *Tai) UnmarshalText(text []byte) error {
	tai, err := TaifromString(string(text))
	if err != nil {
		return err
	}
	*t = tai
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t Tai) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Tai) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	tai, err := TaifromString(s)
	if err != nil {
		return err
	}
	*t = tai
	return nil
}

// Since returns the time elapsed since t.
func (t Tai) Since(u Tai) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until t.
func (t Tai) Until(u Tai) time.Duration {
	return u.Sub(t)
}

// TaifromString returns a Time from an ASCII TAI representation
func TaifromString(str string) (Tai, error) {
	if len(str) != 17 || str[0] != '@' {
		return Tai{}, ErrInvalidTAIFormat
	}

	var buf [8]byte
	_, err := hex.Decode(buf[:], []byte(str[1:]))
	if err != nil {
		return Tai{}, fmt.Errorf("invalid TAI string format: %w", err)
	}

	return Tai{x: binary.BigEndian.Uint64(buf[:])}, nil
}

// TaifromGoTime returns a Time from time.Time
func TaifromGoTime(t time.Time) Tai {
	return Tai{x: TAICONST + lsoffset(t) + uint64(t.Unix())}
}
