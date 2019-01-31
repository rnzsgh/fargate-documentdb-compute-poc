package model

import (
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var taskTestJobId = primitive.NewObjectID()

var testJob *Job

func TestCreateJobWithTasksEntry(t *testing.T) {
	now := time.Now()
	testJob = &Job{Id: &taskTestJobId, Start: &now, Stop: &now}
	testJob.Tasks = make(map[string]*Task)
	for i := 0; i < 2; i++ {
		taskId := primitive.NewObjectID()
		testJob.Tasks[taskId.Hex()] = &Task{Id: &taskId, JobId: &taskTestJobId}
	}
	if err := CreateJob(testJob); err != nil {
		t.Errorf("Problem creating job entry for test task: %v", err)
	}
}

func TestUpdateTaskFailureReason(t *testing.T) {
	t.Run("TestUpdateTaskFailureReason", func(t *testing.T) {
		for _, task := range testJob.Tasks {
			if err := UpdateTaskFailureReason(task, "FAILED"); err != nil {
				t.Errorf("Problem updating task failure reason: %v", err)
			}
		}

		if job, err := FindJobById(&taskTestJobId); err != nil {
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

func TestUpdateTaskStopTime(t *testing.T) {
	t.Run("TestUpdateTaskStopTime", func(t *testing.T) {
		for _, task := range testJob.Tasks {
			now := time.Now()
			task.Stop = &now
			if err := UpdateTaskStopTime(task); err != nil {
				t.Errorf("Problem updating task stop time - reason: %v", err)
			}
		}

		if job, err := FindJobById(&taskTestJobId); err != nil {
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
