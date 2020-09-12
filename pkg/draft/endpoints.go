package draft

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	entities2 "github.com/thethan/fdr-users/pkg/users/entities"
	"go.elastic.co/apm"
)

type Endpoints struct {
	logger                   log.Logger
	service                  *Service
	ImportDraft              endpoint.Endpoint
	GetLeagueDraft           endpoint.Endpoint
	SaveDraftResult          endpoint.Endpoint
	GetTeamRoster            endpoint.Endpoint
	SaveUserPlayerPreference endpoint.Endpoint
	GetUserPlayerPreference  endpoint.Endpoint
	OpenDraft                endpoint.Endpoint
	ShuffleDraftOrder        endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, service *Service, authService *auth.AuthService, authMiddleware endpoint.Middleware, getUserInfoMiddleWare endpoint.Middleware) Endpoints {
	e := Endpoints{
		logger:                   logger,
		ImportDraft:              makeImportDraft(logger, service),
		GetLeagueDraft:           authMiddleware(makeGetDraftInfo(logger, service)),
		SaveDraftResult:          authMiddleware(getUserInfoMiddleWare(makeSaveDraftResult(logger, service))),
		GetTeamRoster:            authMiddleware(makeGetTeamDraftRoster(logger, service)),
		SaveUserPlayerPreference: authMiddleware(getUserInfoMiddleWare(makeSaveUserPlayerPreference(logger, service))),
		OpenDraft:                authMiddleware(getUserInfoMiddleWare(makeOpenDraft(logger, service))),
		ShuffleDraftOrder:        authMiddleware(getUserInfoMiddleWare(makeShuffle(logger, service))),
	}

	return e
}

func makeImportDraft(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "ImportDraft", "endpoint")
		defer span.End()

		return nil, nil
	}
}

type DraftResultResponse struct {
	League       *entities.League       `json:"league"`
	DraftResults []entities.DraftResult `json:"draft_results"`
}

type DraftTeamRostersResponse struct {
	Teams map[string]entities.Roster `json:"rosters"`
}

type LeagueDraftRequest struct {
	LeagueKey string
}

type User struct {
	Email string `json:"email"`
	Guid  string `json:"guid"`
	UID   string `json:"uid"`
}

type SaveDraftResultRequest struct {
	User   entities.User         `json:"user"`
	League entities.League       `json:"league"`
	Player entities.PlayerSeason `json:"player"`
	Team   entities.Team         `json:"team"`
	Pick   int                   `json:"pick"`
}

func makeGetDraftInfo(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "GetDraftInfo", "endpoint")
		defer span.End()

		req, ok := request.(*LeagueDraftRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		league, results, err := service.ListDraftResults(ctx, req.LeagueKey)
		if err != nil {
			return nil, err
		}
		teamKeyToIdx := make(map[string]int, len(league.Teams))
		for idx := range league.Teams {
			teamKeyToIdx[league.Teams[idx].TeamKey] = idx
		}
		teams := make([]entities.Team, len(league.Teams))
		for idx, teamKey := range league.DraftOrder {
			teams[idx] = league.Teams[teamKeyToIdx[teamKey]]
		}
		league.TeamDraftOrder = teams
		return &DraftResultResponse{DraftResults: results, League: league}, err
	}
}

func makeGetTeamDraftRoster(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "GetTeamDraftRoster", "endpoint")
		defer span.End()

		req, ok := request.(*LeagueDraftRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		rosters, err := service.GetTeamsDraftResults(ctx, req.LeagueKey)
		if err != nil {
			return nil, err
		}

		return &DraftTeamRostersResponse{Teams: rosters}, err
	}
}

func makeSaveDraftResult(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveDraftResult", "endpoint")
		defer span.End()

		req, ok := request.(*SaveDraftResultRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		result, err := service.SaveDraftRequest(ctx, req.User, req.League, req.Team, req.Player, req.Pick)
		if err != nil {
			return nil, err
		}
		return &DraftResultResponse{
			DraftResults: []entities.DraftResult{*result},
		}, err
	}
}

func makeSaveUserPlayerPreference(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveUserPlayerPreference", "endpoint")
		defer span.End()

		req, ok := request.(*entities.UserPlayerPreference)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		userInterface := ctx.Value(auth.User)
		user, ok := userInterface.(*entities2.User)
		if !ok {
			return nil, errors.New("bad request")
		}

		req.UserID = user.GUID

		err = service.SaveUserPlayerPreference(ctx, *req)
		if err != nil {
			return nil, err
		}
		return req, err
	}
}

func makeOpenDraft(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "makeOpenDraft", "endpoint")
		defer span.End()


		req, ok := request.(*OpenDraftRequest)
		if !ok {
			return nil, errors.New("Could not get request")
		}
		league, err := service.OpenDraft(ctx, req.LeagueID)
		return league, err
	}
}

func makeShuffle(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "makeShuffle", "endpoint")
		defer span.End()

		req, ok := request.(*ShuffleDraftOrderRequest)
		if !ok {
			return nil, errors.New("Could not get request")
		}
		league, err := service.ShuffleOrder(ctx, req.LeagueID)
		return league, err
	}
}

const LeagueKey = "league_key"

func NewUserHasAccessToDraftMiddleware(logger log.Logger, a *auth.AuthService) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			span, ctx := apm.StartSpan(ctx, "NewAuthMiddleware", "middleware")
			defer span.End()

			tokenIface := ctx.Value(LeagueKey)
			tokenString := tokenIface.(string)

			ctx, err := a.AddFirebaseTokenToContext(ctx, tokenString)
			if err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}
