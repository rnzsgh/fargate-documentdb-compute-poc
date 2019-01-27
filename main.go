package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/work"
)

const taskCount = 10

func main() {

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		response := &response{}

		// Create an example job
		jobId := primitive.NewObjectID()
		tasks := make(map[string]*work.Task)
		for i := 0; i < taskCount; i++ {
			taskId := primitive.NewObjectID()
			tasks[taskId.Hex()] = &work.Task{Id: &taskId, JobId: &jobId}
		}

		now := time.Now()
		job := &work.Job{Id: &jobId, Start: &now, Tasks: tasks}

		work.SubmitJobChannel <- job

		response.Message = "Accepted"
		w.WriteHeader(http.StatusAccepted)

		response.Job = job

		// Normally this would be application/json, but we don't want to prompt downloads
		w.Header().Set("Content-Type", "text/plain")
		out, _ := json.Marshal(response)
		io.WriteString(w, string(out))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	log.Info("Ready to process")

	// This is a blocking call
	http.ListenAndServe(":8080", nil)

	close(work.SubmitJobChannel)
	log.Flush()
}

type response struct {
	Message string    `json:"message"`
	EnvVars []string  `json:"env"`
	Job     *work.Job `json:"job,omitempty"`
}
