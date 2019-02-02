package work

import (
	"testing"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

func TestJobStopStart(t *testing.T) {
	t.Run("TestJobStopStart", func(t *testing.T) {

		// Stop the processing
		close(JobSubmitChannel)

		// Start the job processing again
		JobSubmitChannel = make(chan *model.Job)
		go processJobs(JobSubmitChannel)
	})
}

func TestWaitForTasksToComplete(t *testing.T) {
	t.Run("TestWaitForTasksToComplete", func(t *testing.T) {

		completedTaskChannel := make(chan *model.Task)
		taskCount := 1001

		testCompletedChannel := make(chan bool, 1)

		go func() {
			waitForTasksToComplete(taskCount, completedTaskChannel)
			testCompletedChannel <- true
		}()

		for i := 0; i < taskCount; i++ {
			completedTaskChannel <- &model.Task{}
		}

		for {
			select {
			case _ = <-testCompletedChannel:
				return
			}
		}
	})
}
