package mongo

import (
	"context"
	"fmt"
	"go.elastic.co/apm/module/apmmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func NewMongoDBClient(user, password, host, port string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI(
	//	fmt.Sprintf("mongodb+srv://%s:%s@%s", user, password, host),
	//), options.Client().SetMonitor(apmmongo.CommandMonitor()), )
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s:%s@%s:%s/fdr?authSource=admin&w=majority", user, password, host, port),
	), options.Client().SetMonitor(apmmongo.CommandMonitor()))

	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return nil, err
	}

	return client, nil
}
