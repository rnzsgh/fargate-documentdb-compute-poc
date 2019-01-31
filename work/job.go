package work

import (
	"context"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

var SubmitJobChannel = make(chan *model.Job)

func init() {
	go processJobs(SubmitJobChannel)
}

func processJobs(jobs <-chan *model.Job) {
	for job := range jobs {
		go processJob(job)
	}
}

func processJob(job *model.Job) {

	if err := createJobEntry(job); err != nil {
		log.Errorf("Unable to create job entry: %v", err)
		return
	}

	completedTaskChannel := make(chan *model.Task)

	taskCount := len(job.Tasks)

	for _, task := range job.Tasks {
		go processTask(task, completedTaskChannel)
	}

	count := 0

	for {
		select {
		case _ = <-completedTaskChannel:
			count++
		}

		if count == taskCount {
			break
		}
	}

	log.Infof("Job work completed - id: %s", job.Id.Hex())

	if len(job.FailureReason) > 0 {
		log.Errorf("Job had a failure - id: %s - reason: %s", job.Id.Hex(), job.FailureReason)
		if err := updateJobFailureReason(job.Id, job.FailureReason); err != nil {
			log.Errorf("Job failed to update db failure reason - id: %s - reason: %v", job.Id.Hex(), err)
		}
	}

	for _, task := range job.Tasks {
		if len(task.FailureReason) > 0 {
			log.Errorf("Job task had a failure - id: %s task: %s - reason: %s", job.Id.Hex(), task.Id.Hex(), job.FailureReason)
		}
	}

	if err := updateJobStopTime(job.Id); err != nil {
		log.Errorf("Failed to update job stop time in db - id: %s - reason: %v", job.Id.Hex(), err)
	}
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

func createJobEntry(job *model.Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").InsertOne(ctx, job)
	return
}

func findJobById(id *primitive.ObjectID) (job *model.Job, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	job = &model.Job{}
	err = docdb.Client.Database("work").Collection("jobs").FindOne(ctx, bson.D{{"_id", id}}).Decode(job)
	return
}
