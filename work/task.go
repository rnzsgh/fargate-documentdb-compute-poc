package work

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

func dispatchTask(task *model.Task) error {

	m := make(map[string]string)
	m["taskId"] = task.Id.Hex()
	m["jobId"] = task.JobId.Hex()
	m["dataId"] = task.DataId.Hex()

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Unable to marshal task json - reason %v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := docdb.TaskDispatchQueue.Enqueue(ctx, string(b), 30); err != nil {
		return fmt.Errorf("Unable to enqueue task - reason: %v", err)
	}

	return nil
}
