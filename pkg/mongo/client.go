package mongo

import (
	"context"
	"fmt"
	"go.elastic.co/apm/module/apmmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func NewMongoDBClient(user, password, host string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI(
	//	fmt.Sprintf("mongodb+srv://%s:%s@%s", user, password, host),
	//), options.Client().SetMonitor(apmmongo.CommandMonitor()), )
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		//fmt.Sprintf("mongodb+srv://draft:jnkdNjCBAPbAkPF7@fdr-0hsvx.mongodb.net/drafts?w=majority",),
		fmt.Sprintf("mongodb://draft:asdf@localhost:27017/fdr_drafts?w=majority",),
	), options.Client().SetMonitor(apmmongo.CommandMonitor()), )


	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return nil, err
	}

	return client, nil
}
