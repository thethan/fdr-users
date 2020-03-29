package firebase

import (
	"context"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func NewFirestoreAuthRepo(logger log.Logger, client *firebaseAuth.Client) FirestoreTokenRepo {
	return FirestoreTokenRepo{
		logger: logger,
		client: client,
	}
}

// FirestoreToken
type FirestoreTokenRepo struct {
	logger log.Logger
	client *firebaseAuth.Client
}

func (fT FirestoreTokenRepo) VerifyIDToken(ctx context.Context, idToken string) (context.Context, error) {

	token, err := fT.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		level.Error(fT.logger).Log("message", "error verifying ID token: %v\n", "error", err)
		return ctx, err
	}
	ctx = context.WithValue(ctx, "firebase_token", token)

	return ctx, nil
}
