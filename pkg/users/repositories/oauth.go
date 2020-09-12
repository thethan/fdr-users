package repositories

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"time"
)

const Oauth string = "yahoo_users"

type Repository struct {
	logger log.Logger
	client *mongo.Client
}

func NewMongoOauthRepository(logger log.Logger, client *mongo.Client) Repository {
	return Repository{
		logger: logger,
		client: client,
	}
}

type oauthUser struct {
	Guid string `bson:"guid"`
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `bson:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `bson:"token_type"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `bson:"refresh_token"`

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time `bsxon:"expiry,omitempty"`

	// raw optionally contains extra metadata from the server
	// when updating a token.
	Raw       interface{} `bson:"raw"`
	UpdatedAt time.Time   `bson:"updated_at"`
}

const database string = "fdr"

func oauth2TokenToUserToken(oauth2Token oauth2.Token) oauthUser {
	return oauthUser{
		TokenType:    oauth2Token.TokenType,
		AccessToken:  oauth2Token.AccessToken,
		RefreshToken: oauth2Token.RefreshToken,
		Expiry:       oauth2Token.Expiry,
		UpdatedAt:    time.Now(),
	}
}

func (r *Repository) SaveOauthToken(ctx context.Context, uuid string, token oauth2.Token) error {
	span, ctx := apm.StartSpan(ctx, "SaveOauthToken", "repository.Mongo")
	defer span.End()

	user := oauth2TokenToUserToken(token)
	collection := r.client.Database(database).Collection(Oauth)
	user.Guid = uuid

	res := bson.M{"$set": user}
	filter := bson.M{"guid": uuid}
	results, err := collection.UpdateOne(ctx, filter, res, &options.UpdateOptions{Upsert: aws.Bool(true)})
	if err != nil {
		level.Error(r.logger).Log("error", err, "message", "could not execute query", "guid", uuid)
		return err
	}

	level.Debug(r.logger).Log("message", "results", "results", results)

	return nil
}
