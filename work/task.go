package work

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"

	log "github.com/golang/glog"
)

func processTask(task *Task, completedChannel chan<- *Task) {
	waitForTask(task, launchTask(task), completedChannel)
}

func waitForTask(task *Task, taskArn string, completedChannel chan<- *Task) {
	client := cloud.Ecs.Client()
	for {
		time.Sleep(1 * time.Minute)

		response, err := client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: cloud.Ecs.Name,
			Tasks:   []*string{aws.String(taskArn)},
		})

		if err != nil {
			log.Errorf(
				"Cannot describe task - job: %s - task: %s - task arn: %s - err: %v",
				task.JobId.Hex(),
				task.Id.Hex(),
				taskArn,
				err,
			)
			continue
		}

		if len(response.Failures) > 0 {
			for _, failure := range response.Failures {
				// TODO: Update failure reason in db
				log.Errorf("Task failed - job: %s - task: %s - reason %s", task.JobId.Hex(), task.Id.Hex(), *failure.Reason)
				task.FailureReason = *failure.Reason
			}
		}

		for _, submittedTask := range response.Tasks {
			for _, container := range submittedTask.Containers {
				lastStatus := container.LastStatus
				if lastStatus != nil && *lastStatus == "STOPPED" {
					exitCode := container.ExitCode
					if *exitCode != 0 {
						// TODO: Update reason in database
						task.FailureReason = fmt.Sprintf("Task did not have a zero exit code - %d", *exitCode)
					}

					// TODO: Update stop time in database
					task.Stop = submittedTask.StoppedAt
					log.Infof("Task stopped - job: %s, task: %s", task.JobId.Hex(), task.Id.Hex())
					completedChannel <- task
					return
				}
			}
		}
	}
}

func launchTask(task *Task) string {
	count := int64(1)
	var taskArn *string
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
				log.Errorf("Task run failed - job: %s - task: %s - reason: %s", task.JobId.Hex(), task.Id.Hex(), *failure)
			}
			count = wait(count)
			continue
		}

		for _, submittedTask := range response.Tasks {
			for _, container := range submittedTask.Containers {
				taskArn = container.TaskArn
			}
		}

		log.Infof("Task launched - job: %s, task: %s", task.JobId.Hex(), task.Id.Hex())
		return *taskArn
	}
}

func updateTaskFailureReason(task *Task, reason string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = docdb.Client.Database("work").Collection("jobs").UpdateOne(
		ctx,
		bson.D{{"_id", task.JobId}},
		bson.D{{"$set", bson.D{{fmt.Sprintf("tasks.%s.failureReason", task.Id.Hex()), reason}}}})
	return
}

func wait(count int64) int64 {
	time.Sleep(time.Duration(count) * time.Second)
	return count * int64(2)
}

type Task struct {
	Id            *primitive.ObjectID `json:"id" bson:"id"`
	JobId         *primitive.ObjectID `json:"jobId" bson:"jobId"`
	FailureReason string              `json:"failure" bson:"failure"`
	Start         *time.Time          `json:"start" bson:"start"`
	Stop          *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
