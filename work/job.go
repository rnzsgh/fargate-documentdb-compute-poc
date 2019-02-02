package work

import (
	log "github.com/golang/glog"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

var JobSubmitChannel = make(chan *model.Job)

func init() {
	go processJobs(JobSubmitChannel)

	if jobs, err := model.JobFindRunning(); err != nil {
		log.Errorf("Unable to load running jobs on start: %v", err)
	} else {
		for _, job := range jobs {
			go processJob(job)
		}
	}
}

func processJobs(jobs <-chan *model.Job) {
	for job := range jobs {
		go processJob(job)
	}
}

func waitForTasksToComplete(taskCount int, completedTaskChannel chan *model.Task) {
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
}

func processJob(job *model.Job) {

	if exists, err := model.JobExists(job.Id); err != nil {
		log.Errorf("Unable to check if job exists: %v", err)
		return
	} else if !exists {
		if err := model.JobCreate(job); err != nil {
			log.Errorf("Unable to create job entry: %v", err)
			return
		}
	}

	completedTaskChannel := make(chan *model.Task)
	taskCount := len(job.Tasks)

	for _, task := range job.Tasks {
		go processTask(task, completedTaskChannel)
	}

	waitForTasksToComplete(taskCount, completedTaskChannel)
	jobCompleted(job)
}

func jobCompleted(job *model.Job) {
	log.Infof("Job work completed - id: %s", job.Id.Hex())

	if len(job.FailureReason) > 0 {
		log.Errorf("Job had a failure - id: %s - reason: %s", job.Id.Hex(), job.FailureReason)
		if err := model.JobUpdateFailureReason(job.Id, job.FailureReason); err != nil {
			log.Errorf("Job failed to update db failure reason - id: %s - reason: %v", job.Id.Hex(), err)
		}
	}

	if err := model.JobUpdateStopTime(job.Id); err != nil {
		log.Errorf("Failed to update job stop time in db - id: %s - reason: %v", job.Id.Hex(), err)
	}
}
