package model

import (
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var testJobId = primitive.NewObjectID()

func TestCreateJob(t *testing.T) {
	t.Run("TestCreateJob", func(t *testing.T) {
		now := time.Now()
		if err := CreateJob(&Job{Id: &testJobId, Start: &now, Stop: &now}); err != nil {
			t.Errorf("Problem creating job entry: %v", err)
		} else {
			if j, err := FindJobById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else if j == nil {
				t.Error("Null job entry returned")
			}
		}
	})
}

func TestUpdateJobFailureReason(t *testing.T) {
	t.Run("TestUpdateJobFailureReason", func(t *testing.T) {
		if err := UpdateJobFailureReason(&testJobId, "FAILED"); err != nil {
			t.Errorf("Problem updating job failure reason: %v", err)
		} else {
			if j, err := FindJobById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else {
				if j.FailureReason != "FAILED" {
					t.Errorf("Job failure reason did not update - expected: FAILED - received: %s", j.FailureReason)
				}
			}
		}
	})
}

func TestUpdateJobStopTime(t *testing.T) {
	t.Run("TestUpdateJobStopTime", func(t *testing.T) {
		if err := UpdateJobStopTime(&testJobId); err != nil {
			t.Errorf("Problem updating job stop time: %v", err)
		} else {
			if j, err := FindJobById(&testJobId); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else {
				if j.Stop == nil {
					t.Errorf("Job stop time did not update")
				}
			}
		}
	})
}
