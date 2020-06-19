package auth

import (
	"context"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/firebase"
	"go.elastic.co/apm"
	"net/http"
	"strings"

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
	repo   *firebase.FirestoreTokenRepo
}

func (as AuthService) parseToken(ctx context.Context, idToken string) (context.Context, error) {
	return as.repo.VerifyIDToken(ctx, idToken)
}

// Simple example of server initialization code.
func (as AuthService) ServerAuthentication(ctx context.Context, logger log.Logger) grpcAuth.AuthFunc {
	serverAuthentication := func(ctx context.Context) (context.Context, error) {
		span, ctx := apm.StartSpan(ctx, "ServerAuthentication", "grpc.auth")
		defer span.End()

		tokenString, err := grpcAuth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		return as.addFirebaseTokenToContext(ctx, tokenString)
	}
	return serverAuthentication

}

func (as AuthService) addFirebaseTokenToContext(ctx context.Context, tokenString string) (context.Context, error){
	ctx, err := as.parseToken(ctx, tokenString)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	tokenInterface := ctx.Value(FirebaseToken)

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	grpc_ctxtags.Extract(ctx).Set("auth.sub", token.Claims)
	return ctx, nil
}

const FirebaseToken string = "firebase_token"
const BearerToken string = "bearer"

func (a AuthService) ServerBefore(ctx context.Context, req *http.Request) context.Context {
	span, ctx :=apm.StartSpan(ctx, "authService", "ServerBefore")
	defer span.End()
	authHeader := req.Header.Get("Authorization")
	authHeader = strings.Replace(authHeader, "Bearer ", "", 1)
	ctx = context.WithValue(ctx, BearerToken, authHeader)
	return ctx
}

func (a AuthService) NewAuthMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			span, ctx :=apm.StartSpan(ctx, "NewAuthMiddleware", "middleware")
			defer span.End()

			tokenIface := ctx.Value(BearerToken)
			tokenString := tokenIface.(string)

			ctx, err := a.addFirebaseTokenToContext(ctx, tokenString)
			if err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}
