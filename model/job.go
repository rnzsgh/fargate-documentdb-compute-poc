package model

import (
	"context"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failure" bson:"failure"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}

func updateJobFailureReason(id *primitive.ObjectID, reason string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", id}},
		bson.D{{"$set", bson.D{{"failure", reason}}}})
	return
}

func updateJobStopTime(id *primitive.ObjectID) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", id}},
		bson.D{{"$set", bson.D{{"stop", time.Now()}}}})
	return
}

func createJobEntry(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").InsertOne(ctx, job)
	return
}

func findJobById(id *primitive.ObjectID) (job *Job, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	job = &Job{}
	err = docdb.Client.Database("work").Collection("jobs").FindOne(ctx, bson.D{{"_id", id}}).Decode(job)
	return
}
