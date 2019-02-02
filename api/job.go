package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/work"
)

func init() {
	http.HandleFunc("/start", JobStart)
}

func JobStart(w http.ResponseWriter, r *http.Request) {
	response := &response{}

	// Create an example job
	jobId := primitive.NewObjectID()
	tasks := make(map[string]*model.Task)
	for i := 0; i < 11; i++ {
		taskId := primitive.NewObjectID()
		tasks[taskId.Hex()] = &model.Task{Id: &taskId, JobId: &jobId}
	}

	now := time.Now()
	job := &model.Job{Id: &jobId, Start: &now, Tasks: tasks}

	work.JobSubmitChannel <- job

	response.Message = "Accepted"
	w.WriteHeader(http.StatusAccepted)

	response.Job = job

	// Normally this would be application/json, but we don't want to prompt downloads
	w.Header().Set("Content-Type", "text/plain")
	out, _ := json.Marshal(response)
	io.WriteString(w, string(out))
}
