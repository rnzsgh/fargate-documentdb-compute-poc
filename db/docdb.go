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

	connectionUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/work?ssl=true", user, password, endpoint, port)

	if len(os.Getenv("LOCAL")) == 0 {
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

// Find by id. mongo.ErrNoDocuments is returned if nothing is found
func FindById(
	ctx context.Context,
	collection *mongo.Collection,
	id *primitive.ObjectID,
	doc interface{},
) error {
	return collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(doc)
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

	sleeper := util.ExpoentialSleepSeconds()

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

func ping() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return Client.Ping(ctx, readpref.Primary())
}
