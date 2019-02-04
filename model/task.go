package model

import (
	"context"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
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

	count := 0

	for {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := jobCollection().UpdateOne(
			ctx,
			bson.D{{"_id", task.JobId}},
			bson.D{{"$set", bson.D{{fmt.Sprintf("tasks.%s.%s", task.Id.Hex(), field), value}}}})

		if err == nil {
			if res.MatchedCount != 1 && res.ModifiedCount != 1 {
				return fmt.Errorf(
					"Task field not updated - job: %s - task %s - field: %s",
					task.JobId.Hex(),
					task.Id.Hex(),
					field,
				)
			}
			return nil
		}

		log.Errorf(
			"Task field not updated - job: %s - task %s - field: %s - reason %v",
			task.JobId.Hex(),
			task.Id.Hex(),
			field,
			err,
		)

		count++

		time.Sleep(time.Duration(count*2) * time.Second)
		if count == taskMaxFailureRetry {
			return err
		}
	}
}
