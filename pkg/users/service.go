package users

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/league"
	"github.com/thethan/fdr-users/pkg/users/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"time"

	pb "github.com/thethan/fdr_proto"
)

type GetUserInfo interface {
	GetCredentialInformation(ctx context.Context, session string) (entities.User, error)
}

type SaveUserInfo interface {
	SaveYahooCredential(ctx context.Context, uid, accessToken string, guid string) (entities.User, error)
	SaveYahooInformation(ctx context.Context, uid, accessToken, refreshToken, email, guid string) (entities.User, error)
}

// NewService returns a naïve, stateless implementation of Service.
func NewService(logger log.Logger, info GetUserInfo, saveUserInfo SaveUserInfo, importer league.NewImporterService) usersService {
	return usersService{
		logger:       logger,
		userInfo:     info,
		saveUserInfo: saveUserInfo,
		importer:     importer,
	}
}

type usersService struct {
	logger       log.Logger
	userInfo     GetUserInfo
	saveUserInfo SaveUserInfo
	importer     league.NewImporterService
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

type NoopGetUserInformation struct {
	AccessToken string
}

func (n NoopGetUserInformation) GetCredentialInformation(ctx context.Context, session string) (entities.User, error) {
	return entities.User{AccessToken: n.AccessToken}, nil
}

func (s usersService) SaveFromUserID(ctx context.Context, in *UserCredentialRequest) (*UserCredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "SaveFromUserID", "handlers.service")
	defer span.End()

	var resp UserCredentialResponse
	yahooService := yahoo.NewService(s.logger, NoopGetUserInformation{AccessToken: in.Session})
	userResource, err := yahooService.GetUserResourcesGameLeaguesResponse(ctx)
	if err != nil {
		level.Error(s.logger).Log("message", "Could not get user guid", "error", err)
		return nil, err
	}

	user, err := s.saveUserInfo.SaveYahooInformation(ctx, in.UID, in.Session, in.RefreshToken, in.Email, userResource.Users.User.Guid)
	if err != nil {
		return nil, err
	}

	importer := s.importer.NewImporterWithService(yahooService)

	leagueGroups, err := importer.GetUserLeagues(ctx, userResource.Users.User.Guid)
	if err != nil {
		level.Error(s.logger).Log("message", "could not get user leagues", "error", err, "guid", userResource.Users.User.Guid)
	}
	if len(leagueGroups) == 0 {
		leagueGroups, err = importer.ImportFromUser(ctx, userResource)
		if err != nil {
			level.Error(s.logger).Log("message", "could not get import user data", "error", err, "guid", userResource.Users.User.Guid)
		}
	} else {
		// ignore on not empty
		go importer.ImportFromUser(ctx, userResource)
	}

	resp = UserCredentialResponse{
		UID:          userResource.Users.User.Guid,
		Session:      user.AccessToken,
		RefreshToken: user.RefreshToken,
		Leagues:      leagueGroups,
	}

	return &resp, nil
}

type CredentialRequest struct {
	Session string `protobuf:"bytes,1,opt,name=session,proto3" json:"session,omitempty"`
	GUID    string `protobuf:"bytes,1,opt,name=session,proto3" json:"guid,omitempty"`
}

func (s usersService) SaveYahooCredential(ctx context.Context, in *CredentialRequest) (*pb.CredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "SaveYahooCredential", "handlers.service")
	defer span.End()

	var resp pb.CredentialResponse
	// ctx will have token
	tokenInterface := ctx.Value("firebase_token")

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, errors.New("failure to get user token")
	}

	user, err := s.saveUserInfo.SaveYahooCredential(ctx, token.UID, in.Session, in.GUID)
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
