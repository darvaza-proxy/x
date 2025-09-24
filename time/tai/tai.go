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

// TAICONST is 2^62+10 representing the TAI label of the second Unix started
// 1970-01-01 00:00:00 +0000 UTC
const TAICONST = uint64(4611686018427387914)

// TAILength is the length of a TAI timestamp in bytes
const TAILength = 8

// TAINLength is the length of a TAIN timestamp in bytes
const TAINLength = 12

// TAINALength is the length of a TAINA timestamp in bytes
const TAINALength = 16

// TAI represents a TAI timestamp
type TAI struct {
	x uint64
}

var (
	_ encoding.BinaryMarshaler   = TAI{}
	_ encoding.BinaryUnmarshaler = (*TAI)(nil)
	_ encoding.TextMarshaler     = TAI{}
	_ encoding.TextUnmarshaler   = (*TAI)(nil)
	_ json.Marshaler             = TAI{}
	_ json.Unmarshaler           = (*TAI)(nil)
)

// NowTAI returns the current timestamp as a TAI
func NowTAI() TAI {
	return TAIFromTime(time.Now())
}

// ParseTAI parses a TAI64 formatted string.
func ParseTAI(value string) (TAI, error) {
	if len(value) != 2*TAILength+1 {
		return TAI{}, core.Wrapf(ErrInvalidTAILength, "got %d, want %d", len(value), 2*TAILength+1)
	}
	if value[0] != '@' {
		return TAI{}, ErrInvalidTAIFormat
	}

	var buf [8]byte
	_, err := hex.Decode(buf[:], []byte(value[1:]))
	if err != nil {
		return TAI{}, core.Wrap(err, "invalid TAI string format")
	}

	return TAI{x: binary.BigEndian.Uint64(buf[:])}, nil
}

// Add adds a time.Duration to a TAI timestamp. The duration is
// truncated to whole seconds, discarding any sub-second component.
// It returns ErrOverflow or ErrUnderflow if the result doesn't fit.
func (t TAI) Add(d time.Duration) (TAI, error) {
	seconds := int64(d / time.Second)
	if seconds >= 0 {
		x, err := safeAddUint64(t.x, uint64(seconds))
		if err != nil {
			return TAI{}, err
		}
		return TAI{x: x}, nil
	}

	x, err := safeSubUint64(t.x, uint64(-seconds))
	if err != nil {
		return TAI{}, err
	}
	return TAI{x: x}, nil
}

// Sub subtracts two TAI timestamps, in whole seconds.
// The result is clamped to the representable time.Duration range rather
// than being allowed to wrap or change sign.
func (t TAI) Sub(u TAI) time.Duration {
	neg := t.x < u.x

	var absDiff uint64
	if neg {
		absDiff = u.x - t.x
	} else {
		absDiff = t.x - u.x
	}

	// maxSecs is the largest second count whose conversion to
	// time.Duration (nanoseconds) still fits in int64.
	const maxSecs = uint64(math.MaxInt64) / uint64(time.Second)

	if absDiff > maxSecs {
		if neg {
			return minDuration
		}
		return maxDuration
	}

	d := time.Duration(absDiff) * time.Second
	if neg {
		return -d
	}
	return d
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t TAI) Unix() int64 {
	return int64(t.x - TAICONST)
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
// The result is undefined if it does not fit in an int64.
func (t TAI) UnixMilli() int64 {
	return t.Unix() * 1000
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
// The result is undefined if it does not fit in an int64.
func (t TAI) UnixMicro() int64 {
	return t.Unix() * 1000000
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
// !!! The result is undefined if it does not fit in an int64
// (a date before the year 1678 or after 2262).
func (t TAI) UnixNano() int64 {
	return t.Unix() * 1000000000
}

// Before reports whether the time instant t is before u.
func (t TAI) Before(u TAI) bool {
	return t.x < u.x
}

// After reports whether the time instant t is after u.
func (t TAI) After(u TAI) bool {
	return t.x > u.x
}

// Equal reports whether t and u represent the same time instant.
func (t TAI) Equal(u TAI) bool {
	return t.x == u.x
}

// IsZero reports whether t represents the zero time instant.
func (t TAI) IsZero() bool {
	return t.x == 0
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t TAI) Compare(u TAI) int {
	if t.x < u.x {
		return -1
	}
	if t.x > u.x {
		return 1
	}
	return 0
}

// GoTime returns a time.Time representation of the TAI timestamp.
// Instants inside an inserted leap second cannot be represented by
// time.Time and map to the first second after it.
func (t TAI) GoTime() time.Time {
	tm := time.Unix(int64(t.x-TAICONST), 0).UTC()
	return utcFromTAI(tm)
}

// Format returns a textual representation of the time value formatted
// according to layout by converting to time.Time first.
func (t TAI) Format(layout string) string {
	return t.GoTime().Format(layout)
}

// TAIN converts TAI to TAIN with zero nanoseconds.
func (t TAI) TAIN() TAIN {
	return TAIN{sec: t.x, nano: 0}
}

// TAINA converts TAI to TAINA with zero nanoseconds and attoseconds.
func (t TAI) TAINA() TAINA {
	return TAINA{sec: t.x}
}

// String returns the TAI64 string representation
func (t TAI) String() string {
	var buf [17]byte
	var binBuf [8]byte
	buf[0] = '@'
	binary.BigEndian.PutUint64(binBuf[:], t.x)
	hex.Encode(buf[1:], binBuf[:])
	return string(buf[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t TAI) MarshalBinary() ([]byte, error) {
	result := make([]byte, TAILength)
	binary.BigEndian.PutUint64(result[:], t.x)
	return result, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *TAI) UnmarshalBinary(data []byte) error {
	if len(data) != TAILength {
		return core.Wrapf(ErrInvalidTAIBinaryLength, "got %d, want %d", len(data), TAILength)
	}
	t.x = binary.BigEndian.Uint64(data[:])
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t TAI) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TAI) UnmarshalText(text []byte) error {
	tai, err := ParseTAI(string(text))
	if err != nil {
		return err
	}
	*t = tai
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t TAI) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TAI) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	tai, err := ParseTAI(s)
	if err != nil {
		return err
	}
	*t = tai
	return nil
}

// Since returns the time elapsed since u, as observed at t.
func (t TAI) Since(u TAI) time.Duration {
	return t.Sub(u)
}

// Until returns the duration until u, as observed at t.
func (t TAI) Until(u TAI) time.Duration {
	return u.Sub(t)
}

// TAIFromTime returns a TAI from time.Time
//
//revive:disable-next-line:exported
func TAIFromTime(t time.Time) TAI {
	return TAI{x: TAICONST + lsoffset(t) + uint64(t.Unix())}
}
