package model

import (
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var testJobId = primitive.NewObjectID()

func TestJobCreate(t *testing.T) {
	t.Run("TestJobCreate", func(t *testing.T) {
		now := time.Now()
		if err := JobCreate(&Job{Id: &testJobId, Start: &now, Stop: &now}); err != nil {
			t.Errorf("Problem creating job entry: %v", err)
		} else {
			if j, err := JobFindById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else if j == nil {
				t.Error("Null job entry returned")
			}
		}
	})
}

func TestJobUpdateFailureReason(t *testing.T) {
	t.Run("TestJobUpdateFailureReason", func(t *testing.T) {
		if err := JobUpdateFailureReason(&testJobId, "FAILED"); err != nil {
			t.Errorf("Problem updating job failure reason: %v", err)
		} else {
			if j, err := JobFindById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else {
				if j.FailureReason != "FAILED" {
					t.Errorf("Job failure reason did not update - expected: FAILED - received: %s", j.FailureReason)
				}
			}
		}
	})
}

func TestJobUpdateStopTime(t *testing.T) {
	t.Run("TestJobUpdateStopTime", func(t *testing.T) {
		if err := JobUpdateStopTime(&testJobId); err != nil {
			t.Errorf("Problem updating job stop time: %v", err)
		} else {
			if j, err := JobFindById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else {
				if j.Stop == nil {
					t.Errorf("Job stop time did not update")
				}
			}
		}
	})
}
