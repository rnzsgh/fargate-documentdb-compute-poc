package model

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

const jobMaxFailureRetry = 3

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failure" bson:"failure"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop" bson:"stop"`
}

func JobCreate(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = jobCollection().InsertOne(ctx, job)
	return
}

func JobExists(id *primitive.ObjectID) (bool, error) {
	job, err := JobFindById(id)
	if err != nil {
		return false, err
	}

	if job != nil {
		return true, nil
	}

	return false, nil
}

func JobFindById(id *primitive.ObjectID) (*Job, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	job := &Job{}

	if err := docdb.FindById(
		ctx,
		jobCollection(),
		id,
		job,
	); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("Unable to find job by id: %s - reason: %v", id.Hex(), err)
	}

	return job, nil
}

// Called when the server starts to load currently running jobs
func JobFindRunning() ([]*Job, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, err := jobCollection().Find(ctx, bson.D{{"stop", nil}})

	var jobs []*Job

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		job := &Job{}
		if err := cursor.Decode(job); err != nil {
			return nil, fmt.Errorf("Failed to decode job %v", err)
		} else {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func jobCollection() *mongo.Collection {
	return docdb.Client.Database("work").Collection("jobs")
}

func JobUpdateFailureReason(id *primitive.ObjectID, reason string) error {
	return jobUpdateField(id, "failure", reason)
}

func JobUpdateStopTime(id *primitive.ObjectID) error {
	return jobUpdateField(id, "stop", util.TimeNowUtc())
}

func jobUpdateField(id *primitive.ObjectID, field string, value interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return docdb.UpdateOneFieldById(
		ctx,
		jobCollection(),
		id,
		field,
		value,
		jobMaxFailureRetry,
	)
}
