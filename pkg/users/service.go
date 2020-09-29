package users

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/auth"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/consts"
	"github.com/thethan/fdr-users/pkg/league"
	"github.com/thethan/fdr-users/pkg/users/entities"
	"github.com/thethan/fdr-users/pkg/users/info"
	"github.com/thethan/fdr-users/pkg/users/transports"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	pb "github.com/thethan/fdr_proto"
)

type SaveUserInfo interface {
	SaveYahooCredential(ctx context.Context, uid, accessToken string, guid string) (entities.User, error)
	SaveYahooInformation(ctx context.Context, uid, accessToken, refreshToken, email, guid string) (entities.User, error)
}

// NewService returns a naïve, stateless implementation of Service.
func NewService(logger log.Logger, info info.GetUserInfo, saveUserInfo SaveUserInfo, oauthRepo transports.OauthRepository, importer league.NewImporterService) usersService {
	return usersService{
		logger:       logger,
		userInfo:     info,
		saveUserInfo: saveUserInfo,
		importer:     importer,
		oauthRepo:    oauthRepo,
	}
}

type usersService struct {
	logger       log.Logger
	userInfo     info.GetUserInfo
	saveUserInfo SaveUserInfo
	importer     league.NewImporterService
	oauthRepo    transports.OauthRepository
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

func (s usersService) SaveFromUserID(ctx context.Context, in *transports.UserCredentialRequest) (*transports.UserCredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "SaveFromUserID", "handlers.service")
	defer span.End()

	var resp transports.UserCredentialResponse
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
			level.Error(s.logger).Log("message", "could not get importer user data", "error", err, "guid", userResource.Users.User.Guid)
		}
	} else {
		// ignore on not empty
		go importer.ImportFromUser(ctx, userResource)
	}

	resp = transports.UserCredentialResponse{
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

func (s usersService) GetUsersLeagues(ctx context.Context, in *transports.UserCredentialRequest) (*transports.UserCredentialResponse, error) {
	span, ctx := apm.StartSpan(ctx, "GetUsersLeagues", "handlers.service")
	defer span.End()

	var resp transports.UserCredentialResponse
	tokenInterface := ctx.Value(consts.FirebaseToken)

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", ok)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		user, _ := s.userInfo.GetCredentialInformation(ctx, token.UID)

		token.Claims["yahoo_guid"] = user.GUID
	}()
	var yahooID string
	// get token from uuid
	for key, identifiers := range token.Firebase.Identities {
		if key == "yahoo.com" {
			yahooID, ok = identifiers.([]interface{})[0].(string)
			if !ok {
				return nil, errors.New("could not get authenticator token")
			}
		}
	}

	if yahooID == "" {
		return nil, errors.New("could not get authenticator token")
	}

	yahooService := yahoo.NewService(s.logger, NoopGetUserInformation{})
	importer := s.importer.NewImporterWithService(yahooService)

	leagueGroups, err := importer.GetUserLeagues(ctx, yahooID)
	if err != nil {
		level.Error(s.logger).Log("message", "could not get user leagues", "error", err, "guid", in.Guid)
	}

	if err != nil {
		level.Error(s.logger).Log("message", "could not get user leagues", "error", err, "guid", in.Guid)
	}
	wg.Done()
	wg.Wait()

	resp = transports.UserCredentialResponse{
		UID:   token.UID,
		Email: token.Claims["email"].(string),
		Guid:  token.Claims["yahoo_guid"].(string),

		Leagues: leagueGroups,
	}

	return &resp, nil
}

const guidKey string = "xoauth_yahoo_guid"

func (service usersService) YahooCallback(ctx context.Context, r *http.Request) error {
	span, ctx := apm.StartSpan(ctx, "YahooCallback", "service")
	defer span.End()

	values := r.URL.Query()
	state := values["state"]
	code := values["code"]

	token, err := getUserInfo(ctx, service.logger, state[0], code[0])
	if err != nil {
		level.Error(service.logger).Log("msg", "could not  get User from yahoo")
		return err
	}
	level.Debug(service.logger).Log("msg", "token received", "token", token)
	// todo put into a queue
	content, err := getUserGameInfoFromYahoo(ctx, token)
	level.Debug(service.logger).Log("content", string(content))

	// get the raw key from oauth
	guidInterface := token.Extra(guidKey)
	guid, ok := guidInterface.(string)
	if !ok {
		_ = level.Error(service.logger).Log("msg", "Could not get key from")
	}
	_ = level.Debug(service.logger).Log("msg", "about to save guid", "guid", guid)
	go func() {
		// todo queue importer
		// QueuePlayers
	}()
	return service.oauthRepo.SaveOauthToken(ctx, guid, *token)
}

func getUserInfo(ctx context.Context, logger log.Logger, state string, code string) (*oauth2.Token, error) {
	span, ctx := apm.StartSpan(ctx, "getUserInfo", "repo.oauth")
	defer span.End()
	// @todo make this into an auth state from the actual user and not just
	//if state != oauthStateString {
	if state == "" {
		return nil, fmt.Errorf("invalid OauthConfig state")
	}
	token, err := transports.OauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	level.Debug(logger).Log("msg", "token received", "token", token)
	return token, nil
}

// save token info here.
func getUserGameInfoFromYahoo(ctx context.Context, token *oauth2.Token) ([]byte, error) {
	span, ctx := apm.StartSpan(ctx, "getUserGameInfoFromYahoo", "repo")
	defer span.End()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}
