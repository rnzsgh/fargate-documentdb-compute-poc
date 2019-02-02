package model

import (
	"context"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

const jobMaxFailureRetry = 3

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failure" bson:"failure"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}

func JobUpdateFailureReason(id *primitive.ObjectID, reason string) error {
	return jobUpdateField(id, "failure", reason)
}

func JobUpdateStopTime(id *primitive.ObjectID) error {
	return jobUpdateField(id, "stop", time.Now())
}

func JobCreate(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = jobCollection().InsertOne(ctx, job)
	return
}

func JobFindById(id *primitive.ObjectID) (job *Job, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	job = &Job{}
	err = jobCollection().FindOne(ctx, bson.D{{"_id", id}}).Decode(job)
	return
}

func jobCollection() *mongo.Collection {
	return docdb.Client.Database("work").Collection("jobs")
}

func jobUpdateField(id *primitive.ObjectID, field string, value interface{}) error {
	count := 0

	for {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := jobCollection().UpdateOne(
			ctx,
			bson.D{{"_id", id}},
			bson.D{{"$set", bson.D{{field, value}}}})

		if err == nil {
			if res.MatchedCount != 1 && res.ModifiedCount != 1 {
				return fmt.Errorf(
					"Job field not updated - job: %s - field: %s",
					id.Hex(),
					field,
				)
			}
			return nil
		}

		log.Errorf(
			"Job field not updated - job: %s - field: %s - reason %v",
			id.Hex(),
			field,
			err,
		)

		count++

		time.Sleep(time.Duration(count*2) * time.Second)
		if count == jobMaxFailureRetry {
			return err
		}
	}
}
