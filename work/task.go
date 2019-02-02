package work

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"

	log "github.com/golang/glog"
)

func processTask(task *model.Task, completedChannel chan<- *model.Task) {
	waitForTask(task, launchTask(task), completedChannel)
}

func waitForTask(task *model.Task, taskArn string, completedChannel chan<- *model.Task) {
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
				if err := model.TaskUpdateFailureReason(task, *failure.Reason); err != nil {
					log.Errorf("Could not update task failure reason in db - job: %s - task: %s - reason %s", task.JobId.Hex(), task.Id.Hex(), err)
				}
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
						task.FailureReason = fmt.Sprintf("Task did not have a zero exit code - %d", *exitCode)
						if err := model.TaskUpdateFailureReason(task, task.FailureReason); err != nil {
							log.Errorf("Could not task failure reason in db - job: %s - task: %s - reason %s", task.JobId.Hex(), task.Id.Hex(), err)
						}
					}

					task.Stop = submittedTask.StoppedAt
					if err = model.TaskUpdateStopTime(task); err != nil {
						log.Errorf("Unable to update task stop time in db - job: %s - task: %s - reason %s", task.JobId.Hex(), task.Id.Hex(), err)
					}
					log.Infof("Task stopped - job: %s, task: %s", task.JobId.Hex(), task.Id.Hex())
					completedChannel <- task
					return
				}
			}
		}
	}
}

func launchTask(task *model.Task) string {
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
			count = util.WaitSeconds(count)
			continue
		}

		if len(response.Failures) > 0 {
			for _, failure := range response.Failures {
				log.Errorf("Task run failed - job: %s - task: %s - reason: %s", task.JobId.Hex(), task.Id.Hex(), *failure)
			}
			count = util.WaitSeconds(count)
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
