package users

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/log"
	"go.elastic.co/apm"
	"time"

	pb "github.com/thethan/fdr_proto"
)

type GetUserInfo interface {
	GetCredentialInformation(ctx context.Context, session string) (User, error)
}

type SaveUserInfo interface {
	SaveYahooCredential(ctx context.Context, uid, accessToken string) (User, error)
	SaveYahooInformation(ctx context.Context, uid, accessToken, refreshToken, email string) (User, error)
}

// NewService returns a naïve, stateless implementation of Service.
func NewService(logger log.Logger, info GetUserInfo, saveUserInfo SaveUserInfo) usersService {
	return usersService{
		logger:       logger,
		userInfo:     info,
		saveUserInfo: saveUserInfo,
	}
}

type usersService struct {
	logger       log.Logger
	userInfo     GetUserInfo
	saveUserInfo SaveUserInfo
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
	// ctx will have token
	tokenInterface := ctx.Value("firebase_token")

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, errors.New("failure to get user token")
	}
	user, err := s.userInfo.GetCredentialInformation(ctx, token.UID)
	if err != nil {
		return nil, err
	}

	resp = pb.CredentialResponse{
		Token: &pb.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			ExpiresIn:    int32(time.Now().Sub(user.ExpiresAt).Seconds()),
		},
	}
	return &resp, nil
}




func (s usersService) SaveFromUserID(ctx context.Context, in *UserCredentialRequest) (*pb.CredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "SaveFromUserID", "handlers.service")
	defer span.End()


	var resp pb.CredentialResponse

	user, err := s.saveUserInfo.SaveYahooInformation(ctx, in.UID, in.Session, in.RefreshToken, in.Email)
	if err != nil {
		return nil, err
	}

	resp = pb.CredentialResponse{
		Token: &pb.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			ExpiresIn:    int32(time.Now().Sub(user.ExpiresAt).Seconds()),
		},
	}

	return &resp, nil
}

func (s usersService) SaveYahooCredential(ctx context.Context, in *pb.CredentialRequest) (*pb.CredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "SaveYahooCredential", "handlers.service")
	defer span.End()

	var resp pb.CredentialResponse
	// ctx will have token
	tokenInterface := ctx.Value("firebase_token")

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, errors.New("failure to get user token")
	}
	user, err := s.saveUserInfo.SaveYahooCredential(ctx, token.UID, in.Session)
	if err != nil {
		return nil, err
	}

	resp = pb.CredentialResponse{
		Token: &pb.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			ExpiresIn:    int32(time.Now().Sub(user.ExpiresAt).Seconds()),
		},
	}
	return &resp, nil
}
