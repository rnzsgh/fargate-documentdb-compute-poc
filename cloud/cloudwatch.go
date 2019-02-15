package cloud

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

func MonitorPutMetricData(namespace, metric string, value float64) error {
	svc := cloudwatch.New(session.New())

	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String(metric),
				Value:      aws.Float64(value),
				Timestamp:  util.TimeNowUtc(),
			},
		},
	})

	if err != nil {
		return fmt.Errorf(
			"Unable to put metric data - namespace: %s - metric: %s - reasion: %v",
			namespace,
			metric,
			err,
		)
	}

	return nil
}
