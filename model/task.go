package model

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

const taskMaxFailureRetry = 3

type Task struct {
	Id            *primitive.ObjectID `json:"id" bson:"id"`
	JobId         *primitive.ObjectID `json:"jobId" bson:"jobId"`
	FailureReason string              `json:"failure" bson:"failure"`
	DataId        *primitive.ObjectID `json:"dataId" bson:"dataId"`
	Arn           string              `json:"arn" bson:"arn"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop" bson:"stop"`
}

func TaskUpdateArn(task *Task, arn string) error {
	return taskUpdateField(task, "arn", arn)
}

func TaskUpdateFailureReason(task *Task, reason string) error {
	return taskUpdateField(task, "failure", reason)
}

func TaskUpdateStopTime(task *Task) error {
	return taskUpdateField(task, "stop", task.Stop)
}

func taskUpdateField(task *Task, field string, value interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return docdb.UpdateOneFieldById(
		ctx,
		jobCollection(),
		task.Id,
		fmt.Sprintf("tasks.%s.%s", task.Id.Hex(), field),
		value,
		taskMaxFailureRetry,
	)
}
