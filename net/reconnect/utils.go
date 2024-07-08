package reconnect

import "time"

// TimeoutToAbsoluteTime adds the given [time.Duration] to a
// base [time.Time].
// if the duration is negative, a zero [time.Time] will
// be returned.
// if the base is zero, the current time will be used.
func TimeoutToAbsoluteTime(base time.Time, d time.Duration) time.Time {
	if d > 0 {
		if base.IsZero() {
			base = time.Now()
		}

		return base.Add(d)
	}

	return time.Time{} // isZero()
}
