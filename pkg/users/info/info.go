package info

import (
	"context"
	"github.com/thethan/fdr-users/pkg/users/entities"
)

type GetUserInfo interface {
	GetCredentialInformation(ctx context.Context, session string) (entities.User, error)
}