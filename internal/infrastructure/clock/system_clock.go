package clock

import "time"

type SystemClock struct{}

// Now returns the current wall-clock time.
// Now 返回当前系统墙钟时间。
func (SystemClock) Now() time.Time {
	return time.Now()
}
