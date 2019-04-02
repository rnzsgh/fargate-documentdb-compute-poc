package cloud

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	log "github.com/golang/glog"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

// MonitorPutMetricDataBuffer returns a function that allows for real time metric
// writing, but buffers the value for the duration parameter passed. Every d duration,
// the metric is written to CloudWatch. If there are errors, they are written to the log.
func MonitorPutMetricDataBuffer(namespace, metric string, d time.Duration) func(float64) {

	var value float64
	var m sync.Mutex

	go func() {
		for {
			time.Sleep(d)
			m.Lock()
			v := value
			m.Unlock()

			if err := MonitorPutMetricData(namespace, metric, v); err != nil {
				log.Errorf("Problem with MonitorPutMetricData in MonitorPutMetricDataBuffer - reason: %v", err)
			}
		}
	}()

	return func(v float64) {
		m.Lock()
		value = v
		m.Unlock()
	}
}

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
