package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/work"
)

func init() {
	http.HandleFunc("/start", JobStart)
}

func JobStart(w http.ResponseWriter, r *http.Request) {
	response := &response{}
	w.Header().Set("Content-Type", "text/plain")

	if err := model.DataEnsureTest(); err != nil {
		response.Message = fmt.Sprintf("Unable to ensure test data is created - reason: %v", err)
		log.Errorf(response.Message)
		out, _ := json.Marshal(response)
		io.WriteString(w, string(out))
		return
	}

	dataIds, err := model.DataFindAllIds()
	if err != nil {
		response.Message = fmt.Sprintf("Unable to load test data doc ids - reason: %v", err)
		log.Errorf(response.Message)
		out, _ := json.Marshal(response)
		io.WriteString(w, string(out))
		return
	}

	// Create an sample job
	jobId := primitive.NewObjectID()
	tasks := make(map[string]*model.Task)
	for i := 0; i < len(dataIds); i++ {
		taskId := primitive.NewObjectID()
		tasks[taskId.Hex()] = &model.Task{Id: &taskId, JobId: &jobId, DataId: dataIds[i]}
	}

	now := time.Now()
	job := &model.Job{Id: &jobId, Start: &now, Tasks: tasks}

	work.JobSubmitChannel <- job

	response.Message = "Accepted"
	w.WriteHeader(http.StatusAccepted)

	response.Job = job

	out, _ := json.Marshal(response)
	io.WriteString(w, string(out))
}
