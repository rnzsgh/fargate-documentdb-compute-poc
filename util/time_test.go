package util

import (
	"testing"
)

func TestNowTimeUtc(t *testing.T) {
	t.Run("TestNowTimeUtc", func(t *testing.T) {
		if NowTimeUtc() == nil {
			t.Errorf("Now time in utc did not return a value")
		}
	})
}

func TestWaitSeconds(t *testing.T) {
	t.Run("TestWaitSeconds", func(t *testing.T) {

		completed := make(chan bool)

		t0 := NowTimeUtc()

		var next int64

		go func() {
			next = WaitSeconds(1)
			completed <- true
		}()

		for {

			select {
			case _ = <-completed:
				t1 := NowTimeUtc()
				if t1.Sub(*t0) < 1 {
					t.Errorf("Wait time in seconds did not sleep long enough")
				}
				if next != 2 {
					t.Errorf("Wait time in seconds did not increment the value properly")
				}
				return
			}
		}
	})
}
