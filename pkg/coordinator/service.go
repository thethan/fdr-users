package coordinator

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/auth"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/goff"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	logger     log.Logger
	logrus     logrus.FieldLogger
	goffClient *goff.Client
	userRepo   yahoo.UserInformation
}

type Game struct{}

type GameRepo interface {
	SendGame(context.Context, Game)
}

func NewService(logger log.Logger, fieldLogger logrus.FieldLogger, goffService *goff.Client, userRepo yahoo.UserInformation) Service {
	return Service{logger: logger, logrus: fieldLogger, goffClient: goffService, userRepo: userRepo}
}

type User struct {
	session string
}

func (s Service) importUserLeagues(ctx context.Context, req interface{}) (res interface{}, err error) {
	traceContextFields := apmlogrus.TraceContext(ctx)
	span, ctx := apm.StartSpan(ctx, "service.importUserLeagues", "custom")
	defer span.End()

	// get user from context

	tokenInterface := ctx.Value(auth.FirebaseToken)
	fmt.Printf("Token Iface %v\n", tokenInterface)

	token, ok := tokenInterface.(*firebaseAuth.Token)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	if token == nil {
		s.logrus.WithFields(traceContextFields).Error("could not unmarshal user")
		return nil, errors.New("could not get user")
	}

	// @todo logic
	//user, err := s.userRepo.GetCredentialInformation(ctx, token.UID)
	//if err != nil {
	//	return nil, err
	//}
	//yahooService := yahoo.NewService(s.logger,s.userRepo).WithSession(token.UID)
	//
	//client := goff.NewClient(yahooService)
	////goffClient.GetLeagueSettings()
	//games, err := client.GetUserGames()
	//
	//pbDraft := pb.Season{}

	//for _, user := range content.Users {
	//for _, game := range games {
	//
	//	//for _, league := range game.Leagues {
	//	//	setting, err := goffClient.GetLeagueSettings(league.LeagueKey)
	//	//	assert.Nil(t, err)
	//	//	t.Log(setting)
	//	//}
	//	// get all leagues with game keys
	//	leagues , _ := client.GetUserLeagues(game.GameID)
	//
	//	for _, league := range leagues {
	//		settings , _ := client.GetLeagueSettings(league.LeagueKey)
	//
	//		// this is where we need to transform all the information to a grpc call
	//	}
	//
	//	//}
	//}

	return nil, errors.New("not implemented")
}

// getUserLeagues
func (s *Service) getUserLeagues(ctx context.Context, user User) {
	span, ctx := apm.StartSpan(ctx, "service.getUserLeagues", "custom")
	defer span.End()
	// loop through leagues
	for _, val := range goff.YearKeys {
		s.goffClient.GetUserLeagues(val)
	}
}
