package model

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

type Task struct {
	Id            *primitive.ObjectID `json:"id" bson:"id"`
	JobId         *primitive.ObjectID `json:"jobId" bson:"jobId"`
	FailureReason string              `json:"failure" bson:"failure"`
	Arn           string              `json:"arn,omitempty" bson:"arn,omitempty"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}

// Update a field in the specfic task
func taskUpdateField(task *Task, field string, value interface{}) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", task.JobId}},
		bson.D{{"$set", bson.D{{fmt.Sprintf("tasks.%s.%s", task.Id.Hex(), field), value}}}})
	return
}

func TaskUpdateArn(task *Task, arn string) error {
	return taskUpdateField(task, "arn", arn)
}

func TaskUpdateFailureReason(task *Task, reason string) (err error) {
	return taskUpdateField(task, "failure", reason)
}

func TaskUpdateStopTime(task *Task) (err error) {
	return taskUpdateField(task, "stop", task.Stop)
}
