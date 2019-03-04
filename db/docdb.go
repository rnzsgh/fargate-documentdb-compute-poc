package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rnzsgh/fargate-documentdb-compute-poc/cloud"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"

	log "github.com/golang/glog"
	queue "github.com/rnzsgh/documentdb-queue"
)

const defaultTimeoutInSeconds = 10

var Client *mongo.Client

var TaskDispatchQueue *queue.Queue
var TaskResponseQueue *queue.Queue

func init() {

	endpoint := os.Getenv("DOCUMENT_DB_ENDPOINT")
	port := os.Getenv("DOCUMENT_DB_PORT")
	user := os.Getenv("DOCUMENT_DB_USER")
	caFile := os.Getenv("DOCUMENT_DB_PEM")

	password := cloud.Secrets.DatabasePassword

	connectionUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/work?ssl=true", user, password, endpoint, port)

	if len(os.Getenv("LOCAL")) == 0 {
		connectionUri = connectionUri + "&replicaSet=rs0"
	}

	var err error
	if TaskDispatchQueue, err = queue.NewQueue(
		"work",
		"dispatchQueue",
		connectionUri,
		caFile, 5*time.Second,
	); err != nil {
		log.Errorf("Unable to create work dispatch queue - endpoint: %s - reason: %v", endpoint, err)
	}

	if TaskResponseQueue, err = queue.NewQueue(
		"work",
		"responseQueue",
		connectionUri,
		caFile,
		5*time.Second,
	); err != nil {
		log.Errorf("Unable to create work response queue - endpoint: %s - reason: %v", endpoint, err)
	}

	Client, err = mongo.NewClientWithOptions(
		connectionUri,
		options.Client().SetSSL(
			&options.SSLOpt{
				Enabled:  true,
				Insecure: true,
				CaFile:   caFile,
			},
		),
	)

	if err != nil {
		log.Errorf("Unable to create new db client - endpoint: %s - reason: %v", endpoint, err)
	}

	ctx, _ := context.WithTimeout(context.Background(), defaultTimeoutInSeconds*time.Second)
	if err = Client.Connect(ctx); err != nil {
		log.Errorf("Unable to connect to db - endpoint: %s - reason: %v", endpoint, err)
	}

	ctx, _ = context.WithTimeout(context.Background(), defaultTimeoutInSeconds*time.Second)
	if err = Ping(ctx); err != nil {
		log.Errorf("Unable to ping db - endpoint: %s - reason: %v", endpoint, err)
	}
}

// Find by id. mongo.ErrNoDocuments is returned if nothing is found
func FindOneById(
	ctx context.Context,
	collection *mongo.Collection,
	id *primitive.ObjectID,
	doc interface{},
) error {
	if res := collection.FindOne(ctx, bson.D{{"_id", id}}); res.Err() != nil {
		return res.Err()
	} else {
		return res.Decode(doc)
	}
}

// Delete by id. If no documents are deleted an mongo.ErrNoDocuments error
// is returned.
func DeleteOneById(
	ctx context.Context,
	collection *mongo.Collection,
	id *primitive.ObjectID,
) error {
	if res, err := collection.DeleteOne(ctx, bson.D{{"_id", id}}); err != nil {
		return err
	} else {
		if res.DeletedCount == 0 {
			return mongo.ErrNoDocuments
		}
		return nil
	}
}

// Update a single field in a document. Expects document to be
// present or an error is returned. Support for retries, set to
// zero to disable.
func UpdateOneFieldById(
	ctx context.Context,
	collection *mongo.Collection,
	id *primitive.ObjectID,
	field string,
	value interface{},
	retries int,
) error {

	sleeper := util.TimeExponentialSleepSeconds()

	for {
		err := updateOneFieldById(ctx, collection, id, field, value)
		if err == nil {
			return nil
		}

		if retries < 1 {
			return err
		}

		log.Error(err)

		if sleeper() == retries {
			return err
		}
	}
}

func updateOneFieldById(
	ctx context.Context,
	collection *mongo.Collection,
	id *primitive.ObjectID,
	field string,
	value interface{},
) error {
	res, err := collection.UpdateOne(
		ctx,
		bson.D{{"_id", id}},
		bson.D{{"$set", bson.D{{field, value}}}},
	)

	if err == nil {
		if res.MatchedCount != 1 && res.ModifiedCount != 1 {
			return fmt.Errorf(
				"Doc field not updated - no doc match - collection : %s - id: %s - field: %s",
				collection.Name(),
				id.Hex(),
				field,
			)
		}
		return nil
	}

	return fmt.Errorf(
		"Doc field not updated - collection: %s - id: %s - field: %s - reason %v",
		collection.Name(),
		id.Hex(),
		field,
		err,
	)
}

func Ping(ctx context.Context) error {
	return Client.Ping(ctx, nil)
}
