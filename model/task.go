package model

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

const taskMaxFailureRetry = 3

type Task struct {
	Id     *primitive.ObjectID `json:"id" bson:"id"`
	JobId  *primitive.ObjectID `json:"jobId" bson:"jobId"`
	DataId *primitive.ObjectID `json:"dataId" bson:"dataId"`
	Start  *time.Time          `json:"start" bson:"start"`
	Stop   *time.Time          `json:"stop" bson:"stop"`
}

func TaskUpdateStopTime(task *Task) error {
	task.Stop = util.TimeNowUtc()
	return taskUpdateField(task, "stop", task.Stop)
}

func taskUpdateField(task *Task, field string, value interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return docdb.UpdateOneFieldById(
		ctx,
		jobCollection(),
		task.JobId,
		fmt.Sprintf("tasks.%s.%s", task.Id.Hex(), field),
		value,
		taskMaxFailureRetry,
	)
}
