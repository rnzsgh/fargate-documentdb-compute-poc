package util

import (
	"testing"
)

func TestTimeNowUtc(t *testing.T) {
	t.Run("TestTimeNowUtc", func(t *testing.T) {
		if TimeNowUtc() == nil {
			t.Errorf("Now time in utc did not return a value")
		}
	})
}

func TestTimeWaitSeconds(t *testing.T) {
	t.Run("TestTimeWaitSeconds", func(t *testing.T) {

		completed := make(chan bool)

		t0 := TimeNowUtc()

		var next int64

		go func() {
			next = TimeWaitSeconds(1)
			completed <- true
		}()

		for {

			select {
			case _ = <-completed:
				t1 := TimeNowUtc()
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
