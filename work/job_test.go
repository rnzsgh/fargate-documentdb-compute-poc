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
		go newJobListener(JobSubmitChannel)
	})
}
