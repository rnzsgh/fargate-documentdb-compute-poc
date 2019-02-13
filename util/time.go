package util

import (
	"time"
)

// Wait for a period of time and then increment the seconds by two and return
// the new value.
func TimeWaitSeconds(seconds int64) int64 {
	time.Sleep(time.Duration(seconds) * time.Second)
	return seconds * int64(2)
}

func TimeNowUtc() *time.Time {
	now := time.Now().UTC()
	return &now
}

// Add the param seconds to the current time. If seconds is less than
// zero, this method panics.
func TimeNowUtcPlusSeconds(seconds int64) *time.Time {
	if seconds < 0 {
		panic("Seconds must be equal to or greater than zero")
	}

	now := time.Now().UTC()
	now = now.Add(time.Duration(seconds) * time.Second)
	return &now
}

// Sleeps for two seconds, then four, etc., each time the closure is
// called. Closure function returns the count/times it has slept.
func TimeExpoentialSleepSeconds() func() int {
	count := 0
	return func() int {
		count++
		time.Sleep(time.Duration(count*2) * time.Second)
		return count
	}
}
