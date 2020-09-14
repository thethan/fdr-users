package repositories

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	pkgEntities "github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatsRepo struct {
	logger log.Logger
	client *mongo.Client
}

func NewMongoStatsRepo(logger log.Logger, client *mongo.Client) StatsRepo {
	return StatsRepo{
		logger: logger,
		client: client,
	}
}

func (s StatsRepo) SavePlayerStats(ctx context.Context, playerKey string, season, week string, stats []pkgEntities.PlayerStat) error {
	span, ctx := apm.StartSpan(ctx, "SavePlayerStats", "repository.Mongo")
	defer span.End()

	collection := s.client.Database("fdr").Collection("player_stats")
	findQuert := bson.M{"player_key": playerKey}
	query := bson.M{"$set": bson.M{"player_key": playerKey, "stats_by_season": bson.M{season: bson.M{week: stats}}}}

	var playerStatBson pkgEntities.PlayerStats
	playerStatBson.PlayerID = playerKey

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the
		// transaction.
		if _, err := collection.UpdateOne(sessCtx, findQuert, query, &options.UpdateOptions{Upsert: aws.Bool(true)}); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Step 2: Start a session and run the callback using WithTransaction.
	session, err := s.client.StartSession()
	if err != nil {
		level.Error(s.logger).Log("message", "could not start transaction from mongo", "error", err)
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		level.Error(s.logger).Log("message", "could not save transaction from mongo", "error", err)
		return err
	}

	return nil
}
