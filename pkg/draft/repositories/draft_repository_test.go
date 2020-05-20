package repositories

import (
	"context"
	"github.com/stretchr/testify/assert"

	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/test_helpers"
	"os"
	"testing"
)

func TestMongoRepository_GetTeamsForManagers(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	leagues, err := mongoRepo.GetTeamsForManagers(context.TODO(), "DPPQCXCRV75Z2LKJW5YRC7RAYM")
	assert.Nil(t, err)
	assert.True(t, len(leagues) > 0, "Could not find leagues" )
}
