package auth

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/thethan/fdr-users/pkg/consts"
	"github.com/thethan/fdr-users/pkg/firebase"
	"github.com/thethan/fdr-users/pkg/users/info"
	"go.elastic.co/apm"
	"go.opentelemetry.io/otel"
	"net/http"
	"strings"

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

		return as.AddFirebaseTokenToContext(ctx, tokenString)
	}
	return serverAuthentication

}

func (as AuthService) AddFirebaseTokenToContext(ctx context.Context, tokenString string) (context.Context, error) {
	ctx, err := as.parseToken(ctx, tokenString)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	tokenInterface := ctx.Value(consts.FirebaseToken)

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	grpc_ctxtags.Extract(ctx).Set("auth.sub", token.Claims)
	return ctx, nil
}

const BearerToken string = "bearer"
const User string = "user"

func (a AuthService) ServerBefore(ctx context.Context, req *http.Request) context.Context {
	span, ctx := apm.StartSpan(ctx, "authService", "ServerBefore")
	defer span.End()

	authHeader := req.Header.Get("Authorization")
	authHeader = strings.Replace(authHeader, "Bearer ", "", 1)
	ctx = context.WithValue(ctx, BearerToken, authHeader)
	return ctx
}

func (a AuthService) UserInformationToContext(userInfo info.GetUserInfo) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			span, ctx := apm.StartSpan(ctx, "UserInformationToContext", "middleware")
			defer span.End()

			tokenInterface := ctx.Value(consts.FirebaseToken)

			token, ok := tokenInterface.(*firebaseAuth.Token)
			if !ok {
				return nil, errors.New("could not authenticate user")
			}

			user, _ := userInfo.GetCredentialInformation(ctx, token.UID)

			ctx = context.WithValue(ctx, User, user)
			return next(ctx, request)
		}
	}
}

func (a AuthService) NewAuthMiddleware(tracer otel.Tracer, meter otel.Meter) endpoint.Middleware {
	counter, err := meter.NewInt64Counter("auth_middleware", )
	if err != nil {

		panic("error in init counter")
	}
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			counter.Add(ctx, 1,)
			ctx, span := tracer.Start(ctx, "NewAuthMiddleware", )
			defer span.End()

			tokenIface := ctx.Value(BearerToken)
			tokenString := tokenIface.(string)

			ctx, err := a.AddFirebaseTokenToContext(ctx, tokenString)
			if err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}

func (a AuthService) GetUserInfoFromContextMiddleware(userInfo info.GetUserInfo) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {

			tokenInterface := ctx.Value(consts.FirebaseToken)

			token, ok := tokenInterface.(*firebaseAuth.Token)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
			}

			user, err := userInfo.GetCredentialInformation(ctx, token.UID)
			if err != nil {
				level.Error(a.logger).Log("message", "could not get auth", "error", err)
				return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
			}

			ctx = context.WithValue(ctx, User, &user)

			return next(ctx, request)
		}
	}
}
