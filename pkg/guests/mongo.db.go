package guests

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const database = "puretotten"
const collection = "guest"

type MongoRepository struct {
	logger                 log.Logger
	client                 *mongo.Client

}

func NewMongoRepository(logger log.Logger, client *mongo.Client) MongoRepository {
	return MongoRepository{logger: logger, client: client,}
}

func (m MongoRepository) SaveGuest(ctx context.Context, guest Guest) error {
	span, ctx := apm.StartSpan(ctx, "SaveGuest", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(collection)
	bsonGuest, err := bson.Marshal(guest)
	if err != nil {
		level.Error(m.logger).Log("message", "could not marshal guest", "error", err)
		return err
	}
	insertRes, err := collection.InsertOne(ctx, bsonGuest)
	if err != nil {
		level.Error(m.logger).Log("message", "could not save rsvp", "error", err)
		return err
	}
	level.Info(m.logger).Log("message", "saved guest", "_id", insertRes.InsertedID, "guest", guest)

	return nil
}

func (m MongoRepository) GetGuestList(ctx context.Context) ([]Guest, error) {
	span, ctx := apm.StartSpan(ctx, "GetGuestList", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(collection)
	results, err := collection.Find(ctx, bson.M{})
	if err != nil {
		level.Error(m.logger).Log("message", "could not find rsvp", "error", err)
		return nil, err
	}
	guests := make([]Guest, 0)
	err = results.All(ctx, &guests)
	if err != nil {
		level.Error(m.logger).Log("message", "get not marshal guest into list", "error", err)
		return nil, err
	}

	return guests, nil
}
