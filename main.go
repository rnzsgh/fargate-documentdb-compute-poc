package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/docdb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"

	log "github.com/golang/glog"
)

type appContext struct {
	docDbClient *mongo.Client
	cluster     *cloud.EcsCluster
}

const taskCount = 10

func main() {

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	appCtx := &appContext{
		docDbClient: docdb.Client,
		cluster:     cloud.Ecs,
	}

	jobChannel := make(chan Job)

	// Start the job processor
	go processJobs(appCtx, jobChannel)

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

		jobId := primitive.NewObjectID()

		tasks := make(map[string]*Task)
		for i := 0; i < taskCount; i++ {
			taskId := primitive.NewObjectID()
			tasks[taskId.Hex()] = &Task{Id: &taskId, JobId: &jobId}
		}

		now := time.Now()
		job := &Job{Id: &jobId, Start: &now, Tasks: tasks}

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := docdb.Client.Database("test").Collection("jobs").InsertOne(ctx, job)
		if err != nil {
			log.Errorf("Problem creating job: %v", err)
			response.Message = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Infof("New job id: %v", res.InsertedID)
			response.Message = "Accepted"
			w.WriteHeader(http.StatusAccepted)
			jobChannel <- *job
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

	log.Infof("Task count: %d", taskCount)

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

func processTask(appCtx *appContext, task *Task, completed chan<- Task) {
	count := int64(1)
	for {

		response, err := appCtx.cluster.Client.RunTask(&ecs.RunTaskInput{
			Cluster:              appCtx.cluster.Name,
			Count:                aws.Int64(1),
			EnableECSManagedTags: aws.Bool(true),
			LaunchType:           aws.String("FARGATE"),
			NetworkConfiguration: &ecs.NetworkConfiguration{
				AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
					SecurityGroups: appCtx.cluster.TaskSecurityGroupIds,
					Subnets:        appCtx.cluster.SubnetIds,
				},
			},
			StartedBy: aws.String(os.Getenv("STACK_NAME")),
			Tags: []*ecs.Tag{
				&ecs.Tag{Key: aws.String("Name"), Value: aws.String(os.Getenv("STACK_NAME"))},
				&ecs.Tag{Key: aws.String("Family"), Value: appCtx.cluster.WorkerTaskFamily},
				&ecs.Tag{Key: aws.String("JobId"), Value: aws.String(task.JobId.Hex())},
				&ecs.Tag{Key: aws.String("TaskId"), Value: aws.String(task.Id.Hex())},
			},
			TaskDefinition: appCtx.cluster.WorkerTaskFamily,
		})

		if err != nil {
			log.Errorf("Task run error - job: %s, task: %s - error: %v", task.JobId.Hex(), task.Id.Hex(), err)
			count = wait(count)
			continue
		}

		if len(response.Failures) > 0 {
			log.Info("We have failures")
			for _, failure := range response.Failures {
				log.Errorf("Task run failed - job: %s, task: %s - reason: %s", task.JobId.Hex(), task.Id.Hex(), *failure)
			}
			count = wait(count)
			continue
		}

		log.Info("Task launched - job: %s, task: %s", task.JobId.Hex(), task.Id.Hex())
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
	Id    *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tasks map[string]*Task    `json:"tasks" bson:"tasks"`
	Start *time.Time          `json:"start" bson:"start"`
	Stop  *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}

type Task struct {
	Id    *primitive.ObjectID `json:"id" bson:"id"`
	JobId *primitive.ObjectID `json:"jobId" bson:"jobId"`
	Start *time.Time          `json:"start" bson:"start"`
	Stop  *time.Time          `json:"stop,omitempty" bson:"stop,omitempty"`
}
