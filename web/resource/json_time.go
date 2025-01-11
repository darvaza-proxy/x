package resource

import (
	"time"

	"darvaza.org/core"
)

// RFC9557 if the formatting used by [JSONTime] when
// encoding or decoding JSON.
const RFC9557 = "2006-01-02T15:04:05.999Z07:00"

// jsonRFC9557 is a quoted variant of [RFC9557].
const jsonRFC9557 = "\"" + RFC9557 + "\""

// JSONTime is a [time.Time] but using an [RFC9557]
// with millisecond accuracy, compatible with ISO8601.
type JSONTime struct {
	time.Time
}

// MarshalJSON encodes the timestamp for JSON.
func (ts *JSONTime) MarshalJSON() ([]byte, error) {
	if ts == nil {
		return nil, core.ErrNilReceiver
	}

	// truncate to milliseconds
	t2 := ts.Time.Truncate(time.Millisecond)

	// encode quoted
	return []byte(t2.Format(jsonRFC9557)), nil
}

// UnmarshalJSON decodes the timestamp from JSON.
func (ts *JSONTime) UnmarshalJSON(b []byte) error {
	if ts == nil {
		return core.ErrNilReceiver
	}

	s := string(b)
	// quoted
	t, err := time.Parse(jsonRFC9557, s)
	if err == nil {
		ts.Time = t
		return nil
	}

	// unquoted
	t, err = time.Parse(RFC9557, s)
	if err == nil {
		ts.Time = t
		return nil
	}

	return err
}

// NewJSONTime returns a new [JSONTime] set
// to a copy of the given time, or the current
// UTC timestamp if none is provided.
func NewJSONTime(tp *time.Time) JSONTime {
	var t time.Time
	if tp != nil {
		t = *tp
	}

	if t.IsZero() {
		t = time.Now().UTC()
	}

	return JSONTime{Time: t}
}
