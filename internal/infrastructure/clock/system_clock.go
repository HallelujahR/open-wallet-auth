package clock

import "time"

type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time {
	return time.Now()
}
