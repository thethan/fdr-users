package repositories

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/test_helpers"
	"golang.org/x/oauth2"
	"os"
	"testing"
	"time"
)

func TestRepository_SaveOauthToken(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}

	mongo := NewMongoOauthRepository(logger, client)
	ctx := context.Background()
	err = mongo.SaveOauthToken(ctx, "MFG5HMFDHC634Q7W2FPKJBVTKY", oauth2.Token{
		AccessToken:  "access",
		TokenType:    "tokentype",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(360 * time.Second),
	})

	assert.Nil(t, err)
}
