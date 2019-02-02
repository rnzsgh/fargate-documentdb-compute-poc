package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/model"
)

func init() {
	http.HandleFunc("/", Root)
}

type response struct {
	Message string     `json:"message"`
	EnvVars []string   `json:"env"`
	Job     *model.Job `json:"job,omitempty"`
}

func Root(w http.ResponseWriter, r *http.Request) {
	res := &response{Message: "Hello World"}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		res.EnvVars = append(res.EnvVars, pair[0]+"="+pair[1])
	}
	// Normally this would be application/json, but we don't want to prompt downloads
	w.Header().Set("Content-Type", "text/plain")

	out, _ := json.Marshal(res)
	io.WriteString(w, string(out))
}
