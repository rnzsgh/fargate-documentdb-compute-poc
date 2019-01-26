package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/docdb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"

	log "github.com/golang/glog"
)

type appContext struct {
	mClient          *mongo.Client
	eClient          *ecs.ECS
	cluster          *string
	family           *string
	subnetIds        []*string
	securityGroupIds []*string
}

const taskCount = 10

func main() {

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	region := os.Getenv("AWS_REGION")

	appCtx := &appContext{
		mClient:          docdb.Client,
		eClient:          ecs.New(session.Must(session.NewSession()), aws.NewConfig().WithRegion(region)),
		cluster:          aws.String(os.Getenv("CLUSTER_NAME")),
		family:           aws.String(os.Getenv("TASK_DEFINITION_FAMILY_WORKER")),
		subnetIds:        []*string{aws.String(os.Getenv("SUBNET_0")), aws.String(os.Getenv("SUBNET_1"))},
		securityGroupIds: []*string{aws.String(os.Getenv("APP_SECURITY_GROUP_ID"))},
	}

	jobChannel := make(chan Job)

	// Start the job processor
	go processJobs(appCtx, jobChannel)

	if docdb.Client == nil {
		log.Info("DOCDB CLIENT IS NIL")
	}

	collection := docdb.Client.Database("test").Collection("numbers")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	if err != nil {
		log.Error(err)
	} else {
		id := res.InsertedID
		log.Info(id)
	}

	log.Info("This is a test")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Info("/health")
		io.WriteString(w, "ok")
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		log.Info("/")

		response := &response{}

		tasks := make(map[primitive.ObjectID]Task)
		for i := 0; i < taskCount; i++ {
			taskId := primitive.NewObjectID()
			tasks[taskId] = Task{Id: taskId}
		}

		// hello

		job := Job{Start: time.Now(), Tasks: tasks}

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := docdb.Client.Database("test").Collection("jobs").InsertOne(ctx, job)
		if err != nil {
			log.Error(fmt.Sprintf("Problem creating job: %v", err))
			response.Message = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Info(fmt.Sprintf("New job id: %v", res.InsertedID))
			response.Message = "Accepted"
			w.WriteHeader(http.StatusAccepted)
			jobChannel <- job
		}

		// Normally this would be application/json, but we don't want to prompt downloads
		w.Header().Set("Content-Type", "text/plain")
		out, _ := json.Marshal(response)
		io.WriteString(w, string(out))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("/")

		res := &response{Message: "Hello World"}

		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			res.EnvVars = append(res.EnvVars, pair[0]+"="+pair[1])
		}
		// Normally this would be application/json, but we don't want to prompt downloads
		w.Header().Set("Content-Type", "text/plain")

		out, _ := json.Marshal(res)
		io.WriteString(w, string(out))

	})

	log.Info("ready to serve")
	http.ListenAndServe(":8080", nil)

	close(jobChannel)

	log.Flush()
}

func processJobs(appCtx *appContext, jobs <-chan Job) {
	for job := range jobs {
		go processJob(appCtx, job)
	}
}

func processJob(appCtx *appContext, job Job) {
	completed := make(chan Task)

	taskCount := len(job.Tasks)

	for _, task := range job.Tasks {
		go processTask(appCtx, task, completed)
	}

	count := 0

	for {
		select {
		case _ = <-completed:
			count++
		}

		if count == taskCount {
			break
		}
	}
}

func processTask(appCtx *appContext, task Task, completed chan<- Task) {
	// Run the task
	count := int64(1)
	for {

		response, err := appCtx.eClient.RunTask(&ecs.RunTaskInput{
			Cluster:              appCtx.cluster,
			Count:                aws.Int64(1),
			EnableECSManagedTags: aws.Bool(true),
			LaunchType:           aws.String("FARGATE"),
			NetworkConfiguration: &ecs.NetworkConfiguration{
				AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
					SecurityGroups: appCtx.securityGroupIds,
					Subnets:        appCtx.subnetIds,
				},
			},
			PropagateTags: aws.String("true"),
			StartedBy:     aws.String(os.Getenv("STACK_NAME")),
			Tags: []*ecs.Tag{
				&ecs.Tag{Key: aws.String("Name"), Value: aws.String(os.Getenv("STACK_NAME"))},
				&ecs.Tag{Key: aws.String("Family"), Value: appCtx.family},
				&ecs.Tag{Key: aws.String("JobId"), Value: aws.String(task.JobId.String())},
				&ecs.Tag{Key: aws.String("TaskId"), Value: aws.String(task.Id.String())},
			},
			TaskDefinition: appCtx.family,
		})

		if err != nil {
			log.Error(fmt.Sprintf("Task run error - job: %s, task: %s - error: %v", task.JobId.String(), task.Id.String(), err))
			count = wait(count)
			continue
		}

		if len(response.Failures) > 0 {
			for _, failure := range response.Failures {
				log.Error(fmt.Sprintf("Task run failed - job: %s, task: %s - reason: %s", task.JobId.String(), task.Id.String(), *failure))
			}
			count = wait(count)
			continue
		}

		log.Info(fmt.Sprintf("Task launched - job: %s, task: %s", task.JobId.String(), task.Id.String()))
		break
	}
}

func wait(count int64) int64 {
	time.Sleep(time.Duration(count) * time.Second)
	return count * int64(2)
}

type response struct {
	Message string   `json:"message"`
	EnvVars []string `json:"env"`
	Jobs    []Job    `json:"jobs,omitempty"`
}

type Job struct {
	Id    primitive.ObjectID          `json:"id" bson:"_id,omitempty"`
	Tasks map[primitive.ObjectID]Task `json:"tasks" bson:"tasks"`
	Start time.Time                   `json:"start" bson:"start"`
	Stop  time.Time                   `json:"stop,omitempty" bson:"stop,omitempty"`
}

type Task struct {
	Id    primitive.ObjectID `json:"id" bson:"id"`
	JobId primitive.ObjectID `json:"jobId" bson:"jobId"`
	Start time.Time          `json:"start" bson:"start"`
	Stop  time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
