package service

import "time"

// Clock supplies the current time. Domain rules never read the wall clock
// themselves; they take an injected now func() time.Time, and this is where
// that function comes from. Tests substitute a fixed Clock.
type Clock func() time.Time

// SystemClock reads the real wall clock.
func SystemClock() time.Time { return time.Now() }

// FixedClock returns a Clock that always reports at.
func FixedClock(at time.Time) Clock {
	return func() time.Time { return at }
}
