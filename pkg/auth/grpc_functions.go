package auth

import (
	"context"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/log"
	"go.elastic.co/apm"
	"github.com/thethan/fdr-users/pkg/firebase"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAuthService(logger log.Logger, repo *firebase.FirestoreTokenRepo) AuthService {
	return AuthService{
		logger: logger,
		repo:   repo,
	}
}

type AuthService struct {
	logger log.Logger
	repo 	*firebase.FirestoreTokenRepo
}


func (as AuthService) parseToken(ctx context.Context, idToken string) (context.Context, error) {
	return as.repo.VerifyIDToken(ctx, idToken)
}

// Simple example of server initialization code.
func (as AuthService) ServerAuthentication(ctx context.Context, logger log.Logger) grpcAuth.AuthFunc {
	exampleAuthFunc := func(ctx context.Context) (context.Context, error) {
		span, ctx := apm.StartSpan(ctx, "ServerAuthentication", "grpc.auth")
		defer span.End()

		tokenString, err := grpcAuth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		ctx, err = as.parseToken(ctx, tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}
		tokenInterface := ctx.Value("firebase_token")
		token, ok := tokenInterface.(firebaseAuth.Token)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}
		logger.Log("firebase_token", token.UID)

		grpc_ctxtags.Extract(ctx).Set("auth.sub", token.Claims)
		return ctx, nil
	}
	return exampleAuthFunc

}
