package cloud

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/golang/glog"
)

type EcsCluster struct {
	Name                 *string
	SubnetIds            []*string
	TaskSecurityGroupIds []*string
	WorkerTaskFamily     *string
}

var Ecs *EcsCluster

func (e *EcsCluster) Client() *ecs.ECS {
	return ecs.New(
		session.Must(session.NewSession()),
		aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")),
	)
}

func init() {
	Ecs = &EcsCluster{
		Name:                 aws.String(os.Getenv("CLUSTER_NAME")),
		SubnetIds:            []*string{aws.String(os.Getenv("SUBNET_0")), aws.String(os.Getenv("SUBNET_1"))},
		TaskSecurityGroupIds: []*string{aws.String(os.Getenv("APP_SECURITY_GROUP_ID"))},
		WorkerTaskFamily:     aws.String(os.Getenv("TASK_DEFINITION_FAMILY_WORKER")),
	}

	if err := ecsLongArnRoleWorkaround(); err != nil {
		log.Error(err)
	}
}

// This is a workaround to enable long ARN support for roles.
func ecsLongArnRoleWorkaround() error {
	if len(os.Getenv("LOCAL")) > 0 {
		return nil
	}

	if _, err := Ecs.Client().PutAccountSetting(
		&ecs.PutAccountSettingInput{
			Name:  aws.String("taskLongArnFormat"),
			Value: aws.String("enabled"),
		},
	); err != nil {
		return fmt.Errorf("Problem adding task long arn format setting - reason: %v", err)
	}

	return nil
}
