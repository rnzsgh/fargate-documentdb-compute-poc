package model

import (
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

var taskTestJobId = primitive.NewObjectID()

var testJob *Job

func TestCreateJobWithTasksEntry(t *testing.T) {
	t.Run("TestCreateJobWithTasksEntry", func(t *testing.T) {
		now := util.TimeNowUtc()
		testJob = &Job{Id: &taskTestJobId, Start: now, Stop: now}
		testJob.Tasks = make(map[string]*Task)
		for i := 0; i < 2; i++ {
			taskId := primitive.NewObjectID()
			testJob.Tasks[taskId.Hex()] = &Task{Id: &taskId, JobId: &taskTestJobId}
		}
		if err := JobCreate(testJob); err != nil {
			t.Errorf("Problem creating job entry for test task: %v", err)
		}

	})
}

func TestTaskUpdateFailureReason(t *testing.T) {
	t.Run("TestTaskUpdateFailureReason", func(t *testing.T) {
		for _, task := range testJob.Tasks {
			if err := TaskUpdateFailureReason(task, "FAILED"); err != nil {
				t.Errorf("Problem updating task failure reason: %v", err)
			}
		}

		if job, err := JobFindById(&taskTestJobId); err != nil {
			t.Errorf("Cannot load job entry: %v", err)
		} else {
			for _, task := range job.Tasks {
				if task.FailureReason != "FAILED" {
					t.Errorf("Failed to update the task failure reaason - expected: FAILED - recevied: %s", task.FailureReason)
				}
			}
		}
	})
}

func TestTaskUpdateArn(t *testing.T) {
	t.Run("TestTaskUpdateArn", func(t *testing.T) {
		for _, task := range testJob.Tasks {
			now := time.Now()
			task.Stop = &now
			if err := TaskUpdateArn(task, "test"); err != nil {
				t.Errorf("Problem updating task arn - reason: %v", err)
			}
		}

		if job, err := JobFindById(&taskTestJobId); err != nil {
			t.Errorf("Cannot load job entry: %v", err)
		} else {
			for _, task := range job.Tasks {
				if len(task.Arn) == 0 {
					t.Errorf("Failed to update the task arn")
				}
			}
		}
	})
}

func TestTaskUpdateStopTime(t *testing.T) {
	t.Run("TestTaskUpdateStopTime", func(t *testing.T) {
		for _, task := range testJob.Tasks {
			task.Stop = util.TimeNowUtc()
			if err := TaskUpdateStopTime(task); err != nil {
				t.Errorf("Problem updating task stop time - reason: %v", err)
			}
		}

		if job, err := JobFindById(&taskTestJobId); err != nil {
			t.Errorf("Cannot load job entry: %v", err)
		} else {
			for _, task := range job.Tasks {
				if task.Stop == nil {
					t.Errorf("Failed to update the task stop time")
				}
			}
		}
	})
}
