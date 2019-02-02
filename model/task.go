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
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}

func TaskUpdateFailureReason(task *Task, reason string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", task.JobId}},
		bson.D{{"$set", bson.D{{fmt.Sprintf("tasks.%s.failure", task.Id.Hex()), reason}}}})
	return
}

func TaskUpdateStopTime(task *Task) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", task.JobId}},
		bson.D{{"$set", bson.D{{fmt.Sprintf("tasks.%s.stop", task.Id.Hex()), task.Stop}}}})
	return
}
