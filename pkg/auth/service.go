package auth

import (
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/users"
)

type Service struct {
	log.Logger

	getUserRepo users.GetUserInfo
}

func NewService(logger log.Logger, info users.GetUserInfo) Service {
	return Service{
		Logger:       logger,
		getUserRepo:  info,
	}
}

