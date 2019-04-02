package work

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/golang/glog"
	queue "github.com/rnzsgh/documentdb-queue"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

var JobSubmitChannel = make(chan *model.Job)

var jobs sync.Map

func init() {

	go taskCompletedListener(docdb.TaskResponseQueue.Listen(2))

	go newJobListener(JobSubmitChannel)

	go updateQueueDepthMonitor()

	if jobs, err := model.JobFindRunning(); err != nil {
		log.Errorf("Unable to load running jobs on start: %v", err)
	} else {
		for _, job := range jobs {
			runJob(job)
		}
	}
}

func updateQueueDepthMonitor() {

	monitor := cloud.MonitorPutMetricDataBuffer(
		"Application/compute",
		"WorkerQueueDepth",
		30*time.Second,
	)

	for {
		time.Sleep(15 * time.Second)

		if size, err := docdb.TaskDispatchQueue.Size(context.Background()); err != nil {
			log.Errorf("Unable to to get task dispatch queue size - reason: %v", err)
		} else {
			monitor(float64(size))
		}
	}
}

func newJobListener(jobs <-chan *model.Job) {
	for job := range jobs {
		runJob(job)
	}
}

func taskCompletedListener(msgs <-chan *queue.QueueMessage) {
	for msg := range msgs {

		var vals map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &vals); err != nil {
			log.Errorf("Unable to unmarshal task completed payload - reason: %v", err)
			continue
		}

		job, ok := jobs.Load(vals["jobId"].(string))

		if !ok {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			if err := msg.Done(ctx); err != nil {
				log.Errorf("Unable to mark queue msg as done - reason: %v", err)
			}
			continue
		}

		taskCount := len(job.(*model.Job).Tasks)
		completedCount := 0

		for taskId, task := range job.(*model.Job).Tasks {
			if task.Stop != nil {
				completedCount++
			}

			if taskId == vals["taskId"].(string) {
				if err := model.TaskUpdateStopTime(task); err != nil {
					log.Errorf("Unable to udpate task stop time - reason: %v", err)
				}
			}
		}

		if taskCount == completedCount {
			jobCompleted(job.(*model.Job))
		}
	}
}

func runJob(job *model.Job) {

	if exists, err := model.JobExists(job.Id); err != nil {
		log.Errorf("Unable to check if job exists: %v", err)
		return
	} else if !exists {
		if err := model.JobCreate(job); err != nil {
			log.Errorf("Unable to create job entry: %v", err)
			return
		}
	}

	jobs.Store(job.Id.Hex(), job)

	for _, task := range job.Tasks {
		if err := dispatchTask(task); err != nil {
			log.Errorf(
				"Job task dispatch issue - job: %s - task: %s - reason: %s",
				job.Id.Hex(),
				task.Id.Hex(),
				err,
			)
		}
	}
}

func jobCompleted(job *model.Job) {
	log.Infof("Job work completed - id: %s", job.Id.Hex())

	jobs.Delete(job.Id.Hex())

	if err := model.JobUpdateStopTime(job.Id); err != nil {
		log.Errorf("Failed to update job stop time in db - id: %s - reason: %v", job.Id.Hex(), err)
	}
}
