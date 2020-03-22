package auth

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/handlers"
	"go.elastic.co/apm"
)

type Service struct {
	log.Logger

	getUserRepo  handlers.GetUserInfo
}

func NewService(logger log.Logger, info handlers.GetUserInfo) Service {
	return Service{
		Logger:       logger,
		getUserRepo:  info,
	}
}

func (s Service) GetCredentialInformation(ctx context.Context, session string) (handlers.User, error) {
	span, ctx := apm.StartSpan(ctx, "GetCredentialInformation", "service.method")
	defer span.End()

	return s.getUserRepo.GetCredentialInformation(ctx, session)
}
