package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"

	mongotrace "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver"
)

func NewMongoDBClient(ctx context.Context, user, password, host, port string) (*mongo.Client, error) {
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI(
	//	fmt.Sprintf("mongodb+srv://%s:%s@%s", user, password, host),
	//), options.Client().SetMonitor(apmmongo.CommandMonitor()), )

	opts := options.Client()
	opts.Monitor = mongotrace.NewMonitor(os.Getenv("SERVICE_NAME"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s:%s@%s:%s/fdr?authSource=admin&w=majority", user, password, host, port),
	), opts)

	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return nil, err
	}

	return client, nil
}
