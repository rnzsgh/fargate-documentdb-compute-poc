package work

import (
	"context"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

var SubmitJobChannel = make(chan *Job)

func init() {
	go processJobs(SubmitJobChannel)
}

func processJobs(jobs <-chan *Job) {
	for job := range jobs {
		go processJob(job)
	}
}

func processJob(job *Job) {

	if err := createJobEntry(job); err != nil {
		job.FailureReason = err.Error()
		return
	}

	completedTaskChannel := make(chan *Task)

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
		if err := updateJobFailureReason(job, job.FailureReason); err != nil {
			log.Errorf("Job failed to update db failure reason - id: %s - reason: %v", job.Id.Hex(), err)
		}
	}

	for _, task := range job.Tasks {
		if len(task.FailureReason) > 0 {
			log.Errorf("Job task had a failure - id: %s task: %s - reason: %s", job.Id.Hex(), task.Id.Hex(), job.FailureReason)
		}
	}

	if err := updateJobStopTime(job); err != nil {
		log.Errorf("Failed to update job stop time in db - id: %s - reason: %v", job.Id.Hex(), err)
	}
}

func updateJobFailureReason(job *Job, reason string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", job.Id}},
		bson.D{{"$set", bson.D{{"failure", reason}}}})
	return
}

func updateJobStopTime(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", job.Id}},
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

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failure" bson:"failure"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
