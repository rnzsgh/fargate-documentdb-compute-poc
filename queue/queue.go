package queue

import (
	"context"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
	"github.com/rnzsgh/fargate-documentdb-compute-poc/util"
)

// The queue entry structure.
type QueueEntry struct {
	Id         *primitive.ObjectID `json:"id" bson:"_id"`
	Payload    string              `json:"payload" bson:"payload"`
	Visibility int                 `json:"visibility" bson:"visibility"` // Visibility timeout is in seconds
	Created    *time.Time          `json:"created" bson:"created"`
	Started    *time.Time          `json:"started" bson:"started"`
	queue      *Queue              `json:"-" bson:"-"`
}

type Queue struct {
	collection *mongo.Collection
}

// Create new new queue struct.
func NewQueue(collection *mongo.Collection) *Queue {
	ensureCollectionIndexes(collection)
	return &Queue{collection: collection}
}

func (e *QueueEntry) Delete(ctx context.Context) error {
	return docdb.DeleteOneById(ctx, e.queue.collection, e.Id)
}

// Pull the next item off the queue. You must call the Delete function on
// the message when you are done processing or it will timeout and be made
// visible again.
func (q *Queue) Dequeue(
	ctx context.Context,
) (*QueueEntry, error) {

	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)
	opts.SetSort(bson.D{{"created", 1}})
	opts.SetUpsert(false)

	res := q.collection.FindOneAndUpdate(
		ctx,
		bson.D{{"started", nil}},
		bson.D{{"$set", bson.D{{"started", util.TimeNowUtc()}}}},
		opts,
	)

	if res.Err() != nil {
		return nil, res.Err()
	}

	entry := &QueueEntry{}

	if err := res.Decode(entry); err != nil {
		return nil, err
	}

	entry.collection = q.collection

	return entry, nil
}

// Insert a new item into the queue. This allows for an empty payload.
// If visibility is negative, this will panic.
func (q *Queue) Enqueue(
	ctx context.Context,
	payload string,
	visibility int,
) (*primitive.ObjectID, error) {

	if visibility < 0 {
		panic("Cannot have a negative visibility timeout")
	}

	id := primitive.NewObjectID()

	entry := &QueueEntry{
		Id:         &id,
		Payload:    payload,
		Visibility: visibility,
		Created:    util.TimeNowUtc(),
	}

	if res, err := q.collection.InsertOne(ctx, entry); err != nil {
		return nil, fmt.Errorf("Unable to enqueue doc into collection %s - reason: %v", q.collection.Name(), err)
	} else {
		docId := res.InsertedID.(primitive.ObjectID)
		return &docId, nil
	}
}

// Ensure that the proper indices are on the collection. This is performed once by each
// process when the queue is created.
func ensureCollectionIndexes(collection *mongo.Collection) {
	if _, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{Keys: bson.D{{"started", 1}, {"created", 1}}},
	); err != nil {
		log.Errorf("Unable to create queue started/created index on collection: %s - reason: %v", collection.Name(), err)
	}
}
