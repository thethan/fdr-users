package auth

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/markbates/goth"
	"github.com/thethan/fdr-users/handlers"
	"github.com/thethan/fdr-users/pkg/gothic"
	"go.elastic.co/apm"
	"google.golang.org/grpc/metadata"
	"net/http"
)

type Service struct {
	log.Logger
	providerName string
	provider     goth.Provider
	getUserRepo  handlers.GetUserInfo
}

func NewService(logger log.Logger, provider goth.Provider, info handlers.GetUserInfo) Service {
	return Service{
		providerName: "yahoo",
		Logger:       logger,
		provider:     provider,
		getUserRepo:  info,
	}
}

func (s Service) GetCredentialInformation(ctx context.Context, session string) (goth.User, error) {
	span, ctx := apm.StartSpan(ctx, "GetCredentialInformation", "service.method")
	defer span.End()
	return s.getUserRepo.GetCredentialInformation(ctx, session)
}

func (s Service) getUserFromGrpc(md metadata.MD) (goth.User, error) {
	session := md.Get(gothic.SessionName)

	return s.GetCredentialInformation(context.Background(), session[0])
}

func (s Service) getUserFromHttp(req *http.Request) (goth.User, error) {
	return s.getUserFromValues(req)
}

func (s Service) getUserFromValues(req *http.Request) (goth.User, error) {

	value, err := gothic.GetFromSession(s.providerName, req)
	if err != nil {
		return goth.User{}, err
	}

	sess, err := s.provider.UnmarshalSession(value)
	if err != nil {
		return goth.User{}, err
	}

	user, err := s.provider.FetchUser(sess)
	if err != nil {
		return goth.User{}, err
	}
	//if user.ExpiresAt.Before(time.Now()) {
	//	s.provider.RefreshToken(user.RefreshToken)
	//}

	return user, nil
}
