package docdb

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	log "github.com/golang/glog"
)

var Client *mongo.Client

func init() {

	log.Info("Init called here")

	endpoint := os.Getenv("DOCUMENT_DB_ENDPOINT")
	port := os.Getenv("DOCUMENT_DB_PORT")
	user := os.Getenv("DOCUMENT_DB_USER")
	pemFile := os.Getenv("DOCUMENT_DB_PEM")

	// This is not secure. Waiting for secrets support in CFN
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html
	password := os.Getenv("DOCUMENT_DB_PASSWORD")

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

	err = Client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Errorf("Unable to ping db: %v", err)
	}

	log.Info("Init done")
}
