// Package database contains the local unwindia mongodb database models and client
package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/environment"
	"github.com/gammazero/workerpool"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	COLLECTION_NAME = "dotlan_status"
	DATABASE_NAME   = "unwindia"
	DEFAULT_TIMEOUT = 10 * time.Second
)

// DatabaseClient is the client-interface for the main mongodb database
type DatabaseClient interface {
	Upsert(*DotlanStatus) (*DotlanStatus, error)
	List(ctx context.Context, filter interface{}, c chan Result)
}

type DatabaseClientImpl struct {
	ctx        context.Context
	collection *mongo.Collection
	workerpool *workerpool.WorkerPool
}

func NewClient(ctx context.Context, env *environment.Environment, workerpool *workerpool.WorkerPool) (*DatabaseClientImpl, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(env.MongoDbURI))
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	return NewClientWithDatabase(client.Database(DATABASE_NAME), workerpool)
}

func NewClientWithDatabase(db *mongo.Database, workerpool *workerpool.WorkerPool) (*DatabaseClientImpl, error) {

	dbClient := DatabaseClientImpl{
		collection: db.Collection(COLLECTION_NAME),
		workerpool: workerpool,
	}

	return &dbClient, nil
}

func (d *DatabaseClientImpl) Upsert(doc *DotlanStatus) (*DotlanStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	filter := bson.D{{"_id", doc.DotlanContestID}}

	updateResult, err := d.collection.ReplaceOne(ctx, filter, doc, options.Replace().SetUpsert(true))

	log.Debug().Interface("updateResult", *updateResult).Msg("Update result")

	return nil, err
}

func (d *DatabaseClientImpl) List(ctx context.Context, filter interface{}, c chan Result) {
	var listResult []DotlanStatus

	ctx, cancel := context.WithTimeout(ctx, DEFAULT_TIMEOUT)
	defer cancel()

	if filter == nil {
		filter = bson.D{}
	}

	cur, err := d.collection.Find(ctx, filter)

	if err != nil {
		c <- Result{Result: nil, Error: err}
		return
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result DotlanStatus
		if err := cur.Decode(&result); err != nil {
			log.Error().Err(err).Msg("Error decoding document")
		} else {
			listResult = append(listResult, result)
		}

	}

	c <- Result{Result: listResult, Error: nil}
}
