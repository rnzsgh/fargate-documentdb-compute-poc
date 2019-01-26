package cloud

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type EcsCluster struct {
	Client               *ecs.ECS
	Name                 *string
	SubnetIds            []*string
	TaskSecurityGroupIds []*string
	WorkerTaskFamily     *string
}

var Ecs *EcsCluster

func init() {
	Ecs = &EcsCluster{
		Client:               ecs.New(session.Must(session.NewSession()), aws.NewConfig().WithRegion(os.Getenv("AWS_REGION"))),
		Name:                 aws.String(os.Getenv("CLUSTER_NAME")),
		SubnetIds:            []*string{aws.String(os.Getenv("SUBNET_0")), aws.String(os.Getenv("SUBNET_1"))},
		TaskSecurityGroupIds: []*string{aws.String(os.Getenv("APP_SECURITY_GROUP_ID"))},
		WorkerTaskFamily:     aws.String(os.Getenv("TASK_DEFINITION_FAMILY_WORKER")),
	}
}
