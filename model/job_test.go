package model

import (
	"context"
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var testJobId = primitive.NewObjectID()

func TestCleanup(t *testing.T) {
	t.Run("TestCleanup", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, err := jobCollection().DeleteMany(ctx, &bson.D{}); err != nil {
			t.Errorf("Unable to delete jobs: %v", err)
		}
	})
}

func TestJobCreate(t *testing.T) {
	t.Run("TestJobCreate", func(t *testing.T) {
		now := time.Now()
		if err := JobCreate(&Job{Id: &testJobId, Start: &now}); err != nil {
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

func TestJobFindRunning(t *testing.T) {
	t.Run("TestJobFindRunning", func(t *testing.T) {
		if jobs, err := JobFindRunning(); err != nil {
			t.Errorf("Problem finding running jobs: %v", err)
		} else {
			if len(jobs) != 1 {
				t.Errorf("Running jobs not found - expected: 1 - received: %d", len(jobs))
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
