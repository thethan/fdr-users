package repositories

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thethan/fdr-users/internal/test_helpers"
	"github.com/thethan/fdr-users/pkg/mongo"
	"go.opentelemetry.io/otel"
	"golang.org/x/oauth2"
	"os"
	"testing"
	"time"
)

var tracer otel.Tracer

func TestRepository_SaveOauthToken(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(context.TODO(), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	tracer = test_helpers.TestingTracer()

	mongo := NewMongoOauthRepository(logger, client, tracer)
	ctx := context.Background()
	err = mongo.SaveOauthToken(ctx, "MFG5HMFDHC634Q7W2FPKJBVTKY", oauth2.Token{
		AccessToken:  "access",
		TokenType:    "tokentype",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(360 * time.Second),
	})

	assert.Nil(t, err)
}
