package repositories

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"golang.org/x/oauth2"
	"time"
)

const Oauth string = "yahoo_users"

type Repository struct {
	logger log.Logger
	client *mongo.Client
	tracer otel.Tracer
}

func NewMongoOauthRepository(logger log.Logger, client *mongo.Client, tracer otel.Tracer) Repository {
	return Repository{
		logger: logger,
		client: client,
		tracer: tracer,
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

func oauthUserToOauth(oauth2Token oauthUser) oauth2.Token {
	return oauth2.Token{
		AccessToken:  oauth2Token.AccessToken,
		TokenType:    oauth2Token.TokenType,
		RefreshToken: oauth2Token.RefreshToken,
		Expiry:       oauth2Token.Expiry,
	}
}

func (r *Repository) SaveOauthToken(ctx context.Context, uuid string, token oauth2.Token) error {
	ctx, span := r.tracer.Start(ctx, "SaveOauthToken")
	span.SetAttributes(label.String("user_guid", uuid))
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

func (r *Repository) GetUserOAuthToken(ctx context.Context, guid string) (oauth2.Token, error) {
	ctx, span := r.tracer.Start(ctx, "GetUserOAuthToken")
	span.SetAttributes(label.String("user_guid", guid))
	defer span.End()

	var user oauthUser
	collection := r.client.Database(database).Collection(Oauth)
	filter := bson.M{"guid": guid}
	result := collection.FindOne(ctx, filter, &options.FindOneOptions{})
	if result.Err() != nil {
		level.Error(r.logger).Log("error", result.Err(), "message", "could not execute query", "guid", guid)
		return oauth2.Token{}, result.Err()
	}
	err := result.Decode(&user)

	if err != nil {
		level.Error(r.logger).Log("error", err, "message", "could not execute query", "guid", guid)
		return oauth2.Token{}, err
	}

	level.Debug(r.logger).Log("message", "results", "results", result.Err())
	return oauthUserToOauth(user), nil
}
