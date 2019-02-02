package util

import (
	"time"
)

// Wait for a period of time and then increment the seconds by two and return
// the new value.
func WaitSeconds(seconds int64) int64 {
	time.Sleep(time.Duration(seconds) * time.Second)
	return seconds * int64(2)
}

func NowTimeUtc() *time.Time {
	time := time.Now().UTC()
	return &time
}
