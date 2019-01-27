package work

import (
	"context"
	"time"

	log "github.com/golang/glog"
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
	}

	for _, task := range job.Tasks {
		if len(task.FailureReason) > 0 {
			log.Errorf("Job task had a failure - id: %s task: %s - reason: %s", job.Id.Hex(), task.Id.Hex(), job.FailureReason)
		}
	}

	if err := updateJobEntry(job); err != nil {
		log.Errorf("Failed to update job in db - id: %s task: %s - error: %v", job.Id.Hex(), task.Id.Hex(), err)
	}
}

func updateJobEntry(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").Update(ctx, job)
	return
}

func createJobEntry(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").InsertOne(ctx, job)
	return
}

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failureReason" bson:"failureReason"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
