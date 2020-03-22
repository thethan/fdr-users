package handlers

import (
	"context"
	"github.com/go-kit/kit/log"
	"go.elastic.co/apm"
	"time"

	pb "github.com/thethan/fdr_proto"
)

type GetUserInfo interface {
	GetCredentialInformation(ctx context.Context, session string) (User, error)
}

// NewService returns a naïve, stateless implementation of Service.
func NewService(logger log.Logger, info GetUserInfo) pb.UsersServer {
	return usersService{
		logger:   logger,
		userInfo: info,
	}
}

type usersService struct {
	logger   log.Logger
	userInfo GetUserInfo
}

// Create implements Service.
func (s usersService) Create(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var resp pb.CreateUserResponse
	resp = pb.CreateUserResponse{
		// Status:
	}
	return &resp, nil
}

// Search implements Service.
func (s usersService) Search(ctx context.Context, in *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	var resp pb.ListUserResponse
	resp = pb.ListUserResponse{
		// User:
		// Metadata:
	}
	return &resp, nil
}

// Login implements Service.
func (s usersService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var resp pb.LoginResponse
	resp = pb.LoginResponse{
		// Token:
	}
	return &resp, nil
}

func (s usersService) Credentials(ctx context.Context, in *pb.CredentialRequest) (*pb.CredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "Credentials", "handlers.service")
	defer span.End()

	var resp pb.CredentialResponse
	user, err := s.userInfo.GetCredentialInformation(ctx, in.Session)
	if err != nil {
		return nil, err
	}

	resp = pb.CredentialResponse{
		Token: &pb.Token{
			AccessToken:          user.AccessToken,
			RefreshToken:         user.RefreshToken,
			ExpiresIn:            int32(time.Now().Sub(user.ExpiresAt).Seconds()),
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
	}
	return &resp, nil
}
