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

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	log "github.com/golang/glog"
)

func main() {

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	endpoint := os.Getenv("DOCUMENT_DB_ENDPOINT")
	port := os.Getenv("DOCUMENT_DB_PORT")
	user := os.Getenv("DOCUMENT_DB_USER")

	// This is not secure. Waiting for secrets support in CFN
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html
	password := os.Getenv("DOCUMENT_DB_PASSWORD")

	connectionUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?ssl_ca_certs=rds-combined-ca-bundle.pem&replicaSet=rs0", user, password, endpoint, port)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, connectionUri)

	if err != nil {
		log.Error(err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Error(err)
	}

	log.Info("This is a test")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		f := fib()

		res := &response{Message: "Hello World"}

		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			res.EnvVars = append(res.EnvVars, pair[0]+"="+pair[1])
		}

		for i := 1; i <= 90; i++ {
			res.Fib = append(res.Fib, f())
		}

		// Beautify the JSON output
		out, _ := json.MarshalIndent(res, "", "  ")

		// Normally this would be application/json, but we don't want to prompt downloads
		w.Header().Set("Content-Type", "text/plain")

		io.WriteString(w, string(out))

		fmt.Println("Hello world - the log message")
	})
	http.ListenAndServe(":8080", nil)

	log.Flush()
}

type response struct {
	Message string   `json:"message"`
	EnvVars []string `json:"env"`
	Fib     []int    `json:"fib"`
}

func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}
