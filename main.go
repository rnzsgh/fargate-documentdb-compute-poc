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

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
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

	connectionUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/test?ssl=true&replicaSet=rs0", user, password, endpoint, port)
	log.Info(connectionUri)

	client, err := mongo.NewClientWithOptions(
		connectionUri,
		options.Client().SetSSL(
			&options.SSLOpt{
				Enabled:  true,
				Insecure: true,
				CaFile:   "/rds-combined-ca-bundle.pem",
			},
		),
	)

	if err != nil {
		log.Error(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)

	if err != nil {
		log.Error(err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Error(err)
	}

	collection := client.Database("test").Collection("numbers")

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	if err != nil {
		log.Error(err)
	} else {
		id := res.InsertedID
		log.Info(id)
	}

	log.Info("This is a test")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		res := &response{Message: "healthy"}
		out, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(out))
	})

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

	})
	http.ListenAndServe(":80", nil)

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
