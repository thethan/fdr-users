package handlers

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/auth"
	"time"

	pb "github.com/thethan/fdr_proto"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService(logger log.Logger, service *auth.Service) pb.UsersServer {
	return usersService{
		logger:      logger,
		authService: service,
	}
}

type usersService struct {
	logger      log.Logger
	authService *auth.Service
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
	var resp pb.CredentialResponse

	user, err := s.authService.GetCredentialInformation(ctx, in.Session)
	if err != nil {
		return nil, err
	}

	resp = pb.CredentialResponse{
		Token: &pb.Token{
			AccessToken:          user.AccessToken,
			RefreshToken:         user.AccessToken,
			ExpiresIn:            int32(time.Now().Sub(user.ExpiresAt).Seconds()),
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
	}
	return &resp, nil
}
