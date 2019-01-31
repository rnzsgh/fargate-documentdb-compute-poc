package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	log "github.com/golang/glog"
)

var Client *mongo.Client

func init() {

	endpoint := os.Getenv("DOCUMENT_DB_ENDPOINT")
	port := os.Getenv("DOCUMENT_DB_PORT")
	user := os.Getenv("DOCUMENT_DB_USER")
	pemFile := os.Getenv("DOCUMENT_DB_PEM")

	password := cloud.Secrets.DatabasePassword

	connectionUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/test?ssl=true", user, password, endpoint, port)

	if len(os.Getenv("DOCUMENT_DB_LOCAL")) == 0 {
		connectionUri = connectionUri + "&replicaSet=rs0"
	}

	var err error
	Client, err = mongo.NewClientWithOptions(
		connectionUri,
		options.Client().SetSSL(
			&options.SSLOpt{
				Enabled:  true,
				Insecure: true,
				CaFile:   pemFile,
			},
		),
	)

	if err != nil {
		log.Errorf("Unable to create new db client: %v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = Client.Connect(ctx)

	if err != nil {
		log.Errorf("Unable to connect to db: %v", err)
	}

	if err = ping(); err != nil {
		log.Errorf("Unable to ping db: %v", err)
	}
}

func ping() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return Client.Ping(ctx, readpref.Primary())

}
