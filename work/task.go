package work

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"

	log "github.com/golang/glog"
)

func init() {
}

func processTask(task *Task, completed chan<- *Task) {
	count := int64(1)
	for {
		response, err := cloud.Ecs.Client().RunTask(&ecs.RunTaskInput{
			Cluster:              cloud.Ecs.Name,
			Count:                aws.Int64(1),
			EnableECSManagedTags: aws.Bool(true),
			LaunchType:           aws.String("FARGATE"),
			NetworkConfiguration: &ecs.NetworkConfiguration{
				AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
					SecurityGroups: cloud.Ecs.TaskSecurityGroupIds,
					Subnets:        cloud.Ecs.SubnetIds,
					AssignPublicIp: aws.String("ENABLED"),
				},
			},
			StartedBy: aws.String(os.Getenv("STACK_NAME")),
			Tags: []*ecs.Tag{
				&ecs.Tag{Key: aws.String("Name"), Value: aws.String(os.Getenv("STACK_NAME"))},
				&ecs.Tag{Key: aws.String("Family"), Value: cloud.Ecs.WorkerTaskFamily},
				&ecs.Tag{Key: aws.String("JobId"), Value: aws.String(task.JobId.Hex())},
				&ecs.Tag{Key: aws.String("TaskId"), Value: aws.String(task.Id.Hex())},
			},
			TaskDefinition: cloud.Ecs.WorkerTaskFamily,
		})

		if err != nil {
			log.Errorf("Task run error - job: %s, task: %s - error: %v", task.JobId.Hex(), task.Id.Hex(), err)
			count = wait(count)
			continue
		}

		if len(response.Failures) > 0 {
			for _, failure := range response.Failures {
				log.Errorf("Task run failed - job: %s, task: %s - reason: %s", task.JobId.Hex(), task.Id.Hex(), *failure)
			}
			count = wait(count)
			continue
		}

		log.Infof("Task launched - job: %s, task: %s", task.JobId.Hex(), task.Id.Hex())
		break
	}

	// Wait for task to complete
	for {
		time.Sleep(1 * time.Minute)

	}

	// TODO: Poll and wait for task to complete - update stop time
}

func wait(count int64) int64 {
	time.Sleep(time.Duration(count) * time.Second)
	return count * int64(2)
}

type Task struct {
	Id      *primitive.ObjectID `json:"id" bson:"id"`
	JobId   *primitive.ObjectID `json:"jobId" bson:"jobId"`
	Failure string              `json:"failure" bson:"failure"`
	Start   *time.Time          `json:"start" bson:"start"`
	Stop    *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
