package work

import (
	"context"
	"time"

	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var SubmitJobChannel = make(chan *Job)
var JobResultChannel = make(chan *Job)

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
		JobResultChannel <- job
		return
	}

	completed := make(chan *Task)

	taskCount := len(job.Tasks)

	for _, task := range job.Tasks {
		go processTask(task, completed)
	}

	count := 0

	for {
		select {
		case _ = <-completed:
			count++
		}

		if count == taskCount {
			break
		}
	}

	// TODO: Update job struct if there are errors
	// Update job in db

	JobResultChannel <- job
}

func createJobEntry(job *Job) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("test").Collection("jobs").InsertOne(ctx, job)
	return
}

type Job struct {
	Id            *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks         map[string]*Task    `json:"tasks" bson:"tasks"`
	FailureReason string              `json:"failureReason" bson:"failureReason"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
